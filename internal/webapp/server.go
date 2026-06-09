package webapp

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"scoreador/internal/loader"
	"scoreador/internal/model"
	"scoreador/internal/report"
	"scoreador/internal/sim"
	"scoreador/internal/store"
)

const defaultLambdaPath = "examples/demo/lambda.csv"

type PageData struct {
	LambdaPath        string
	TeamA             string
	TeamB             string
	ShotsA            int
	ShotsB            int
	MotivationA       string
	MotivationB       string
	Tiebreaker        string
	Simulations       int
	Seed              string
	Error             string
	Result            *MatchReport
	History           []store.SingleMatchRun
	BatchLambdaPath   string
	BatchSeed         string
	BatchError        string
	BatchMessage      string
	BatchDownloadURL  string
	BatchDownloadName string
}

type MatchReport struct {
	Run                store.SingleMatchRun
	Simulations        int
	WinPctA            float64
	WinPctB            float64
	DrawPct            float64
	RegulationDraws    int
	RegulationPct      float64
	PenaltyPct         float64
	RandomTiePct       float64
	AvgGoalsA          float64
	AvgGoalsB          float64
	AvgTotalGoals      float64
	AvgGoalDiff        float64
	AvgGoalsAWidth     float64
	AvgGoalsBWidth     float64
	AvgTotalGoalsWidth float64
	AvgGoalDiffWidth   float64
	TotalGoals         int
	GoalDiff           int
	FavoriteTeam       string
	FavoritePct        float64
	FavoriteAdvantage  float64
	MostRepeatedLabel  string
	MostRepeatedPct    float64
	Highlighted        []ScorelineView
	TopScores          []ScorelineView
}

type ScorelineView struct {
	GoalsA  int
	GoalsB  int
	Count   int
	Percent float64
	Label   string
}

type Server struct {
	store          *store.SQLiteStore
	defaultPath    string
	batchOutputDir string
	tmpl           *template.Template
}

func NewServer(db *store.SQLiteStore, lambdaPath, batchOutputDir string) *Server {
	if strings.TrimSpace(lambdaPath) == "" {
		lambdaPath = defaultLambdaPath
	}
	if strings.TrimSpace(batchOutputDir) == "" {
		batchOutputDir = filepath.Join("out", "web_batch")
	}
	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"pct": func(count, total int) string {
			if total <= 0 {
				return "0.00"
			}
			return fmt.Sprintf("%.2f", float64(count)*100/float64(total))
		},
		"js": func(value string) template.JS {
			return template.JS(strconv.Quote(value))
		},
		"avg": func(total, sims int) string {
			if sims <= 0 {
				return "0.00"
			}
			return fmt.Sprintf("%.2f", float64(total)/float64(sims))
		},
	}).Parse(pageTemplateWeb))
	return &Server{
		store:          db,
		defaultPath:    lambdaPath,
		batchOutputDir: batchOutputDir,
		tmpl:           tmpl,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/simulate", s.handleSimulate)
	mux.HandleFunc("/process-batch", s.handleProcessBatch)
	mux.HandleFunc("/download-batch", s.handleDownloadBatch)
	mux.HandleFunc("/download-batch-template", s.handleDownloadBatchTemplate)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := s.baseData()
	if err := s.render(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "formulario invalido", http.StatusBadRequest)
		return
	}

	data := s.baseData()
	data.LambdaPath = strings.TrimSpace(r.FormValue("lambda_path"))
	if data.LambdaPath == "" {
		data.LambdaPath = s.defaultPath
	}
	data.TeamA = strings.TrimSpace(r.FormValue("team_a"))
	data.TeamB = strings.TrimSpace(r.FormValue("team_b"))
	data.MotivationA = normalizeMotivation(r.FormValue("motivation_a"), model.MotivationMedium.String())
	data.MotivationB = normalizeMotivation(r.FormValue("motivation_b"), model.MotivationMedium.String())
	data.Tiebreaker = normalizeTiebreaker(r.FormValue("tiebreaker"))
	data.ShotsA = parseIntOrDefault(r.FormValue("shots_a"), 5)
	data.ShotsB = parseIntOrDefault(r.FormValue("shots_b"), 5)
	data.Simulations = parseIntOrDefault(r.FormValue("simulations"), 6000)
	data.Seed = strings.TrimSpace(r.FormValue("seed"))

	if data.TeamA == "" || data.TeamB == "" {
		data.Error = "debes completar team A y team B"
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	rules, err := loader.LoadLambdaRules(data.LambdaPath)
	if err != nil {
		data.Error = fmt.Sprintf("no se pudo cargar lambda: %v", err)
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	motA, err := model.ParseMotivation(data.MotivationA)
	if err != nil {
		data.Error = err.Error()
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	motB, err := model.ParseMotivation(data.MotivationB)
	if err != nil {
		data.Error = err.Error()
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	seed, err := parseSeed(data.Seed)
	if err != nil {
		data.Error = err.Error()
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
		data.Seed = strconv.FormatInt(seed, 10)
	}

	input := sim.SingleMatchInput{
		TeamA:       data.TeamA,
		TeamB:       data.TeamB,
		ShotsA:      data.ShotsA,
		ShotsB:      data.ShotsB,
		MotivationA: motA,
		MotivationB: motB,
		Tiebreaker:  data.Tiebreaker,
	}

	summary, err := sim.RunSingleMatchSeries(seed, data.Simulations, rules, input)
	if err != nil {
		data.Error = err.Error()
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	report := buildMatchReport(store.SingleMatchRun{
		Seed:        seed,
		LambdaPath:  data.LambdaPath,
		TeamA:       input.TeamA,
		TeamB:       input.TeamB,
		ShotsA:      input.ShotsA,
		ShotsB:      input.ShotsB,
		MotivationA: input.MotivationA.String(),
		MotivationB: input.MotivationB.String(),
		Tiebreaker:  input.Tiebreaker,
		Simulations: summary.Simulations,
		WinsA:       summary.WinsA,
		WinsB:       summary.WinsB,
		GoalsA:      summary.GoalsA,
		GoalsB:      summary.GoalsB,
		Regulation:  summary.Regulation,
		Penalties:   summary.Penalties,
		RandomTie:   summary.RandomTie,
	}, summary)

	record, err := s.store.SaveSingleMatchRun(seed, data.LambdaPath, input, summary)
	if err != nil {
		data.Error = fmt.Sprintf("no se pudo guardar en sqlite: %v", err)
		data.Result = report
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	report.Run = record
	data.Result = report
	data.History, err = s.store.ListSingleMatchRuns(10)
	if err != nil {
		data.Error = fmt.Sprintf("no se pudo cargar historial: %v", err)
	}

	if err := s.render(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleProcessBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "archivo invalido o demasiado grande", http.StatusBadRequest)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	data := s.baseData()
	data.BatchLambdaPath = strings.TrimSpace(r.FormValue("batch_lambda_path"))
	if data.BatchLambdaPath == "" {
		data.BatchLambdaPath = s.defaultPath
	}
	data.BatchSeed = strings.TrimSpace(r.FormValue("batch_seed"))

	file, header, err := r.FormFile("batch_file")
	if err != nil {
		data.BatchError = "debes subir un archivo Excel .xlsx"
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	if header == nil || strings.TrimSpace(header.Filename) == "" {
		data.BatchError = "el archivo subido no tiene nombre valido"
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !strings.EqualFold(filepath.Ext(header.Filename), ".xlsx") {
		data.BatchError = "el archivo debe ser un .xlsx"
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	seed, err := parseSeed(data.BatchSeed)
	if err != nil {
		data.BatchError = err.Error()
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	rules, err := loader.LoadLambdaRules(data.BatchLambdaPath)
	if err != nil {
		data.BatchError = fmt.Sprintf("no se pudo cargar lambda: %v", err)
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := os.MkdirAll(s.batchOutputDir, 0o755); err != nil {
		data.BatchError = fmt.Sprintf("no se pudo crear el directorio de salida: %v", err)
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	baseName := sanitizeUploadedName(header.Filename)
	stamp := time.Now().Format("20060102_150405")
	inputPath := filepath.Join(s.batchOutputDir, fmt.Sprintf("%s_%s.xlsx", baseName, stamp))
	outputPath := report.BatchOutputPath(s.batchOutputDir, inputPath)

	if err := saveMultipartFile(inputPath, file); err != nil {
		data.BatchError = fmt.Sprintf("no se pudo guardar el archivo subido: %v", err)
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer func() {
		_ = os.Remove(inputPath)
	}()

	if err := report.ProcessBatchWorkbook(inputPath, outputPath, rules, seed); err != nil {
		data.BatchError = fmt.Sprintf("no se pudo procesar el Excel: %v", err)
		if err := s.render(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	downloadName := filepath.Base(outputPath)
	data.BatchMessage = fmt.Sprintf("Archivo procesado correctamente. Se genero %s.", downloadName)
	data.BatchDownloadName = downloadName
	data.BatchDownloadURL = "/download-batch?file=" + url.QueryEscape(downloadName)

	if err := s.render(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleDownloadBatch(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("file"))
	name = filepath.Base(name)
	if name == "" || name == "." {
		http.Error(w, "archivo invalido", http.StatusBadRequest)
		return
	}
	if !strings.EqualFold(filepath.Ext(name), ".xlsx") {
		http.Error(w, "archivo invalido", http.StatusBadRequest)
		return
	}

	path := filepath.Join(s.batchOutputDir, name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	http.ServeFile(w, r, path)
}

func (s *Server) handleDownloadBatchTemplate(w http.ResponseWriter, r *http.Request) {
	if err := os.MkdirAll(s.batchOutputDir, 0o755); err != nil {
		http.Error(w, fmt.Sprintf("no se pudo crear el directorio de salida: %v", err), http.StatusInternalServerError)
		return
	}

	templateName := "plantilla_batch_ejemplo.xlsx"
	templatePath := filepath.Join(s.batchOutputDir, templateName)
	if err := report.WriteBatchTemplateXLSX(templatePath, "Grupo Ejemplo", []string{"Ecuador", "Alemania", "Polonia", "Corea"}); err != nil {
		http.Error(w, fmt.Sprintf("no se pudo generar la plantilla: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, templateName))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	http.ServeFile(w, r, templatePath)
}

func (s *Server) baseData() PageData {
	history, err := s.store.ListSingleMatchRuns(10)
	if err != nil {
		history = nil
	}
	return PageData{
		LambdaPath:      s.defaultPath,
		TeamA:           "Equipo A",
		TeamB:           "Equipo B",
		ShotsA:          10,
		ShotsB:          4,
		MotivationA:     model.MotivationHigh.String(),
		MotivationB:     model.MotivationMedium.String(),
		Tiebreaker:      "penalties",
		Simulations:     6000,
		History:         history,
		BatchLambdaPath: s.defaultPath,
		BatchSeed:       "",
	}
}

func (s *Server) render(w http.ResponseWriter, data PageData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return s.tmpl.Execute(w, data)
}

func normalizeMotivation(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	motivation, err := model.ParseMotivation(value)
	if err != nil {
		return fallback
	}
	return motivation.String()
}

func normalizeTiebreaker(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "penalties", "random":
		return value
	default:
		return "penalties"
	}
}

func parseIntOrDefault(value string, fallback int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func parseSeed(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	seed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("semilla invalida: %q", value)
	}
	return seed, nil
}

func buildMatchReport(record store.SingleMatchRun, summary sim.SingleMatchSeries) *MatchReport {
	sims := summary.Simulations
	if sims <= 0 {
		sims = 1
	}

	report := &MatchReport{
		Run:             record,
		Simulations:     sims,
		WinPctA:         percentOf(summary.WinsA, sims),
		WinPctB:         percentOf(summary.WinsB, sims),
		DrawPct:         percentOf(summary.RegulationDraws, sims),
		RegulationDraws: summary.RegulationDraws,
		RegulationPct:   percentOf(summary.Regulation, sims),
		PenaltyPct:      percentOf(summary.Penalties, sims),
		RandomTiePct:    percentOf(summary.RandomTie, sims),
		AvgGoalsA:       float64(summary.GoalsA) / float64(sims),
		AvgGoalsB:       float64(summary.GoalsB) / float64(sims),
		AvgTotalGoals:   float64(summary.GoalsA+summary.GoalsB) / float64(sims),
		AvgGoalDiff:     float64(summary.GoalsA-summary.GoalsB) / float64(sims),
		TotalGoals:      summary.GoalsA + summary.GoalsB,
		GoalDiff:        summary.GoalsA - summary.GoalsB,
	}

	report.AvgGoalsAWidth = scaledWidth(report.AvgGoalsA, 5)
	report.AvgGoalsBWidth = scaledWidth(report.AvgGoalsB, 5)
	report.AvgTotalGoalsWidth = scaledWidth(report.AvgTotalGoals, 5)
	report.AvgGoalDiffWidth = scaledWidth(absFloat(report.AvgGoalDiff), 5)

	switch {
	case summary.WinsA > summary.WinsB:
		report.FavoriteTeam = record.TeamA
		report.FavoritePct = report.WinPctA
		report.FavoriteAdvantage = report.WinPctA - report.WinPctB
	case summary.WinsB > summary.WinsA:
		report.FavoriteTeam = record.TeamB
		report.FavoritePct = report.WinPctB
		report.FavoriteAdvantage = report.WinPctB - report.WinPctA
	default:
		report.FavoriteTeam = "Empate tecnico"
		report.FavoritePct = 50
		report.FavoriteAdvantage = 0
	}

	if summary.MostRepeatedCount > 0 {
		report.MostRepeatedLabel = fmt.Sprintf("%s %d - %d %s", record.TeamA, summary.MostRepeatedGoalsA, summary.MostRepeatedGoalsB, record.TeamB)
		report.MostRepeatedPct = summary.MostRepeatedPercent
	}

	report.Highlighted = buildHighlightedScorelines(record, summary)
	report.TopScores = make([]ScorelineView, 0, len(summary.TopScores))
	for _, score := range summary.TopScores {
		report.TopScores = append(report.TopScores, ScorelineView{
			GoalsA:  score.GoalsA,
			GoalsB:  score.GoalsB,
			Count:   score.Count,
			Percent: score.Percent,
			Label:   formatScorelineLabel(record.TeamA, score.GoalsA, score.GoalsB, record.TeamB),
		})
	}

	return report
}

func buildHighlightedScorelines(record store.SingleMatchRun, summary sim.SingleMatchSeries) []ScorelineView {
	candidates := [][2]int{
		{0, 0},
		{1, 0},
		{0, 1},
		{1, 1},
		{2, 0},
		{0, 2},
		{2, 1},
		{1, 2},
		{2, 2},
		{3, 1},
		{1, 3},
		{3, 2},
		{2, 3},
	}

	views := make([]ScorelineView, 0, len(candidates))
	for _, candidate := range candidates {
		count := summary.ScoreCounts[candidate]
		views = append(views, ScorelineView{
			GoalsA:  candidate[0],
			GoalsB:  candidate[1],
			Count:   count,
			Percent: percentOf(count, summary.Simulations),
			Label:   formatScorelineLabel(record.TeamA, candidate[0], candidate[1], record.TeamB),
		})
	}
	return views
}

func formatScorelineLabel(teamA string, goalsA int, goalsB int, teamB string) string {
	return fmt.Sprintf("%s %d - %d %s", teamA, goalsA, goalsB, teamB)
}

func percentOf(count, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(count) * 100 / float64(total)
}

func scaledWidth(value, max float64) float64 {
	if max <= 0 {
		return 0
	}
	width := value * 100 / max
	if width < 0 {
		return 0
	}
	if width > 100 {
		return 100
	}
	return width
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func saveMultipartFile(path string, src io.Reader) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func sanitizeUploadedName(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" || name == "." {
		return "batch"
	}
	name = strings.TrimSuffix(name, filepath.Ext(name))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		default:
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_-")
	if out == "" {
		return "batch"
	}
	return out
}

const pageTemplate = `
<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Scoreador - Partido unico</title>
  <style>
    :root {
      --bg: #0f172a;
      --panel: #111827;
      --panel-2: #1f2937;
      --text: #e5e7eb;
      --muted: #9ca3af;
      --accent: #22c55e;
      --accent-2: #38bdf8;
      --line: #334155;
      --danger: #f87171;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Segoe UI, Arial, sans-serif;
      background: radial-gradient(circle at top, #1d4ed8 0%, var(--bg) 50%);
      color: var(--text);
    }
    .wrap {
      max-width: 1200px;
      margin: 0 auto;
      padding: 24px;
    }
    .hero {
      display: grid;
      gap: 12px;
      grid-template-columns: 1.3fr 0.7fr;
      align-items: end;
      margin-bottom: 24px;
    }
    .card {
      background: rgba(17, 24, 39, 0.92);
      border: 1px solid rgba(148, 163, 184, 0.18);
      border-radius: 18px;
      padding: 20px;
      box-shadow: 0 20px 40px rgba(0,0,0,0.25);
      backdrop-filter: blur(10px);
    }
    h1, h2, h3 { margin: 0 0 12px; }
    h1 { font-size: 30px; }
    p { color: var(--muted); margin: 0; line-height: 1.5; }
    .grid {
      display: grid;
      gap: 18px;
      grid-template-columns: 1fr 1fr;
    }
    .form-grid {
      display: grid;
      gap: 12px;
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .field { display: grid; gap: 6px; }
    .field-help {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.35;
    }
    label { font-size: 12px; color: var(--muted); text-transform: uppercase; letter-spacing: .06em; }
    input, select {
      width: 100%;
      padding: 11px 12px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: var(--panel-2);
      color: var(--text);
      outline: none;
    }
    input:focus, select:focus { border-color: var(--accent-2); }
    .full { grid-column: 1 / -1; }
    .actions { display: flex; gap: 12px; align-items: center; margin-top: 14px; }
    button {
      border: 0;
      background: linear-gradient(135deg, var(--accent), #16a34a);
      color: #052e16;
      font-weight: 700;
      padding: 12px 16px;
      border-radius: 12px;
      cursor: pointer;
    }
    .hint { color: var(--muted); font-size: 13px; }
    .error {
      margin-bottom: 18px;
      padding: 12px 14px;
      border-radius: 12px;
      background: rgba(248, 113, 113, .12);
      border: 1px solid rgba(248, 113, 113, .35);
      color: #fecaca;
    }
    .result-grid {
      display: grid;
      gap: 12px;
      grid-template-columns: repeat(4, minmax(0, 1fr));
    }
    .metric {
      background: rgba(31, 41, 55, 0.8);
      border: 1px solid rgba(148, 163, 184, 0.14);
      border-radius: 14px;
      padding: 14px;
    }
    .metric span { display: block; color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .06em; }
    .metric strong { display: block; margin-top: 6px; font-size: 20px; }
    table {
      width: 100%;
      border-collapse: collapse;
      overflow: hidden;
      border-radius: 14px;
    }
    th, td {
      padding: 10px 12px;
      border-bottom: 1px solid rgba(148, 163, 184, 0.14);
      text-align: left;
      vertical-align: top;
    }
    th { color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .06em; }
    tr:hover td { background: rgba(59, 130, 246, 0.06); }
    .badge {
      display: inline-block;
      padding: 4px 8px;
      border-radius: 999px;
      background: rgba(56, 189, 248, .12);
      color: #bae6fd;
      font-size: 12px;
    }
    .section { margin-top: 18px; }
    .footer-note { color: var(--muted); font-size: 12px; margin-top: 10px; }
    @media (max-width: 900px) {
      .hero, .grid, .result-grid, .form-grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="hero">
      <div class="card">
        <h1>Simulador de partido unico</h1>
        <p>Corre una serie de simulaciones, guarda el resultado en SQLite y revisa el historial desde el navegador.</p>
      </div>
      <div class="card">
        <span class="badge">GUI local</span>
        <p class="footer-note">Abre esta pagina en tu navegador y ajusta equipos, tiros, motivacion 0-10, semilla y cantidad de corridas.</p>
      </div>
    </div>

    {{if .Error}}
    <div class="error">{{.Error}}</div>
    {{end}}

    <div class="grid">
      <div class="card">
        <h2>Nuevo partido</h2>
        <form method="post" action="/simulate">
          <div class="form-grid">
            <div class="field">
              <label>Ruta lambda</label>
              <input name="lambda_path" value="{{.LambdaPath}}">
            </div>
            <div class="field">
              <label>Semilla</label>
              <input name="seed" value="{{.Seed}}" placeholder="Opcional">
            </div>
            <div class="field">
              <label>Equipo A</label>
              <input name="team_a" value="{{.TeamA}}">
            </div>
            <div class="field">
              <label>Equipo B</label>
              <input name="team_b" value="{{.TeamB}}">
            </div>
            <div class="field">
              <label>Tiros A</label>
              <input name="shots_a" type="number" min="1" value="{{.ShotsA}}">
            </div>
            <div class="field">
              <label>Tiros B</label>
              <input name="shots_b" type="number" min="1" value="{{.ShotsB}}">
            </div>
            <div class="field">
              <label>Motivacion A</label>
              <input name="motivation_a" type="number" min="0" max="10" step="1" value="{{.MotivationA}}" placeholder="0-10" title="0 = nada motivado, 10 = maxima motivacion">
              <span class="field-help">0 = nada motivado, 10 = maxima motivacion.</span>
            </div>
            <div class="field">
              <label>Motivacion B</label>
              <input name="motivation_b" type="number" min="0" max="10" step="1" value="{{.MotivationB}}" placeholder="0-10" title="0 = nada motivado, 10 = maxima motivacion">
              <span class="field-help">0 = nada motivado, 10 = maxima motivacion.</span>
            </div>
            <div class="field">
              <label>Desempate</label>
              <select name="tiebreaker">
                <option value="penalties" {{if eq .Tiebreaker "penalties"}}selected{{end}}>penalties</option>
                <option value="random" {{if eq .Tiebreaker "random"}}selected{{end}}>random</option>
              </select>
            </div>
            <div class="field">
              <label>Simulaciones</label>
              <input name="simulations" type="number" min="1" value="{{.Simulations}}">
            </div>
          </div>
          <div class="actions">
            <button type="submit">Simular y guardar</button>
            <span class="hint">Ejemplo: 6000 corridas para un resumen estable.</span>
          </div>
        </form>
      </div>

      <div class="card">
        <h2>Resultado actual</h2>
        {{if .Result}}
        <p><strong>{{.Result.TeamA}}</strong> vs <strong>{{.Result.TeamB}}</strong></p>
        <div class="section result-grid">
          <div class="metric"><span>Victorias A</span><strong>{{.Result.WinsA}}</strong></div>
          <div class="metric"><span>Victorias B</span><strong>{{.Result.WinsB}}</strong></div>
          <div class="metric"><span>Promedio goles A</span><strong>{{avg .Result.GoalsA .Result.Simulations}}</strong></div>
          <div class="metric"><span>Promedio goles B</span><strong>{{avg .Result.GoalsB .Result.Simulations}}</strong></div>
          <div class="metric"><span>Tiempo regular</span><strong>{{.Result.Regulation}}</strong></div>
          <div class="metric"><span>Penales</span><strong>{{.Result.Penalties}}</strong></div>
          <div class="metric"><span>Sorteo</span><strong>{{.Result.RandomTie}}</strong></div>
          <div class="metric"><span>Simulaciones</span><strong>{{.Result.Simulations}}</strong></div>
        </div>
        <div class="section">
          <p>Semilla: {{.Result.Seed}}</p>
          <p>Lambda: {{.Result.LambdaPath}}</p>
          <p>Motivacion: {{.Result.MotivationA}}/10 vs {{.Result.MotivationB}}/10</p>
          <p>Tiros: {{.Result.ShotsA}} vs {{.Result.ShotsB}}</p>
          <p>Desempate: {{.Result.Tiebreaker}}</p>
          <p class="footer-note">Victorias: {{pct .Result.WinsA .Result.Simulations}}% / {{pct .Result.WinsB .Result.Simulations}}%</p>
        </div>
        {{else}}
        <p>Aun no hay una simulacion guardada en esta sesion.</p>
        {{end}}
      </div>
    </div>

    <div class="card section">
      <h2>Historial reciente</h2>
      {{if .History}}
      <table>
        <thead>
          <tr>
            <th>Fecha</th>
            <th>Partido</th>
            <th>Simulaciones</th>
            <th>Victorias</th>
            <th>Goles promedio</th>
            <th>Definicion</th>
            <th>Semilla</th>
          </tr>
        </thead>
        <tbody>
          {{range .History}}
          <tr>
            <td>{{.DisplayCreatedAt}}</td>
            <td>{{.TeamA}} vs {{.TeamB}}</td>
            <td>{{.Simulations}}</td>
            <td>{{.WinsA}} - {{.WinsB}}</td>
            <td>{{avg .GoalsA .Simulations}} - {{avg .GoalsB .Simulations}}</td>
            <td>{{.Regulation}} reg / {{.Penalties}} pen / {{.RandomTie}} sort</td>
            <td>{{.Seed}}</td>
          </tr>
          {{end}}
        </tbody>
      </table>
      {{else}}
      <p>No hay partidas guardadas todavia.</p>
      {{end}}
    </div>
  </div>
</body>
</html>`

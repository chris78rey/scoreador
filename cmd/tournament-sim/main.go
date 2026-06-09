package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"scoreador/internal/loader"
	"scoreador/internal/model"
	"scoreador/internal/preset/worldcup2026"
	"scoreador/internal/report"
	"scoreador/internal/sim"
)

func main() {
	var (
		preset          = flag.String("preset", "", "Preset de datos a cargar (por ejemplo: worldcup2026)")
		configPath      = flag.String("config", "", "Ruta al archivo JSON de configuracion")
		matchesPath     = flag.String("matches", "", "Ruta al archivo CSV de partidos")
		lambdaPath      = flag.String("lambda", "", "Ruta al archivo CSV de tabla lambda")
		outDir          = flag.String("outdir", "out", "Directorio de salida")
		seedFlag        = flag.Int64("seed", 0, "Semilla opcional para reproducibilidad")
		simulations     = flag.Int("simulations", 1, "Cantidad de simulaciones para -single-match")
		groupTemplate   = flag.Bool("group-template", false, "Genera una plantilla Excel para ingresar resultados de un grupo manualmente")
		groupDemo       = flag.Bool("group-demo", false, "Genera un grupo demo completo con marcadores de ejemplo")
		groupName       = flag.String("group-name", "Grupo", "Nombre del grupo para la plantilla manual")
		groupTeams      = flag.String("group-teams", "", "Lista de equipos separada por coma para la plantilla manual")
		batchTemplate   = flag.Bool("batch-template", false, "Genera una plantilla Excel para varios partidos con simulaciones por fila")
		batchInputPath  = flag.String("batch-xlsx", "", "Ruta al Excel de entrada batch a procesar")
		batchOutputPath = flag.String("batch-out", "", "Ruta del Excel de salida batch")

		singleMatch = flag.Bool("single-match", false, "Simula un partido aislado sin pasar por el torneo")
		teamA       = flag.String("team-a", "", "Nombre del equipo A")
		teamB       = flag.String("team-b", "", "Nombre del equipo B")
		shotsA      = flag.Int("shots-a", 5, "Tiros esperados del equipo A")
		shotsB      = flag.Int("shots-b", 5, "Tiros esperados del equipo B")
		motivationA = flag.Int("motivation-a", int(model.MotivationMedium), "Motivacion del equipo A: 0 a 10")
		motivationB = flag.Int("motivation-b", int(model.MotivationMedium), "Motivacion del equipo B: 0 a 10")
		tiebreaker  = flag.String("tiebreaker", "penalties", "Desempate en caso de empate: penalties o random")
	)
	flag.Parse()

	if *groupTemplate && *groupDemo {
		fmt.Fprintln(os.Stderr, "Debes usar solo uno: -group-template o -group-demo")
		os.Exit(1)
	}
	if *batchTemplate && *batchInputPath != "" {
		fmt.Fprintln(os.Stderr, "Debes usar solo uno: -batch-template o -batch-xlsx")
		os.Exit(1)
	}
	if *batchTemplate && *singleMatch {
		fmt.Fprintln(os.Stderr, "Debes usar solo uno: -batch-template o -single-match")
		os.Exit(1)
	}
	if *batchInputPath != "" && *singleMatch {
		fmt.Fprintln(os.Stderr, "Debes usar solo uno: -batch-xlsx o -single-match")
		os.Exit(1)
	}

	if *groupDemo {
		var teams []string
		if strings.TrimSpace(*groupTeams) == "" {
			teams = []string{"Ecuador", "Alemania", "Polonia", "Corea"}
		} else {
			teams = parseTeamList(*groupTeams)
		}
		if len(teams) < 2 {
			fmt.Fprintln(os.Stderr, "En -group-demo debes indicar al menos 2 equipos validos en -group-teams")
			os.Exit(1)
		}
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fatal(err)
		}
		results := report.DefaultGroupDemoResults(teams)
		path := report.GroupDemoPath(*outDir, *groupName)
		if err := report.WriteGroupDemoXLSX(path, *groupName, teams, results); err != nil {
			fatal(err)
		}
		standings := report.GroupDemoStandings(*groupName, results)
		printGroupDemoOutput(path, *groupName, teams, results, standings)
		return
	}

	if *groupTemplate {
		if strings.TrimSpace(*groupTeams) == "" {
			fmt.Fprintln(os.Stderr, "En -group-template debes indicar -group-teams con al menos 2 equipos separados por coma")
			os.Exit(1)
		}
		teams := parseTeamList(*groupTeams)
		if len(teams) < 2 {
			fmt.Fprintln(os.Stderr, "En -group-template debes indicar al menos 2 equipos validos en -group-teams")
			os.Exit(1)
		}
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fatal(err)
		}
		path := report.GroupTemplatePath(*outDir, *groupName)
		if err := report.WriteGroupTemplateXLSX(path, *groupName, teams); err != nil {
			fatal(err)
		}
		fmt.Printf("Plantilla de grupo generada: %s\n", path)
		return
	}

	if *batchTemplate {
		if strings.TrimSpace(*groupTeams) == "" {
			fmt.Fprintln(os.Stderr, "En -batch-template debes indicar -group-teams con al menos 2 equipos separados por coma")
			os.Exit(1)
		}
		teams := parseTeamList(*groupTeams)
		if len(teams) < 2 {
			fmt.Fprintln(os.Stderr, "En -batch-template debes indicar al menos 2 equipos validos en -group-teams")
			os.Exit(1)
		}
		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			fatal(err)
		}
		path := report.BatchTemplatePath(*outDir, *groupName)
		if err := report.WriteBatchTemplateXLSX(path, *groupName, teams); err != nil {
			fatal(err)
		}
		fmt.Printf("Plantilla batch generada: %s\n", path)
		return
	}

	if strings.TrimSpace(*batchInputPath) != "" {
		if strings.TrimSpace(*lambdaPath) == "" {
			fmt.Fprintln(os.Stderr, "En -batch-xlsx debes indicar -lambda")
			os.Exit(1)
		}
		rules, err := loader.LoadLambdaRules(*lambdaPath)
		if err != nil {
			fatal(err)
		}
		outputPath := strings.TrimSpace(*batchOutputPath)
		if outputPath == "" {
			if err := os.MkdirAll(*outDir, 0o755); err != nil {
				fatal(err)
			}
			outputPath = report.BatchOutputPath(*outDir, *batchInputPath)
		}
		if err := report.ProcessBatchWorkbook(*batchInputPath, outputPath, rules, *seedFlag); err != nil {
			fatal(err)
		}
		fmt.Printf("Batch procesado: %s\n", outputPath)
		return
	}

	if *singleMatch {
		runSingleMatchMode(*lambdaPath, *seedFlag, *simulations, *teamA, *teamB, *shotsA, *shotsB, *motivationA, *motivationB, *tiebreaker)
		return
	}

	var (
		cfg         model.Config
		matches     []model.MatchInput
		lambdaRules []model.LambdaRule
		err         error
	)

	switch *preset {
	case "worldcup2026":
		data := worldcup2026.Load(*seedFlag)
		cfg = data.Config
		if *seedFlag != 0 {
			cfg.Seed = *seedFlag
		}
		matches = data.Matches
		lambdaRules = data.Rules
	default:
		if *preset != "" {
			fmt.Fprintf(os.Stderr, "Preset no soportado: %s\n", *preset)
			os.Exit(1)
		}
		if *configPath == "" || *matchesPath == "" || *lambdaPath == "" {
			fmt.Fprintln(os.Stderr, "Uso: go run ./cmd/tournament-sim -config config.json -matches matches.csv -lambda lambda.csv -outdir out [-seed 123]")
			fmt.Fprintln(os.Stderr, "   o: go run ./cmd/tournament-sim -preset worldcup2026 -outdir out [-seed 123]")
			fmt.Fprintln(os.Stderr, "   o: go run ./cmd/tournament-sim -single-match -team-a \"A\" -team-b \"B\" -lambda lambda.csv [-simulations 6000] [-shots-a 5 -shots-b 5 -motivation-a 5 -motivation-b 5 -tiebreaker penalties] [-seed 123]")
			fmt.Fprintln(os.Stderr, "   o: go run ./cmd/tournament-sim -group-template -group-name \"Grupo A\" -group-teams \"Ecuador,Alemania,Polonia,Corea\" -outdir out")
			fmt.Fprintln(os.Stderr, "   o: go run ./cmd/tournament-sim -group-demo -group-name \"Grupo A\" -group-teams \"Ecuador,Alemania,Polonia,Corea\" -outdir out")
			fmt.Fprintln(os.Stderr, "   o: go run ./cmd/tournament-sim -batch-template -group-name \"Grupo A\" -group-teams \"Ecuador,Alemania,Polonia,Corea\" -outdir out")
			fmt.Fprintln(os.Stderr, "   o: go run ./cmd/tournament-sim -batch-xlsx grupo_a_input.xlsx -lambda lambda.csv [-batch-out salida.xlsx] [-seed 123]")
			os.Exit(1)
		}

		cfg, err = loader.LoadConfig(*configPath)
		if err != nil {
			fatal(err)
		}
		if *seedFlag != 0 {
			cfg.Seed = *seedFlag
		}

		matches, err = loader.LoadMatches(*matchesPath)
		if err != nil {
			fatal(err)
		}

		lambdaRules, err = loader.LoadLambdaRules(*lambdaPath)
		if err != nil {
			fatal(err)
		}
	}

	cfg.ApplyDefaults()
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	summary, err := sim.RunMonteCarlo(cfg, matches, lambdaRules)
	if err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal(err)
	}

	csvPath := filepath.Join(*outDir, "summary.csv")
	jsonPath := filepath.Join(*outDir, "summary.json")
	xlsxPath := filepath.Join(*outDir, "summary.xlsx")

	if err := report.WriteCSV(csvPath, summary); err != nil {
		fatal(err)
	}
	if err := report.WriteJSON(jsonPath, summary); err != nil {
		fatal(err)
	}
	if err := report.WriteXLSX(xlsxPath, cfg, matches, lambdaRules, summary); err != nil {
		fatal(err)
	}

	fmt.Printf("Simulacion completada: %s\n", summary.Name)
	fmt.Printf("CSV: %s\n", csvPath)
	fmt.Printf("JSON: %s\n", jsonPath)
	fmt.Printf("XLSX: %s\n", xlsxPath)
	fmt.Printf("Semilla: %d\n", cfg.Seed)
}

func parseTeamList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case ',', ';', '\n', '\t':
			return true
		default:
			return false
		}
	})
	teams := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		teams = append(teams, name)
	}
	return teams
}

func runSingleMatchMode(lambdaPath string, seed int64, simulations int, teamA, teamB string, shotsA, shotsB int, motivationA, motivationB int, tiebreaker string) {
	if strings.TrimSpace(teamA) == "" || strings.TrimSpace(teamB) == "" {
		fmt.Fprintln(os.Stderr, "En -single-match debes indicar -team-a y -team-b")
		os.Exit(1)
	}
	if strings.TrimSpace(lambdaPath) == "" {
		fmt.Fprintln(os.Stderr, "En -single-match debes indicar -lambda")
		os.Exit(1)
	}

	rules, err := loader.LoadLambdaRules(lambdaPath)
	if err != nil {
		fatal(err)
	}

	motA, err := model.MotivationFromInt(motivationA)
	if err != nil {
		fatal(err)
	}
	motB, err := model.MotivationFromInt(motivationB)
	if err != nil {
		fatal(err)
	}

	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	input := sim.SingleMatchInput{
		TeamA:       teamA,
		TeamB:       teamB,
		ShotsA:      shotsA,
		ShotsB:      shotsB,
		MotivationA: motA,
		MotivationB: motB,
		Tiebreaker:  tiebreaker,
	}
	if simulations <= 1 {
		result, err := sim.RunSingleMatch(seed, rules, input)
		if err != nil {
			fatal(err)
		}

		fmt.Println("Partido unico")
		printMatchContext(lambdaPath, seed, 1, input)
		fmt.Println("Resultado")
		fmt.Printf("%s %d - %d %s\n", result.TeamA, result.GoalsA, result.GoalsB, result.TeamB)
		fmt.Printf("Ganador: %s\n", result.Winner)
		fmt.Printf("Decidido por: %s\n", result.DecidedBy)
		fmt.Printf("Lambda aplicada: %.2f vs %.2f\n", result.LambdaA, result.LambdaB)
		fmt.Printf("Diferencia de goles: %+d\n", result.GoalsA-result.GoalsB)
		return
	}

	summary, err := sim.RunSingleMatchSeries(seed, simulations, rules, input)
	if err != nil {
		fatal(err)
	}

	fmt.Println("Resumen partido unico")
	printMatchContext(lambdaPath, seed, simulations, input)
	printSingleMatchSeriesReport(summary)
	if len(summary.TopScores) > 0 {
		fmt.Println("Top marcadores")
		for i, score := range summary.TopScores {
			fmt.Printf("%2d. %s %d - %d %s | %d veces | %.2f%%\n", i+1, summary.TeamA, score.GoalsA, score.GoalsB, summary.TeamB, score.Count, score.Percent)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func printGroupDemoOutput(path, groupName string, teams []string, results []report.GroupMatchResult, standings []model.GroupStanding) {
	type record struct {
		Wins   int
		Draws  int
		Losses int
	}

	records := map[string]record{}
	for _, result := range results {
		switch {
		case result.GoalsA > result.GoalsB:
			ra := records[result.TeamA]
			ra.Wins++
			records[result.TeamA] = ra
			rb := records[result.TeamB]
			rb.Losses++
			records[result.TeamB] = rb
		case result.GoalsB > result.GoalsA:
			ra := records[result.TeamA]
			ra.Losses++
			records[result.TeamA] = ra
			rb := records[result.TeamB]
			rb.Wins++
			records[result.TeamB] = rb
		default:
			ra := records[result.TeamA]
			ra.Draws++
			records[result.TeamA] = ra
			rb := records[result.TeamB]
			rb.Draws++
			records[result.TeamB] = rb
		}
	}

	fmt.Printf("Grupo demo generado: %s\n", path)
	fmt.Printf("Grupo: %s\n", groupName)
	fmt.Printf("Equipos: %s\n", strings.Join(teams, ", "))
	fmt.Println("Partidos")
	for i, result := range results {
		fmt.Printf("%2d. %s %d - %d %s\n", i+1, result.TeamA, result.GoalsA, result.GoalsB, result.TeamB)
	}
	fmt.Println("Tabla de posiciones")
	fmt.Println("Pos  Equipo           PJ PG PE PP GF GC DG Pts")
	for i, standing := range standings {
		fmt.Printf("%-4d %-15s %-2d %-2d %-2d %-2d %-2d %-2d %+3d %-3d\n",
			i+1,
			standing.Team,
			standing.Played,
			records[standing.Team].Wins,
			records[standing.Team].Draws,
			records[standing.Team].Losses,
			standing.GoalsFor,
			standing.GoalsAgainst,
			standing.GoalDifference,
			standing.Points,
		)
	}
}

func printMatchContext(lambdaPath string, seed int64, simulations int, input sim.SingleMatchInput) {
	fmt.Println("Contexto")
	fmt.Printf("Equipos: %s vs %s\n", input.TeamA, input.TeamB)
	fmt.Printf("Simulaciones: %d\n", simulations)
	fmt.Printf("Tiros esperados: %d vs %d\n", input.ShotsA, input.ShotsB)
	fmt.Printf("Motivacion: %s/10 vs %s/10\n", input.MotivationA, input.MotivationB)
	fmt.Printf("Desempate: %s\n", input.Tiebreaker)
	fmt.Printf("Lambda: %s\n", lambdaPath)
	fmt.Printf("Semilla: %d\n", seed)
}

func printSingleMatchSeriesReport(summary sim.SingleMatchSeries) {
	total := summary.Simulations
	if total <= 0 {
		return
	}

	avgGoalsA := float64(summary.GoalsA) / float64(total)
	avgGoalsB := float64(summary.GoalsB) / float64(total)
	avgTotal := avgGoalsA + avgGoalsB
	goalDiffAvg := avgGoalsA - avgGoalsB
	winPctA := percentage(summary.WinsA, total)
	winPctB := percentage(summary.WinsB, total)
	drawPct := percentage(summary.RegulationDraws, total)
	regPct := percentage(summary.Regulation, total)
	penPct := percentage(summary.Penalties, total)
	randomPct := percentage(summary.RandomTie, total)

	favoredTeam := "Empate estadistico"
	favoredPct := 0.0
	if summary.WinsA > summary.WinsB {
		favoredTeam = summary.TeamA
		favoredPct = winPctA
	} else if summary.WinsB > summary.WinsA {
		favoredTeam = summary.TeamB
		favoredPct = winPctB
	}

	fmt.Println("Lectura rapida")
	fmt.Printf("Favorito estadistico: %s\n", favoredTeam)
	fmt.Printf("Ventaja en victorias: %.2f puntos porcentuales\n", absFloat(winPctA-winPctB))
	if favoredPct > 0 {
		fmt.Printf("Peso del favorito: %.2f%%\n", favoredPct)
	}
	fmt.Printf("Marcador mas repetido: %s %d - %d %s (%d veces, %.2f%%)\n", summary.TeamA, summary.MostRepeatedGoalsA, summary.MostRepeatedGoalsB, summary.TeamB, summary.MostRepeatedCount, summary.MostRepeatedPercent)
	fmt.Printf("Promedio total de goles: %.2f\n", avgTotal)
	fmt.Printf("Diferencia promedio de goles: %+0.2f\n", goalDiffAvg)

	fmt.Println("Distribucion de resultados")
	fmt.Printf("%s: %d victorias (%.2f%%)\n", summary.TeamA, summary.WinsA, winPctA)
	fmt.Printf("%s: %d victorias (%.2f%%)\n", summary.TeamB, summary.WinsB, winPctB)
	fmt.Printf("Empates en tiempo regular: %d (%.2f%%)\n", summary.RegulationDraws, drawPct)
	fmt.Printf("Definidos en tiempo regular: %d (%.2f%%)\n", summary.Regulation, regPct)
	fmt.Printf("Definidos por penales: %d (%.2f%%)\n", summary.Penalties, penPct)
	fmt.Printf("Definidos por sorteo: %d (%.2f%%)\n", summary.RandomTie, randomPct)

	fmt.Println("Produccion ofensiva")
	fmt.Printf("%s: promedio %.2f goles\n", summary.TeamA, avgGoalsA)
	fmt.Printf("%s: promedio %.2f goles\n", summary.TeamB, avgGoalsB)
	printExactScorelineReport(summary)
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func printExactScorelineReport(summary sim.SingleMatchSeries) {
	if len(summary.ScoreCounts) == 0 {
		return
	}

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

	fmt.Println("Marcadores exactos destacados")
	for _, score := range candidates {
		count := summary.ScoreCounts[score]
		percent := percentage(count, summary.Simulations)
		fmt.Printf("%s %d - %d %s | %d veces | %.2f%%\n", summary.TeamA, score[0], score[1], summary.TeamB, count, percent)
	}
}

func percentage(count, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(count) * 100 / float64(total)
}

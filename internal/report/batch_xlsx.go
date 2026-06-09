package report

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"scoreador/internal/model"
	"scoreador/internal/sim"
)

type BatchMatchInput struct {
	RowNumber   int
	MatchID     int
	Group       string
	TeamA       string
	TeamB       string
	ShotsA      int
	ShotsB      int
	MotivationA model.Motivation
	MotivationB model.Motivation
	Simulations int
	Seed        int64
	Tiebreaker  string
}

type BatchMatchReport struct {
	Input                   BatchMatchInput
	SeedUsed                int64
	Series                  sim.SingleMatchSeries
	RepresentativeGoalsA    int
	RepresentativeGoalsB    int
	RepresentativeWinner    string
	RepresentativeDecidedBy string
}

type batchStyles struct {
	title   int
	header  int
	body    int
	altBody int
	input   int
	note    int
	result  int
}

func BatchTemplatePath(outDir, groupName string) string {
	fileName := fmt.Sprintf("grupo_%s_input.xlsx", sanitizeTemplateFileName(groupName))
	return filepath.Join(outDir, fileName)
}

func BatchOutputPath(outDir, inputPath string) string {
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if base == "" {
		base = "batch"
	}
	return filepath.Join(outDir, base+"_procesado.xlsx")
}

// WriteBatchTemplateXLSX creates an Excel template that the user fills row by row.
// Each row represents one match to be simulated by Go.
func WriteBatchTemplateXLSX(path string, groupName string, teams []string) error {
	teams = cleanTeams(teams)
	if len(teams) < 2 {
		return fmt.Errorf("debes indicar al menos 2 equipos para la plantilla batch")
	}

	pairings := buildPairings(teams)
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	if err := f.SetSheetName("Sheet1", "Instrucciones"); err != nil {
		return err
	}
	if _, err := f.NewSheet("Partidos"); err != nil {
		return err
	}

	styles, err := newBatchStyles(f)
	if err != nil {
		return err
	}

	if err := buildBatchInstructionsSheet(f, groupName, teams, styles); err != nil {
		return err
	}
	templateRows := make([]BatchMatchInput, 0, len(pairings))
	for _, pair := range pairings {
		templateRows = append(templateRows, BatchMatchInput{
			MatchID:     pair.Index,
			Group:       groupName,
			TeamA:       pair.TeamA,
			TeamB:       pair.TeamB,
			ShotsA:      5,
			ShotsB:      5,
			MotivationA: model.MotivationMedium,
			MotivationB: model.MotivationMedium,
			Simulations: 1,
			Tiebreaker:  "penalties",
		})
	}
	if err := buildBatchInputSheet(f, groupName, templateRows, styles, true); err != nil {
		return err
	}

	if idx, err := f.GetSheetIndex("Partidos"); err == nil {
		f.SetActiveSheet(idx)
	}

	if err := ensureDir(path); err != nil {
		return err
	}
	return f.SaveAs(path)
}

// ProcessBatchWorkbook reads the input workbook, simulates each row and writes a new
// Excel file with the same input sheet plus statistics, top scores and standings.
func ProcessBatchWorkbook(inputPath, outputPath string, rules []model.LambdaRule, globalSeed int64) error {
	inputRows, groupName, err := readBatchInputs(inputPath)
	if err != nil {
		return err
	}
	if len(inputRows) == 0 {
		return errors.New("no se encontraron partidos para procesar")
	}
	if len(rules) == 0 {
		return errors.New("no hay reglas lambda cargadas")
	}

	baseSeed := globalSeed
	if baseSeed == 0 {
		baseSeed = time.Now().UnixNano()
	}

	reports := make([]BatchMatchReport, 0, len(inputRows))
	representativeResults := make([]GroupMatchResult, 0, len(inputRows))

	for idx, row := range inputRows {
		seed := row.Seed
		if seed == 0 {
			seed = baseSeed + int64(idx+1)*7919
		}

		series, err := sim.RunSingleMatchSeries(seed, row.Simulations, rules, sim.SingleMatchInput{
			TeamA:       row.TeamA,
			TeamB:       row.TeamB,
			ShotsA:      row.ShotsA,
			ShotsB:      row.ShotsB,
			MotivationA: row.MotivationA,
			MotivationB: row.MotivationB,
			Tiebreaker:  row.Tiebreaker,
		})
		if err != nil {
			return fmt.Errorf("fila %d: %w", row.RowNumber, err)
		}

		repWinner, repDecidedBy := batchRepresentativeOutcome(row.TeamA, row.TeamB, series.MostRepeatedGoalsA, series.MostRepeatedGoalsB)
		reports = append(reports, BatchMatchReport{
			Input:                   row,
			SeedUsed:                seed,
			Series:                  series,
			RepresentativeGoalsA:    series.MostRepeatedGoalsA,
			RepresentativeGoalsB:    series.MostRepeatedGoalsB,
			RepresentativeWinner:    repWinner,
			RepresentativeDecidedBy: repDecidedBy,
		})
		representativeResults = append(representativeResults, GroupMatchResult{
			TeamA:  row.TeamA,
			TeamB:  row.TeamB,
			GoalsA: series.MostRepeatedGoalsA,
			GoalsB: series.MostRepeatedGoalsB,
		})
	}

	standings := computeGroupStandings(groupName, representativeResults)
	records := buildStandingRecords(representativeResults)

	wb := excelize.NewFile()
	defer func() {
		_ = wb.Close()
	}()

	if err := wb.SetSheetName("Sheet1", "Instrucciones"); err != nil {
		return err
	}
	if _, err := wb.NewSheet("Partidos"); err != nil {
		return err
	}
	if _, err := wb.NewSheet("Resultados"); err != nil {
		return err
	}
	if _, err := wb.NewSheet("Top Marcadores"); err != nil {
		return err
	}
	if _, err := wb.NewSheet("Posiciones"); err != nil {
		return err
	}
	if _, err := wb.NewSheet("Resumen"); err != nil {
		return err
	}

	styles, err := newBatchStyles(wb)
	if err != nil {
		return err
	}

	if err := buildBatchInstructionsSheet(wb, groupName, extractTeamsFromBatchInputs(inputRows), styles); err != nil {
		return err
	}
	if err := buildBatchInputSheet(wb, groupName, inputRows, styles, false); err != nil {
		return err
	}
	if err := buildBatchResultsSheet(wb, groupName, reports, styles); err != nil {
		return err
	}
	if err := buildBatchTopScoresSheet(wb, reports, styles); err != nil {
		return err
	}
	if err := buildBatchPositionsSheet(wb, groupName, standings, records, styles); err != nil {
		return err
	}
	if err := buildBatchSummarySheet(wb, groupName, reports, standings, styles); err != nil {
		return err
	}

	if idx, err := wb.GetSheetIndex("Posiciones"); err == nil {
		wb.SetActiveSheet(idx)
	}

	if err := ensureDir(outputPath); err != nil {
		return err
	}
	return wb.SaveAs(outputPath)
}

func newBatchStyles(f *excelize.File) (batchStyles, error) {
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#1F4E78"}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return batchStyles{}, err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#1F4E78"}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return batchStyles{}, err
	}
	bodyStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#1F1F1F"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9E1F2", Style: 1},
			{Type: "right", Color: "#D9E1F2", Style: 1},
			{Type: "top", Color: "#D9E1F2", Style: 1},
			{Type: "bottom", Color: "#D9E1F2", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return batchStyles{}, err
	}
	altBodyStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#1F1F1F"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F7FAFD"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9E1F2", Style: 1},
			{Type: "right", Color: "#D9E1F2", Style: 1},
			{Type: "top", Color: "#D9E1F2", Style: 1},
			{Type: "bottom", Color: "#D9E1F2", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return batchStyles{}, err
	}
	inputStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#1F1F1F"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FFF2CC"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9E1F2", Style: 1},
			{Type: "right", Color: "#D9E1F2", Style: 1},
			{Type: "top", Color: "#D9E1F2", Style: 1},
			{Type: "bottom", Color: "#D9E1F2", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return batchStyles{}, err
	}
	noteStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Color: "#6B6B6B"},
	})
	if err != nil {
		return batchStyles{}, err
	}
	resultStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#1F1F1F"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2F0D9"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9E1F2", Style: 1},
			{Type: "right", Color: "#D9E1F2", Style: 1},
			{Type: "top", Color: "#D9E1F2", Style: 1},
			{Type: "bottom", Color: "#D9E1F2", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return batchStyles{}, err
	}

	return batchStyles{
		title:   titleStyle,
		header:  headerStyle,
		body:    bodyStyle,
		altBody: altBodyStyle,
		input:   inputStyle,
		note:    noteStyle,
		result:  resultStyle,
	}, nil
}

func buildBatchInstructionsSheet(f *excelize.File, groupName string, teams []string, styles batchStyles) error {
	sheet := "Instrucciones"
	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Plantilla batch del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Completa una fila por partido. Go leerá este Excel, simulará cada fila y generará otro libro con resultados, top 10 y posiciones."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", styles.note); err != nil {
		return err
	}

	info := [][]interface{}{
		{"grupo", groupName},
		{"equipos", len(teams)},
		{"partidos", len(teams) * (len(teams) - 1) / 2},
		{"simulaciones_maximas_por_fila", 10000},
		{"motivacion", "0 a 10 o baja/media/alta"},
		{"seed", "opcional; si queda vacio se deriva una semilla"},
	}
	if err := writeHeaderRow(f, sheet, 4, []string{"campo", "valor"}, styles.header); err != nil {
		return err
	}
	for idx, row := range info {
		excelRow := 5 + idx
		if err := writeRow(f, sheet, excelRow, row); err != nil {
			return err
		}
		style := styles.body
		if idx%2 == 1 {
			style = styles.altBody
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow), style); err != nil {
			return err
		}
	}

	if err := f.SetCellValue(sheet, "A13", "Columnas de la hoja Partidos"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A13", "A13", styles.header); err != nil {
		return err
	}
	headers := []string{"partido", "grupo", "equipo_a", "equipo_b", "tiros_a", "tiros_b", "motivacion_a", "motivacion_b", "simulaciones", "seed", "tiebreaker"}
	if err := writeHeaderRow(f, sheet, 14, headers, styles.header); err != nil {
		return err
	}

	if err := f.SetColWidth(sheet, "A", "A", 30); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 30); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "A", "B", 30); err != nil {
		return err
	}
	return nil
}

func buildBatchInputSheet(f *excelize.File, groupName string, rows []BatchMatchInput, styles batchStyles, template bool) error {
	sheet := "Partidos"
	title := fmt.Sprintf("Partidos de entrada del grupo %s", groupName)
	if err := f.SetCellValue(sheet, "A1", title); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if template {
		if err := f.SetCellValue(sheet, "A2", "Rellena las celdas amarillas. simulaciones admite un valor entre 1 y 10000 por fila."); err != nil {
			return err
		}
	} else {
		if err := f.SetCellValue(sheet, "A2", "Libro procesado por Go. Las filas de entrada se conservan y los resultados se escriben en otras hojas."); err != nil {
			return err
		}
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", styles.note); err != nil {
		return err
	}

	headers := []string{"partido", "grupo", "equipo_a", "equipo_b", "tiros_a", "tiros_b", "motivacion_a", "motivacion_b", "simulaciones", "seed", "tiebreaker"}
	if err := writeHeaderRow(f, sheet, 4, headers, styles.header); err != nil {
		return err
	}

	for idx, row := range rows {
		rowNum := 5 + idx
		values := []interface{}{
			chooseInt(row.MatchID, idx+1),
			row.Group,
			row.TeamA,
			row.TeamB,
			chooseInt(row.ShotsA, 5),
			chooseInt(row.ShotsB, 5),
			int(row.MotivationA),
			int(row.MotivationB),
			chooseInt(row.Simulations, 1),
			row.Seed,
			chooseString(row.Tiebreaker, "penalties"),
		}
		if err := writeRow(f, sheet, rowNum, values); err != nil {
			return err
		}
		style := styles.body
		if idx%2 == 1 {
			style = styles.altBody
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("K%d", rowNum), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("K%d", rowNum), styles.input); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), style); err != nil {
			return err
		}
	}

	widths := map[string]float64{
		"A": 12,
		"B": 24,
		"C": 22,
		"D": 22,
		"E": 10,
		"F": 10,
		"G": 14,
		"H": 14,
		"I": 12,
		"J": 12,
		"K": 14,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if len(rows) > 0 {
		if err := f.AutoFilter(sheet, fmt.Sprintf("A4:K%d", 4+len(rows)), nil); err != nil {
			return err
		}
	}
	return nil
}

func buildBatchResultsSheet(f *excelize.File, groupName string, reports []BatchMatchReport, styles batchStyles) error {
	sheet := "Resultados"
	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Resultados simulados del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Cada fila resume las simulaciones de un partido y deja el marcador mas probable para la tabla de posiciones."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", styles.note); err != nil {
		return err
	}

	headers := []string{
		"partido", "grupo", "equipo_a", "equipo_b", "tiros_a", "tiros_b", "motivacion_a", "motivacion_b",
		"simulaciones", "seed", "tiebreaker", "goles_rep_a", "goles_rep_b", "ganador_rep", "decidido_por_rep",
		"victorias_a", "victorias_b", "empates_reg", "prom_goles_a", "prom_goles_b", "marcador_mas_probable", "veces", "porcentaje",
	}
	if err := writeHeaderRow(f, sheet, 4, headers, styles.header); err != nil {
		return err
	}

	for idx, report := range reports {
		row := 5 + idx
		summary := report.Series
		values := []interface{}{
			report.Input.MatchID,
			report.Input.Group,
			report.Input.TeamA,
			report.Input.TeamB,
			report.Input.ShotsA,
			report.Input.ShotsB,
			int(report.Input.MotivationA),
			int(report.Input.MotivationB),
			report.Input.Simulations,
			report.SeedUsed,
			report.Input.Tiebreaker,
			report.RepresentativeGoalsA,
			report.RepresentativeGoalsB,
			report.RepresentativeWinner,
			report.RepresentativeDecidedBy,
			summary.WinsA,
			summary.WinsB,
			summary.RegulationDraws,
			formatBatchAverage(summary.GoalsA, summary.Simulations),
			formatBatchAverage(summary.GoalsB, summary.Simulations),
			fmt.Sprintf("%s %d - %d %s", summary.TeamA, summary.MostRepeatedGoalsA, summary.MostRepeatedGoalsB, summary.TeamB),
			summary.MostRepeatedCount,
			summary.MostRepeatedPercent,
		}
		if err := writeRow(f, sheet, row, values); err != nil {
			return err
		}
		style := styles.body
		if idx%2 == 1 {
			style = styles.altBody
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("W%d", row), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("L%d", row), fmt.Sprintf("W%d", row), styles.result); err != nil {
			return err
		}
	}

	widths := map[string]float64{
		"A": 10, "B": 20, "C": 18, "D": 18, "E": 8, "F": 8, "G": 12, "H": 12, "I": 12, "J": 12, "K": 12,
		"L": 10, "M": 10, "N": 12, "O": 12, "P": 12, "Q": 12, "R": 10, "S": 18, "T": 14, "U": 18, "V": 10, "W": 12,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if len(reports) > 0 {
		if err := f.AutoFilter(sheet, fmt.Sprintf("A4:W%d", 4+len(reports)), nil); err != nil {
			return err
		}
	}
	return nil
}

func buildBatchTopScoresSheet(f *excelize.File, reports []BatchMatchReport, styles batchStyles) error {
	sheet := "Top Marcadores"
	if err := f.SetCellValue(sheet, "A1", "Top marcadores exactos por partido"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Cada fila representa un marcador exacto dentro del top 10 de su partido."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", styles.note); err != nil {
		return err
	}

	headers := []string{"partido", "grupo", "rank", "equipo_a", "goles_a", "goles_b", "equipo_b", "veces", "porcentaje"}
	if err := writeHeaderRow(f, sheet, 4, headers, styles.header); err != nil {
		return err
	}

	row := 5
	for _, report := range reports {
		for idx, score := range report.Series.TopScores {
			values := []interface{}{
				report.Input.MatchID,
				report.Input.Group,
				idx + 1,
				report.Series.TeamA,
				score.GoalsA,
				score.GoalsB,
				report.Series.TeamB,
				score.Count,
				score.Percent,
			}
			if err := writeRow(f, sheet, row, values); err != nil {
				return err
			}
			style := styles.body
			if row%2 == 0 {
				style = styles.altBody
			}
			if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row), style); err != nil {
				return err
			}
			row++
		}
	}

	widths := map[string]float64{
		"A": 10, "B": 20, "C": 8, "D": 18, "E": 10, "F": 10, "G": 18, "H": 10, "I": 12,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if row > 5 {
		if err := f.AutoFilter(sheet, fmt.Sprintf("A4:I%d", row-1), nil); err != nil {
			return err
		}
	}
	return nil
}

func buildBatchPositionsSheet(f *excelize.File, groupName string, standings []model.GroupStanding, records map[string]standingRecord, styles batchStyles) error {
	sheet := "Posiciones"
	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Tabla de posiciones del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Se ordena por puntos, diferencia de gol y goles a favor usando el marcador mas probable de cada partido."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", styles.note); err != nil {
		return err
	}

	headers := []string{"pos", "equipo", "pj", "pg", "pe", "pp", "gf", "gc", "dg", "pts"}
	if err := writeHeaderRow(f, sheet, 4, headers, styles.header); err != nil {
		return err
	}

	for idx, standing := range standings {
		row := 5 + idx
		rec := records[standing.Team]
		values := []interface{}{
			idx + 1,
			standing.Team,
			standing.Played,
			rec.Wins,
			rec.Draws,
			rec.Losses,
			standing.GoalsFor,
			standing.GoalsAgainst,
			standing.GoalDifference,
			standing.Points,
		}
		if err := writeRow(f, sheet, row, values); err != nil {
			return err
		}
		style := styles.body
		if idx%2 == 1 {
			style = styles.altBody
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("J%d", row), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("J%d", row), styles.result); err != nil {
			return err
		}
	}

	widths := map[string]float64{
		"A": 8, "B": 24, "C": 8, "D": 8, "E": 8, "F": 8, "G": 8, "H": 8, "I": 8, "J": 8,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if len(standings) > 0 {
		if err := f.AutoFilter(sheet, fmt.Sprintf("A4:J%d", 4+len(standings)), nil); err != nil {
			return err
		}
	}
	return nil
}

func buildBatchSummarySheet(f *excelize.File, groupName string, reports []BatchMatchReport, standings []model.GroupStanding, styles batchStyles) error {
	sheet := "Resumen"
	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Resumen batch del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", styles.title); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Consolida el volumen de simulaciones y la fotografia general del grupo."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", styles.note); err != nil {
		return err
	}

	totalMatches := len(reports)
	totalSimulations := 0
	totalGoals := 0
	totalDraws := 0
	totalPenalties := 0
	totalRandom := 0
	minSimulations := 0
	maxSimulations := 0
	for idx, report := range reports {
		totalSimulations += report.Series.Simulations
		totalGoals += report.Series.GoalsA + report.Series.GoalsB
		totalDraws += report.Series.RegulationDraws
		totalPenalties += report.Series.Penalties
		totalRandom += report.Series.RandomTie
		if idx == 0 || report.Series.Simulations < minSimulations {
			minSimulations = report.Series.Simulations
		}
		if report.Series.Simulations > maxSimulations {
			maxSimulations = report.Series.Simulations
		}
	}

	leader := ""
	leaderPoints := 0
	leaderGoals := 0
	if len(standings) > 0 {
		leader = standings[0].Team
		leaderPoints = standings[0].Points
		leaderGoals = standings[0].GoalsFor
	}

	rows := []struct {
		label string
		value interface{}
	}{
		{"grupo", groupName},
		{"partidos_procesados", totalMatches},
		{"equipos_unicos", len(standings)},
		{"simulaciones_totales", totalSimulations},
		{"promedio_simulaciones_por_partido", safeDivideFloat(totalSimulations, totalMatches)},
		{"simulaciones_minimas_por_partido", minSimulations},
		{"simulaciones_maximas_por_partido", maxSimulations},
		{"goles_totales_representativos", totalGoals},
		{"promedio_goles_representativos_por_partido", safeDivideFloat(totalGoals, totalMatches)},
		{"empates_en_tiempo_regular", totalDraws},
		{"definidos_por_penales", totalPenalties},
		{"definidos_por_sorteo", totalRandom},
		{"lider", leader},
		{"puntos_lider", leaderPoints},
		{"goles_a_favor_lider", leaderGoals},
	}
	if err := writeHeaderRow(f, sheet, 4, []string{"estadistica", "valor"}, styles.header); err != nil {
		return err
	}
	for idx, row := range rows {
		excelRow := 5 + idx
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.label); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.value); err != nil {
			return err
		}
		style := styles.body
		if idx%2 == 1 {
			style = styles.altBody
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow), style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "A", 38); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 18); err != nil {
		return err
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, fmt.Sprintf("A4:B%d", 4+len(rows)), nil); err != nil {
		return err
	}
	return nil
}

func readBatchInputs(path string) ([]BatchMatchInput, string, error) {
	wb, err := excelize.OpenFile(path)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = wb.Close()
	}()

	sheet := "Partidos"
	if idx, err := wb.GetSheetIndex(sheet); err != nil || idx < 0 {
		if idx, err := wb.GetSheetIndex("Entrada"); err != nil || idx < 0 {
			return nil, "", fmt.Errorf("no se encontro la hoja Partidos ni Entrada en %s", path)
		}
		sheet = "Entrada"
	}

	var (
		rows         []BatchMatchInput
		groupName    string
		seenGroupSet = map[string]struct{}{}
		baseTeams    = map[string]struct{}{}
	)

	for rowNum := 5; ; rowNum++ {
		teamA := strings.TrimSpace(cellValue(wb, sheet, fmt.Sprintf("C%d", rowNum)))
		teamB := strings.TrimSpace(cellValue(wb, sheet, fmt.Sprintf("D%d", rowNum)))
		if teamA == "" && teamB == "" {
			break
		}
		if teamA == "" || teamB == "" {
			return nil, "", fmt.Errorf("fila %d: equipo_a y equipo_b son obligatorios", rowNum)
		}

		rowGroup := strings.TrimSpace(cellValue(wb, sheet, fmt.Sprintf("B%d", rowNum)))
		if rowGroup != "" {
			seenGroupSet[strings.ToLower(rowGroup)] = struct{}{}
			if groupName == "" {
				groupName = rowGroup
			}
		}

		matchID, err := parseIntCell(cellValue(wb, sheet, fmt.Sprintf("A%d", rowNum)), rowNum-4)
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: %w", rowNum, err)
		}
		shotsA, err := parseIntCell(cellValue(wb, sheet, fmt.Sprintf("E%d", rowNum)), 5)
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: tiros_a invalidos: %w", rowNum, err)
		}
		shotsB, err := parseIntCell(cellValue(wb, sheet, fmt.Sprintf("F%d", rowNum)), 5)
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: tiros_b invalidos: %w", rowNum, err)
		}
		motA, err := parseMotivationCell(cellValue(wb, sheet, fmt.Sprintf("G%d", rowNum)))
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: motivacion_a invalida: %w", rowNum, err)
		}
		motB, err := parseMotivationCell(cellValue(wb, sheet, fmt.Sprintf("H%d", rowNum)))
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: motivacion_b invalida: %w", rowNum, err)
		}
		simulations, err := parseIntCell(cellValue(wb, sheet, fmt.Sprintf("I%d", rowNum)), 1)
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: simulaciones invalidas: %w", rowNum, err)
		}
		if simulations <= 0 {
			simulations = 1
		}
		if simulations > 10000 {
			return nil, "", fmt.Errorf("fila %d: simulaciones excede el maximo permitido de 10000", rowNum)
		}
		seed, err := parseInt64Cell(cellValue(wb, sheet, fmt.Sprintf("J%d", rowNum)), 0)
		if err != nil {
			return nil, "", fmt.Errorf("fila %d: seed invalida: %w", rowNum, err)
		}
		tiebreaker := strings.TrimSpace(cellValue(wb, sheet, fmt.Sprintf("K%d", rowNum)))
		if tiebreaker == "" {
			tiebreaker = "penalties"
		}

		baseTeams[strings.ToLower(teamA)] = struct{}{}
		baseTeams[strings.ToLower(teamB)] = struct{}{}

		rows = append(rows, BatchMatchInput{
			RowNumber:   rowNum,
			MatchID:     matchID,
			Group:       rowGroup,
			TeamA:       teamA,
			TeamB:       teamB,
			ShotsA:      shotsA,
			ShotsB:      shotsB,
			MotivationA: motA,
			MotivationB: motB,
			Simulations: simulations,
			Seed:        seed,
			Tiebreaker:  tiebreaker,
		})
	}

	if len(rows) == 0 {
		return nil, "", fmt.Errorf("no se encontraron filas en la hoja %s", sheet)
	}
	if len(seenGroupSet) > 1 {
		return nil, "", fmt.Errorf("este proceso batch soporta un solo grupo por archivo de entrada")
	}
	if groupName == "" {
		groupName = "Grupo"
	}
	_ = baseTeams
	return rows, groupName, nil
}

type standingRecord struct {
	Wins   int
	Draws  int
	Losses int
}

func buildStandingRecords(results []GroupMatchResult) map[string]standingRecord {
	records := map[string]standingRecord{}
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
	return records
}

func batchRepresentativeOutcome(teamA, teamB string, goalsA, goalsB int) (string, string) {
	switch {
	case goalsA > goalsB:
		return teamA, "tiempo regular"
	case goalsB > goalsA:
		return teamB, "tiempo regular"
	default:
		return "Empate", "tiempo regular"
	}
}

func extractTeamsFromBatchInputs(rows []BatchMatchInput) []string {
	seen := map[string]struct{}{}
	teams := make([]string, 0, len(rows)*2)
	for _, row := range rows {
		for _, team := range []string{row.TeamA, row.TeamB} {
			key := strings.ToLower(strings.TrimSpace(team))
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			teams = append(teams, team)
		}
	}
	return teams
}

func chooseInt(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func chooseString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func parseIntCell(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("valor entero invalido: %q", value)
	}
	return n, nil
}

func parseInt64Cell(value string, fallback int64) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("valor entero invalido: %q", value)
	}
	return n, nil
}

func parseMotivationCell(value string) (model.Motivation, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return model.MotivationMedium, nil
	}
	return model.ParseMotivation(value)
}

func cellValue(wb *excelize.File, sheet, cell string) string {
	value, err := wb.GetCellValue(sheet, cell)
	if err != nil {
		return ""
	}
	return value
}

func safeDivideFloat(value, divisor int) float64 {
	if divisor <= 0 {
		return 0
	}
	return float64(value) / float64(divisor)
}

func formatBatchAverage(sum, simulations int) float64 {
	if simulations <= 0 {
		return 0
	}
	return float64(sum) / float64(simulations)
}

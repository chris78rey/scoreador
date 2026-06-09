package report_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"scoreador/internal/loader"
	"scoreador/internal/model"
	"scoreador/internal/report"
	"scoreador/internal/sim"
)

const batchTestLambda = `tiros_min,tiros_max,motivacion,lambda
0,1,baja,0.2
0,1,media,0.4
0,1,alta,0.7
2,3,baja,0.5
2,3,media,0.9
2,3,alta,1.3
4,5,baja,0.8
4,5,media,1.4
4,5,alta,1.9
6,7,baja,1.1
6,7,media,1.8
6,7,alta,2.4
8,9,baja,1.5
8,9,media,2.3
8,9,alta,3.0
10,99,baja,2.0
10,99,media,2.8
10,99,alta,3.6
`

type batchInputRow struct {
	MatchID     int
	Group       string
	TeamA       string
	TeamB       string
	ShotsA      int
	ShotsB      int
	MotivationA int
	MotivationB int
	Simulations int
	Seed        int64
	Tiebreaker  string
}

func TestProcessBatchWorkbookFromExcelInput(t *testing.T) {
	dir := t.TempDir()

	teams := []string{"Ecuador", "Alemania", "Polonia", "Corea"}
	groupName := "Grupo A"
	templatePath := filepath.Join(dir, "batch_input.xlsx")
	if err := report.WriteBatchTemplateXLSX(templatePath, groupName, teams); err != nil {
		t.Fatalf("write template: %v", err)
	}

	rows := []batchInputRow{
		{MatchID: 1, Group: groupName, TeamA: "Ecuador", TeamB: "Alemania", ShotsA: 4, ShotsB: 5, MotivationA: 5, MotivationB: 8, Simulations: 120, Seed: 101, Tiebreaker: "penalties"},
		{MatchID: 2, Group: groupName, TeamA: "Ecuador", TeamB: "Polonia", ShotsA: 6, ShotsB: 4, MotivationA: 7, MotivationB: 3, Simulations: 120, Seed: 102, Tiebreaker: "penalties"},
		{MatchID: 3, Group: groupName, TeamA: "Ecuador", TeamB: "Corea", ShotsA: 5, ShotsB: 5, MotivationA: 4, MotivationB: 6, Simulations: 120, Seed: 103, Tiebreaker: "penalties"},
		{MatchID: 4, Group: groupName, TeamA: "Alemania", TeamB: "Polonia", ShotsA: 7, ShotsB: 6, MotivationA: 8, MotivationB: 2, Simulations: 120, Seed: 104, Tiebreaker: "penalties"},
		{MatchID: 5, Group: groupName, TeamA: "Alemania", TeamB: "Corea", ShotsA: 6, ShotsB: 4, MotivationA: 9, MotivationB: 1, Simulations: 120, Seed: 105, Tiebreaker: "penalties"},
		{MatchID: 6, Group: groupName, TeamA: "Polonia", TeamB: "Corea", ShotsA: 4, ShotsB: 3, MotivationA: 5, MotivationB: 7, Simulations: 120, Seed: 106, Tiebreaker: "penalties"},
	}
	fillBatchInputWorkbook(t, templatePath, rows)

	lambdaPath := filepath.Join(dir, "lambda.csv")
	if err := os.WriteFile(lambdaPath, []byte(batchTestLambda), 0o644); err != nil {
		t.Fatalf("write lambda: %v", err)
	}
	rules, err := loader.LoadLambdaRules(lambdaPath)
	if err != nil {
		t.Fatalf("load lambda rules: %v", err)
	}

	outputPath := filepath.Join(dir, "batch_output.xlsx")
	if err := report.ProcessBatchWorkbook(templatePath, outputPath, rules, 2026); err != nil {
		t.Fatalf("process batch workbook: %v", err)
	}

	wb, err := excelize.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("open output workbook: %v", err)
	}
	defer func() { _ = wb.Close() }()

	expectedSheets := []string{"Instrucciones", "Partidos", "Resultados", "Top Marcadores", "Posiciones", "Resumen"}
	if got := wb.GetSheetList(); !sameStringSlices(got, expectedSheets) {
		t.Fatalf("unexpected sheets: got %v want %v", got, expectedSheets)
	}

	expectedReports := make([]sim.SingleMatchSeries, 0, len(rows))
	expectedStandingsInput := make([]report.GroupMatchResult, 0, len(rows))
	for _, row := range rows {
		series, err := sim.RunSingleMatchSeries(row.Seed, row.Simulations, rules, sim.SingleMatchInput{
			TeamA:       row.TeamA,
			TeamB:       row.TeamB,
			ShotsA:      row.ShotsA,
			ShotsB:      row.ShotsB,
			MotivationA: model.Motivation(row.MotivationA),
			MotivationB: model.Motivation(row.MotivationB),
			Tiebreaker:  row.Tiebreaker,
		})
		if err != nil {
			t.Fatalf("expected series for match %d: %v", row.MatchID, err)
		}
		expectedReports = append(expectedReports, series)
		expectedStandingsInput = append(expectedStandingsInput, report.GroupMatchResult{
			TeamA:  row.TeamA,
			TeamB:  row.TeamB,
			GoalsA: series.MostRepeatedGoalsA,
			GoalsB: series.MostRepeatedGoalsB,
		})
	}

	// Validate representative results sheet.
	for idx, row := range rows {
		excelRow := 5 + idx
		series := expectedReports[idx]
		mustEqualCell(t, wb, "Resultados", excelRow, "A", row.MatchID)
		mustEqualCell(t, wb, "Resultados", excelRow, "B", row.Group)
		mustEqualCell(t, wb, "Resultados", excelRow, "C", row.TeamA)
		mustEqualCell(t, wb, "Resultados", excelRow, "D", row.TeamB)
		mustEqualCell(t, wb, "Resultados", excelRow, "E", row.ShotsA)
		mustEqualCell(t, wb, "Resultados", excelRow, "F", row.ShotsB)
		mustEqualCell(t, wb, "Resultados", excelRow, "G", row.MotivationA)
		mustEqualCell(t, wb, "Resultados", excelRow, "H", row.MotivationB)
		mustEqualCell(t, wb, "Resultados", excelRow, "I", row.Simulations)
		mustEqualCell(t, wb, "Resultados", excelRow, "J", row.Seed)
		mustEqualCell(t, wb, "Resultados", excelRow, "K", row.Tiebreaker)
		mustEqualCell(t, wb, "Resultados", excelRow, "L", series.MostRepeatedGoalsA)
		mustEqualCell(t, wb, "Resultados", excelRow, "M", series.MostRepeatedGoalsB)
		mustEqualCell(t, wb, "Resultados", excelRow, "N", representativeWinner(row.TeamA, row.TeamB, series.MostRepeatedGoalsA, series.MostRepeatedGoalsB))
		mustEqualCell(t, wb, "Resultados", excelRow, "O", "tiempo regular")
		mustEqualCell(t, wb, "Resultados", excelRow, "P", series.WinsA)
		mustEqualCell(t, wb, "Resultados", excelRow, "Q", series.WinsB)
		mustEqualCell(t, wb, "Resultados", excelRow, "R", series.RegulationDraws)
		mustFloatCell(t, wb, "Resultados", excelRow, "S", float64(series.GoalsA)/float64(series.Simulations))
		mustFloatCell(t, wb, "Resultados", excelRow, "T", float64(series.GoalsB)/float64(series.Simulations))
		mustEqualCell(t, wb, "Resultados", excelRow, "U", fmt.Sprintf("%s %d - %d %s", series.TeamA, series.MostRepeatedGoalsA, series.MostRepeatedGoalsB, series.TeamB))
		mustEqualCell(t, wb, "Resultados", excelRow, "V", series.MostRepeatedCount)
		mustFloatCell(t, wb, "Resultados", excelRow, "W", series.MostRepeatedPercent)
	}

	// Validate top score rows for the first match.
	first := expectedReports[0]
	firstTop := first.TopScores[0]
	mustEqualCell(t, wb, "Top Marcadores", 5, "A", rows[0].MatchID)
	mustEqualCell(t, wb, "Top Marcadores", 5, "B", rows[0].Group)
	mustEqualCell(t, wb, "Top Marcadores", 5, "C", 1)
	mustEqualCell(t, wb, "Top Marcadores", 5, "D", first.TeamA)
	mustEqualCell(t, wb, "Top Marcadores", 5, "E", firstTop.GoalsA)
	mustEqualCell(t, wb, "Top Marcadores", 5, "F", firstTop.GoalsB)
	mustEqualCell(t, wb, "Top Marcadores", 5, "G", first.TeamB)
	mustEqualCell(t, wb, "Top Marcadores", 5, "H", firstTop.Count)
	mustFloatCell(t, wb, "Top Marcadores", 5, "I", firstTop.Percent)

	// Validate standings.
	expectedStandings := report.GroupDemoStandings(groupName, expectedStandingsInput)
	for idx, standing := range expectedStandings {
		row := 5 + idx
		mustEqualCell(t, wb, "Posiciones", row, "A", idx+1)
		mustEqualCell(t, wb, "Posiciones", row, "B", standing.Team)
		mustEqualCell(t, wb, "Posiciones", row, "C", standing.Played)
		mustEqualCell(t, wb, "Posiciones", row, "D", recordWins(expectedStandingsInput, standing.Team))
		mustEqualCell(t, wb, "Posiciones", row, "E", recordDraws(expectedStandingsInput, standing.Team))
		mustEqualCell(t, wb, "Posiciones", row, "F", recordLosses(expectedStandingsInput, standing.Team))
		mustEqualCell(t, wb, "Posiciones", row, "G", standing.GoalsFor)
		mustEqualCell(t, wb, "Posiciones", row, "H", standing.GoalsAgainst)
		mustEqualCell(t, wb, "Posiciones", row, "I", standing.GoalDifference)
		mustEqualCell(t, wb, "Posiciones", row, "J", standing.Points)
	}

	summary := readKeyValueSheet(t, wb, "Resumen")
	mustSummaryValue(t, summary, "grupo", groupName)
	mustSummaryValue(t, summary, "partidos_procesados", len(rows))
	mustSummaryValue(t, summary, "simulaciones_totales", totalSimulations(rows))
	mustSummaryValue(t, summary, "lider", expectedStandings[0].Team)
}

func TestProcessBatchWorkbookRejectsTooManySimulations(t *testing.T) {
	dir := t.TempDir()

	templatePath := filepath.Join(dir, "batch_input.xlsx")
	if err := report.WriteBatchTemplateXLSX(templatePath, "Grupo A", []string{"Ecuador", "Alemania", "Polonia", "Corea"}); err != nil {
		t.Fatalf("write template: %v", err)
	}
	fillBatchInputWorkbook(t, templatePath, []batchInputRow{
		{MatchID: 1, Group: "Grupo A", TeamA: "Ecuador", TeamB: "Alemania", ShotsA: 4, ShotsB: 5, MotivationA: 5, MotivationB: 8, Simulations: 10001, Seed: 101, Tiebreaker: "penalties"},
	})

	lambdaPath := filepath.Join(dir, "lambda.csv")
	if err := os.WriteFile(lambdaPath, []byte(batchTestLambda), 0o644); err != nil {
		t.Fatalf("write lambda: %v", err)
	}
	rules, err := loader.LoadLambdaRules(lambdaPath)
	if err != nil {
		t.Fatalf("load lambda rules: %v", err)
	}

	err = report.ProcessBatchWorkbook(templatePath, filepath.Join(dir, "output.xlsx"), rules, 2026)
	if err == nil {
		t.Fatal("expected error for simulations > 10000")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "10000") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func fillBatchInputWorkbook(t *testing.T, path string, rows []batchInputRow) {
	t.Helper()

	wb, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer func() { _ = wb.Close() }()

	for idx, row := range rows {
		excelRow := 5 + idx
		values := map[string]interface{}{
			"A": row.MatchID,
			"B": row.Group,
			"C": row.TeamA,
			"D": row.TeamB,
			"E": row.ShotsA,
			"F": row.ShotsB,
			"G": row.MotivationA,
			"H": row.MotivationB,
			"I": row.Simulations,
			"J": row.Seed,
			"K": row.Tiebreaker,
		}
		for col, value := range values {
			if err := wb.SetCellValue("Partidos", fmt.Sprintf("%s%d", col, excelRow), value); err != nil {
				t.Fatalf("set cell %s%d: %v", col, excelRow, err)
			}
		}
	}

	if err := wb.SaveAs(path); err != nil {
		t.Fatalf("save filled template: %v", err)
	}
}

func mustEqualCell(t *testing.T, wb *excelize.File, sheet string, row int, col string, want interface{}) {
	t.Helper()

	got, err := wb.GetCellValue(sheet, fmt.Sprintf("%s%d", col, row))
	if err != nil {
		t.Fatalf("get cell %s%d on %s: %v", col, row, sheet, err)
	}
	if fmt.Sprint(want) != got {
		t.Fatalf("unexpected value at %s!%s%d: got %q want %q", sheet, col, row, got, fmt.Sprint(want))
	}
}

func mustFloatCell(t *testing.T, wb *excelize.File, sheet string, row int, col string, want float64) {
	t.Helper()

	got, err := wb.GetCellValue(sheet, fmt.Sprintf("%s%d", col, row))
	if err != nil {
		t.Fatalf("get cell %s%d on %s: %v", col, row, sheet, err)
	}
	gotFloat, err := strconv.ParseFloat(got, 64)
	if err != nil {
		t.Fatalf("parse float at %s!%s%d: %v (value=%q)", sheet, col, row, err, got)
	}
	if diff := gotFloat - want; diff < -0.0001 || diff > 0.0001 {
		t.Fatalf("unexpected float at %s!%s%d: got %f want %f", sheet, col, row, gotFloat, want)
	}
}

func readKeyValueSheet(t *testing.T, wb *excelize.File, sheet string) map[string]string {
	t.Helper()

	rows, err := wb.GetRows(sheet)
	if err != nil {
		t.Fatalf("get rows from %s: %v", sheet, err)
	}
	values := map[string]string{}
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		key := strings.TrimSpace(row[0])
		if key == "" || key == "estadistica" {
			continue
		}
		values[key] = strings.TrimSpace(row[1])
	}
	return values
}

func mustSummaryValue(t *testing.T, got map[string]string, key string, want interface{}) {
	t.Helper()

	value, ok := got[key]
	if !ok {
		t.Fatalf("missing summary key %q", key)
	}
	if value != fmt.Sprint(want) {
		t.Fatalf("unexpected summary value for %s: got %q want %q", key, value, fmt.Sprint(want))
	}
}

func totalSimulations(rows []batchInputRow) int {
	total := 0
	for _, row := range rows {
		total += row.Simulations
	}
	return total
}

func recordWins(results []report.GroupMatchResult, team string) int {
	wins := 0
	for _, result := range results {
		if result.TeamA == team && result.GoalsA > result.GoalsB {
			wins++
		}
		if result.TeamB == team && result.GoalsB > result.GoalsA {
			wins++
		}
	}
	return wins
}

func recordDraws(results []report.GroupMatchResult, team string) int {
	draws := 0
	for _, result := range results {
		if result.TeamA == team && result.GoalsA == result.GoalsB {
			draws++
		}
		if result.TeamB == team && result.GoalsA == result.GoalsB {
			draws++
		}
	}
	return draws
}

func recordLosses(results []report.GroupMatchResult, team string) int {
	losses := 0
	for _, result := range results {
		if result.TeamA == team && result.GoalsA < result.GoalsB {
			losses++
		}
		if result.TeamB == team && result.GoalsB < result.GoalsA {
			losses++
		}
	}
	return losses
}

func representativeWinner(teamA, teamB string, goalsA, goalsB int) string {
	switch {
	case goalsA > goalsB:
		return teamA
	case goalsB > goalsA:
		return teamB
	default:
		return "Empate"
	}
}

func sameStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

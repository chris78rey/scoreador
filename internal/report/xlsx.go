package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/xuri/excelize/v2"

	"scoreador/internal/model"
)

func WriteXLSX(path string, cfg model.Config, matches []model.MatchInput, rules []model.LambdaRule, summary model.TournamentSummary) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	if err := f.SetSheetName("Sheet1", "Resumen"); err != nil {
		return err
	}
	if isWorldCup2026(cfg) {
		if err := createWorldCup2026Sheets(f); err != nil {
			return err
		}
	}
	if _, err := f.NewSheet("Configuracion"); err != nil {
		return err
	}
	if _, err := f.NewSheet("Partidos"); err != nil {
		return err
	}
	if _, err := f.NewSheet("Lambda"); err != nil {
		return err
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#1F4E78"}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return err
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
		return err
	}
	subHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#1F1F1F"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E1F2"}},
	})
	if err != nil {
		return err
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
		return err
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
		return err
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
		return err
	}
	formulaStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#1F1F1F"},
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
		return err
	}
	resultStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#1F1F1F"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9EAD3"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#A9D18E", Style: 1},
			{Type: "right", Color: "#A9D18E", Style: 1},
			{Type: "top", Color: "#A9D18E", Style: 1},
			{Type: "bottom", Color: "#A9D18E", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return err
	}
	noteStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Italic: true, Color: "#666666"},
	})
	if err != nil {
		return err
	}

	if err := buildSummarySheet(f, summary, titleStyle, headerStyle, bodyStyle, altBodyStyle); err != nil {
		return err
	}
	if err := buildConfigSheet(f, cfg, subHeaderStyle, bodyStyle, altBodyStyle); err != nil {
		return err
	}
	if err := buildMatchesSheet(f, matches, headerStyle, bodyStyle, altBodyStyle); err != nil {
		return err
	}
	if err := buildLambdaSheet(f, rules, headerStyle, bodyStyle, altBodyStyle); err != nil {
		return err
	}
	if isWorldCup2026(cfg) {
		if err := buildWorldCup2026Sheets(f, cfg, matches, summary, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
			return err
		}
	}

	if idx, err := f.GetSheetIndex("Resumen"); err == nil {
		f.SetActiveSheet(idx)
	}

	if err := ensureDir(path); err != nil {
		return err
	}
	return f.SaveAs(path)
}

func buildSummarySheet(f *excelize.File, summary model.TournamentSummary, titleStyle, headerStyle, bodyStyle, altBodyStyle int) error {
	sheet := "Resumen"

	if err := f.SetCellValue(sheet, "A1", summary.Name); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Simulaciones"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "B2", summary.Simulations); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "B2", headerStyle); err != nil {
		return err
	}

	headers := []string{
		"equipo",
		"simulaciones",
		"veces_clasifico",
		"porcentaje_clasifico",
		"veces_llego_dieciseisavos",
		"porcentaje_dieciseisavos",
		"veces_llego_octavos",
		"porcentaje_octavos",
		"veces_llego_cuartos",
		"porcentaje_cuartos",
		"veces_llego_semifinal",
		"porcentaje_semifinal",
		"veces_llego_final",
		"porcentaje_final",
		"veces_campeon",
		"porcentaje_campeon",
	}
	if err := writeHeaderRow(f, sheet, 4, headers, headerStyle); err != nil {
		return err
	}

	for rowIndex, team := range summary.Teams {
		row := []interface{}{
			team.Team,
			team.Simulations,
			team.Qualified,
			team.QualifiedPct,
			team.Dieciseisavos,
			team.DieciseisavosPct,
			team.Octavos,
			team.OctavosPct,
			team.Cuartos,
			team.CuartosPct,
			team.Semifinal,
			team.SemifinalPct,
			team.Final,
			team.FinalPct,
			team.Campeon,
			team.CampeonPct,
		}
		if err := writeRow(f, sheet, 5+rowIndex, row); err != nil {
			return err
		}
		style := bodyStyle
		if rowIndex%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", 5+rowIndex), fmt.Sprintf("P%d", 5+rowIndex), style); err != nil {
			return err
		}
	}

	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, "A4:P"+strconv.Itoa(4+len(summary.Teams)), nil); err != nil {
		return err
	}
	setSummaryWidths(f, sheet)
	return nil
}

func buildConfigSheet(f *excelize.File, cfg model.Config, headerStyle, bodyStyle, altBodyStyle int) error {
	sheet := "Configuracion"
	if err := f.SetCellValue(sheet, "A1", "campo"); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "B1", "valor"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "B1", headerStyle); err != nil {
		return err
	}

	rows := [][]interface{}{
		{"name", cfg.Name},
		{"simulations", cfg.Simulations},
		{"groups", cfg.Groups},
		{"teams_per_group", cfg.TeamsPerGroup},
		{"qualified_per_group", cfg.QualifiedPerGroup},
		{"best_thirds", cfg.BestThirds},
		{"knockout", cfg.Knockout},
		{"knockout_tiebreaker", cfg.KnockoutTiebreaker},
		{"seed", cfg.Seed},
	}
	for idx, row := range rows {
		if err := writeRow(f, sheet, idx+2, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", idx+2), fmt.Sprintf("B%d", idx+2), style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "A", 24); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 28); err != nil {
		return err
	}
	return nil
}

func buildMatchesSheet(f *excelize.File, matches []model.MatchInput, headerStyle, bodyStyle, altBodyStyle int) error {
	sheet := "Partidos"
	headers := []string{
		"match_id",
		"stage",
		"grupo",
		"equipo_a",
		"equipo_b",
		"tiros_a",
		"tiros_b",
		"motivacion_a",
		"motivacion_b",
	}
	if err := writeHeaderRow(f, sheet, 1, headers, headerStyle); err != nil {
		return err
	}
	for idx, match := range matches {
		row := []interface{}{
			match.MatchID,
			match.Stage,
			match.Group,
			match.TeamA,
			match.TeamB,
			match.ShotsA,
			match.ShotsB,
			match.MotivationA.String(),
			match.MotivationB.String(),
		}
		if err := writeRow(f, sheet, idx+2, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", idx+2), fmt.Sprintf("I%d", idx+2), style); err != nil {
			return err
		}
	}

	widths := map[string]float64{
		"A": 12,
		"B": 14,
		"C": 10,
		"D": 22,
		"E": 22,
		"F": 10,
		"G": 10,
		"H": 14,
		"I": 14,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 1, "A2"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, "A1:I"+strconv.Itoa(len(matches)+1), nil); err != nil {
		return err
	}
	return nil
}

func buildLambdaSheet(f *excelize.File, rules []model.LambdaRule, headerStyle, bodyStyle, altBodyStyle int) error {
	sheet := "Lambda"
	headers := []string{"tiros_min", "tiros_max", "motivacion", "lambda"}
	if err := writeHeaderRow(f, sheet, 1, headers, headerStyle); err != nil {
		return err
	}

	orderedRules := append([]model.LambdaRule(nil), rules...)
	sort.SliceStable(orderedRules, func(i, j int) bool {
		if orderedRules[i].ShotsMin != orderedRules[j].ShotsMin {
			return orderedRules[i].ShotsMin < orderedRules[j].ShotsMin
		}
		if orderedRules[i].ShotsMax != orderedRules[j].ShotsMax {
			return orderedRules[i].ShotsMax < orderedRules[j].ShotsMax
		}
		return orderedRules[i].Motivation < orderedRules[j].Motivation
	})

	for idx, rule := range orderedRules {
		row := []interface{}{rule.ShotsMin, rule.ShotsMax, rule.Motivation.String(), rule.Lambda}
		if err := writeRow(f, sheet, idx+2, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", idx+2), fmt.Sprintf("D%d", idx+2), style); err != nil {
			return err
		}
	}
	if err := f.SetColWidth(sheet, "A", "D", 16); err != nil {
		return err
	}
	if err := freezeTopRows(f, sheet, 1, "A2"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, "A1:D"+strconv.Itoa(len(orderedRules)+1), nil); err != nil {
		return err
	}
	return nil
}

func freezeTopRows(f *excelize.File, sheet string, rows int, topLeftCell string) error {
	return f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		YSplit:      rows,
		TopLeftCell: topLeftCell,
		ActivePane:  "bottomLeft",
	})
}

func writeHeaderRow(f *excelize.File, sheet string, row int, headers []string, style int) error {
	return writeHeaderRowAt(f, sheet, row, 1, headers, style)
}

func writeHeaderRowAt(f *excelize.File, sheet string, row int, startCol int, headers []string, style int) error {
	values := make([]interface{}, len(headers))
	for i, header := range headers {
		values[i] = header
	}
	if err := writeRowAt(f, sheet, row, startCol, values); err != nil {
		return err
	}
	lastCell, err := excelize.CoordinatesToCellName(startCol+len(headers)-1, row)
	if err != nil {
		return err
	}
	startCell, err := excelize.CoordinatesToCellName(startCol, row)
	if err != nil {
		return err
	}
	return f.SetCellStyle(sheet, startCell, lastCell, style)
}

func writeRow(f *excelize.File, sheet string, row int, values []interface{}) error {
	return writeRowAt(f, sheet, row, 1, values)
}

func writeRowAt(f *excelize.File, sheet string, row int, startCol int, values []interface{}) error {
	startCell, err := excelize.CoordinatesToCellName(startCol, row)
	if err != nil {
		return err
	}
	return f.SetSheetRow(sheet, startCell, &values)
}

func setSummaryWidths(f *excelize.File, sheet string) error {
	widths := map[string]float64{
		"A": 24,
		"B": 14,
		"C": 16,
		"D": 18,
		"E": 22,
		"F": 22,
		"G": 18,
		"H": 18,
		"I": 18,
		"J": 18,
		"K": 20,
		"L": 20,
		"M": 16,
		"N": 16,
		"O": 14,
		"P": 16,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	return nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func isWorldCup2026(cfg model.Config) bool {
	return cfg.Name == "Mundial 2026" || (cfg.Groups == 12 && cfg.TeamsPerGroup == 4 && cfg.QualifiedPerGroup == 2 && cfg.BestThirds == 8)
}

package report

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// WriteGroupTemplateXLSX creates a manual Excel template for a single group.
// The user fills match scores in the Partidos sheet and Excel recalculates the standings.
func WriteGroupTemplateXLSX(path string, groupName string, teams []string) error {
	teams = cleanTeams(teams)
	if len(teams) < 2 {
		return fmt.Errorf("debes indicar al menos 2 equipos para la plantilla del grupo")
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
	if _, err := f.NewSheet("Tabla Base"); err != nil {
		return err
	}
	if _, err := f.NewSheet("Posiciones"); err != nil {
		return err
	}
	if _, err := f.NewSheet("Resumen"); err != nil {
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

	if err := buildGroupInstructionsSheet(f, groupName, teams, titleStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle); err != nil {
		return err
	}
	if err := buildGroupMatchesTemplateSheet(f, groupName, pairings, titleStyle, headerStyle, bodyStyle, altBodyStyle, inputStyle, noteStyle); err != nil {
		return err
	}
	if err := buildGroupBaseTableSheet(f, groupName, teams, pairings, titleStyle, headerStyle, bodyStyle, altBodyStyle, formulaStyle, noteStyle); err != nil {
		return err
	}
	if err := buildGroupPositionsSheet(f, groupName, teams, titleStyle, headerStyle, bodyStyle, altBodyStyle, resultStyle, noteStyle); err != nil {
		return err
	}
	if err := buildGroupSummarySheet(f, groupName, pairings, titleStyle, headerStyle, bodyStyle, altBodyStyle, formulaStyle, noteStyle); err != nil {
		return err
	}

	if idx, err := f.GetSheetIndex("Posiciones"); err == nil {
		f.SetActiveSheet(idx)
	}

	if err := ensureDir(path); err != nil {
		return err
	}
	return f.SaveAs(path)
}

type groupPairing struct {
	Index  int
	TeamA  string
	TeamB  string
	RowNum int
}

func cleanTeams(teams []string) []string {
	cleaned := make([]string, 0, len(teams))
	seen := map[string]struct{}{}
	for _, team := range teams {
		name := strings.TrimSpace(team)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cleaned = append(cleaned, name)
	}
	return cleaned
}

func buildPairings(teams []string) []groupPairing {
	pairings := make([]groupPairing, 0, len(teams)*(len(teams)-1)/2)
	index := 1
	for i := 0; i < len(teams); i++ {
		for j := i + 1; j < len(teams); j++ {
			pairings = append(pairings, groupPairing{
				Index: index,
				TeamA: teams[i],
				TeamB: teams[j],
			})
			index++
		}
	}
	return pairings
}

func buildGroupInstructionsSheet(f *excelize.File, groupName string, teams []string, titleStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle int) error {
	sheet := "Instrucciones"

	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Plantilla manual del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Completa los goles en la hoja Partidos. La tabla de posiciones se recalcula al abrir Excel."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	if err := f.SetCellValue(sheet, "A4", "Configuracion"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A4", "A4", subHeaderStyle); err != nil {
		return err
	}
	rows := [][]interface{}{
		{"grupo", groupName},
		{"equipos", len(teams)},
		{"partidos_generados", len(teams) * (len(teams) - 1) / 2},
		{"desempate", "Puntos, diferencia de gol y goles a favor"},
		{"nota", "Si hay empate total, se conserva el orden de captura de los equipos"},
	}
	for idx, row := range rows {
		if err := writeRow(f, sheet, 5+idx, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", 5+idx), fmt.Sprintf("B%d", 5+idx), style); err != nil {
			return err
		}
	}

	if err := f.SetCellValue(sheet, "D4", "Equipos cargados"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "D4", "D4", subHeaderStyle); err != nil {
		return err
	}
	if err := writeHeaderRowAt(f, sheet, 5, 4, []string{"orden", "equipo"}, subHeaderStyle); err != nil {
		return err
	}
	for idx, team := range teams {
		if err := writeRowAt(f, sheet, 6+idx, 4, []interface{}{idx + 1, team}); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("D%d", 6+idx), fmt.Sprintf("E%d", 6+idx), style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "A", 24); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 38); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "D", "D", 12); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "E", "E", 24); err != nil {
		return err
	}
	return nil
}

func buildGroupMatchesTemplateSheet(f *excelize.File, groupName string, pairings []groupPairing, titleStyle, headerStyle, bodyStyle, altBodyStyle, inputStyle, noteStyle int) error {
	sheet := "Partidos"

	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Partidos del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Ingresa goles_a y goles_b. Los equipos ya vienen propuestos en cada cruce."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	headers := []string{"partido", "equipo_a", "goles_a", "goles_b", "equipo_b"}
	if err := writeHeaderRow(f, sheet, 4, headers, headerStyle); err != nil {
		return err
	}

	for idx, pair := range pairings {
		rowNum := 5 + idx
		if err := writeRow(f, sheet, rowNum, []interface{}{pair.Index, pair.TeamA, "", "", pair.TeamB}); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("E%d", rowNum), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("C%d", rowNum), fmt.Sprintf("D%d", rowNum), inputStyle); err != nil {
			return err
		}
	}

	widths := map[string]float64{
		"A": 12,
		"B": 24,
		"C": 12,
		"D": 12,
		"E": 24,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if len(pairings) > 0 {
		if err := f.AutoFilter(sheet, fmt.Sprintf("A4:E%d", 4+len(pairings)), nil); err != nil {
			return err
		}
	}
	return nil
}

func buildGroupBaseTableSheet(f *excelize.File, groupName string, teams []string, pairings []groupPairing, titleStyle, headerStyle, bodyStyle, altBodyStyle, formulaStyle, noteStyle int) error {
	sheet := "Tabla Base"
	startRow := 5
	endRow := 4 + len(pairings)
	teamEndRow := startRow + len(teams) - 1

	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Tabla base del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Las celdas verdes calculan la estadistica de cada equipo segun los resultados de la hoja Partidos."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	headers := []string{"equipo", "pj", "pg", "pe", "pp", "gf", "gc", "dg", "pts", "clave_orden"}
	if err := writeHeaderRow(f, sheet, 4, headers, headerStyle); err != nil {
		return err
	}

	partidosTeamA := fmt.Sprintf("'Partidos'!$B$5:$B$%d", endRow)
	partidosGoalsA := fmt.Sprintf("'Partidos'!$C$5:$C$%d", endRow)
	partidosGoalsB := fmt.Sprintf("'Partidos'!$D$5:$D$%d", endRow)
	partidosTeamB := fmt.Sprintf("'Partidos'!$E$5:$E$%d", endRow)

	for idx, team := range teams {
		row := startRow + idx
		teamCell := fmt.Sprintf("A%d", row)
		if err := f.SetCellValue(sheet, teamCell, team); err != nil {
			return err
		}
		rowStyleStart := fmt.Sprintf("A%d", row)
		rowStyleEnd := fmt.Sprintf("J%d", row)
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, rowStyleStart, rowStyleEnd, style); err != nil {
			return err
		}

		pjFormula := fmt.Sprintf(`=SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>""))+SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>""))`, partidosTeamA, row, partidosGoalsA, partidosGoalsB, partidosTeamB, row, partidosGoalsA, partidosGoalsB)
		pgFormula := fmt.Sprintf(`=SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>"")*(%s>%s))+SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>"")*(%s>%s))`, partidosTeamA, row, partidosGoalsA, partidosGoalsB, partidosGoalsA, partidosGoalsB, partidosTeamB, row, partidosGoalsA, partidosGoalsB, partidosGoalsB, partidosGoalsA)
		peFormula := fmt.Sprintf(`=SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>"")*(%s=%s))+SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>"")*(%s=%s))`, partidosTeamA, row, partidosGoalsA, partidosGoalsB, partidosGoalsA, partidosGoalsB, partidosTeamB, row, partidosGoalsA, partidosGoalsB, partidosGoalsB, partidosGoalsA)
		ppFormula := fmt.Sprintf(`=SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>"")*(%s<%s))+SUMPRODUCT((%s=A%d)*(%s<>"")*(%s<>"")*(%s<%s))`, partidosTeamA, row, partidosGoalsA, partidosGoalsB, partidosGoalsA, partidosGoalsB, partidosTeamB, row, partidosGoalsA, partidosGoalsB, partidosGoalsB, partidosGoalsA)
		gfFormula := fmt.Sprintf(`=SUMPRODUCT((%s=A%d)*N(%s))+SUMPRODUCT((%s=A%d)*N(%s))`, partidosTeamA, row, partidosGoalsA, partidosTeamB, row, partidosGoalsB)
		gcFormula := fmt.Sprintf(`=SUMPRODUCT((%s=A%d)*N(%s))+SUMPRODUCT((%s=A%d)*N(%s))`, partidosTeamA, row, partidosGoalsB, partidosTeamB, row, partidosGoalsA)
		dgFormula := fmt.Sprintf(`=F%d-G%d`, row, row)
		ptsFormula := fmt.Sprintf(`=C%d*3+D%d`, row, row)
		keyFormula := fmt.Sprintf(`=I%d*1000000+H%d*1000+F%d+(1000-ROW())/1000000`, row, row, row)

		for _, item := range []struct {
			cell string
			form string
		}{
			{fmt.Sprintf("B%d", row), pjFormula},
			{fmt.Sprintf("C%d", row), pgFormula},
			{fmt.Sprintf("D%d", row), peFormula},
			{fmt.Sprintf("E%d", row), ppFormula},
			{fmt.Sprintf("F%d", row), gfFormula},
			{fmt.Sprintf("G%d", row), gcFormula},
			{fmt.Sprintf("H%d", row), dgFormula},
			{fmt.Sprintf("I%d", row), ptsFormula},
			{fmt.Sprintf("J%d", row), keyFormula},
		} {
			if err := f.SetCellFormula(sheet, item.cell, item.form); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, item.cell, item.cell, formulaStyle); err != nil {
				return err
			}
		}
		if err := f.SetCellStyle(sheet, teamCell, teamCell, style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "A", 24); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "J", 14); err != nil {
		return err
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, fmt.Sprintf("A4:J%d", teamEndRow), nil); err != nil {
		return err
	}
	return nil
}

func buildGroupPositionsSheet(f *excelize.File, groupName string, teams []string, titleStyle, headerStyle, bodyStyle, altBodyStyle, resultStyle, noteStyle int) error {
	sheet := "Posiciones"
	startRow := 5
	teamEndRow := startRow + len(teams) - 1

	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Tabla de posiciones del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Se ordena por puntos, diferencia de gol, goles a favor y orden de captura como ultimo desempate."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	headers := []string{"pos", "equipo", "pj", "pg", "pe", "pp", "gf", "gc", "dg", "pts"}
	if err := writeHeaderRow(f, sheet, 4, headers, headerStyle); err != nil {
		return err
	}

	baseTeams := fmt.Sprintf("'Tabla Base'!$A$5:$A$%d", teamEndRow)
	basePJ := fmt.Sprintf("'Tabla Base'!$B$5:$B$%d", teamEndRow)
	basePG := fmt.Sprintf("'Tabla Base'!$C$5:$C$%d", teamEndRow)
	basePE := fmt.Sprintf("'Tabla Base'!$D$5:$D$%d", teamEndRow)
	basePP := fmt.Sprintf("'Tabla Base'!$E$5:$E$%d", teamEndRow)
	baseGF := fmt.Sprintf("'Tabla Base'!$F$5:$F$%d", teamEndRow)
	baseGC := fmt.Sprintf("'Tabla Base'!$G$5:$G$%d", teamEndRow)
	baseDG := fmt.Sprintf("'Tabla Base'!$H$5:$H$%d", teamEndRow)
	basePTS := fmt.Sprintf("'Tabla Base'!$I$5:$I$%d", teamEndRow)
	baseKey := fmt.Sprintf("'Tabla Base'!$J$5:$J$%d", teamEndRow)

	for idx := range teams {
		row := startRow + idx
		rank := idx + 1
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row), rank); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, baseTeams, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, basePJ, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, basePG, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, basePE, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, basePP, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("G%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, baseGF, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, baseGC, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("I%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, baseDG, baseKey, rank, baseKey)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("J%d", row), fmt.Sprintf(`=INDEX(%s,MATCH(LARGE(%s,%d),%s,0))`, basePTS, baseKey, rank, baseKey)); err != nil {
			return err
		}

		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("J%d", row), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("J%d", row), resultStyle); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), bodyStyle); err != nil {
			return err
		}
	}

	widths := map[string]float64{
		"A": 8,
		"B": 24,
		"C": 8,
		"D": 8,
		"E": 8,
		"F": 8,
		"G": 8,
		"H": 8,
		"I": 8,
		"J": 8,
	}
	for col, width := range widths {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return err
		}
	}
	if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, fmt.Sprintf("A4:J%d", teamEndRow), nil); err != nil {
		return err
	}
	return nil
}

func buildGroupSummarySheet(f *excelize.File, groupName string, pairings []groupPairing, titleStyle, headerStyle, bodyStyle, altBodyStyle, formulaStyle, noteStyle int) error {
	sheet := "Resumen"
	endRow := 4 + len(pairings)

	if err := f.SetCellValue(sheet, "A1", fmt.Sprintf("Resumen del grupo %s", groupName)); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Calcula el estado general del grupo a partir de los resultados ingresados."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	headers := []string{"estadistica", "valor"}
	if err := writeHeaderRow(f, sheet, 4, headers, headerStyle); err != nil {
		return err
	}

	rows := []struct {
		label string
		form  string
	}{
		{"partidos_programados", strconv.Itoa(len(pairings))},
		{"partidos_con_resultado", fmt.Sprintf(`=COUNTIFS('Partidos'!$C$5:$C$%d,"<>",'Partidos'!$D$5:$D$%d,"<>")`, endRow, endRow)},
		{"goles_totales", fmt.Sprintf(`=SUMPRODUCT(N('Partidos'!$C$5:$C$%d))+SUMPRODUCT(N('Partidos'!$D$5:$D$%d))`, endRow, endRow)},
		{"promedio_goles_por_partido", fmt.Sprintf(`=IF(B6=0,0,B7/B6)`)},
		{"empates", fmt.Sprintf(`=SUMPRODUCT(('Partidos'!$C$5:$C$%d='Partidos'!$D$5:$D$%d)*('Partidos'!$C$5:$C$%d<>"")*('Partidos'!$D$5:$D$%d<>""))`, endRow, endRow, endRow, endRow)},
	}
	for idx, row := range rows {
		excelRow := 5 + idx
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.label); err != nil {
			return err
		}
		if strings.HasPrefix(row.form, "=") {
			if err := f.SetCellFormula(sheet, fmt.Sprintf("B%d", excelRow), row.form); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), formulaStyle); err != nil {
				return err
			}
		} else {
			if err := f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.form); err != nil {
				return err
			}
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "A", 28); err != nil {
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

func sanitizeTemplateFileName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "grupo"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		default:
			builder.WriteRune('_')
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "grupo"
	}
	return result
}

func GroupTemplatePath(outDir, groupName string) string {
	fileName := fmt.Sprintf("grupo_%s_template.xlsx", sanitizeTemplateFileName(groupName))
	return filepath.Join(outDir, fileName)
}

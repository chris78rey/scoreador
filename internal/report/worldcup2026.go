package report

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"scoreador/internal/model"
)

type teamInsight struct {
	Team             string
	Group            string
	AvgShots         float64
	Motivation       model.Motivation
	Matches          int
	QualifiedCount   int
	QualificationPct float64
}

func createWorldCup2026Sheets(f *excelize.File) error {
	sheets := []string{"Formato 2026", "Grupos"}
	for _, group := range worldCupGroupOrder() {
		sheets = append(sheets, "Grupo "+group)
	}
	sheets = append(sheets, "Clasificados", "Dieciseisavos", "Octavos", "Cuartos", "Semis", "Final")
	for _, sheet := range sheets {
		if _, err := f.NewSheet(sheet); err != nil {
			return err
		}
	}
	return nil
}

func buildWorldCup2026Sheets(f *excelize.File, cfg model.Config, matches []model.MatchInput, summary model.TournamentSummary, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle int) error {
	insights := buildTeamInsights(matches, summary)
	groupMatches := groupMatchesByGroup(matches)
	groupTeams := teamsByGroup(matches)

	if err := buildWorldCupFormatSheet(f, cfg, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle); err != nil {
		return err
	}
	if err := buildGroupOverviewSheet(f, groupTeams, insights, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle); err != nil {
		return err
	}
	if err := buildGroupDetailSheets(f, groupMatches, groupTeams, insights, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle); err != nil {
		return err
	}
	if err := buildQualifiedSheet(f, summary, insights, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle); err != nil {
		return err
	}
	if err := buildKnockoutTemplates(f, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
		return err
	}

	return nil
}

func buildTeamInsights(matches []model.MatchInput, summary model.TournamentSummary) map[string]teamInsight {
	type accumulator struct {
		group       string
		shots       int
		appearances int
		motivation  map[model.Motivation]int
	}

	acc := map[string]*accumulator{}
	teamGroup := map[string]string{}
	for _, match := range matches {
		if _, ok := teamGroup[match.TeamA]; !ok {
			teamGroup[match.TeamA] = match.Group
		}
		if _, ok := teamGroup[match.TeamB]; !ok {
			teamGroup[match.TeamB] = match.Group
		}
		add := func(team string, shots int, mot model.Motivation) {
			entry := acc[team]
			if entry == nil {
				entry = &accumulator{group: match.Group, motivation: map[model.Motivation]int{}}
				acc[team] = entry
			}
			entry.shots += shots
			entry.appearances++
			entry.motivation[mot]++
		}
		add(match.TeamA, match.ShotsA, match.MotivationA)
		add(match.TeamB, match.ShotsB, match.MotivationB)
	}

	summaryByTeam := map[string]model.TeamStats{}
	for _, team := range summary.Teams {
		summaryByTeam[team.Team] = team
	}

	insights := make(map[string]teamInsight, len(acc))
	for team, entry := range acc {
		avgShots := 0.0
		if entry.appearances > 0 {
			avgShots = float64(entry.shots) / float64(entry.appearances)
		}
		stat := summaryByTeam[team]
		insights[team] = teamInsight{
			Team:             team,
			Group:            teamGroup[team],
			AvgShots:         avgShots,
			Motivation:       modeMotivation(entry.motivation),
			Matches:          entry.appearances,
			QualifiedCount:   stat.Qualified,
			QualificationPct: stat.QualifiedPct,
		}
	}

	return insights
}

func modeMotivation(values map[model.Motivation]int) model.Motivation {
	if len(values) == 0 {
		return model.MotivationMedium
	}
	type pair struct {
		mot   model.Motivation
		count int
	}
	var best pair
	for mot, count := range values {
		if count > best.count || (count == best.count && mot > best.mot) {
			best = pair{mot: mot, count: count}
		}
	}
	return best.mot
}

func groupMatchesByGroup(matches []model.MatchInput) map[string][]model.MatchInput {
	grouped := map[string][]model.MatchInput{}
	for _, match := range matches {
		if strings.EqualFold(match.Stage, "group") {
			grouped[match.Group] = append(grouped[match.Group], match)
		}
	}
	return grouped
}

func teamsByGroup(matches []model.MatchInput) map[string][]string {
	grouped := map[string][]string{}
	seen := map[string]map[string]struct{}{}
	for _, match := range matches {
		if !strings.EqualFold(match.Stage, "group") {
			continue
		}
		if seen[match.Group] == nil {
			seen[match.Group] = map[string]struct{}{}
		}
		for _, team := range []string{match.TeamA, match.TeamB} {
			if _, ok := seen[match.Group][team]; ok {
				continue
			}
			seen[match.Group][team] = struct{}{}
			grouped[match.Group] = append(grouped[match.Group], team)
		}
	}
	return grouped
}

func buildWorldCupFormatSheet(f *excelize.File, cfg model.Config, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle int) error {
	sheet := "Formato 2026"

	if err := f.SetCellValue(sheet, "A1", "Mundial 2026"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Plantilla de fases del Mundial 2026. Los cruces de eliminación se calculan en Excel."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	rows := [][]interface{}{
		{"equipo_total", 48},
		{"grupos", 12},
		{"equipos_por_grupo", 4},
		{"clasifican_por_grupo", 2},
		{"mejores_terceros", 8},
		{"partidos_fase_grupos", 72},
		{"partidos_totales_torneo", 104},
		{"fase_cargada_por_el_motor", "solo fase de grupos"},
		{"fase_restante", "se resuelve por pestañas en Excel"},
	}
	if err := writeHeaderRow(f, sheet, 3, []string{"campo", "valor"}, headerStyle); err != nil {
		return err
	}
	for idx, row := range rows {
		if err := writeRow(f, sheet, idx+4, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", idx+4), fmt.Sprintf("B%d", idx+4), style); err != nil {
			return err
		}
	}
	if err := f.SetCellValue(sheet, "D3", "Llave de dieciseisavos"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "D3", "D3", subHeaderStyle); err != nil {
		return err
	}

	bracketHeaders := []string{"partido", "origen_a", "origen_b", "equipo_a", "equipo_b", "marcador_a", "marcador_b", "ganador"}
	if err := writeHeaderRowAt(f, sheet, 4, 4, bracketHeaders, headerStyle); err != nil {
		return err
	}
	for idx, row := range worldCupR32Rows() {
		if err := writeRowAt(f, sheet, idx+5, 4, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("D%d", idx+5), fmt.Sprintf("K%d", idx+5), style); err != nil {
			return err
		}
	}

	if err := f.SetCellValue(sheet, "A15", "Siguientes fases"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A15", "A15", subHeaderStyle); err != nil {
		return err
	}
	if err := writeHeaderRow(f, sheet, 16, []string{"fase", "partidos", "descripcion"}, headerStyle); err != nil {
		return err
	}
	phases := [][]interface{}{
		{"Dieciseisavos", 16, "Partidos 73-88"},
		{"Octavos", 8, "Partidos 89-96"},
		{"Cuartos", 4, "Partidos 97-100"},
		{"Semis", 2, "Partidos 101-102"},
		{"Final", 2, "Tercer puesto y final"},
	}
	for idx, row := range phases {
		if err := writeRow(f, sheet, idx+17, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", idx+17), fmt.Sprintf("C%d", idx+17), style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "A", 26); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "C", "C", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "D", "K", 18); err != nil {
		return err
	}
	return nil
}

func buildGroupOverviewSheet(f *excelize.File, teamsByGroup map[string][]string, insights map[string]teamInsight, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle int) error {
	sheet := "Grupos"
	if err := f.SetCellValue(sheet, "A1", "Resumen de grupos"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}

	headers := []string{"grupo", "equipo", "tiros_promedio", "motivacion_moda", "clasifica_pct"}
	if err := writeHeaderRow(f, sheet, 3, headers, headerStyle); err != nil {
		return err
	}

	row := 4
	for _, group := range worldCupGroupOrder() {
		teams := append([]string(nil), teamsByGroup[group]...)
		sort.Strings(teams)
		if len(teams) == 0 {
			continue
		}
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row-1), fmt.Sprintf("Grupo %s", group)); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row-1), fmt.Sprintf("A%d", row-1), subHeaderStyle); err != nil {
			return err
		}
		for _, team := range teams {
			insight := insights[team]
			values := []interface{}{group, team, insight.AvgShots, insight.Motivation.String(), insight.QualificationPct}
			if err := writeRow(f, sheet, row, values); err != nil {
				return err
			}
			style := bodyStyle
			if row%2 == 0 {
				style = altBodyStyle
			}
			if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("E%d", row), style); err != nil {
				return err
			}
			row++
		}
		row++
	}

	if err := f.SetColWidth(sheet, "A", "B", 16); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "C", "E", 18); err != nil {
		return err
	}
	if err := freezeTopRows(f, sheet, 3, "A4"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, "A3:E"+fmt.Sprintf("%d", row-1), nil); err != nil {
		return err
	}
	return nil
}

func buildGroupDetailSheets(f *excelize.File, groupMatches map[string][]model.MatchInput, groupTeams map[string][]string, insights map[string]teamInsight, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle int) error {
	for _, group := range worldCupGroupOrder() {
		sheet := "Grupo " + group
		if err := f.SetCellValue(sheet, "A1", "Grupo "+group); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, "A2", "Partidos simulados de fase de grupos y perfil de equipos."); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
			return err
		}

		if err := f.SetCellValue(sheet, "A3", "Partidos"); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, "A3", "A3", subHeaderStyle); err != nil {
			return err
		}
		if err := writeHeaderRow(f, sheet, 4, []string{"match_id", "equipo_a", "equipo_b", "tiros_a", "tiros_b", "mot_a", "mot_b"}, headerStyle); err != nil {
			return err
		}
		row := 5
		matches := append([]model.MatchInput(nil), groupMatches[group]...)
		sort.SliceStable(matches, func(i, j int) bool { return matches[i].MatchID < matches[j].MatchID })
		for _, match := range matches {
			values := []interface{}{match.MatchID, match.TeamA, match.TeamB, match.ShotsA, match.ShotsB, match.MotivationA.String(), match.MotivationB.String()}
			if err := writeRow(f, sheet, row, values); err != nil {
				return err
			}
			style := bodyStyle
			if row%2 == 0 {
				style = altBodyStyle
			}
			if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), style); err != nil {
				return err
			}
			row++
		}

		row += 2
		if err := f.SetCellValue(sheet, fmt.Sprintf("A%d", row-1), "Perfil de equipos"); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row-1), fmt.Sprintf("A%d", row-1), subHeaderStyle); err != nil {
			return err
		}
		if err := writeHeaderRow(f, sheet, row, []string{"equipo", "tiros_promedio", "motivacion_moda", "clasifica_pct"}, headerStyle); err != nil {
			return err
		}
		row++
		teams := append([]string(nil), groupTeams[group]...)
		sort.Strings(teams)
		for _, team := range teams {
			insight := insights[team]
			values := []interface{}{team, insight.AvgShots, insight.Motivation.String(), insight.QualificationPct}
			if err := writeRow(f, sheet, row, values); err != nil {
				return err
			}
			style := bodyStyle
			if row%2 == 0 {
				style = altBodyStyle
			}
			if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), style); err != nil {
				return err
			}
			row++
		}

		if err := f.SetColWidth(sheet, "A", "G", 16); err != nil {
			return err
		}
		if err := freezeTopRows(f, sheet, 4, "A5"); err != nil {
			return err
		}
	}
	return nil
}

func buildQualifiedSheet(f *excelize.File, summary model.TournamentSummary, insights map[string]teamInsight, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, noteStyle int) error {
	sheet := "Clasificados"
	if err := f.SetCellValue(sheet, "A1", "Ranking de clasificacion"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Los primeros 32 equipos de este ranking alimentan la llave de dieciseisavos."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}

	headers := []string{"pos", "equipo", "grupo", "clasifico_pct", "veces_clasifico", "tiros_promedio", "motivacion"}
	if err := writeHeaderRow(f, sheet, 3, headers, headerStyle); err != nil {
		return err
	}

	teams := append([]model.TeamStats(nil), summary.Teams...)
	sort.SliceStable(teams, func(i, j int) bool {
		if teams[i].QualifiedPct != teams[j].QualifiedPct {
			return teams[i].QualifiedPct > teams[j].QualifiedPct
		}
		return teams[i].Team < teams[j].Team
	})

	for idx, team := range teams {
		insight := insights[team.Team]
		values := []interface{}{idx + 1, team.Team, insight.Group, team.QualifiedPct, team.Qualified, insight.AvgShots, insight.Motivation.String()}
		if err := writeRow(f, sheet, idx+4, values); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", idx+4), fmt.Sprintf("G%d", idx+4), style); err != nil {
			return err
		}
	}

	if err := f.SetCellValue(sheet, "I3", "Clasificados por fase"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "I3", "I3", subHeaderStyle); err != nil {
		return err
	}
	if err := writeHeaderRowAt(f, sheet, 4, 9, []string{"fase", "cupo"}, headerStyle); err != nil {
		return err
	}
	phaseRows := [][]interface{}{
		{"Top 2 por grupo", 24},
		{"8 mejores terceros", 8},
		{"Total", 32},
	}
	for idx, row := range phaseRows {
		if err := writeRowAt(f, sheet, idx+5, 9, row); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("I%d", idx+5), fmt.Sprintf("J%d", idx+5), style); err != nil {
			return err
		}
	}

	if err := f.SetColWidth(sheet, "A", "G", 18); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "I", "J", 20); err != nil {
		return err
	}
	if err := freezeTopRows(f, sheet, 3, "A4"); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, "A3:G"+strconv.Itoa(3+len(teams)), nil); err != nil {
		return err
	}
	return nil
}

func buildKnockoutTemplates(f *excelize.File, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle int) error {
	if err := buildPhaseTemplate(f, phaseTemplateSpec{
		sheet:      "Dieciseisavos",
		firstMatch: 73,
		rows:       16,
		prevSheet:  "Clasificados",
		sourceRow:  4,
		rowsPerSrc: 1,
		usesLoser:  false,
		templates:  worldCupR32Rows(),
	}, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
		return err
	}
	if err := buildPhaseTemplate(f, phaseTemplateSpec{
		sheet:      "Octavos",
		firstMatch: 89,
		rows:       8,
		prevSheet:  "Dieciseisavos",
		sourceRow:  5,
		rowsPerSrc: 2,
		usesLoser:  false,
		templates:  worldCupOctavosRows(),
	}, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
		return err
	}
	if err := buildPhaseTemplate(f, phaseTemplateSpec{
		sheet:      "Cuartos",
		firstMatch: 97,
		rows:       4,
		prevSheet:  "Octavos",
		sourceRow:  5,
		rowsPerSrc: 2,
		usesLoser:  false,
		templates:  worldCupCuartosRows(),
	}, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
		return err
	}
	if err := buildPhaseTemplate(f, phaseTemplateSpec{
		sheet:      "Semis",
		firstMatch: 101,
		rows:       2,
		prevSheet:  "Cuartos",
		sourceRow:  5,
		rowsPerSrc: 2,
		usesLoser:  false,
		templates:  worldCupSemisRows(),
	}, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
		return err
	}
	if err := buildPhaseTemplate(f, phaseTemplateSpec{
		sheet:      "Final",
		firstMatch: 103,
		rows:       2,
		prevSheet:  "Semis",
		sourceRow:  5,
		rowsPerSrc: 1,
		usesLoser:  true,
		templates:  worldCupFinalRows(),
	}, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle); err != nil {
		return err
	}
	return nil
}

type phaseTemplateSpec struct {
	sheet      string
	firstMatch int
	rows       int
	prevSheet  string
	sourceRow  int
	rowsPerSrc int
	usesLoser  bool
	templates  [][]interface{}
}

func buildPhaseTemplate(f *excelize.File, spec phaseTemplateSpec, titleStyle, headerStyle, subHeaderStyle, bodyStyle, altBodyStyle, inputStyle, formulaStyle, resultStyle, noteStyle int) error {
	sheet := spec.sheet
	if err := f.SetCellValue(sheet, "A1", sheet); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A2", "Ingresa los marcadores en las columnas F y G. H e I se calculan automáticamente."); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A2", "A2", noteStyle); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, "A3", "Fase eliminatoria"); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A3", "A3", subHeaderStyle); err != nil {
		return err
	}

	headers := []string{"partido", "origen_a", "origen_b", "equipo_a", "equipo_b", "marcador_a", "marcador_b", "ganador", "perdedor"}
	if err := writeHeaderRow(f, sheet, 4, headers, headerStyle); err != nil {
		return err
	}
	for idx := 0; idx < spec.rows && idx < len(spec.templates); idx++ {
		rowNum := idx + 5
		template := spec.templates[idx]
		values := make([]interface{}, 9)
		if len(template) > 0 {
			values[0] = template[0]
		}
		if len(template) > 1 {
			values[1] = template[1]
		}
		if len(template) > 2 {
			values[2] = template[2]
		}
		if len(template) > 5 {
			values[5] = template[5]
		}
		if len(template) > 6 {
			values[6] = template[6]
		}
		if err := writeRow(f, sheet, rowNum, values); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("D%d", rowNum), knockoutTeamFormula(spec, idx, false)); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("E%d", rowNum), knockoutTeamFormula(spec, idx, true)); err != nil {
			return err
		}
		winnerFormula := fmt.Sprintf(`IF(OR(F%d="",G%d=""),"",IF(F%d>G%d,D%d,IF(G%d>F%d,E%d,"DESEMPATE")))`, rowNum, rowNum, rowNum, rowNum, rowNum, rowNum, rowNum, rowNum)
		loserFormula := fmt.Sprintf(`IF(OR(F%d="",G%d=""),"",IF(F%d>G%d,E%d,IF(G%d>F%d,D%d,"DESEMPATE")))`, rowNum, rowNum, rowNum, rowNum, rowNum, rowNum, rowNum, rowNum)
		if err := f.SetCellFormula(sheet, fmt.Sprintf("H%d", rowNum), winnerFormula); err != nil {
			return err
		}
		if err := f.SetCellFormula(sheet, fmt.Sprintf("I%d", rowNum), loserFormula); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("C%d", rowNum), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("D%d", rowNum), fmt.Sprintf("E%d", rowNum), formulaStyle); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("F%d", rowNum), fmt.Sprintf("G%d", rowNum), inputStyle); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, fmt.Sprintf("H%d", rowNum), fmt.Sprintf("I%d", rowNum), resultStyle); err != nil {
			return err
		}
	}
	if err := f.SetCellValue(sheet, "J3", "Origen de equipos"); err == nil {
		_ = f.SetCellStyle(sheet, "J3", "J3", subHeaderStyle)
	}
	if err := f.SetColWidth(sheet, "A", "A", 12); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "C", 18); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "D", "E", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "F", "G", 12); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "H", "I", 18); err != nil {
		return err
	}
	if err := f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		YSplit:      4,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return err
	}
	if err := f.AutoFilter(sheet, "A4:I"+strconv.Itoa(4+spec.rows), nil); err != nil {
		return err
	}
	_ = spec.firstMatch
	return nil
}

func worldCupR32Rows() [][]interface{} {
	return [][]interface{}{
		{73, "A2", "B2", "A2 vs B2", "A2 vs B2", "", "", "A2/B2"},
		{74, "E1", "3rd A/B/C/D/F", "E1 vs 3rd", "E1 vs 3rd", "", "", "E1/3rd"},
		{75, "F1", "C2", "F1 vs C2", "F1 vs C2", "", "", "F1/C2"},
		{76, "C1", "F2", "C1 vs F2", "C1 vs F2", "", "", "C1/F2"},
		{77, "I1", "3rd C/D/F/G/H", "I1 vs 3rd", "I1 vs 3rd", "", "", "I1/3rd"},
		{78, "E2", "I2", "E2 vs I2", "E2 vs I2", "", "", "E2/I2"},
		{79, "A1", "3rd C/E/F/H/I", "A1 vs 3rd", "A1 vs 3rd", "", "", "A1/3rd"},
		{80, "L1", "3rd E/H/I/J/K", "L1 vs 3rd", "L1 vs 3rd", "", "", "L1/3rd"},
		{81, "D1", "3rd B/E/F/I/J", "D1 vs 3rd", "D1 vs 3rd", "", "", "D1/3rd"},
		{82, "G1", "3rd A/E/H/I/J", "G1 vs 3rd", "G1 vs 3rd", "", "", "G1/3rd"},
		{83, "K2", "L2", "K2 vs L2", "K2 vs L2", "", "", "K2/L2"},
		{84, "H1", "J2", "H1 vs J2", "H1 vs J2", "", "", "H1/J2"},
		{85, "B1", "3rd E/F/G/I/J", "B1 vs 3rd", "B1 vs 3rd", "", "", "B1/3rd"},
		{86, "J1", "H2", "J1 vs H2", "J1 vs H2", "", "", "J1/H2"},
		{87, "K1", "3rd D/E/I/J/L", "K1 vs 3rd", "K1 vs 3rd", "", "", "K1/3rd"},
		{88, "D2", "G2", "D2 vs G2", "D2 vs G2", "", "", "D2/G2"},
	}
}

func worldCupOctavosRows() [][]interface{} {
	return [][]interface{}{
		{89, "Winner 73", "Winner 74", "", "", "", "", ""},
		{90, "Winner 75", "Winner 76", "", "", "", "", ""},
		{91, "Winner 77", "Winner 78", "", "", "", "", ""},
		{92, "Winner 79", "Winner 80", "", "", "", "", ""},
		{93, "Winner 81", "Winner 82", "", "", "", "", ""},
		{94, "Winner 83", "Winner 84", "", "", "", "", ""},
		{95, "Winner 85", "Winner 86", "", "", "", "", ""},
		{96, "Winner 87", "Winner 88", "", "", "", "", ""},
	}
}

func worldCupCuartosRows() [][]interface{} {
	return [][]interface{}{
		{97, "Winner 89", "Winner 90", "", "", "", "", ""},
		{98, "Winner 91", "Winner 92", "", "", "", "", ""},
		{99, "Winner 93", "Winner 94", "", "", "", "", ""},
		{100, "Winner 95", "Winner 96", "", "", "", "", ""},
	}
}

func worldCupSemisRows() [][]interface{} {
	return [][]interface{}{
		{101, "Winner 97", "Winner 98", "", "", "", "", ""},
		{102, "Winner 99", "Winner 100", "", "", "", "", ""},
	}
}

func worldCupFinalRows() [][]interface{} {
	return [][]interface{}{
		{103, "Loser 101", "Loser 102", "", "", "", "", "", ""},
		{104, "Winner 101", "Winner 102", "", "", "", "", "", ""},
	}
}

func worldCupGroupOrder() []string {
	return []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
}

func knockoutTeamFormula(spec phaseTemplateSpec, index int, second bool) string {
	if spec.prevSheet == "Clasificados" {
		row := spec.sourceRow + index*2
		if second {
			row++
		}
		return fmt.Sprintf(`'Clasificados'!$B$%d`, row)
	}
	if spec.sheet == "Final" {
		previousRow := spec.sourceRow
		if second {
			previousRow++
		}
		if index == 0 {
			return fmt.Sprintf(`'%s'!$I$%d`, spec.prevSheet, previousRow)
		}
		return fmt.Sprintf(`'%s'!$H$%d`, spec.prevSheet, previousRow)
	}
	previousRow := spec.sourceRow + index*spec.rowsPerSrc
	if second {
		previousRow++
	}
	return fmt.Sprintf(`'%s'!$H$%d`, spec.prevSheet, previousRow)
}

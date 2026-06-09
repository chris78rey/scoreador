package report

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"

	"scoreador/internal/model"
)

type GroupMatchResult struct {
	TeamA  string
	TeamB  string
	GoalsA int
	GoalsB int
}

// WriteGroupDemoXLSX creates a filled group workbook for quick testing.
// It reuses the manual template, writes all match scores, and adds a static demo standings sheet.
func WriteGroupDemoXLSX(path string, groupName string, teams []string, results []GroupMatchResult) error {
	teams = cleanTeams(teams)
	if len(teams) < 2 {
		return fmt.Errorf("debes indicar al menos 2 equipos para la demo del grupo")
	}

	pairings := buildPairings(teams)
	if len(results) != len(pairings) {
		return fmt.Errorf("cantidad de resultados invalida: got %d want %d", len(results), len(pairings))
	}

	if err := WriteGroupTemplateXLSX(path, groupName, teams); err != nil {
		return err
	}

	wb, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = wb.Close()
	}()

	for idx, pair := range pairings {
		result := results[idx]
		if !samePair(pair.TeamA, pair.TeamB, result.TeamA, result.TeamB) {
			return fmt.Errorf("resultado fuera de orden en el partido %d: esperado %s vs %s, recibido %s vs %s", pair.Index, pair.TeamA, pair.TeamB, result.TeamA, result.TeamB)
		}
		row := 5 + idx
		if err := wb.SetCellValue("Partidos", fmt.Sprintf("C%d", row), result.GoalsA); err != nil {
			return err
		}
		if err := wb.SetCellValue("Partidos", fmt.Sprintf("D%d", row), result.GoalsB); err != nil {
			return err
		}
	}

	standings := computeGroupStandings(groupName, results)
	if err := writeGroupDemoSheet(wb, standings, results); err != nil {
		return err
	}

	if idx, err := wb.GetSheetIndex("Salida"); err == nil && idx > 0 {
		wb.SetActiveSheet(idx)
	}

	return wb.SaveAs(path)
}

// GroupDemoPath returns the output file name for a filled group demo workbook.
func GroupDemoPath(outDir, groupName string) string {
	fileName := fmt.Sprintf("grupo_%s_demo.xlsx", sanitizeTemplateFileName(groupName))
	return filepath.Join(outDir, fileName)
}

func DefaultGroupDemoResults(teams []string) []GroupMatchResult {
	teams = cleanTeams(teams)
	pairings := buildPairings(teams)
	results := make([]GroupMatchResult, 0, len(pairings))

	for _, pair := range pairings {
		var goalsA, goalsB int
		switch pair.Index {
		case 1:
			goalsA, goalsB = 1, 2
		case 2:
			goalsA, goalsB = 2, 1
		case 3:
			goalsA, goalsB = 0, 1
		case 4:
			goalsA, goalsB = 3, 0
		case 5:
			goalsA, goalsB = 2, 1
		case 6:
			goalsA, goalsB = 1, 1
		default:
			goalsA = 1 + (pair.Index % 3)
			goalsB = pair.Index % 2
		}
		results = append(results, GroupMatchResult{
			TeamA:  pair.TeamA,
			TeamB:  pair.TeamB,
			GoalsA: goalsA,
			GoalsB: goalsB,
		})
	}

	return results
}

func computeGroupStandings(groupName string, results []GroupMatchResult) []model.GroupStanding {
	byTeam := map[string]*model.GroupStanding{}
	order := make([]string, 0)

	getStanding := func(team string) *model.GroupStanding {
		if standing, ok := byTeam[team]; ok {
			return standing
		}
		standing := &model.GroupStanding{Team: team, Group: groupName}
		byTeam[team] = standing
		order = append(order, team)
		return standing
	}

	for _, result := range results {
		teamA := getStanding(result.TeamA)
		teamB := getStanding(result.TeamB)

		teamA.Played++
		teamB.Played++
		teamA.GoalsFor += result.GoalsA
		teamA.GoalsAgainst += result.GoalsB
		teamB.GoalsFor += result.GoalsB
		teamB.GoalsAgainst += result.GoalsA

		switch {
		case result.GoalsA > result.GoalsB:
			teamA.Points += 3
		case result.GoalsB > result.GoalsA:
			teamB.Points += 3
		default:
			teamA.Points++
			teamB.Points++
		}
	}

	standings := make([]model.GroupStanding, 0, len(order))
	for _, team := range order {
		standing := byTeam[team]
		standing.GoalDifference = standing.GoalsFor - standing.GoalsAgainst
		standings = append(standings, *standing)
	}

	sort.SliceStable(standings, func(i, j int) bool {
		return compareGroupStandings(standings[i], standings[j])
	})

	return standings
}

// GroupDemoStandings returns the computed standings for a filled demo group.
func GroupDemoStandings(groupName string, results []GroupMatchResult) []model.GroupStanding {
	return computeGroupStandings(groupName, results)
}

func compareGroupStandings(a, b model.GroupStanding) bool {
	if a.Points != b.Points {
		return a.Points > b.Points
	}
	if a.GoalDifference != b.GoalDifference {
		return a.GoalDifference > b.GoalDifference
	}
	if a.GoalsFor != b.GoalsFor {
		return a.GoalsFor > b.GoalsFor
	}
	if a.GoalsAgainst != b.GoalsAgainst {
		return a.GoalsAgainst < b.GoalsAgainst
	}
	return a.Team < b.Team
}

func writeGroupDemoSheet(wb *excelize.File, standings []model.GroupStanding, results []GroupMatchResult) error {
	const sheet = "Salida"

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

	if idx, err := wb.GetSheetIndex(sheet); err == nil && idx >= 0 {
		if err := wb.DeleteSheet(sheet); err != nil {
			return err
		}
	}
	if _, err := wb.NewSheet(sheet); err != nil {
		return err
	}

	titleStyle, err := wb.NewStyle(&excelize.Style{
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
	headerStyle, err := wb.NewStyle(&excelize.Style{
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
	bodyStyle, err := wb.NewStyle(&excelize.Style{
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
	altBodyStyle, err := wb.NewStyle(&excelize.Style{
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

	if err := wb.SetCellValue(sheet, "A1", "Salida demo del grupo"); err != nil {
		return err
	}
	if err := wb.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "A2", "Tabla calculada en Go para validar el grupo completo con marcadores de ejemplo."); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "A4", "pos"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "B4", "equipo"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "C4", "pj"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "D4", "pg"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "E4", "pe"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "F4", "pp"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "G4", "gf"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "H4", "gc"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "I4", "dg"); err != nil {
		return err
	}
	if err := wb.SetCellValue(sheet, "J4", "pts"); err != nil {
		return err
	}
	if err := wb.SetCellStyle(sheet, "A4", "J4", headerStyle); err != nil {
		return err
	}

	for idx, standing := range standings {
		row := 5 + idx
		rec := records[standing.Team]
		if err := wb.SetCellValue(sheet, fmt.Sprintf("A%d", row), idx+1); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("B%d", row), standing.Team); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("C%d", row), standing.Played); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("D%d", row), rec.Wins); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("E%d", row), rec.Draws); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("F%d", row), rec.Losses); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("G%d", row), standing.GoalsFor); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("H%d", row), standing.GoalsAgainst); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("I%d", row), standing.GoalDifference); err != nil {
			return err
		}
		if err := wb.SetCellValue(sheet, fmt.Sprintf("J%d", row), standing.Points); err != nil {
			return err
		}
		style := bodyStyle
		if idx%2 == 1 {
			style = altBodyStyle
		}
		if err := wb.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("J%d", row), style); err != nil {
			return err
		}
	}

	if err := wb.SetColWidth(sheet, "A", "A", 8); err != nil {
		return err
	}
	if err := wb.SetColWidth(sheet, "B", "B", 24); err != nil {
		return err
	}
	if err := wb.SetColWidth(sheet, "C", "J", 12); err != nil {
		return err
	}
	return nil
}

func samePair(expectedA, expectedB, gotA, gotB string) bool {
	return (sameTeam(expectedA, gotA) && sameTeam(expectedB, gotB)) || (sameTeam(expectedA, gotB) && sameTeam(expectedB, gotA))
}

func sameTeam(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

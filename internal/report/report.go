package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"scoreador/internal/model"
)

func WriteCSV(path string, summary model.TournamentSummary) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
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
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, team := range summary.Teams {
		row := []string{
			team.Team,
			fmt.Sprintf("%d", team.Simulations),
			fmt.Sprintf("%d", team.Qualified),
			formatPct(team.QualifiedPct),
			fmt.Sprintf("%d", team.Dieciseisavos),
			formatPct(team.DieciseisavosPct),
			fmt.Sprintf("%d", team.Octavos),
			formatPct(team.OctavosPct),
			fmt.Sprintf("%d", team.Cuartos),
			formatPct(team.CuartosPct),
			fmt.Sprintf("%d", team.Semifinal),
			formatPct(team.SemifinalPct),
			fmt.Sprintf("%d", team.Final),
			formatPct(team.FinalPct),
			fmt.Sprintf("%d", team.Campeon),
			formatPct(team.CampeonPct),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteJSON(path string, summary model.TournamentSummary) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

func formatPct(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

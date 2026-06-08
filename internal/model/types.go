package model

import (
	"fmt"
	"strings"
)

type Motivation string

const (
	MotivationLow    Motivation = "baja"
	MotivationMedium Motivation = "media"
	MotivationHigh   Motivation = "alta"
)

func ParseMotivation(value string) (Motivation, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(MotivationLow):
		return MotivationLow, nil
	case string(MotivationMedium):
		return MotivationMedium, nil
	case string(MotivationHigh):
		return MotivationHigh, nil
	default:
		return "", fmt.Errorf("motivacion invalida: %q", value)
	}
}

func (m Motivation) String() string {
	return string(m)
}

type Config struct {
	Name               string `json:"name"`
	Simulations        int    `json:"simulations"`
	Groups             int    `json:"groups"`
	TeamsPerGroup      int    `json:"teams_per_group"`
	QualifiedPerGroup  int    `json:"qualified_per_group"`
	BestThirds         int    `json:"best_thirds"`
	Knockout           bool   `json:"knockout"`
	KnockoutTiebreaker string `json:"knockout_tiebreaker"`
	Seed               int64  `json:"seed"`
}

func (c *Config) ApplyDefaults() {
	if strings.TrimSpace(c.Name) == "" {
		c.Name = "Torneo"
	}
	if c.Simulations <= 0 {
		c.Simulations = 1000
	}
	if strings.TrimSpace(c.KnockoutTiebreaker) == "" {
		c.KnockoutTiebreaker = "penalties"
	}
}

type MatchInput struct {
	MatchID     int
	Stage       string
	Group       string
	TeamA       string
	TeamB       string
	ShotsA      int
	ShotsB      int
	MotivationA Motivation
	MotivationB Motivation
}

type LambdaRule struct {
	ShotsMin   int
	ShotsMax   int
	Motivation Motivation
	Lambda     float64
}

type TeamProfile struct {
	Team       string
	AvgShots   float64
	Motivation Motivation
}

type GroupStanding struct {
	Team           string `json:"team"`
	Group          string `json:"group"`
	Played         int    `json:"played"`
	Points         int    `json:"points"`
	GoalsFor       int    `json:"goals_for"`
	GoalsAgainst   int    `json:"goals_against"`
	GoalDifference int    `json:"goal_difference"`
}

type TeamStats struct {
	Team             string  `json:"equipo"`
	Simulations      int     `json:"simulaciones"`
	Qualified        int     `json:"veces_clasifico"`
	QualifiedPct     float64 `json:"porcentaje_clasifico"`
	Dieciseisavos    int     `json:"veces_llego_dieciseisavos"`
	DieciseisavosPct float64 `json:"porcentaje_dieciseisavos"`
	Octavos          int     `json:"veces_llego_octavos"`
	OctavosPct       float64 `json:"porcentaje_octavos"`
	Cuartos          int     `json:"veces_llego_cuartos"`
	CuartosPct       float64 `json:"porcentaje_cuartos"`
	Semifinal        int     `json:"veces_llego_semifinal"`
	SemifinalPct     float64 `json:"porcentaje_semifinal"`
	Final            int     `json:"veces_llego_final"`
	FinalPct         float64 `json:"porcentaje_final"`
	Campeon          int     `json:"veces_campeon"`
	CampeonPct       float64 `json:"porcentaje_campeon"`
}

type TournamentSummary struct {
	Name        string      `json:"name"`
	Simulations int         `json:"simulations"`
	Teams       []TeamStats `json:"teams"`
}

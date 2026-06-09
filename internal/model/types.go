package model

import (
	"fmt"
	"strconv"
	"strings"
)

type Motivation int

const (
	MotivationLow    Motivation = 0
	MotivationMedium Motivation = 5
	MotivationHigh   Motivation = 10
)

func ParseMotivation(value string) (Motivation, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "baja":
		return MotivationLow, nil
	case "media":
		return MotivationMedium, nil
	case "alta":
		return MotivationHigh, nil
	default:
		n, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("motivacion invalida: %q", value)
		}
		return MotivationFromInt(n)
	}
}

func (m Motivation) String() string {
	return strconv.Itoa(int(m))
}

func (m Motivation) Valid() bool {
	return m >= 0 && m <= 10
}

func MotivationFromInt(value int) (Motivation, error) {
	motivation := Motivation(value)
	if !motivation.Valid() {
		return 0, fmt.Errorf("motivacion invalida: %d", value)
	}
	return motivation, nil
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

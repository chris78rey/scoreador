package sim

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"scoreador/internal/model"
	"scoreador/internal/poisson"
)

type SingleMatchInput struct {
	TeamA       string
	TeamB       string
	ShotsA      int
	ShotsB      int
	MotivationA model.Motivation
	MotivationB model.Motivation
	Tiebreaker  string
}

type SingleMatchResult struct {
	TeamA       string
	TeamB       string
	ShotsA      int
	ShotsB      int
	MotivationA model.Motivation
	MotivationB model.Motivation
	LambdaA     float64
	LambdaB     float64
	GoalsA      int
	GoalsB      int
	Winner      string
	DecidedBy   string
}

type SingleMatchSeries struct {
	TeamA string
	TeamB string

	Simulations         int
	WinsA               int
	WinsB               int
	GoalsA              int
	GoalsB              int
	Regulation          int
	RegulationDraws     int
	Penalties           int
	RandomTie           int
	MostRepeatedGoalsA  int
	MostRepeatedGoalsB  int
	MostRepeatedCount   int
	MostRepeatedPercent float64
	ScoreCounts         map[[2]int]int
	TopScores           []ScorelineStat
}

type ScorelineStat struct {
	GoalsA  int
	GoalsB  int
	Count   int
	Percent float64
}

func RunSingleMatch(seed int64, rules []model.LambdaRule, input SingleMatchInput) (SingleMatchResult, error) {
	if strings.TrimSpace(input.TeamA) == "" {
		return SingleMatchResult{}, errors.New("team_a no puede estar vacio")
	}
	if strings.TrimSpace(input.TeamB) == "" {
		return SingleMatchResult{}, errors.New("team_b no puede estar vacio")
	}
	if len(rules) == 0 {
		return SingleMatchResult{}, errors.New("no hay reglas lambda cargadas")
	}
	if input.ShotsA <= 0 {
		input.ShotsA = 5
	}
	if input.ShotsB <= 0 {
		input.ShotsB = 5
	}
	if strings.TrimSpace(input.Tiebreaker) == "" {
		input.Tiebreaker = "penalties"
	}

	rng := rand.New(rand.NewSource(seed))
	lambdaA, err := lookupLambda(rules, input.ShotsA, input.MotivationA)
	if err != nil {
		return SingleMatchResult{}, err
	}
	lambdaB, err := lookupLambda(rules, input.ShotsB, input.MotivationB)
	if err != nil {
		return SingleMatchResult{}, err
	}

	goalsA := poisson.Sample(rng, lambdaA)
	goalsB := poisson.Sample(rng, lambdaB)
	result := SingleMatchResult{
		TeamA:       input.TeamA,
		TeamB:       input.TeamB,
		ShotsA:      input.ShotsA,
		ShotsB:      input.ShotsB,
		MotivationA: input.MotivationA,
		MotivationB: input.MotivationB,
		LambdaA:     lambdaA,
		LambdaB:     lambdaB,
		GoalsA:      goalsA,
		GoalsB:      goalsB,
		DecidedBy:   "tiempo regular",
	}

	switch {
	case goalsA > goalsB:
		result.Winner = input.TeamA
		return result, nil
	case goalsB > goalsA:
		result.Winner = input.TeamB
		return result, nil
	}

	winner, decidedBy, err := resolveKnockoutTie(rng, input.TeamA, input.TeamB, lambdaA, lambdaB, input.Tiebreaker)
	if err != nil {
		return SingleMatchResult{}, err
	}
	result.Winner = winner
	result.DecidedBy = decidedBy
	return result, nil
}

func RunSingleMatchSeries(seed int64, simulations int, rules []model.LambdaRule, input SingleMatchInput) (SingleMatchSeries, error) {
	if simulations <= 0 {
		simulations = 1
	}

	summary := SingleMatchSeries{
		TeamA:       input.TeamA,
		TeamB:       input.TeamB,
		Simulations: simulations,
		ScoreCounts: make(map[[2]int]int),
	}

	for i := 0; i < simulations; i++ {
		result, err := RunSingleMatch(seed+int64(i)*7919, rules, input)
		if err != nil {
			return SingleMatchSeries{}, err
		}
		summary.GoalsA += result.GoalsA
		summary.GoalsB += result.GoalsB
		summary.ScoreCounts[[2]int{result.GoalsA, result.GoalsB}]++
		switch result.Winner {
		case input.TeamA:
			summary.WinsA++
		case input.TeamB:
			summary.WinsB++
		}
		switch strings.ToLower(strings.TrimSpace(result.DecidedBy)) {
		case "tiempo regular":
			summary.Regulation++
		case "penales":
			summary.Penalties++
		case "sorteo":
			summary.RandomTie++
		}
		if result.GoalsA == result.GoalsB {
			summary.RegulationDraws++
		}
	}

	topScores := make([]ScorelineStat, 0, len(summary.ScoreCounts))
	for score, count := range summary.ScoreCounts {
		stat := ScorelineStat{
			GoalsA: score[0],
			GoalsB: score[1],
			Count:  count,
		}
		if summary.Simulations > 0 {
			stat.Percent = float64(count) * 100 / float64(summary.Simulations)
		}
		topScores = append(topScores, stat)
		if count > summary.MostRepeatedCount ||
			(count == summary.MostRepeatedCount && betterScore(score, [2]int{summary.MostRepeatedGoalsA, summary.MostRepeatedGoalsB})) {
			summary.MostRepeatedGoalsA = score[0]
			summary.MostRepeatedGoalsB = score[1]
			summary.MostRepeatedCount = count
		}
	}
	sort.Slice(topScores, func(i, j int) bool {
		if topScores[i].Count != topScores[j].Count {
			return topScores[i].Count > topScores[j].Count
		}
		return betterScore([2]int{topScores[i].GoalsA, topScores[i].GoalsB}, [2]int{topScores[j].GoalsA, topScores[j].GoalsB})
	})
	if len(topScores) > 10 {
		topScores = topScores[:10]
	}
	summary.TopScores = topScores
	if summary.Simulations > 0 {
		summary.MostRepeatedPercent = float64(summary.MostRepeatedCount) * 100 / float64(summary.Simulations)
	}

	return summary, nil
}

func betterScore(candidate, current [2]int) bool {
	candidateTotal := candidate[0] + candidate[1]
	currentTotal := current[0] + current[1]
	if candidateTotal != currentTotal {
		return candidateTotal < currentTotal
	}
	if candidate[0] != current[0] {
		return candidate[0] < current[0]
	}
	return candidate[1] < current[1]
}

func resolveKnockoutTie(rng *rand.Rand, teamA, teamB string, lambdaA, lambdaB float64, tiebreaker string) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(tiebreaker)) {
	case "", "penalties":
		return penaltyShootout(rng, teamA, teamB, lambdaA, lambdaB), "penales", nil
	case "random":
		if rng.Intn(2) == 0 {
			return teamA, "sorteo", nil
		}
		return teamB, "sorteo", nil
	default:
		return "", "", fmt.Errorf("tiebreaker invalido: %q", tiebreaker)
	}
}

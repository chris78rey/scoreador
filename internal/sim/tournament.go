package sim

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"scoreador/internal/model"
	"scoreador/internal/poisson"
)

type qualifier struct {
	Team     string
	Group    string
	Rank     int
	Standing model.GroupStanding
}

type profileAccumulator struct {
	shots       int
	appearances int
	motivation  map[model.Motivation]int
}

func RunMonteCarlo(cfg model.Config, matches []model.MatchInput, rules []model.LambdaRule) (model.TournamentSummary, error) {
	cfg.ApplyDefaults()
	if len(matches) == 0 {
		return model.TournamentSummary{}, errors.New("no hay partidos cargados")
	}
	if len(rules) == 0 {
		return model.TournamentSummary{}, errors.New("no hay reglas lambda cargadas")
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}
	if err := validateInput(cfg, matches); err != nil {
		return model.TournamentSummary{}, err
	}

	_, teamNames := buildProfiles(matches)
	tracker := newTracker(teamNames, cfg.Simulations)

	for sim := 0; sim < cfg.Simulations; sim++ {
		rng := rand.New(rand.NewSource(cfg.Seed + int64(sim)*7919))

		groupTables, err := simulateGroupStage(rng, matches, rules)
		if err != nil {
			return model.TournamentSummary{}, err
		}

		qualifiers, err := selectQualifiers(groupTables, cfg)
		if err != nil {
			return model.TournamentSummary{}, err
		}
		for _, q := range qualifiers {
			tracker.AddQualified(q.Team)
		}

	}

	return tracker.Summary(cfg.Name, cfg.Simulations), nil
}

func validateInput(cfg model.Config, matches []model.MatchInput) error {
	groupNames := map[string]struct{}{}
	teamsByGroup := map[string]map[string]struct{}{}
	for _, match := range matches {
		if strings.EqualFold(match.Stage, "group") {
			if match.Group == "" {
				return fmt.Errorf("partido de fase de grupos sin grupo: match_id=%d", match.MatchID)
			}
			groupNames[match.Group] = struct{}{}
			if teamsByGroup[match.Group] == nil {
				teamsByGroup[match.Group] = map[string]struct{}{}
			}
			teamsByGroup[match.Group][match.TeamA] = struct{}{}
			teamsByGroup[match.Group][match.TeamB] = struct{}{}
		}
	}

	if len(groupNames) == 0 {
		return errors.New("no hay partidos de fase de grupos")
	}
	if cfg.Groups > 0 && len(groupNames) != cfg.Groups {
		return fmt.Errorf("se esperaban %d grupos y se encontraron %d", cfg.Groups, len(groupNames))
	}
	if cfg.TeamsPerGroup > 0 {
		for group, teams := range teamsByGroup {
			if len(teams) != cfg.TeamsPerGroup {
				return fmt.Errorf("el grupo %s tiene %d equipos y se esperaban %d", group, len(teams), cfg.TeamsPerGroup)
			}
		}
	}
	if cfg.QualifiedPerGroup <= 0 {
		return errors.New("qualified_per_group debe ser mayor a 0")
	}
	if cfg.BestThirds < 0 {
		return errors.New("best_thirds no puede ser negativo")
	}
	return nil
}

func buildProfiles(matches []model.MatchInput) (map[string]model.TeamProfile, []string) {
	acc := map[string]*profileAccumulator{}
	teams := map[string]struct{}{}

	add := func(team string, shots int, motivation model.Motivation) {
		teams[team] = struct{}{}
		if acc[team] == nil {
			acc[team] = &profileAccumulator{motivation: map[model.Motivation]int{}}
		}
		acc[team].shots += shots
		acc[team].appearances++
		acc[team].motivation[motivation]++
	}

	for _, match := range matches {
		add(match.TeamA, match.ShotsA, match.MotivationA)
		add(match.TeamB, match.ShotsB, match.MotivationB)
	}

	profiles := make(map[string]model.TeamProfile, len(acc))
	for team, a := range acc {
		avgShots := 5.0
		if a.appearances > 0 {
			avgShots = float64(a.shots) / float64(a.appearances)
		}
		profiles[team] = model.TeamProfile{
			Team:       team,
			AvgShots:   avgShots,
			Motivation: modeMotivation(a.motivation),
		}
	}

	teamNames := make([]string, 0, len(teams))
	for team := range teams {
		teamNames = append(teamNames, team)
	}
	sort.Strings(teamNames)

	return profiles, teamNames
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

func simulateGroupStage(rng *rand.Rand, matches []model.MatchInput, rules []model.LambdaRule) (map[string][]model.GroupStanding, error) {
	groupTables := map[string]map[string]*model.GroupStanding{}

	ensure := func(group, team string) *model.GroupStanding {
		if groupTables[group] == nil {
			groupTables[group] = map[string]*model.GroupStanding{}
		}
		if groupTables[group][team] == nil {
			groupTables[group][team] = &model.GroupStanding{Team: team, Group: group}
		}
		return groupTables[group][team]
	}

	for _, match := range matches {
		if !strings.EqualFold(match.Stage, "group") {
			continue
		}
		lambdaA, err := lookupLambda(rules, match.ShotsA, match.MotivationA)
		if err != nil {
			return nil, err
		}
		lambdaB, err := lookupLambda(rules, match.ShotsB, match.MotivationB)
		if err != nil {
			return nil, err
		}
		goalsA := poisson.Sample(rng, lambdaA)
		goalsB := poisson.Sample(rng, lambdaB)

		standingA := ensure(match.Group, match.TeamA)
		standingB := ensure(match.Group, match.TeamB)
		standingA.Played++
		standingB.Played++
		standingA.GoalsFor += goalsA
		standingA.GoalsAgainst += goalsB
		standingB.GoalsFor += goalsB
		standingB.GoalsAgainst += goalsA

		switch {
		case goalsA > goalsB:
			standingA.Points += 3
		case goalsB > goalsA:
			standingB.Points += 3
		default:
			standingA.Points++
			standingB.Points++
		}
	}

	result := make(map[string][]model.GroupStanding, len(groupTables))
	for group, teams := range groupTables {
		table := make([]model.GroupStanding, 0, len(teams))
		for _, standing := range teams {
			standing.GoalDifference = standing.GoalsFor - standing.GoalsAgainst
			table = append(table, *standing)
		}
		sortStandings(table)
		result[group] = table
	}

	return result, nil
}

func sortStandings(table []model.GroupStanding) {
	sort.SliceStable(table, func(i, j int) bool {
		if table[i].Points != table[j].Points {
			return table[i].Points > table[j].Points
		}
		if table[i].GoalDifference != table[j].GoalDifference {
			return table[i].GoalDifference > table[j].GoalDifference
		}
		if table[i].GoalsFor != table[j].GoalsFor {
			return table[i].GoalsFor > table[j].GoalsFor
		}
		if table[i].GoalsAgainst != table[j].GoalsAgainst {
			return table[i].GoalsAgainst < table[j].GoalsAgainst
		}
		return table[i].Team < table[j].Team
	})
}

func selectQualifiers(groupTables map[string][]model.GroupStanding, cfg model.Config) ([]qualifier, error) {
	groupNames := make([]string, 0, len(groupTables))
	for group := range groupTables {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	qualifiers := make([]qualifier, 0)
	selected := map[string]struct{}{}
	thirdCandidates := make([]qualifier, 0)

	for _, group := range groupNames {
		table := groupTables[group]
		limit := cfg.QualifiedPerGroup
		if limit > len(table) {
			limit = len(table)
		}
		for idx := 0; idx < limit; idx++ {
			team := table[idx]
			qual := qualifier{
				Team:     team.Team,
				Group:    group,
				Rank:     idx + 1,
				Standing: team,
			}
			qualifiers = append(qualifiers, qual)
			selected[team.Team] = struct{}{}
		}
		if cfg.BestThirds > 0 && len(table) >= 3 && cfg.QualifiedPerGroup < 3 {
			third := table[2]
			thirdCandidates = append(thirdCandidates, qualifier{
				Team:     third.Team,
				Group:    group,
				Rank:     3,
				Standing: third,
			})
		}
	}

	if cfg.BestThirds > 0 && cfg.QualifiedPerGroup < 3 {
		sort.SliceStable(thirdCandidates, func(i, j int) bool {
			return compareStandings(thirdCandidates[i].Standing, thirdCandidates[j].Standing)
		})
		for _, candidate := range thirdCandidates {
			if len(qualifiers) >= len(groupTables)*cfg.QualifiedPerGroup+cfg.BestThirds {
				break
			}
			if _, ok := selected[candidate.Team]; ok {
				continue
			}
			qualifiers = append(qualifiers, candidate)
			selected[candidate.Team] = struct{}{}
		}
	}

	sort.SliceStable(qualifiers, func(i, j int) bool {
		if qualifiers[i].Rank != qualifiers[j].Rank {
			return qualifiers[i].Rank < qualifiers[j].Rank
		}
		return compareStandings(qualifiers[i].Standing, qualifiers[j].Standing)
	})

	return qualifiers, nil
}

func compareStandings(a, b model.GroupStanding) bool {
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

func simulateKnockout(rng *rand.Rand, qualifiers []qualifier, profiles map[string]model.TeamProfile, rules []model.LambdaRule, tiebreaker string, tracker *tracker) (string, error) {
	if len(qualifiers) == 0 {
		return "", errors.New("no hay clasificados para la fase eliminatoria")
	}

	participants := make([]string, len(qualifiers))
	for i, q := range qualifiers {
		participants[i] = q.Team
	}

	for len(participants) > 1 {
		roundSize := len(participants)
		tracker.AddRoundBatch(participants, roundSize)

		nextRound := make([]string, 0, (len(participants)+1)/2)
		for i := 0; i < len(participants); i += 2 {
			if i+1 >= len(participants) {
				nextRound = append(nextRound, participants[i])
				continue
			}

			a := participants[i]
			b := participants[i+1]
			winner, err := playKnockoutMatch(rng, a, b, profiles, rules, tiebreaker)
			if err != nil {
				return "", err
			}
			nextRound = append(nextRound, winner)
		}

		participants = nextRound
	}

	return participants[0], nil
}

func playKnockoutMatch(rng *rand.Rand, teamA, teamB string, profiles map[string]model.TeamProfile, rules []model.LambdaRule, tiebreaker string) (string, error) {
	profileA := profileForTeam(teamA, profiles)
	profileB := profileForTeam(teamB, profiles)

	shotsA := int(math.Round(profileA.AvgShots))
	shotsB := int(math.Round(profileB.AvgShots))
	lambdaA, err := lookupLambda(rules, shotsA, profileA.Motivation)
	if err != nil {
		return "", err
	}
	lambdaB, err := lookupLambda(rules, shotsB, profileB.Motivation)
	if err != nil {
		return "", err
	}

	goalsA := poisson.Sample(rng, lambdaA)
	goalsB := poisson.Sample(rng, lambdaB)
	if goalsA > goalsB {
		return teamA, nil
	}
	if goalsB > goalsA {
		return teamB, nil
	}

	switch strings.ToLower(strings.TrimSpace(tiebreaker)) {
	case "", "penalties":
		return penaltyShootout(rng, teamA, teamB, lambdaA, lambdaB), nil
	case "random":
		if rng.Intn(2) == 0 {
			return teamA, nil
		}
		return teamB, nil
	default:
		return "", fmt.Errorf("knockout_tiebreaker invalido: %q", tiebreaker)
	}
}

func profileForTeam(team string, profiles map[string]model.TeamProfile) model.TeamProfile {
	profile, ok := profiles[team]
	if ok {
		return profile
	}
	return model.TeamProfile{
		Team:       team,
		AvgShots:   5,
		Motivation: model.MotivationMedium,
	}
}

func penaltyShootout(rng *rand.Rand, teamA, teamB string, lambdaA, lambdaB float64) string {
	probA := penaltySuccessProbability(lambdaA)
	probB := penaltySuccessProbability(lambdaB)

	scoreA, scoreB := 0, 0
	for i := 0; i < 5; i++ {
		if rng.Float64() < probA {
			scoreA++
		}
		if rng.Float64() < probB {
			scoreB++
		}
	}
	if scoreA != scoreB {
		if scoreA > scoreB {
			return teamA
		}
		return teamB
	}

	for i := 0; i < 20; i++ {
		if rng.Float64() < probA {
			scoreA++
		}
		if rng.Float64() < probB {
			scoreB++
		}
		if scoreA != scoreB {
			if scoreA > scoreB {
				return teamA
			}
			return teamB
		}
	}

	if rng.Intn(2) == 0 {
		return teamA
	}
	return teamB
}

func penaltySuccessProbability(lambda float64) float64 {
	prob := 0.72 + 0.04*(lambda-1.5)
	if prob < 0.55 {
		return 0.55
	}
	if prob > 0.9 {
		return 0.9
	}
	return prob
}

func lookupLambda(rules []model.LambdaRule, shots int, motivation model.Motivation) (float64, error) {
	var (
		bestLambda float64
		bestDelta  = int(^uint(0) >> 1)
		found      bool
	)

	for _, rule := range rules {
		if shots < rule.ShotsMin || shots > rule.ShotsMax {
			continue
		}
		if rule.Motivation == motivation {
			return rule.Lambda, nil
		}
		delta := absInt(int(rule.Motivation) - int(motivation))
		if !found || delta < bestDelta || (delta == bestDelta && rule.Motivation > motivation) {
			bestLambda = rule.Lambda
			bestDelta = delta
			found = true
		}
	}
	if found {
		return bestLambda, nil
	}
	return 0, fmt.Errorf("no se encontro lambda para tiros=%d motivacion=%s", shots, motivation)
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func roundLabel(size int) string {
	switch size {
	case 32:
		return "dieciseisavos"
	case 16:
		return "octavos"
	case 8:
		return "cuartos"
	case 4:
		return "semifinal"
	case 2:
		return "final"
	default:
		return fmt.Sprintf("ronda_%d", size)
	}
}

type tracker struct {
	stats map[string]*model.TeamStats
}

func newTracker(teams []string, simulations int) *tracker {
	stats := make(map[string]*model.TeamStats, len(teams))
	for _, team := range teams {
		stats[team] = &model.TeamStats{
			Team:        team,
			Simulations: simulations,
		}
	}
	return &tracker{stats: stats}
}

func (t *tracker) ensure(team string) *model.TeamStats {
	if entry, ok := t.stats[team]; ok {
		return entry
	}
	entry := &model.TeamStats{Team: team}
	t.stats[team] = entry
	return entry
}

func (t *tracker) AddQualified(team string) {
	t.ensure(team).Qualified++
}

func (t *tracker) AddRoundBatch(teams []string, roundSize int) {
	switch roundLabel(roundSize) {
	case "dieciseisavos":
		for _, team := range teams {
			t.ensure(team).Dieciseisavos++
		}
	case "octavos":
		for _, team := range teams {
			t.ensure(team).Octavos++
		}
	case "cuartos":
		for _, team := range teams {
			t.ensure(team).Cuartos++
		}
	case "semifinal":
		for _, team := range teams {
			t.ensure(team).Semifinal++
		}
	case "final":
		for _, team := range teams {
			t.ensure(team).Final++
		}
	}
}

func (t *tracker) AddChampion(team string) {
	t.ensure(team).Campeon++
}

func (t *tracker) Summary(name string, simulations int) model.TournamentSummary {
	result := make([]model.TeamStats, 0, len(t.stats))
	for _, stats := range t.stats {
		stats.Simulations = simulations
		stats.QualifiedPct = percentage(stats.Qualified, simulations)
		stats.DieciseisavosPct = percentage(stats.Dieciseisavos, simulations)
		stats.OctavosPct = percentage(stats.Octavos, simulations)
		stats.CuartosPct = percentage(stats.Cuartos, simulations)
		stats.SemifinalPct = percentage(stats.Semifinal, simulations)
		stats.FinalPct = percentage(stats.Final, simulations)
		stats.CampeonPct = percentage(stats.Campeon, simulations)
		result = append(result, *stats)
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Campeon != result[j].Campeon {
			return result[i].Campeon > result[j].Campeon
		}
		if result[i].Final != result[j].Final {
			return result[i].Final > result[j].Final
		}
		if result[i].Semifinal != result[j].Semifinal {
			return result[i].Semifinal > result[j].Semifinal
		}
		if result[i].Cuartos != result[j].Cuartos {
			return result[i].Cuartos > result[j].Cuartos
		}
		if result[i].Octavos != result[j].Octavos {
			return result[i].Octavos > result[j].Octavos
		}
		if result[i].Dieciseisavos != result[j].Dieciseisavos {
			return result[i].Dieciseisavos > result[j].Dieciseisavos
		}
		if result[i].Qualified != result[j].Qualified {
			return result[i].Qualified > result[j].Qualified
		}
		return result[i].Team < result[j].Team
	})

	return model.TournamentSummary{
		Name:        name,
		Simulations: simulations,
		Teams:       result,
	}
}

func percentage(count, simulations int) float64 {
	if simulations <= 0 {
		return 0
	}
	return float64(count) * 100 / float64(simulations)
}

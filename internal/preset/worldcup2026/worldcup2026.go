package worldcup2026

import "scoreador/internal/model"

type Group struct {
	Name  string
	Teams []string
}

type Data struct {
	Config  model.Config
	Groups  []Group
	Matches []model.MatchInput
	Rules   []model.LambdaRule
}

func Load(seed int64) Data {
	groups := []Group{
		{Name: "A", Teams: []string{"Mexico", "South Africa", "Korea Republic", "Czechia"}},
		{Name: "B", Teams: []string{"Canada", "Bosnia and Herzegovina", "Qatar", "Switzerland"}},
		{Name: "C", Teams: []string{"Brazil", "Morocco", "Haiti", "Scotland"}},
		{Name: "D", Teams: []string{"USA", "Australia", "Paraguay", "Turkiye"}},
		{Name: "E", Teams: []string{"Germany", "Cote d'Ivoire", "Ecuador", "Curacao"}},
		{Name: "F", Teams: []string{"Netherlands", "Japan", "Sweden", "Tunisia"}},
		{Name: "G", Teams: []string{"Belgium", "Egypt", "IR Iran", "New Zealand"}},
		{Name: "H", Teams: []string{"Spain", "Cabo Verde", "Saudi Arabia", "Uruguay"}},
		{Name: "I", Teams: []string{"France", "Senegal", "Iraq", "Norway"}},
		{Name: "J", Teams: []string{"Argentina", "Algeria", "Austria", "Jordan"}},
		{Name: "K", Teams: []string{"Portugal", "Colombia", "DR Congo", "Uzbekistan"}},
		{Name: "L", Teams: []string{"England", "Croatia", "Panama", "Ghana"}},
	}

	cfg := model.Config{
		Name:               "Mundial 2026",
		Simulations:        1000,
		Groups:             len(groups),
		TeamsPerGroup:      4,
		QualifiedPerGroup:  2,
		BestThirds:         8,
		Knockout:           false,
		KnockoutTiebreaker: "penalties",
		Seed:               seed,
	}

	return Data{
		Config:  cfg,
		Groups:  groups,
		Matches: buildMatches(groups),
		Rules:   defaultLambdaRules(),
	}
}

func buildMatches(groups []Group) []model.MatchInput {
	type strength struct {
		shots      int
		motivation model.Motivation
	}

	teamStrength := func(position int) strength {
		switch position {
		case 0:
			return strength{shots: 9, motivation: model.MotivationHigh}
		case 1:
			return strength{shots: 8, motivation: model.MotivationMedium}
		case 2:
			return strength{shots: 6, motivation: model.MotivationMedium}
		default:
			return strength{shots: 4, motivation: model.MotivationLow}
		}
	}

	clamp := func(value, min, max int) int {
		if value < min {
			return min
		}
		if value > max {
			return max
		}
		return value
	}

	pattern := []int{1, 0, -1, 1, 0, -1}
	pairings := [][2]int{
		{0, 1},
		{2, 3},
		{0, 2},
		{1, 3},
		{0, 3},
		{1, 2},
	}

	matches := make([]model.MatchInput, 0, len(groups)*len(pairings))
	matchID := 1
	for _, group := range groups {
		for index, pairing := range pairings {
			teamA := group.Teams[pairing[0]]
			teamB := group.Teams[pairing[1]]
			strA := teamStrength(pairing[0])
			strB := teamStrength(pairing[1])
			offset := pattern[index]

			matches = append(matches, model.MatchInput{
				MatchID:     matchID,
				Stage:       "group",
				Group:       group.Name,
				TeamA:       teamA,
				TeamB:       teamB,
				ShotsA:      clamp(strA.shots+offset, 1, 12),
				ShotsB:      clamp(strB.shots-offset, 1, 12),
				MotivationA: strA.motivation,
				MotivationB: strB.motivation,
			})
			matchID++
		}
	}

	return matches
}

func defaultLambdaRules() []model.LambdaRule {
	return []model.LambdaRule{
		{ShotsMin: 0, ShotsMax: 1, Motivation: model.MotivationLow, Lambda: 0.2},
		{ShotsMin: 0, ShotsMax: 1, Motivation: model.MotivationMedium, Lambda: 0.4},
		{ShotsMin: 0, ShotsMax: 1, Motivation: model.MotivationHigh, Lambda: 0.7},
		{ShotsMin: 2, ShotsMax: 3, Motivation: model.MotivationLow, Lambda: 0.5},
		{ShotsMin: 2, ShotsMax: 3, Motivation: model.MotivationMedium, Lambda: 0.9},
		{ShotsMin: 2, ShotsMax: 3, Motivation: model.MotivationHigh, Lambda: 1.3},
		{ShotsMin: 4, ShotsMax: 5, Motivation: model.MotivationLow, Lambda: 0.8},
		{ShotsMin: 4, ShotsMax: 5, Motivation: model.MotivationMedium, Lambda: 1.4},
		{ShotsMin: 4, ShotsMax: 5, Motivation: model.MotivationHigh, Lambda: 1.9},
		{ShotsMin: 6, ShotsMax: 7, Motivation: model.MotivationLow, Lambda: 1.1},
		{ShotsMin: 6, ShotsMax: 7, Motivation: model.MotivationMedium, Lambda: 1.8},
		{ShotsMin: 6, ShotsMax: 7, Motivation: model.MotivationHigh, Lambda: 2.4},
		{ShotsMin: 8, ShotsMax: 9, Motivation: model.MotivationLow, Lambda: 1.5},
		{ShotsMin: 8, ShotsMax: 9, Motivation: model.MotivationMedium, Lambda: 2.3},
		{ShotsMin: 8, ShotsMax: 9, Motivation: model.MotivationHigh, Lambda: 3.0},
		{ShotsMin: 10, ShotsMax: 99, Motivation: model.MotivationLow, Lambda: 2.0},
		{ShotsMin: 10, ShotsMax: 99, Motivation: model.MotivationMedium, Lambda: 2.8},
		{ShotsMin: 10, ShotsMax: 99, Motivation: model.MotivationHigh, Lambda: 3.6},
	}
}

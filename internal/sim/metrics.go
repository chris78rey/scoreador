package sim

import "math"

func (e *Engine) estimateShotXG(shooter playerState, assistQuality float64) float64 {
	defendingTeam := 1 - shooter.Team
	keeperIndex := e.teams[defendingTeam].Goalkeeper
	keeper := e.players[keeperIndex]

	goalLine := goalLineFor(shooter.Team)
	goalCenter := Vec2{X: goalLine, Y: PitchWidth * 0.5}
	distance := math.Hypot(goalCenter.X-shooter.X, goalCenter.Y-shooter.Y)
	lateralOffset := math.Abs(shooter.Y - goalCenter.Y)
	distanceScore := clamp(1-distance/32.0, 0, 1)
	angleScore := clamp(1-lateralOffset/18.0, 0, 1)
	spaceScore := e.localSpaceScore(shooter.X, shooter.Y, shooter.Team)
	crowdingPenalty := e.localCrowding(shooter.X, shooter.Y, defendingTeam)
	boxBonus := 0.0
	if distance < 18 {
		boxBonus = 0.12
	}

	raw := -2.1 +
		distanceScore*2.8 +
		angleScore*1.4 +
		spaceScore*0.9 +
		shooter.Finishing*1.3 +
		assistQuality*0.85 +
		boxBonus -
		crowdingPenalty*1.5

	xg := 1.0 / (1.0 + math.Exp(-raw))
	keeperPenalty := clamp(1.0-keeper.GoalkeeperSkill*0.45-keeper.Stamina*0.05, 0.42, 1)
	xg *= keeperPenalty
	return clamp(xg, 0.01, 0.98)
}

func (e *Engine) computeLBS(origin, target playerState) float64 {
	lowY := math.Min(origin.Y, target.Y)
	highY := math.Max(origin.Y, target.Y)
	opponentTeam := 1 - origin.Team

	count := 0.0
	for _, idx := range e.teamPlayers(opponentTeam) {
		defenderY := e.players[idx].Y
		if defenderY > lowY && defenderY <= highY {
			count++
		}
	}
	return count
}

func (e *Engine) computeSGM(origin, target playerState) float64 {
	originCrowding := e.localCrowding(origin.X, origin.Y, origin.Team)
	targetSpace := e.localSpaceScore(target.X, target.Y, target.Team)
	return clamp(targetSpace-originCrowding, -1, 1)
}

func (e *Engine) findInterceptionCandidate(targetTeam int, targetPos Vec2) int {
	opponentTeam := 1 - targetTeam
	start := e.ball.Position
	end := targetPos
	if end.Sub(start).Len() < 1e-6 {
		return -1
	}

	bestIndex := -1
	bestScore := 0.58
	for _, idx := range e.teamPlayers(opponentTeam) {
		player := e.players[idx]
		block := e.teams[player.Team].Blocks[player.Block]
		playerPos := Vec2{X: player.X, Y: player.Y}
		distance, ratio := distanceToSegment(playerPos, start, end)
		if ratio < 0.06 || ratio > 0.96 {
			continue
		}

		corridor := 0.9 + block.InterceptFactor*1.4 + player.Reaction*0.25
		if distance > corridor {
			continue
		}

		pathFactor := 1 - math.Abs(0.5-ratio)*1.25
		score := block.InterceptFactor*0.45 +
			player.Interception*0.3 +
			player.Reaction*0.15 +
			player.Stamina*0.08 +
			pathFactor*0.2 -
			distance*0.18
		if score > bestScore {
			bestScore = score
			bestIndex = idx
		}
	}

	return bestIndex
}

func (e *Engine) updateBlockDynamics() {
	const dt = 0.1

	for team := 0; team < TeamsPerMatch; team++ {
		for block := 0; block < BlocksPerTeam; block++ {
			state := &e.teams[team].Blocks[block]
			cfg := state.Config
			travelled := 0.0
			for _, idx := range state.PlayerList {
				travelled += e.players[idx].LastMovement.Len()
			}

			state.Travelled = travelled

			anchor := state.Anchor
			ballDistance := math.Hypot(anchor.X-e.ball.Position.X, anchor.Y-e.ball.Position.Y)
			ballProximity := clamp(1-ballDistance/32.0, 0, 1)
			opponentCrowding := e.localCrowding(anchor.X, anchor.Y, 1-team)
			pressure := e.config.PressingIntensity*0.18 +
				cfg.PressureAggression*0.36 +
				cfg.LineHeight*0.14 +
				cfg.Coverage*0.08 +
				ballProximity*0.16 +
				opponentCrowding*0.22

			if e.possessor >= 0 {
				carrierTeam := e.players[e.possessor].Team
				if carrierTeam == team {
					pressure *= 0.58
				} else {
					pressure *= 1.18
				}
			}

			state.Pressure = clamp(pressure, 0, 1)

			gamma := 0.035 + e.config.PressingIntensity*0.014 + cfg.PressureAggression*0.012
			mu := 0.008 + e.config.Tempo*0.003 + (1-cfg.Compactness)*0.004
			state.Fatigue = saturate(state.Fatigue + (gamma*math.Pow(state.Pressure, 1.5)+mu*travelled)*dt)
			state.PassingFactor = clamp(1-state.Fatigue*0.42-state.Pressure*0.12+cfg.Coverage*0.06, 0.35, 1)
			state.InterceptFactor = clamp(1-state.Fatigue*0.38+state.Pressure*0.08+cfg.PressureAggression*0.08, 0.35, 1)
		}
	}
}

func distanceToSegment(point, start, end Vec2) (float64, float64) {
	segment := end.Sub(start)
	lengthSquared := segment.X*segment.X + segment.Y*segment.Y
	if lengthSquared < 1e-9 {
		return point.Sub(start).Len(), 0
	}

	t := ((point.X-start.X)*segment.X + (point.Y-start.Y)*segment.Y) / lengthSquared
	t = clamp(t, 0, 1)
	projection := Vec2{X: start.X + segment.X*t, Y: start.Y + segment.Y*t}
	return point.Sub(projection).Len(), t
}

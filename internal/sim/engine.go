package sim

import (
	"math"
	"math/rand"
)

type eventKind int

const (
	eventNone eventKind = iota
	eventPass
	eventShot
)

type inFlightEvent struct {
	kind          eventKind
	originIndex   int
	targetIndex   int
	originBlock   BlockKind
	targetBlock   BlockKind
	targetTeam    int
	accuracy      float64
	xg            float64
	assistQuality float64
}

type Engine struct {
	tick              int
	config            TacticalConfig
	players           [MaxPlayers]playerState
	teams             [TeamsPerMatch]teamState
	ball              ballState
	possessionTeam    int
	possessor         int
	homeXG            float64
	awayXG            float64
	homeLBS           float64
	awayLBS           float64
	homeSGM           float64
	awaySGM           float64
	averageFatigue    float64
	lastPassAccuracy  float64
	lastPassCompleted bool
	lastShotXG        float64
	passMatrix        [BlocksPerTeam][BlocksPerTeam]float64
	inFlight          inFlightEvent
	rng               *rand.Rand
}

func New() *Engine {
	return NewWithSeed(17)
}

func NewWithSeed(seed int64) *Engine {
	e := &Engine{
		config: TacticalConfig{
			Tempo:             0.55,
			PressingIntensity: 0.48,
			BlockHeight:       0.42,
			RiskAppetite:      0.35,
			Blocks:            defaultBlockConfigs(),
		},
		possessionTeam: HomeTeam,
		possessor:      9,
		ball: ballState{
			Position: Vec2{X: PitchLength * 0.5, Y: PitchWidth * 0.5},
			Velocity: Vec2{},
			Spin:     0,
			Height:   0,
			Vertical: 0,
		},
		rng: rand.New(rand.NewSource(seed)),
	}
	e.passMatrix = defaultPassMatrix()
	e.bootstrapTeams()
	e.applyBlockConfigs(e.config.Blocks)
	e.syncBallToPossessor()
	e.updateBlockDynamics()
	e.updateMetrics()
	return e
}

func defaultBlockConfigs() [BlocksPerTeam]BlockTacticalConfig {
	return [BlocksPerTeam]BlockTacticalConfig{
		{
			LineHeight:         0.34,
			Compactness:        0.68,
			PressureAggression: 0.52,
			OffsideTrap:        0.30,
			Coverage:           0.76,
		},
		{
			LineHeight:         0.50,
			Compactness:        0.62,
			PressureAggression: 0.48,
			OffsideTrap:        0.24,
			Coverage:           0.64,
		},
		{
			LineHeight:         0.66,
			Compactness:        0.55,
			PressureAggression: 0.44,
			OffsideTrap:        0.18,
			Coverage:           0.56,
		},
		{
			LineHeight:         0.82,
			Compactness:        0.48,
			PressureAggression: 0.38,
			OffsideTrap:        0.12,
			Coverage:           0.48,
		},
	}
}

func sanitizeBlockConfig(cfg BlockTacticalConfig) BlockTacticalConfig {
	return BlockTacticalConfig{
		LineHeight:         clamp(cfg.LineHeight, 0, 1),
		Compactness:        clamp(cfg.Compactness, 0, 1),
		PressureAggression: clamp(cfg.PressureAggression, 0, 1),
		OffsideTrap:        clamp(cfg.OffsideTrap, 0, 1),
		Coverage:           clamp(cfg.Coverage, 0, 1),
	}
}

func (e *Engine) applyBlockConfigs(configs [BlocksPerTeam]BlockTacticalConfig) {
	for team := 0; team < TeamsPerMatch; team++ {
		for block := 0; block < BlocksPerTeam; block++ {
			e.teams[team].Blocks[block].Config = sanitizeBlockConfig(configs[block])
		}
	}
}

func defaultPassMatrix() [BlocksPerTeam][BlocksPerTeam]float64 {
	return [BlocksPerTeam][BlocksPerTeam]float64{
		{0.72, 0.65, 0.56, 0.42},
		{0.78, 0.74, 0.68, 0.54},
		{0.82, 0.79, 0.76, 0.63},
		{0.86, 0.82, 0.79, 0.73},
	}
}

func (e *Engine) bootstrapTeams() {
	homeFormation := [PlayersPerTeam]Vec2{
		{10, 34}, {18, 12}, {18, 28}, {18, 40}, {18, 56},
		{32, 18}, {34, 34}, {32, 50}, {48, 16}, {50, 34}, {48, 52},
	}
	awayFormation := [PlayersPerTeam]Vec2{
		{95, 34}, {87, 12}, {87, 28}, {87, 40}, {87, 56},
		{73, 18}, {71, 34}, {73, 50}, {57, 16}, {55, 34}, {57, 52},
	}

	teamLayout := [TeamsPerMatch][BlocksPerTeam][]int{
		{
			{0, 1, 2, 3},
			{4, 5, 6},
			{7},
			{8, 9, 10},
		},
		{
			{11, 12, 13, 14},
			{15, 16, 17},
			{18},
			{19, 20, 21},
		},
	}

	for team := 0; team < TeamsPerMatch; team++ {
		e.teams[team] = teamState{}
		e.teams[team].Goalkeeper = team * PlayersPerTeam
		for block := 0; block < BlocksPerTeam; block++ {
			blockState := blockState{
				Team:   team,
				Kind:   BlockKind(block),
				Anchor: Vec2{},
			}
			for _, playerIndex := range teamLayout[team][block] {
				blockState.PlayerList = append(blockState.PlayerList, playerIndex)
				if playerIndex >= team*PlayersPerTeam && playerIndex < (team+1)*PlayersPerTeam {
					blockState.PlayerIndices[playerIndex-team*PlayersPerTeam] = true
				}
			}
			e.teams[team].Blocks[block] = blockState
		}
	}

	for i := 0; i < PlayersPerTeam; i++ {
		e.players[i] = e.newPlayerState(i, HomeTeam, homeFormation[i], e.blockForIndex(i))
		e.players[i+PlayersPerTeam] = e.newPlayerState(i+PlayersPerTeam, AwayTeam, awayFormation[i], e.blockForIndex(i+PlayersPerTeam))
	}

	e.recalculateAnchors()
}

func (e *Engine) newPlayerState(id, team int, home Vec2, block BlockKind) playerState {
	reaction := 0.56 + float64(id%4)*0.06
	passing := 0.64 + float64(id%3)*0.05
	finishing := 0.50 + float64(id%5)*0.04
	interception := 0.58 + float64(id%4)*0.05
	gk := 0.0
	if id%PlayersPerTeam == 0 {
		gk = 0.84
		finishing = 0.14
		passing = 0.30
		interception = 0.22
	}

	return playerState{
		PlayerSnapshot: PlayerSnapshot{
			ID:      id,
			Team:    team,
			X:       home.X,
			Y:       home.Y,
			Speed:   0,
			Stamina: 1,
		},
		Home:            home,
		BaseSpeed:       7.0 + float64(id%3)*0.2,
		Passing:         passing,
		Finishing:       finishing,
		Interception:    interception,
		Reaction:        reaction,
		GoalkeeperSkill: gk,
		Block:           block,
		RelativeIndex:   id % PlayersPerTeam,
	}
}

func (e *Engine) blockForIndex(index int) BlockKind {
	relative := index % PlayersPerTeam
	switch {
	case relative <= 3:
		return BlockDefense
	case relative <= 6:
		return BlockMidfield
	case relative == 7:
		return BlockEnganche
	default:
		return BlockAttack
	}
}

func (e *Engine) Configure(cfg TacticalConfig) {
	e.config = TacticalConfig{
		Tempo:             clamp(cfg.Tempo, 0, 1),
		PressingIntensity: clamp(cfg.PressingIntensity, 0, 1),
		BlockHeight:       clamp(cfg.BlockHeight, 0, 1),
		RiskAppetite:      clamp(cfg.RiskAppetite, 0, 1),
		Blocks:            cfg.Blocks,
	}
	e.applyBlockConfigs(e.config.Blocks)
}

func (e *Engine) SetPassingMatrix(matrix [BlocksPerTeam][BlocksPerTeam]float64) {
	for i := 0; i < BlocksPerTeam; i++ {
		for j := 0; j < BlocksPerTeam; j++ {
			e.passMatrix[i][j] = clamp(matrix[i][j], 0.05, 0.99)
		}
	}
}

func (e *Engine) Config() TacticalConfig {
	return e.config
}

func (e *Engine) PassingMatrix() [BlocksPerTeam][BlocksPerTeam]float64 {
	return e.passMatrix
}

func (e *Engine) RunLaboratory(matches, ticksPerMatch int, seed int64) LaboratorySummary {
	if matches < 1 {
		matches = 1
	}
	if ticksPerMatch < 1 {
		ticksPerMatch = 1
	}

	baseConfig := e.Config()
	baseMatrix := e.PassingMatrix()
	var summary LaboratorySummary
	summary.Matches = matches
	summary.AverageTicks = float64(ticksPerMatch)

	for match := 0; match < matches; match++ {
		matchEngine := NewWithSeed(seed + int64(match)*7919)
		matchEngine.Configure(baseConfig)
		matchEngine.SetPassingMatrix(baseMatrix)

		homeScore := 0
		awayScore := 0
		homeXGTotal := 0.0
		awayXGTotal := 0.0
		homeLBSTotal := 0.0
		awayLBSTotal := 0.0
		homeSGMTotal := 0.0
		awaySGMTotal := 0.0
		fatigueTotal := 0.0

		for tick := 0; tick < ticksPerMatch; tick++ {
			snap := matchEngine.Step()
			homeXGTotal += snap.HomeXG
			awayXGTotal += snap.AwayXG
			homeLBSTotal += snap.HomeLBS
			awayLBSTotal += snap.AwayLBS
			homeSGMTotal += snap.HomeSGM
			awaySGMTotal += snap.AwaySGM
			fatigueTotal += snap.AverageFatigue

			if snap.LastShotXG <= 0 {
				continue
			}

			scoreBias := snap.LastShotXG + (snap.HomeXG-snap.AwayXG)*0.12 + matchEngine.rng.Float64()*0.35
			if snap.PossessionTeam == HomeTeam {
				scoreBias += 0.04
			} else {
				scoreBias -= 0.04
			}
			if scoreBias > 0.72 {
				if snap.PossessionTeam == HomeTeam {
					homeScore++
				} else {
					awayScore++
				}
			}
		}

		summary.AverageHomeXG += homeXGTotal / float64(ticksPerMatch)
		summary.AverageAwayXG += awayXGTotal / float64(ticksPerMatch)
		summary.AverageHomeLBS += homeLBSTotal / float64(ticksPerMatch)
		summary.AverageAwayLBS += awayLBSTotal / float64(ticksPerMatch)
		summary.AverageHomeSGM += homeSGMTotal / float64(ticksPerMatch)
		summary.AverageAwaySGM += awaySGMTotal / float64(ticksPerMatch)
		summary.AverageFatigue += fatigueTotal / float64(ticksPerMatch)

		switch {
		case homeScore > awayScore:
			summary.HomeWins++
		case awayScore > homeScore:
			summary.AwayWins++
		default:
			summary.Draws++
		}
	}

	invMatches := 1.0 / float64(matches)
	summary.HomeWinRate = float64(summary.HomeWins) * invMatches
	summary.AwayWinRate = float64(summary.AwayWins) * invMatches
	summary.DrawRate = float64(summary.Draws) * invMatches
	summary.AverageHomeXG *= invMatches
	summary.AverageAwayXG *= invMatches
	summary.AverageHomeLBS *= invMatches
	summary.AverageAwayLBS *= invMatches
	summary.AverageHomeSGM *= invMatches
	summary.AverageAwaySGM *= invMatches
	summary.AverageFatigue *= invMatches

	return summary
}

func (e *Engine) Step() Snapshot {
	e.tick++
	e.updateBlockAnchors()
	e.stepPlayers()
	e.updateBlockDynamics()
	e.maybeLaunchAction()
	e.stepBall()
	e.resolveLooseBallAndInterceptions()
	e.updateMetrics()
	return e.Snapshot()
}

func (e *Engine) Snapshot() Snapshot {
	var snap Snapshot
	snap.Tick = e.tick
	snap.PossessionTeam = e.possessionTeam
	snap.Possessor = e.possessor
	snap.HomeXG = e.homeXG
	snap.AwayXG = e.awayXG
	snap.HomeLBS = e.homeLBS
	snap.AwayLBS = e.awayLBS
	snap.HomeSGM = e.homeSGM
	snap.AwaySGM = e.awaySGM
	snap.AverageFatigue = e.averageFatigue
	snap.LastPassAccuracy = e.lastPassAccuracy
	snap.LastPassCompleted = e.lastPassCompleted
	snap.LastShotXG = e.lastShotXG
	snap.Ball = BallSnapshot{
		X:    e.ball.Position.X,
		Y:    e.ball.Position.Y,
		VX:   e.ball.Velocity.X,
		VY:   e.ball.Velocity.Y,
		Spin: e.ball.Spin,
	}

	for i := range e.players {
		snap.Players[i] = e.players[i].PlayerSnapshot
	}
	for team := 0; team < TeamsPerMatch; team++ {
		for block := 0; block < BlocksPerTeam; block++ {
			state := e.teams[team].Blocks[block]
			snap.Blocks[team][block] = BlockSnapshot{
				Team:               state.Team,
				Kind:               state.Kind,
				X:                  state.Anchor.X,
				Y:                  state.Anchor.Y,
				Pressure:           state.Pressure,
				Fatigue:            state.Fatigue,
				Travelled:          state.Travelled,
				PassingFactor:      state.PassingFactor,
				InterceptFactor:    state.InterceptFactor,
				LineHeight:         state.Config.LineHeight,
				Compactness:        state.Config.Compactness,
				PressureAggression: state.Config.PressureAggression,
				OffsideTrap:        state.Config.OffsideTrap,
				Coverage:           state.Config.Coverage,
			}
		}
	}
	return snap
}

func (e *Engine) stepPlayers() {
	const dt = 0.1
	ballPos := e.ball.Position

	for i := range e.players {
		player := &e.players[i]
		target := e.playerTarget(i, ballPos)
		delta := target.Sub(Vec2{X: player.X, Y: player.Y})
		distance := delta.Len()

		fatigue := e.blockFatigue(player.Team, player.Block)
		speedPenalty := 1 - fatigue*0.28
		tempoBoost := 1 + e.config.Tempo*0.35
		fatigueFactor := 0.55 + player.Stamina*0.45
		maxSpeed := player.BaseSpeed * tempoBoost * fatigueFactor * speedPenalty

		step := math.Min(maxSpeed*dt, distance)
		movement := Vec2{}
		if distance > 1e-6 {
			movement = delta.Normalize().Scale(step)
			player.X += movement.X
			player.Y += movement.Y
			player.Speed = step / dt
		} else {
			player.Speed = 0
		}

		player.X = clamp(player.X, 0, PitchLength)
		player.Y = clamp(player.Y, 0, PitchWidth)
		player.LastMovement = movement
		player.AccumulatedLoad += movement.Len()
		player.Stamina = clamp(player.Stamina-(0.0008+e.config.Tempo*0.0006+e.config.PressingIntensity*0.0009+fatigue*0.0007), 0.25, 1)
	}
}

func (e *Engine) maybeLaunchAction() {
	if e.inFlight.kind != eventNone || e.possessor < 0 {
		return
	}

	carrier := &e.players[e.possessor]
	teamDirection := e.teamDirection(carrier.Team)
	goalLine := PitchLength
	if carrier.Team == AwayTeam {
		goalLine = 0
	}

	progressToGoal := (carrier.X / PitchLength)
	if carrier.Team == AwayTeam {
		progressToGoal = 1 - progressToGoal
	}
	shouldShoot := progressToGoal > 0.72 && e.config.RiskAppetite > 0.35 && e.rng.Float64() < 0.32+e.config.RiskAppetite*0.25
	if shouldShoot {
		e.launchShot(e.possessor, goalLine)
		return
	}

	passChance := 0.18 + e.config.Tempo*0.3 + e.config.RiskAppetite*0.2 + e.config.PressingIntensity*0.06
	if e.rng.Float64() >= passChance {
		return
	}

	targetTeam := carrier.Team
	originBlock := carrier.Block
	targetBlock := e.nextBlockForPass(originBlock, teamDirection)
	targetIndex, ok := e.pickPassTarget(targetTeam, targetBlock, teamDirection)
	if !ok {
		return
	}

	baseAccuracy := e.passMatrix[originBlock][targetBlock]
	passAccuracy := e.evaluatePassAccuracy(*carrier, e.players[targetIndex], baseAccuracy)
	if passAccuracy < 0.12 {
		return
	}
	e.launchPass(e.possessor, targetIndex, passAccuracy)
}

func (e *Engine) nextBlockForPass(origin BlockKind, direction int) BlockKind {
	switch origin {
	case BlockDefense:
		return BlockMidfield
	case BlockMidfield:
		if direction > 0 {
			return BlockEnganche
		}
		return BlockDefense
	case BlockEnganche:
		if direction > 0 {
			return BlockAttack
		}
		return BlockMidfield
	case BlockAttack:
		if direction > 0 {
			return BlockAttack
		}
		return BlockEnganche
	default:
		return BlockMidfield
	}
}

func (e *Engine) pickPassTarget(team int, block BlockKind, direction int) (int, bool) {
	bestIndex := -1
	bestScore := -1.0
	for _, idx := range e.teamBlockIndices(team, block) {
		player := e.players[idx]
		if idx == e.possessor {
			continue
		}

		progress := float64(direction) * (player.X / PitchLength)
		if team == AwayTeam {
			progress = float64(direction) * (1 - player.X/PitchLength)
		}
		spaceScore := e.localSpaceScore(player.X, player.Y, team)
		supportScore := player.Passing*0.5 + player.Reaction*0.25 + player.Stamina*0.15 + progress*0.4 + spaceScore*0.6
		if supportScore > bestScore {
			bestScore = supportScore
			bestIndex = idx
		}
	}
	if bestIndex < 0 {
		return -1, false
	}
	return bestIndex, true
}

func (e *Engine) evaluatePassAccuracy(origin, target playerState, baseAccuracy float64) float64 {
	distance := math.Hypot(origin.X-target.X, origin.Y-target.Y)
	distancePenalty := clamp(distance/42.0, 0, 1) * 0.35
	originBlock := e.teams[origin.Team].Blocks[origin.Block]
	targetCrowding := e.localCrowding(target.X, target.Y, origin.Team)
	fatiguePenalty := originBlock.Fatigue * 0.32
	tempoBonus := e.config.Tempo * 0.08
	riskPenalty := e.config.RiskAppetite * 0.1
	spaceBonus := e.localSpaceScore(target.X, target.Y, origin.Team) * 0.14
	originFactor := originBlock.PassingFactor
	return clamp(baseAccuracy*originFactor+tempoBonus+spaceBonus-distancePenalty-fatiguePenalty-riskPenalty-targetCrowding*0.12, 0.05, 0.99)
}

func (e *Engine) launchPass(originIndex, targetIndex int, accuracy float64) {
	origin := e.players[originIndex]
	target := e.players[targetIndex]
	originPos := Vec2{X: origin.X, Y: origin.Y}
	targetPos := Vec2{X: target.X, Y: target.Y}
	deviation := clamp((1-accuracy)*(0.06+e.config.RiskAppetite*0.04), 0, 0.12)
	if deviation > 0 {
		side := 1.0
		if e.rng.Float64() < 0.5 {
			side = -1.0
		}
		if targetPos.Y > PitchWidth*0.5 {
			side *= -1
		}
		targetPos.Y = clamp(targetPos.Y+side*PitchWidth*deviation, 1.0, PitchWidth-1.0)
	}
	direction := targetPos.Sub(originPos)
	distance := direction.Len()
	if distance < 0.001 {
		return
	}

	originFactor := e.teams[origin.Team].Blocks[origin.Block].PassingFactor
	speed := lerp(10.0, 22.0, accuracy*originFactor)
	normalized := direction.Normalize()
	curve := clamp((targetPos.Y-origin.Y)/PitchWidth*(0.55+e.config.RiskAppetite*0.2), -0.28, 0.28)
	e.ball.Position = originPos
	e.ball.Velocity = normalized.Scale(speed)
	e.ball.Spin = clamp(curve*4.5, -1.8, 1.8)
	e.ball.Height = 0.15
	e.ball.Vertical = 2.2
	e.possessor = -1
	e.inFlight = inFlightEvent{
		kind:          eventPass,
		originIndex:   originIndex,
		targetIndex:   targetIndex,
		originBlock:   origin.Block,
		targetBlock:   target.Block,
		targetTeam:    target.Team,
		accuracy:      accuracy,
		assistQuality: accuracy,
	}
	e.lastPassAccuracy = accuracy
	e.lastPassCompleted = false
}

func (e *Engine) launchShot(shooterIndex int, goalLine float64) {
	shooter := e.players[shooterIndex]
	goalCenter := Vec2{X: goalLine, Y: PitchWidth * 0.5}
	shootFrom := Vec2{X: shooter.X, Y: shooter.Y}
	direction := goalCenter.Sub(shootFrom)
	distance := direction.Len()
	if distance < 0.001 {
		return
	}

	assistQuality := e.lastPassAccuracy
	xg := e.estimateShotXG(shooter, assistQuality)
	e.homeXG, e.awayXG = e.addXG(shooter.Team, xg)
	e.lastShotXG = xg

	speed := lerp(16.0, 28.0, xg)
	e.ball.Position = shootFrom
	e.ball.Velocity = direction.Normalize().Scale(speed)
	e.ball.Spin = clamp((shooter.Y-goalCenter.Y)*0.03, -5, 5)
	e.ball.Height = 0.45
	e.ball.Vertical = 6.5 + xg*2.5
	e.possessor = -1
	e.inFlight = inFlightEvent{
		kind:          eventShot,
		originIndex:   shooterIndex,
		targetIndex:   -1,
		originBlock:   shooter.Block,
		targetBlock:   BlockAttack,
		targetTeam:    shooter.Team,
		accuracy:      xg,
		xg:            xg,
		assistQuality: assistQuality,
	}
}

func (e *Engine) addXG(team int, xg float64) (float64, float64) {
	if team == HomeTeam {
		return e.homeXG + xg, e.awayXG
	}
	return e.homeXG, e.awayXG + xg
}

func (e *Engine) stepBall() {
	const dt = 0.1

	if e.possessor >= 0 {
		e.syncBallToPossessor()
		e.ball.Velocity = e.ball.Velocity.Scale(0.25)
		e.ball.Spin = clamp(e.ball.Spin+e.config.RiskAppetite*0.05, -4, 4)
		e.ball.Height = 0
		e.ball.Vertical = 0
		return
	}

	e.integrateBall(dt)
}

func (e *Engine) syncBallToPossessor() {
	if e.possessor < 0 || e.possessor >= len(e.players) {
		return
	}
	carrier := e.players[e.possessor]
	offsetX := 1.2
	if carrier.Team == AwayTeam {
		offsetX = -1.2
	}
	e.ball.Position = Vec2{X: clamp(carrier.X+offsetX, 0, PitchLength), Y: clamp(carrier.Y, 0, PitchWidth)}
	e.ball.Velocity = Vec2{X: carrier.Speed * 0.25, Y: 0}
}

func (e *Engine) resolveLooseBallAndInterceptions() {
	if e.possessor >= 0 {
		e.resolvePressureRecovery()
		return
	}

	if e.inFlight.kind == eventPass {
		if e.tryResolvePassReception() {
			return
		}
	}
	if e.inFlight.kind == eventShot {
		e.tryResolveShot()
		return
	}

	e.captureLooseBall()
}

func (e *Engine) resolvePressureRecovery() {
	carrier := e.players[e.possessor]
	opponentTeam := 1 - carrier.Team
	challenger := e.closestPlayerToPoint(carrier.X, carrier.Y, opponentTeam)
	if challenger < 0 {
		return
	}
	challengerBlock := e.players[challenger].Block
	blockPressure := e.blockPressure(opponentTeam, challengerBlock)
	threat := e.localCrowding(carrier.X, carrier.Y, opponentTeam)
	recoveryChance := clamp(blockPressure*0.45+threat*0.4+e.players[challenger].Interception*0.1, 0, 0.95)
	if e.rng.Float64() < recoveryChance && recoveryChance > 0.45 {
		e.transferPossession(challenger)
	}
}

func (e *Engine) tryResolvePassReception() bool {
	if e.inFlight.targetIndex < 0 || e.inFlight.targetIndex >= len(e.players) {
		return false
	}
	target := e.players[e.inFlight.targetIndex]
	targetPos := Vec2{X: target.X, Y: target.Y}
	if e.ball.Position.Sub(targetPos).Len() <= e.controlRadius(target.Team, target.Block, target.Reaction) {
		e.transferPossession(e.inFlight.targetIndex)
		e.lastPassCompleted = true
		lbs := e.computeLBS(e.players[e.inFlight.originIndex], target)
		sgm := e.computeSGM(e.players[e.inFlight.originIndex], target)
		if target.Team == HomeTeam {
			e.homeLBS += lbs
			e.homeSGM += sgm
		} else {
			e.awayLBS += lbs
			e.awaySGM += sgm
		}
		e.inFlight = inFlightEvent{}
		return true
	}

	interceptor := e.findInterceptionCandidate(e.inFlight.targetTeam, e.ball.Position)
	if interceptor >= 0 {
		e.transferPossession(interceptor)
		e.inFlight = inFlightEvent{}
		return true
	}
	return false
}

func (e *Engine) tryResolveShot() {
	attackingTeam := e.inFlight.targetTeam
	defendingTeam := 1 - attackingTeam
	keeper := e.teams[defendingTeam].Goalkeeper
	keeperPos := Vec2{X: e.players[keeper].X, Y: e.players[keeper].Y}
	if math.Abs(e.ball.Position.X-goalLineFor(attackingTeam)) < 2.0 && math.Abs(e.ball.Position.Y-PitchWidth*0.5) < 8.0 {
		saveChance := e.players[keeper].GoalkeeperSkill + e.players[keeper].Stamina*0.1 - e.inFlight.xg*0.45
		if e.rng.Float64() < clamp(saveChance, 0.08, 0.92) {
			e.transferPossession(keeper)
			e.inFlight = inFlightEvent{}
			return
		}
	}

	if e.ball.Position.X < 0 || e.ball.Position.X > PitchLength || e.ball.Position.Y < 0 || e.ball.Position.Y > PitchWidth {
		e.captureLooseBall()
		e.inFlight = inFlightEvent{}
		return
	}

	if e.ball.Position.Sub(keeperPos).Len() <= e.controlRadius(defendingTeam, BlockDefense, e.players[keeper].Reaction) {
		e.transferPossession(keeper)
		e.inFlight = inFlightEvent{}
	}
}

func goalLineFor(team int) float64 {
	if team == HomeTeam {
		return PitchLength
	}
	return 0
}

func (e *Engine) captureLooseBall() {
	bestIndex := -1
	bestScore := -1.0
	for i := range e.players {
		player := e.players[i]
		distance := math.Hypot(player.X-e.ball.Position.X, player.Y-e.ball.Position.Y)
		block := e.teams[player.Team].Blocks[player.Block]
		score := player.Reaction*0.4 + player.Interception*0.3 + player.Stamina*0.15 - distance*0.05
		score += block.InterceptFactor * 0.15
		score += (1 - block.Fatigue) * 0.1
		if score > bestScore && distance <= e.controlRadius(player.Team, player.Block, player.Reaction) {
			bestScore = score
			bestIndex = i
		}
	}
	if bestIndex >= 0 {
		e.transferPossession(bestIndex)
	}
}

func (e *Engine) transferPossession(playerIndex int) {
	if playerIndex < 0 || playerIndex >= len(e.players) {
		return
	}
	e.possessor = playerIndex
	e.possessionTeam = e.players[playerIndex].Team
	e.inFlight = inFlightEvent{}
	e.syncBallToPossessor()
}

func (e *Engine) updateMetrics() {
	e.possessionTeam = e.estimatePossession()

	total := 0.0
	for i := range e.players {
		total += 1 - e.players[i].Stamina
	}
	e.averageFatigue = total / float64(len(e.players))
}

func (e *Engine) estimatePossession() int {
	if e.possessor >= 0 {
		return e.players[e.possessor].Team
	}

	bestTeam := e.possessionTeam
	bestScore := -1.0
	for _, team := range []int{HomeTeam, AwayTeam} {
		score := 0.0
		for _, idx := range e.teamPlayers(team) {
			player := e.players[idx]
			distance := math.Hypot(player.X-e.ball.Position.X, player.Y-e.ball.Position.Y)
			score += 1 / (1 + distance)
			score += player.Stamina * 0.02
			score += player.Reaction * 0.01
		}
		if score > bestScore {
			bestScore = score
			bestTeam = team
		}
	}
	return bestTeam
}

func (e *Engine) teamDirection(team int) int {
	if team == HomeTeam {
		return 1
	}
	return -1
}

func (e *Engine) teamPlayers(team int) []int {
	indices := make([]int, 0, PlayersPerTeam)
	for i := 0; i < PlayersPerTeam; i++ {
		indices = append(indices, team*PlayersPerTeam+i)
	}
	return indices
}

func (e *Engine) teamBlockIndices(team int, block BlockKind) []int {
	return e.teams[team].Blocks[block].PlayerList
}

func directionalBounds(base, backward, forward float64, direction int) (float64, float64) {
	if direction >= 0 {
		return base - backward, base + forward
	}
	return base - forward, base + backward
}

func orderedClamp(minValue, maxValue float64) (float64, float64) {
	if minValue > maxValue {
		return maxValue, minValue
	}
	return minValue, maxValue
}

func (e *Engine) constrainPlayerTarget(player playerState, target, ballPos Vec2) Vec2 {
	blockCfg := e.teams[player.Team].Blocks[player.Block].Config
	direction := e.teamDirection(player.Team)

	if player.GoalkeeperSkill > 0.5 {
		goalX := 4.2
		if player.Team == AwayTeam {
			goalX = PitchLength - 4.2
		}
		minX := goalX - 2.2
		maxX := goalX + 2.4
		minY := PitchWidth*0.5 - 5.5
		maxY := PitchWidth*0.5 + 5.5
		minX, maxX = orderedClamp(clamp(minX, 0, PitchLength), clamp(maxX, 0, PitchLength))
		minY, maxY = orderedClamp(clamp(minY, 0, PitchWidth), clamp(maxY, 0, PitchWidth))
		return Vec2{
			X: clamp(target.X, minX, maxX),
			Y: clamp(target.Y, minY, maxY),
		}
	}

	base := player.Home
	centerY := base.Y
	switch player.Block {
	case BlockDefense:
		centerY = lerp(centerY, ballPos.Y, 0.05+blockCfg.Coverage*0.05)
	case BlockMidfield:
		centerY = lerp(centerY, ballPos.Y, 0.07+blockCfg.Coverage*0.06)
	case BlockEnganche:
		centerY = lerp(centerY, ballPos.Y, 0.1+blockCfg.Coverage*0.08)
	case BlockAttack:
		centerY = lerp(centerY, ballPos.Y, 0.12+blockCfg.Coverage*0.1)
	}

	var backward, forward, lateral float64
	switch player.Block {
	case BlockDefense:
		forward = 2.8 + blockCfg.LineHeight*5.5 + blockCfg.PressureAggression*1.4 + e.config.BlockHeight*1.2
		backward = 4.2 + (1-blockCfg.LineHeight)*6.0 + blockCfg.Coverage*2.4
		lateral = 3.4 + blockCfg.Compactness*3.6 + blockCfg.Coverage*2.8
	case BlockMidfield:
		forward = 4.6 + blockCfg.LineHeight*6.4 + e.config.Tempo*2.0 + blockCfg.PressureAggression*1.0
		backward = 4.8 + (1-blockCfg.LineHeight)*4.8 + blockCfg.Compactness*2.0
		lateral = 5.0 + blockCfg.Compactness*5.4 + blockCfg.Coverage*2.2
	case BlockEnganche:
		forward = 6.2 + e.config.RiskAppetite*8.0 + blockCfg.LineHeight*3.2 + blockCfg.PressureAggression*1.2
		backward = 3.8 + (1-blockCfg.LineHeight)*4.0 + blockCfg.Coverage*1.8
		lateral = 6.2 + blockCfg.Coverage*5.2 + blockCfg.Compactness*2.2
	case BlockAttack:
		forward = 8.4 + e.config.RiskAppetite*9.6 + blockCfg.LineHeight*4.0 + blockCfg.PressureAggression*1.0
		backward = 5.0 + (1-blockCfg.LineHeight)*3.6 + blockCfg.Coverage*1.6
		lateral = 7.0 + blockCfg.Coverage*4.8 + blockCfg.Compactness*3.0
	default:
		forward = 4
		backward = 4
		lateral = 4
	}

	minX, maxX := directionalBounds(base.X, backward, forward, direction)
	minY := centerY - lateral
	maxY := centerY + lateral

	minX, maxX = orderedClamp(clamp(minX, 0, PitchLength), clamp(maxX, 0, PitchLength))
	minY, maxY = orderedClamp(clamp(minY, 0, PitchWidth), clamp(maxY, 0, PitchWidth))

	return Vec2{
		X: clamp(target.X, minX, maxX),
		Y: clamp(target.Y, minY, maxY),
	}
}

func (e *Engine) playerTarget(index int, ballPos Vec2) Vec2 {
	player := e.players[index]
	team := player.Team
	direction := e.teamDirection(team)
	pressure := e.blockPressure(team, player.Block)
	blockState := e.teams[team].Blocks[player.Block]
	blockCfg := blockState.Config
	tempoShift := e.config.Tempo * 0.25
	blockHeightShift := e.config.BlockHeight * 0.18
	lineAnchor := lerp(0.16, 0.88, blockCfg.LineHeight)
	compactness := lerp(0.55, 1.25, blockCfg.Compactness)
	pressureBias := 0.6 + blockCfg.PressureAggression*0.7
	coverageBias := 0.7 + blockCfg.Coverage*0.55

	switch player.Block {
	case BlockDefense:
		base := player.Home
		base.X = lerp(base.X, lineAnchor*PitchLength, 0.14)
		base.X += float64(direction) * (pressure*3 - 1.5 + blockCfg.OffsideTrap*2.2)
		base.Y += math.Sin(float64(e.tick)/10.0+float64(index)) * (1.2 * compactness)
		base.Y = lerp(base.Y, ballPos.Y, blockCfg.Coverage*0.12)
		if team == e.possessionTeam {
			base.X += float64(direction) * (tempoShift + blockHeightShift + pressureBias)
		}
		return e.constrainPlayerTarget(player, base, ballPos)
	case BlockMidfield:
		base := player.Home
		base.X += float64(direction) * (pressure*4 + blockHeightShift*1.2 + blockCfg.Compactness)
		base.Y += math.Cos(float64(e.tick)/12.0+float64(index)) * (1.5 * compactness)
		if team == e.possessionTeam {
			base.X = lerp(base.X, ballPos.X-float64(direction)*8, 0.25)
			base.Y = lerp(base.Y, ballPos.Y, 0.2+blockCfg.Coverage*0.08)
		}
		return e.constrainPlayerTarget(player, base, ballPos)
	case BlockEnganche:
		base := player.Home
		base.X += float64(direction) * (8 + e.config.RiskAppetite*8 + pressureBias)
		base.Y = lerp(base.Y, ballPos.Y, 0.3+blockCfg.Coverage*0.12)
		return e.constrainPlayerTarget(player, base, ballPos)
	case BlockAttack:
		base := player.Home
		base.X += float64(direction) * (12 + e.config.RiskAppetite*10 + pressureBias)
		base.Y += math.Sin(float64(e.tick)/8.0+float64(index)) * (1.8 * compactness)
		if team == e.possessionTeam {
			base.X = lerp(base.X, ballPos.X+float64(direction)*6, 0.35)
			base.Y = lerp(base.Y, ballPos.Y, (0.25+blockCfg.Coverage*0.05)*coverageBias)
		}
		return e.constrainPlayerTarget(player, base, ballPos)
	default:
		return e.constrainPlayerTarget(player, player.Home, ballPos)
	}
}

func (e *Engine) controlRadius(team int, block BlockKind, reaction float64) float64 {
	blockCfg := e.teams[team].Blocks[block].Config
	fatigue := e.blockFatigue(team, block)
	return clamp(
		1.9+
			reaction*0.95+
			blockCfg.Coverage*1.15+
			blockCfg.Compactness*0.55-
			fatigue*0.35,
		1.25,
		5.4,
	)
}

func (e *Engine) localSpaceScore(x, y float64, team int) float64 {
	opponent := 1 - team
	return 1 - e.localCrowding(x, y, opponent)
}

func (e *Engine) localCrowding(x, y float64, opponentTeam int) float64 {
	count := 0.0
	for _, idx := range e.teamPlayers(opponentTeam) {
		player := e.players[idx]
		distance := math.Hypot(player.X-x, player.Y-y)
		if distance < 12 {
			count += 1 - distance/12
		}
	}
	return clamp(count/8.0, 0, 1)
}

func (e *Engine) closestPlayerToPoint(x, y float64, team int) int {
	best := -1
	bestDistance := math.MaxFloat64
	for _, idx := range e.teamPlayers(team) {
		player := e.players[idx]
		distance := math.Hypot(player.X-x, player.Y-y)
		if distance < bestDistance {
			bestDistance = distance
			best = idx
		}
	}
	return best
}

func (e *Engine) teamClosestToPlayer(playerIndex int) int {
	if e.players[playerIndex].Team == HomeTeam {
		return AwayTeam
	}
	return HomeTeam
}

func (e *Engine) blockPressure(team int, block BlockKind) float64 {
	state := e.teams[team].Blocks[block]
	pressure := state.Pressure
	if pressure <= 0 {
		pressure = e.config.PressingIntensity*0.25 +
			state.Config.PressureAggression*0.35 +
			state.Config.LineHeight*0.2 +
			state.Fatigue*0.25
	}
	return clamp(pressure, 0, 1)
}

func (e *Engine) blockFatigue(team int, block BlockKind) float64 {
	return e.teams[team].Blocks[block].Fatigue
}

func (e *Engine) recalculateAnchors() {
	for team := 0; team < TeamsPerMatch; team++ {
		for block := 0; block < BlocksPerTeam; block++ {
			state := &e.teams[team].Blocks[block]
			sum := Vec2{}
			count := 0.0
			for _, idx := range state.PlayerList {
				sum = sum.Add(e.players[idx].Home)
				count++
			}
			if count > 0 {
				state.Anchor = Vec2{X: sum.X / count, Y: sum.Y / count}
			}
		}
	}
}

func (e *Engine) updateBlockAnchors() {
	for team := 0; team < TeamsPerMatch; team++ {
		for block := 0; block < BlocksPerTeam; block++ {
			state := &e.teams[team].Blocks[block]
			sum := Vec2{}
			count := 0.0
			for _, idx := range state.PlayerList {
				player := e.players[idx]
				sum = sum.Add(Vec2{X: player.X, Y: player.Y})
				count++
			}
			if count > 0 {
				state.Anchor = Vec2{X: sum.X / count, Y: sum.Y / count}
			}
		}
	}
}

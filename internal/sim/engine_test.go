package sim

import (
	"math"
	"testing"
)

func TestNewEngineInitialState(t *testing.T) {
	e := New()
	snap := e.Snapshot()

	if snap.Tick != 0 {
		t.Fatalf("expected initial tick 0, got %d", snap.Tick)
	}
	if len(snap.Players) != MaxPlayers {
		t.Fatalf("expected %d players, got %d", MaxPlayers, len(snap.Players))
	}
	if snap.PossessionTeam != HomeTeam {
		t.Fatalf("expected home possession, got %d", snap.PossessionTeam)
	}
	if snap.Ball.X < 0 || snap.Ball.X > PitchLength || snap.Ball.Y < 0 || snap.Ball.Y > PitchWidth {
		t.Fatalf("ball out of pitch bounds: %+v", snap.Ball)
	}
}

func TestStepAdvancesTickAndMetrics(t *testing.T) {
	e := New()
	before := e.Snapshot()
	after := e.Step()

	if after.Tick != before.Tick+1 {
		t.Fatalf("expected tick %d, got %d", before.Tick+1, after.Tick)
	}
	if after.AverageFatigue < 0 || after.AverageFatigue > 1 {
		t.Fatalf("average fatigue out of range: %v", after.AverageFatigue)
	}
	if after.HomeXG < 0 || after.AwayXG < 0 {
		t.Fatalf("xG should never be negative: home=%v away=%v", after.HomeXG, after.AwayXG)
	}
}

func TestComputeLBSCountsDefendersInVerticalBand(t *testing.T) {
	e := New()
	origin := e.players[9]
	target := e.players[10]

	origin.Y = 10
	target.Y = 45

	for _, idx := range e.teamPlayers(AwayTeam) {
		e.players[idx].Y = 60
	}
	e.players[11].Y = 12
	e.players[12].Y = 20
	e.players[13].Y = 41

	lbs := e.computeLBS(origin, target)
	if lbs != 3 {
		t.Fatalf("expected 3 defenders in band, got %v", lbs)
	}
}

func TestBlockDynamicsAppliesFatigueAndFactors(t *testing.T) {
	e := New()
	for _, idx := range e.teams[HomeTeam].Blocks[BlockMidfield].PlayerList {
		e.players[idx].LastMovement = Vec2{X: 2, Y: 0}
	}

	e.updateBlockDynamics()
	state := e.teams[HomeTeam].Blocks[BlockMidfield]

	if state.Travelled <= 0 {
		t.Fatalf("expected travelled distance to increase, got %v", state.Travelled)
	}
	if state.Fatigue <= 0 {
		t.Fatalf("expected fatigue to increase, got %v", state.Fatigue)
	}
	if state.PassingFactor > 1 || state.InterceptFactor > 1 {
		t.Fatalf("block factors out of range: passing=%v intercept=%v", state.PassingFactor, state.InterceptFactor)
	}
}

func TestPlayerTargetsStayInsideBlockLanes(t *testing.T) {
	e := New()
	ballPos := Vec2{X: PitchLength, Y: PitchWidth}

	cases := []struct {
		name string
		idx  int
		minX float64
		maxX float64
	}{
		{name: "home_defender", idx: 1, minX: 8, maxX: 28},
		{name: "home_midfielder", idx: 5, minX: 22, maxX: 44},
		{name: "home_enganche", idx: 7, minX: 38, maxX: 66},
		{name: "home_attacker", idx: 9, minX: 40, maxX: 75},
		{name: "away_defender", idx: 12, minX: 78, maxX: 97},
		{name: "away_attacker", idx: 20, minX: 35, maxX: 65},
	}

	for _, tc := range cases {
		target := e.playerTarget(tc.idx, ballPos)
		if target.X < tc.minX || target.X > tc.maxX {
			t.Fatalf("%s target out of lane: x=%v not in [%v,%v]", tc.name, target.X, tc.minX, tc.maxX)
		}
		if target.Y < 0 || target.Y > PitchWidth {
			t.Fatalf("%s target out of pitch on y-axis: %v", tc.name, target.Y)
		}
	}
}

func TestBallPhysicsCapsAccelerationAndSpeed(t *testing.T) {
	e := New()

	acceleration := e.ballAcceleration(Vec2{X: 400, Y: 240})
	if got := acceleration.Len(); got > ballMaxAcceleration+1e-9 {
		t.Fatalf("ball acceleration exceeded cap: %v > %v", got, ballMaxAcceleration)
	}

	e.ball.Position = Vec2{X: PitchLength * 0.5, Y: PitchWidth * 0.5}
	e.ball.Velocity = Vec2{X: 120, Y: 80}
	e.ball.Spin = 10
	e.integrateBall(0.1)

	if got := e.ball.Velocity.Len(); got > ballMaxSpeed+1e-9 {
		t.Fatalf("ball speed exceeded cap: %v > %v", got, ballMaxSpeed)
	}
	if math.IsNaN(e.ball.Position.X) || math.IsNaN(e.ball.Position.Y) {
		t.Fatalf("ball position became invalid: %+v", e.ball.Position)
	}
}

package sim

import "math"

const (
	PitchLength = 105.0
	PitchWidth  = 68.0
	MaxPlayers  = 22

	PlayersPerTeam = 11
	TeamsPerMatch  = 2
	BlocksPerTeam  = 4

	HomeTeam = 0
	AwayTeam = 1

	DefensiveBlock = 0
	MidfieldBlock  = 1
	EngancheBlock  = 2
	AttackingBlock = 3
)

type Vec2 struct {
	X float64
	Y float64
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vec2) Scale(factor float64) Vec2 {
	return Vec2{X: v.X * factor, Y: v.Y * factor}
}

func (v Vec2) Len() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vec2) Normalize() Vec2 {
	length := v.Len()
	if length < 1e-9 {
		return Vec2{}
	}
	return Vec2{X: v.X / length, Y: v.Y / length}
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func saturate(value float64) float64 {
	return clamp(value, 0, 1)
}

type BlockKind int

const (
	BlockDefense BlockKind = iota
	BlockMidfield
	BlockEnganche
	BlockAttack
)

func (k BlockKind) String() string {
	switch k {
	case BlockDefense:
		return "defense"
	case BlockMidfield:
		return "midfield"
	case BlockEnganche:
		return "enganche"
	case BlockAttack:
		return "attack"
	default:
		return "unknown"
	}
}

type TacticalConfig struct {
	Tempo             float64
	PressingIntensity float64
	BlockHeight       float64
	RiskAppetite      float64
	Blocks            [BlocksPerTeam]BlockTacticalConfig
}

type PlayerSnapshot struct {
	ID      int
	Team    int
	X       float64
	Y       float64
	Speed   float64
	Stamina float64
}

type BallSnapshot struct {
	X    float64
	Y    float64
	VX   float64
	VY   float64
	Spin float64
}

type BlockSnapshot struct {
	Team               int
	Kind               BlockKind
	X                  float64
	Y                  float64
	Pressure           float64
	Fatigue            float64
	Travelled          float64
	PassingFactor      float64
	InterceptFactor    float64
	LineHeight         float64
	Compactness        float64
	PressureAggression float64
	OffsideTrap        float64
	Coverage           float64
}

type BlockTacticalConfig struct {
	LineHeight         float64
	Compactness        float64
	PressureAggression float64
	OffsideTrap        float64
	Coverage           float64
}

type Snapshot struct {
	Tick              int
	PossessionTeam    int
	Possessor         int
	HomeXG            float64
	AwayXG            float64
	HomeLBS           float64
	AwayLBS           float64
	HomeSGM           float64
	AwaySGM           float64
	AverageFatigue    float64
	LastPassAccuracy  float64
	LastPassCompleted bool
	LastShotXG        float64
	Ball              BallSnapshot
	Players           [MaxPlayers]PlayerSnapshot
	Blocks            [TeamsPerMatch][BlocksPerTeam]BlockSnapshot
}

type LaboratorySummary struct {
	Matches          int
	HomeWins         int
	AwayWins         int
	Draws            int
	HomeWinRate      float64
	AwayWinRate      float64
	DrawRate         float64
	AverageHomeXG    float64
	AverageAwayXG    float64
	AverageHomeLBS   float64
	AverageAwayLBS   float64
	AverageHomeSGM   float64
	AverageAwaySGM   float64
	AverageFatigue   float64
	AverageTicks     float64
}

type playerState struct {
	PlayerSnapshot
	Home            Vec2
	BaseSpeed       float64
	Passing         float64
	Finishing       float64
	Interception    float64
	Reaction        float64
	GoalkeeperSkill float64
	Block           BlockKind
	RelativeIndex   int
	LastMovement    Vec2
	AccumulatedLoad float64
}

type blockState struct {
	Team            int
	Kind            BlockKind
	Config          BlockTacticalConfig
	PlayerIndices   [PlayersPerTeam]bool
	PlayerList      []int
	Pressure        float64
	Fatigue         float64
	Travelled       float64
	PassingFactor   float64
	InterceptFactor float64
	Anchor          Vec2
}

type teamState struct {
	Blocks     [BlocksPerTeam]blockState
	Goalkeeper int
}

type ballState struct {
	Position Vec2
	Velocity Vec2
	Spin     float64
	Height   float64
	Vertical float64
}

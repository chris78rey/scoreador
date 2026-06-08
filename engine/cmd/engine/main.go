package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct PlayerState {
    int32_t id;
    int32_t team;
    double x;
    double y;
    double speed;
    double stamina;
} PlayerState;

typedef struct BallState {
    double x;
    double y;
    double vx;
    double vy;
    double spin;
} BallState;

typedef struct TacticalConfig {
    double tempo;
    double pressing_intensity;
    double block_height;
    double risk_appetite;
    struct BlockTacticalConfig* blocks;
} TacticalConfig;

typedef struct BlockTacticalConfig {
    double line_height;
    double compactness;
    double pressure_aggression;
    double offside_trap;
    double coverage;
} BlockTacticalConfig;

typedef struct SimSnapshot {
    int32_t tick;
    int32_t possession_team;
    double home_xg;
    double away_xg;
    double average_fatigue;
    BallState ball;
    PlayerState players[22];
} SimSnapshot;

typedef struct EngineFrame {
    int32_t tick;
    int32_t possession_team;
    int32_t possessor;
    int32_t last_pass_completed;
    double home_xg;
    double away_xg;
    double home_lbs;
    double away_lbs;
    double home_sgm;
    double away_sgm;
    double average_fatigue;
    double last_pass_accuracy;
    double last_shot_xg;
    double ball_x;
    double ball_y;
    double ball_vx;
    double ball_vy;
    double ball_spin;
    double player_x[22];
    double player_y[22];
    double player_speed[22];
    double player_stamina[22];
    struct BlockFrame* blocks;
} EngineFrame;

typedef struct BlockFrame {
    int32_t team;
    int32_t kind;
    double x;
    double y;
    double pressure;
    double fatigue;
    double travelled;
    double passing_factor;
    double intercept_factor;
    double line_height;
    double compactness;
    double pressure_aggression;
    double offside_trap;
    double coverage;
} BlockFrame;

typedef struct EngineSession {
    uint64_t handle;
    EngineFrame* frame;
    double playback_rate;
    double accumulator_seconds;
} EngineSession;

typedef struct LaboratorySummary {
    int32_t matches;
    int32_t home_wins;
    int32_t away_wins;
    int32_t draws;
    double home_win_rate;
    double away_win_rate;
    double draw_rate;
    double average_home_xg;
    double average_away_xg;
    double average_home_lbs;
    double average_away_lbs;
    double average_home_sgm;
    double average_away_sgm;
    double average_fatigue;
    double average_ticks;
} LaboratorySummary;
*/
import "C"

import (
	"sync"
	"unsafe"

	"scoreador/engine/internal/sim"
)

var (
	registryMu sync.RWMutex
	registry   = map[uint64]*sim.Engine{}
	nextHandle uint64 = 1

	publishingMu sync.RWMutex
	publishing   = map[uint64]bool{}
)

const blockCount = sim.TeamsPerMatch * sim.BlocksPerTeam

func main() {}

func registerEngine(engine *sim.Engine) uint64 {
	registryMu.Lock()
	defer registryMu.Unlock()

	handle := nextHandle
	nextHandle++
	registry[handle] = engine
	return handle
}

func removeEngine(handle uint64) {
	registryMu.Lock()
	delete(registry, handle)
	registryMu.Unlock()

	publishingMu.Lock()
	delete(publishing, handle)
	publishingMu.Unlock()
}

func setPublishing(handle uint64, enabled bool) {
	publishingMu.Lock()
	publishing[handle] = enabled
	publishingMu.Unlock()
}

func isPublishing(handle uint64) bool {
	publishingMu.RLock()
	defer publishingMu.RUnlock()
	return publishing[handle]
}

func refreshFrame(frame *C.struct_EngineFrame, snapshot sim.Snapshot) {
	if frame == nil {
		return
	}

	frame.tick = C.int(snapshot.Tick)
	frame.possession_team = C.int(snapshot.PossessionTeam)
	frame.possessor = C.int(snapshot.Possessor)
	if snapshot.LastPassCompleted {
		frame.last_pass_completed = 1
	} else {
		frame.last_pass_completed = 0
	}
	frame.home_xg = C.double(snapshot.HomeXG)
	frame.away_xg = C.double(snapshot.AwayXG)
	frame.home_lbs = C.double(snapshot.HomeLBS)
	frame.away_lbs = C.double(snapshot.AwayLBS)
	frame.home_sgm = C.double(snapshot.HomeSGM)
	frame.away_sgm = C.double(snapshot.AwaySGM)
	frame.average_fatigue = C.double(snapshot.AverageFatigue)
	frame.last_pass_accuracy = C.double(snapshot.LastPassAccuracy)
	frame.last_shot_xg = C.double(snapshot.LastShotXG)
	frame.ball_x = C.double(snapshot.Ball.X)
	frame.ball_y = C.double(snapshot.Ball.Y)
	frame.ball_vx = C.double(snapshot.Ball.VX)
	frame.ball_vy = C.double(snapshot.Ball.VY)
	frame.ball_spin = C.double(snapshot.Ball.Spin)

	for i := 0; i < sim.MaxPlayers; i++ {
		player := snapshot.Players[i]
		frame.player_x[i] = C.double(player.X)
		frame.player_y[i] = C.double(player.Y)
		frame.player_speed[i] = C.double(player.Speed)
		frame.player_stamina[i] = C.double(player.Stamina)
	}

	if frame.blocks != nil {
		blocks := (*[blockCount]C.struct_BlockFrame)(unsafe.Pointer(frame.blocks))
		index := 0
		for team := 0; team < sim.TeamsPerMatch; team++ {
			for block := 0; block < sim.BlocksPerTeam; block++ {
				state := snapshot.Blocks[team][block]
				blocks[index].team = C.int32_t(state.Team)
				blocks[index].kind = C.int32_t(state.Kind)
				blocks[index].x = C.double(state.X)
				blocks[index].y = C.double(state.Y)
				blocks[index].pressure = C.double(state.Pressure)
				blocks[index].fatigue = C.double(state.Fatigue)
				blocks[index].travelled = C.double(state.Travelled)
				blocks[index].passing_factor = C.double(state.PassingFactor)
				blocks[index].intercept_factor = C.double(state.InterceptFactor)
				blocks[index].line_height = C.double(state.LineHeight)
				blocks[index].compactness = C.double(state.Compactness)
				blocks[index].pressure_aggression = C.double(state.PressureAggression)
				blocks[index].offside_trap = C.double(state.OffsideTrap)
				blocks[index].coverage = C.double(state.Coverage)
				index++
			}
		}
	}
}

func refreshSessionFrame(session *C.struct_EngineSession) {
	if session == nil || session.frame == nil {
		return
	}

	engine := lookup(uint64(session.handle))
	if engine == nil {
		return
	}

	refreshFrame(session.frame, engine.Snapshot())
}

func shouldPublish(session *C.struct_EngineSession) bool {
	if session == nil {
		return false
	}
	return isPublishing(uint64(session.handle))
}

func readTacticalConfig(cfg *C.struct_TacticalConfig) sim.TacticalConfig {
	out := sim.TacticalConfig{
		Tempo:             0.55,
		PressingIntensity: 0.48,
		BlockHeight:       0.42,
		RiskAppetite:      0.35,
		Blocks:            [sim.BlocksPerTeam]sim.BlockTacticalConfig{},
	}
	if cfg == nil {
		return out
	}

	out.Tempo = float64(cfg.tempo)
	out.PressingIntensity = float64(cfg.pressing_intensity)
	out.BlockHeight = float64(cfg.block_height)
	out.RiskAppetite = float64(cfg.risk_appetite)

	defaultBlocks := [sim.BlocksPerTeam]sim.BlockTacticalConfig{
		{LineHeight: 0.34, Compactness: 0.68, PressureAggression: 0.52, OffsideTrap: 0.30, Coverage: 0.76},
		{LineHeight: 0.50, Compactness: 0.62, PressureAggression: 0.48, OffsideTrap: 0.24, Coverage: 0.64},
		{LineHeight: 0.66, Compactness: 0.55, PressureAggression: 0.44, OffsideTrap: 0.18, Coverage: 0.56},
		{LineHeight: 0.82, Compactness: 0.48, PressureAggression: 0.38, OffsideTrap: 0.12, Coverage: 0.48},
	}
	out.Blocks = defaultBlocks
	if cfg.blocks == nil {
		return out
	}

	blockCfgs := (*[sim.BlocksPerTeam]C.struct_BlockTacticalConfig)(unsafe.Pointer(cfg.blocks))
	for i := 0; i < sim.BlocksPerTeam; i++ {
		out.Blocks[i] = sim.BlockTacticalConfig{
			LineHeight:         float64(blockCfgs[i].line_height),
			Compactness:        float64(blockCfgs[i].compactness),
			PressureAggression: float64(blockCfgs[i].pressure_aggression),
			OffsideTrap:        float64(blockCfgs[i].offside_trap),
			Coverage:           float64(blockCfgs[i].coverage),
		}
	}
	return out
}

func clampPlaybackRate(rate float64) float64 {
	if rate < 0.25 {
		return 0.25
	}
	if rate > 10.0 {
		return 10.0
	}
	return rate
}

//export EngineCreate
func EngineCreate() C.uint64_t {
	engine := sim.New()
	handle := registerEngine(engine)
	return C.uint64_t(handle)
}

//export EngineDestroy
func EngineDestroy(handle C.uint64_t) {
	removeEngine(uint64(handle))
}

//export EngineSetConfig
func EngineSetConfig(handle C.uint64_t, cfg *C.struct_TacticalConfig) C.int {
	engine := lookup(uint64(handle))
	if engine == nil || cfg == nil {
		return 0
	}

	engine.Configure(readTacticalConfig(cfg))
	return 1
}

//export EngineStep
func EngineStep(handle C.uint64_t, ticks C.int) C.int {
	engine := lookup(uint64(handle))
	if engine == nil {
		return 0
	}

	steps := int(ticks)
	if steps < 1 {
		steps = 1
	}
	for i := 0; i < steps; i++ {
		engine.Step()
	}
	return 1
}

//export EngineCreateSession
func EngineCreateSession() *C.struct_EngineSession {
	engine := sim.New()
	handle := registerEngine(engine)

	session := (*C.struct_EngineSession)(C.calloc(1, C.size_t(C.sizeof_EngineSession)))
	if session == nil {
		removeEngine(handle)
		return nil
	}

	frame := (*C.struct_EngineFrame)(C.calloc(1, C.size_t(C.sizeof_EngineFrame)))
	if frame == nil {
		removeEngine(handle)
		C.free(unsafe.Pointer(session))
		return nil
	}

	blocks := (*C.struct_BlockFrame)(C.calloc(C.size_t(blockCount), C.size_t(C.sizeof_BlockFrame)))
	if blocks == nil {
		removeEngine(handle)
		C.free(unsafe.Pointer(frame))
		C.free(unsafe.Pointer(session))
		return nil
	}

	session.handle = C.uint64_t(handle)
	session.frame = frame
	session.playback_rate = 1.0
	session.accumulator_seconds = 0.0
	frame.blocks = blocks
	setPublishing(handle, true)
	refreshSessionFrame(session)
	return session
}

//export EngineDestroySession
func EngineDestroySession(session *C.struct_EngineSession) {
	if session == nil {
		return
	}

	handle := uint64(session.handle)
	if session.frame != nil && session.frame.blocks != nil {
		C.free(unsafe.Pointer(session.frame.blocks))
		session.frame.blocks = nil
	}
	if session.frame != nil {
		C.free(unsafe.Pointer(session.frame))
		session.frame = nil
	}
	C.free(unsafe.Pointer(session))
	removeEngine(handle)
}

//export EngineSetFramePublishingSession
func EngineSetFramePublishingSession(session *C.struct_EngineSession, enabled C.int) C.int {
	if session == nil {
		return 0
	}

	setPublishing(uint64(session.handle), enabled != 0)
	return 1
}

//export EngineSetConfigSession
func EngineSetConfigSession(session *C.struct_EngineSession, cfg *C.struct_TacticalConfig) C.int {
	if session == nil || cfg == nil {
		return 0
	}

	engine := lookup(uint64(session.handle))
	if engine == nil {
		return 0
	}

	engine.Configure(readTacticalConfig(cfg))
	if shouldPublish(session) {
		refreshSessionFrame(session)
	}
	return 1
}

//export EngineStepSession
func EngineStepSession(session *C.struct_EngineSession, ticks C.int) C.int {
	if session == nil {
		return 0
	}

	engine := lookup(uint64(session.handle))
	if engine == nil {
		return 0
	}

	steps := int(ticks)
	if steps < 1 {
		steps = 1
	}
	for i := 0; i < steps; i++ {
		engine.Step()
	}
	if shouldPublish(session) {
		refreshSessionFrame(session)
	}
	return 1
}

//export EngineSetPlaybackRateSession
func EngineSetPlaybackRateSession(session *C.struct_EngineSession, rate C.double) C.int {
	if session == nil {
		return 0
	}

	session.playback_rate = C.double(clampPlaybackRate(float64(rate)))
	return 1
}

//export EngineAdvanceSession
func EngineAdvanceSession(session *C.struct_EngineSession, elapsedSeconds C.double) C.int {
	if session == nil {
		return 0
	}

	engine := lookup(uint64(session.handle))
	if engine == nil {
		return 0
	}

	delta := float64(elapsedSeconds)
	if delta <= 0 {
		if shouldPublish(session) {
			refreshSessionFrame(session)
		}
		return 1
	}

	session.accumulator_seconds += C.double(delta * float64(session.playback_rate))
	const tickSeconds = 0.1
	steps := 0
	for session.accumulator_seconds >= tickSeconds {
		engine.Step()
		session.accumulator_seconds -= C.double(tickSeconds)
		steps++
	}
	if steps == 0 {
		refreshSessionFrame(session)
		return 1
	}

	if shouldPublish(session) {
		refreshSessionFrame(session)
	}
	return 1
}

//export EngineSessionFrame
func EngineSessionFrame(session *C.struct_EngineSession) *C.struct_EngineFrame {
	if session == nil {
		return nil
	}
	return session.frame
}

//export EngineRunLaboratorySession
func EngineRunLaboratorySession(
	session *C.struct_EngineSession,
	matches C.int,
	ticksPerMatch C.int,
	seed C.int64_t,
	outSummary *C.struct_LaboratorySummary,
) C.int {
	if session == nil || outSummary == nil {
		return 0
	}

	engine := lookup(uint64(session.handle))
	if engine == nil {
		return 0
	}

	summary := engine.RunLaboratory(int(matches), int(ticksPerMatch), int64(seed))
	outSummary.matches = C.int32_t(summary.Matches)
	outSummary.home_wins = C.int32_t(summary.HomeWins)
	outSummary.away_wins = C.int32_t(summary.AwayWins)
	outSummary.draws = C.int32_t(summary.Draws)
	outSummary.home_win_rate = C.double(summary.HomeWinRate)
	outSummary.away_win_rate = C.double(summary.AwayWinRate)
	outSummary.draw_rate = C.double(summary.DrawRate)
	outSummary.average_home_xg = C.double(summary.AverageHomeXG)
	outSummary.average_away_xg = C.double(summary.AverageAwayXG)
	outSummary.average_home_lbs = C.double(summary.AverageHomeLBS)
	outSummary.average_away_lbs = C.double(summary.AverageAwayLBS)
	outSummary.average_home_sgm = C.double(summary.AverageHomeSGM)
	outSummary.average_away_sgm = C.double(summary.AverageAwaySGM)
	outSummary.average_fatigue = C.double(summary.AverageFatigue)
	outSummary.average_ticks = C.double(summary.AverageTicks)
	return 1
}

//export EngineSnapshot
func EngineSnapshot(handle C.uint64_t) *C.struct_SimSnapshot {
	engine := lookup(uint64(handle))
	if engine == nil {
		return nil
	}

	snapshot := engine.Snapshot()
	raw := (*C.struct_SimSnapshot)(C.calloc(1, C.size_t(C.sizeof_SimSnapshot)))
	if raw == nil {
		return nil
	}

	raw.tick = C.int(snapshot.Tick)
	raw.possession_team = C.int(snapshot.PossessionTeam)
	raw.home_xg = C.double(snapshot.HomeXG)
	raw.away_xg = C.double(snapshot.AwayXG)
	raw.average_fatigue = C.double(snapshot.AverageFatigue)
	raw.ball.x = C.double(snapshot.Ball.X)
	raw.ball.y = C.double(snapshot.Ball.Y)
	raw.ball.vx = C.double(snapshot.Ball.VX)
	raw.ball.vy = C.double(snapshot.Ball.VY)
	raw.ball.spin = C.double(snapshot.Ball.Spin)

	for i := 0; i < len(snapshot.Players); i++ {
		player := snapshot.Players[i]
		raw.players[i].id = C.int(player.ID)
		raw.players[i].team = C.int(player.Team)
		raw.players[i].x = C.double(player.X)
		raw.players[i].y = C.double(player.Y)
		raw.players[i].speed = C.double(player.Speed)
		raw.players[i].stamina = C.double(player.Stamina)
	}

	return raw
}

//export EngineFreeSnapshot
func EngineFreeSnapshot(ptr *C.struct_SimSnapshot) {
	if ptr == nil {
		return
	}
	C.free(unsafe.Pointer(ptr))
}

func lookup(handle uint64) *sim.Engine {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[handle]
}

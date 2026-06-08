import 'dart:math';

import '../domain/career.dart';

enum EngineTeam { home, away }

class PlayerSnapshot {
  const PlayerSnapshot({
    required this.index,
    required this.team,
    required this.x,
    required this.y,
    required this.speed,
    required this.stamina,
  });

  final int index;
  final EngineTeam team;
  final double x;
  final double y;
  final double speed;
  final double stamina;
}

class BlockSnapshotView {
  const BlockSnapshotView({
    required this.team,
    required this.kind,
    required this.x,
    required this.y,
    required this.pressure,
    required this.fatigue,
    required this.travelled,
    required this.passingFactor,
    required this.interceptFactor,
    required this.lineHeight,
    required this.compactness,
    required this.pressureAggression,
    required this.offsideTrap,
    required this.coverage,
  });

  final int team;
  final int kind;
  final double x;
  final double y;
  final double pressure;
  final double fatigue;
  final double travelled;
  final double passingFactor;
  final double interceptFactor;
  final double lineHeight;
  final double compactness;
  final double pressureAggression;
  final double offsideTrap;
  final double coverage;

  bool get isHome => team == 0;
}

class EngineFrameView {
  const EngineFrameView({
    required this.tick,
    required this.possessionTeam,
    required this.possessor,
    required this.lastPassCompleted,
    required this.homeXg,
    required this.awayXg,
    required this.homeLbs,
    required this.awayLbs,
    required this.homeSgm,
    required this.awaySgm,
    required this.averageFatigue,
    required this.lastPassAccuracy,
    required this.lastShotXg,
    required this.ballX,
    required this.ballY,
    required this.ballVx,
    required this.ballVy,
    required this.ballSpin,
    required this.players,
    required this.blocks,
  });

  final int tick;
  final int possessionTeam;
  final int possessor;
  final bool lastPassCompleted;
  final double homeXg;
  final double awayXg;
  final double homeLbs;
  final double awayLbs;
  final double homeSgm;
  final double awaySgm;
  final double averageFatigue;
  final double lastPassAccuracy;
  final double lastShotXg;
  final double ballX;
  final double ballY;
  final double ballVx;
  final double ballVy;
  final double ballSpin;
  final List<PlayerSnapshot> players;
  final List<BlockSnapshotView> blocks;

  bool get homePossession => possessionTeam == 0;
}

class LaboratorySummaryView {
  const LaboratorySummaryView({
    required this.matches,
    required this.homeWins,
    required this.awayWins,
    required this.draws,
    required this.homeWinRate,
    required this.awayWinRate,
    required this.drawRate,
    required this.averageHomeXg,
    required this.averageAwayXg,
    required this.averageHomeLbs,
    required this.averageAwayLbs,
    required this.averageHomeSgm,
    required this.averageAwaySgm,
    required this.averageFatigue,
    required this.averageTicks,
  });

  final int matches;
  final int homeWins;
  final int awayWins;
  final int draws;
  final double homeWinRate;
  final double awayWinRate;
  final double drawRate;
  final double averageHomeXg;
  final double averageAwayXg;
  final double averageHomeLbs;
  final double averageAwayLbs;
  final double averageHomeSgm;
  final double averageAwaySgm;
  final double averageFatigue;
  final double averageTicks;

  Map<String, Object?> toJson() => <String, Object?>{
    'matches': matches,
    'homeWins': homeWins,
    'awayWins': awayWins,
    'draws': draws,
    'homeWinRate': homeWinRate,
    'awayWinRate': awayWinRate,
    'drawRate': drawRate,
    'averageHomeXg': averageHomeXg,
    'averageAwayXg': averageAwayXg,
    'averageHomeLbs': averageHomeLbs,
    'averageAwayLbs': averageAwayLbs,
    'averageHomeSgm': averageHomeSgm,
    'averageAwaySgm': averageAwaySgm,
    'averageFatigue': averageFatigue,
    'averageTicks': averageTicks,
  };
}

class BlockTacticalConfig {
  const BlockTacticalConfig({
    required this.lineHeight,
    required this.compactness,
    required this.pressureAggression,
    required this.offsideTrap,
    required this.coverage,
  });

  final double lineHeight;
  final double compactness;
  final double pressureAggression;
  final double offsideTrap;
  final double coverage;

  BlockTacticalConfig copyWith({
    double? lineHeight,
    double? compactness,
    double? pressureAggression,
    double? offsideTrap,
    double? coverage,
  }) {
    return BlockTacticalConfig(
      lineHeight: lineHeight ?? this.lineHeight,
      compactness: compactness ?? this.compactness,
      pressureAggression: pressureAggression ?? this.pressureAggression,
      offsideTrap: offsideTrap ?? this.offsideTrap,
      coverage: coverage ?? this.coverage,
    );
  }
}

class TacticalConfig {
  const TacticalConfig({
    required this.tempo,
    required this.pressingIntensity,
    required this.blockHeight,
    required this.riskAppetite,
    required this.blocks,
  });

  factory TacticalConfig.defaults() {
    return const TacticalConfig(
      tempo: 0.55,
      pressingIntensity: 0.45,
      blockHeight: 0.5,
      riskAppetite: 0.5,
      blocks: [
        BlockTacticalConfig(
          lineHeight: 0.34,
          compactness: 0.68,
          pressureAggression: 0.52,
          offsideTrap: 0.30,
          coverage: 0.74,
        ),
        BlockTacticalConfig(
          lineHeight: 0.48,
          compactness: 0.62,
          pressureAggression: 0.47,
          offsideTrap: 0.20,
          coverage: 0.64,
        ),
        BlockTacticalConfig(
          lineHeight: 0.58,
          compactness: 0.55,
          pressureAggression: 0.44,
          offsideTrap: 0.18,
          coverage: 0.58,
        ),
        BlockTacticalConfig(
          lineHeight: 0.72,
          compactness: 0.50,
          pressureAggression: 0.36,
          offsideTrap: 0.12,
          coverage: 0.46,
        ),
      ],
    );
  }

  final double tempo;
  final double pressingIntensity;
  final double blockHeight;
  final double riskAppetite;
  final List<BlockTacticalConfig> blocks;

  TacticalConfig copyWith({
    double? tempo,
    double? pressingIntensity,
    double? blockHeight,
    double? riskAppetite,
    List<BlockTacticalConfig>? blocks,
  }) {
    return TacticalConfig(
      tempo: tempo ?? this.tempo,
      pressingIntensity: pressingIntensity ?? this.pressingIntensity,
      blockHeight: blockHeight ?? this.blockHeight,
      riskAppetite: riskAppetite ?? this.riskAppetite,
      blocks: blocks ?? this.blocks,
    );
  }
}

abstract class ScoreadorEngineBridge {
  factory ScoreadorEngineBridge.create() => MockScoreadorEngineBridge();

  EngineFrameView readFrame();
  void applyConfig(TacticalConfig config);
  void setPlaybackRate(double rate);
  void setFramePublishing(bool enabled);
  void advance(Duration elapsed);
  void step([int ticks = 1]);
  LaboratorySummaryView runLaboratory({
    required int matches,
    required int ticksPerMatch,
    required int seed,
  });
  void reset();
  void dispose();
  bool get isNative;
}

class MockScoreadorEngineBridge implements ScoreadorEngineBridge {
  MockScoreadorEngineBridge()
    : _config = TacticalConfig.defaults(),
      _frame = _buildFrame(
        tick: 0,
        config: TacticalConfig.defaults(),
        ballX: 52.5,
        ballY: 34.0,
        ballVx: 0.0,
        ballVy: 0.0,
        ballSpin: 0.0,
        possessionTeam: 0,
        possessor: 8,
        lastPassCompleted: false,
        homeXg: 0.12,
        awayXg: 0.08,
        homeLbs: 2.0,
        awayLbs: 1.0,
        homeSgm: 0.9,
        awaySgm: 0.7,
        averageFatigue: 0.18,
        lastPassAccuracy: 0.74,
        lastShotXg: 0.15,
      );

  static const double _tickSeconds = 0.1;
  static const double _minPlaybackRate = 0.25;
  static const double _maxPlaybackRate = 10.0;

  TacticalConfig _config;
  EngineFrameView _frame;
  double _playbackRate = 1.0;
  double _accumulator = 0.0;
  bool _publishing = true;

  @override
  bool get isNative => false;

  @override
  void applyConfig(TacticalConfig config) {
    _config = config;
    _frame = _buildFrame(
      tick: _frame.tick,
      config: _config,
      ballX: _frame.ballX,
      ballY: _frame.ballY,
      ballVx: _frame.ballVx,
      ballVy: _frame.ballVy,
      ballSpin: _frame.ballSpin,
      possessionTeam: _frame.possessionTeam,
      possessor: _frame.possessor,
      lastPassCompleted: _frame.lastPassCompleted,
      homeXg: _frame.homeXg,
      awayXg: _frame.awayXg,
      homeLbs: _frame.homeLbs,
      awayLbs: _frame.awayLbs,
      homeSgm: _frame.homeSgm,
      awaySgm: _frame.awaySgm,
      averageFatigue: _frame.averageFatigue,
      lastPassAccuracy: _frame.lastPassAccuracy,
      lastShotXg: _frame.lastShotXg,
      players: _frame.players,
    );
  }

  @override
  void setPlaybackRate(double rate) {
    _playbackRate = rate.clamp(_minPlaybackRate, _maxPlaybackRate).toDouble();
  }

  @override
  void setFramePublishing(bool enabled) {
    _publishing = enabled;
  }

  @override
  void advance(Duration elapsed) {
    if (!_publishing) {
      return;
    }
    _accumulator += (elapsed.inMicroseconds / 1000000.0) * _playbackRate;
    while (_accumulator >= _tickSeconds) {
      _simulateTick();
      _accumulator -= _tickSeconds;
    }
  }

  @override
  void step([int ticks = 1]) {
    if (!_publishing) {
      return;
    }
    for (var index = 0; index < ticks; index++) {
      _simulateTick();
    }
  }

  @override
  LaboratorySummaryView runLaboratory({
    required int matches,
    required int ticksPerMatch,
    required int seed,
  }) {
    final random = Random(seed);
    var homeWins = 0;
    var awayWins = 0;
    var draws = 0;
    var totalHomeXg = 0.0;
    var totalAwayXg = 0.0;
    var totalHomeLbs = 0.0;
    var totalAwayLbs = 0.0;
    var totalHomeSgm = 0.0;
    var totalAwaySgm = 0.0;
    var totalFatigue = 0.0;

    final frame = _frame;
    for (var matchIndex = 0; matchIndex < matches; matchIndex++) {
      final matchNoise = (random.nextDouble() - 0.5) * 0.12;
      final tempoFactor = _config.tempo * 0.3;
      final pressureFactor = _config.pressingIntensity * 0.18;
      final riskFactor = _config.riskAppetite * 0.14;
      final homeXg = (frame.homeXg + tempoFactor + matchNoise).clamp(0.05, 4.5);
      final awayXg = (frame.awayXg + pressureFactor - matchNoise / 2).clamp(
        0.05,
        4.5,
      );
      final homeLbs =
          (frame.homeLbs + _config.blockHeight * 1.8 + random.nextDouble())
              .clamp(0.0, 12.0);
      final awayLbs =
          (frame.awayLbs +
                  (1.0 - _config.blockHeight) * 1.4 +
                  random.nextDouble())
              .clamp(0.0, 12.0);
      final homeSgm = (frame.homeSgm + _config.tempo * 0.8 + riskFactor).clamp(
        0.0,
        12.0,
      );
      final awaySgm =
          (frame.awaySgm + _config.pressingIntensity * 0.7 + riskFactor / 2)
              .clamp(0.0, 12.0);
      final fatigue =
          (frame.averageFatigue +
                  (ticksPerMatch / 240.0) * 0.08 +
                  random.nextDouble() * 0.05)
              .clamp(0.0, 1.0);
      final homeStrength =
          homeXg + homeLbs * 0.06 + homeSgm * 0.04 - fatigue * 0.3;
      final awayStrength =
          awayXg + awayLbs * 0.06 + awaySgm * 0.04 - fatigue * 0.3;

      totalHomeXg += homeXg;
      totalAwayXg += awayXg;
      totalHomeLbs += homeLbs;
      totalAwayLbs += awayLbs;
      totalHomeSgm += homeSgm;
      totalAwaySgm += awaySgm;
      totalFatigue += fatigue;

      final difference = (homeStrength - awayStrength).abs();
      if (difference <= 0.14) {
        draws += 1;
      } else if (homeStrength > awayStrength) {
        homeWins += 1;
      } else {
        awayWins += 1;
      }
    }

    final totalMatches = matches <= 0 ? 1 : matches;
    return LaboratorySummaryView(
      matches: matches,
      homeWins: homeWins,
      awayWins: awayWins,
      draws: draws,
      homeWinRate: homeWins / totalMatches,
      awayWinRate: awayWins / totalMatches,
      drawRate: draws / totalMatches,
      averageHomeXg: totalHomeXg / totalMatches,
      averageAwayXg: totalAwayXg / totalMatches,
      averageHomeLbs: totalHomeLbs / totalMatches,
      averageAwayLbs: totalAwayLbs / totalMatches,
      averageHomeSgm: totalHomeSgm / totalMatches,
      averageAwaySgm: totalAwaySgm / totalMatches,
      averageFatigue: totalFatigue / totalMatches,
      averageTicks: ticksPerMatch.toDouble(),
    );
  }

  @override
  EngineFrameView readFrame() => _frame;

  @override
  void reset() {
    _accumulator = 0.0;
    _publishing = true;
    _frame = _buildFrame(
      tick: 0,
      config: _config,
      ballX: 52.5,
      ballY: 34.0,
      ballVx: 0.0,
      ballVy: 0.0,
      ballSpin: 0.0,
      possessionTeam: 0,
      possessor: 8,
      lastPassCompleted: false,
      homeXg: 0.12,
      awayXg: 0.08,
      homeLbs: 2.0,
      awayLbs: 1.0,
      homeSgm: 0.9,
      awaySgm: 0.7,
      averageFatigue: 0.18,
      lastPassAccuracy: 0.74,
      lastShotXg: 0.15,
    );
  }

  @override
  void dispose() {}

  void _simulateTick() {
    final nextTick = _frame.tick + 1;
    final tempo = _config.tempo;
    final pressure = _config.pressingIntensity;
    final risk = _config.riskAppetite;
    final ballX = (_frame.ballX + 0.9 + tempo * 0.45).clamp(0.0, kPitchLength);
    final ballY =
        (_frame.ballY + (nextTick.isEven ? 0.28 : -0.18) + pressure * 0.12)
            .clamp(0.0, kPitchWidth);
    final players = List<PlayerSnapshot>.generate(kPlayerCount, (index) {
      final player = _frame.players[index];
      final isHome = player.team == EngineTeam.home;
      final direction = isHome ? 1.0 : -1.0;
      final driftX = direction * (0.05 + tempo * 0.03);
      final driftY = (index % 3 - 1) * 0.05 + pressure * 0.01;
      return PlayerSnapshot(
        index: player.index,
        team: player.team,
        x: (player.x + driftX).clamp(0.0, kPitchLength),
        y: (player.y + driftY).clamp(0.0, kPitchWidth),
        speed: 4.0 + (index % 4) * 0.5,
        stamina: (player.stamina - 0.0009 - risk * 0.0003).clamp(0.0, 1.0),
      );
    });
    final blocks = _rebuildBlocks(players, _config, nextTick);
    _frame = _buildFrame(
      tick: nextTick,
      config: _config,
      ballX: ballX,
      ballY: ballY,
      ballVx: 0.85 + tempo * 0.25,
      ballVy: (nextTick.isEven ? 0.2 : -0.15) + pressure * 0.08,
      ballSpin: pressure * 1.2,
      possessionTeam: nextTick.isEven ? 0 : 1,
      possessor: nextTick % kPlayerCount,
      lastPassCompleted: nextTick % 3 != 0,
      homeXg: _frame.homeXg + 0.001,
      awayXg: _frame.awayXg + 0.0005,
      homeLbs: _frame.homeLbs + 0.01,
      awayLbs: _frame.awayLbs + 0.01,
      homeSgm: _frame.homeSgm + 0.005,
      awaySgm: _frame.awaySgm + 0.004,
      averageFatigue: (_frame.averageFatigue + 0.002).clamp(0.0, 1.0),
      lastPassAccuracy: 0.62 + tempo * 0.2,
      lastShotXg: 0.1 + risk * 0.1,
      players: players,
      blocks: blocks,
    );
  }
}

const double kPitchLength = 105.0;
const double kPitchWidth = 68.0;
const int kPlayerCount = 22;
const int kBlockCount = 4;

EngineFrameView _buildFrame({
  required int tick,
  required TacticalConfig config,
  required double ballX,
  required double ballY,
  required double ballVx,
  required double ballVy,
  required double ballSpin,
  required int possessionTeam,
  required int possessor,
  required bool lastPassCompleted,
  required double homeXg,
  required double awayXg,
  required double homeLbs,
  required double awayLbs,
  required double homeSgm,
  required double awaySgm,
  required double averageFatigue,
  required double lastPassAccuracy,
  required double lastShotXg,
  List<PlayerSnapshot>? players,
  List<BlockSnapshotView>? blocks,
}) {
  return EngineFrameView(
    tick: tick,
    possessionTeam: possessionTeam,
    possessor: possessor,
    lastPassCompleted: lastPassCompleted,
    homeXg: homeXg,
    awayXg: awayXg,
    homeLbs: homeLbs,
    awayLbs: awayLbs,
    homeSgm: homeSgm,
    awaySgm: awaySgm,
    averageFatigue: averageFatigue,
    lastPassAccuracy: lastPassAccuracy,
    lastShotXg: lastShotXg,
    ballX: ballX,
    ballY: ballY,
    ballVx: ballVx,
    ballVy: ballVy,
    ballSpin: ballSpin,
    players:
        players ??
        List<PlayerSnapshot>.generate(kPlayerCount, (index) {
          final isHome = index < 11;
          final team = isHome ? EngineTeam.home : EngineTeam.away;
          final formationIndex = index % 11;
          final row = formationIndex ~/ 4;
          final column = formationIndex % 4;
          final x = isHome
              ? 8.0 + column * 7.0 + row * 3.0
              : 97.0 - column * 7.0 - row * 3.0;
          final y = 8.0 + row * 18.0 + column * 5.0;
          return PlayerSnapshot(
            index: index,
            team: team,
            x: x.clamp(0.0, kPitchLength),
            y: y.clamp(0.0, kPitchWidth),
            speed: 4.5 + (index % 5) * 0.3,
            stamina: 0.92 - index * 0.01,
          );
        }),
    blocks: blocks ?? _buildBlocks(config),
  );
}

List<BlockSnapshotView> _buildBlocks(TacticalConfig config) {
  return List<BlockSnapshotView>.generate(kBlockCount, (index) {
    final blockConfig = config.blocks[index];
    final isHome = index < 2;
    final x = isHome ? 24.0 + index * 8.0 : 81.0 - (index - 2) * 8.0;
    final y = 15.0 + index * 11.0;
    return BlockSnapshotView(
      team: isHome ? 0 : 1,
      kind: index,
      x: x,
      y: y,
      pressure: config.pressingIntensity * (0.7 + index * 0.06),
      fatigue: 0.1 + index * 0.03,
      travelled: 2.0 + index * 0.5,
      passingFactor: 0.7 + config.tempo * 0.2 - index * 0.03,
      interceptFactor: 0.5 + config.pressingIntensity * 0.25 + index * 0.04,
      lineHeight: blockConfig.lineHeight,
      compactness: blockConfig.compactness,
      pressureAggression: blockConfig.pressureAggression,
      offsideTrap: blockConfig.offsideTrap,
      coverage: blockConfig.coverage,
    );
  });
}

List<BlockSnapshotView> _rebuildBlocks(
  List<PlayerSnapshot> players,
  TacticalConfig config,
  int tick,
) {
  final blocks = _buildBlocks(config);
  return List<BlockSnapshotView>.generate(blocks.length, (index) {
    final block = blocks[index];
    final teamPlayers = players
        .where(
          (player) => index < 2
              ? player.team == EngineTeam.home
              : player.team == EngineTeam.away,
        )
        .toList(growable: false);
    final averageStamina = teamPlayers.isEmpty
        ? 0.0
        : teamPlayers.map((player) => player.stamina).reduce((a, b) => a + b) /
              teamPlayers.length;
    return BlockSnapshotView(
      team: block.team,
      kind: block.kind,
      x: block.x,
      y: block.y,
      pressure: (block.pressure + tick * 0.001).clamp(0.0, 1.5),
      fatigue: (1.0 - averageStamina).clamp(0.0, 1.0),
      travelled: block.travelled + tick * 0.02,
      passingFactor: block.passingFactor,
      interceptFactor: block.interceptFactor,
      lineHeight: block.lineHeight,
      compactness: block.compactness,
      pressureAggression: block.pressureAggression,
      offsideTrap: block.offsideTrap,
      coverage: block.coverage,
    );
  });
}

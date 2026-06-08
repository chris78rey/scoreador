import 'dart:ffi';
import 'dart:math';
import 'dart:io';

import 'package:ffi/ffi.dart';

import '../domain/career.dart';
import 'engine_ffi.dart';

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
          coverage: 0.76,
        ),
        BlockTacticalConfig(
          lineHeight: 0.50,
          compactness: 0.62,
          pressureAggression: 0.48,
          offsideTrap: 0.24,
          coverage: 0.64,
        ),
        BlockTacticalConfig(
          lineHeight: 0.66,
          compactness: 0.55,
          pressureAggression: 0.44,
          offsideTrap: 0.18,
          coverage: 0.56,
        ),
        BlockTacticalConfig(
          lineHeight: 0.82,
          compactness: 0.48,
          pressureAggression: 0.38,
          offsideTrap: 0.12,
          coverage: 0.48,
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
  factory ScoreadorEngineBridge.create() {
    if (Platform.isAndroid) {
      try {
        return NativeScoreadorEngineBridge();
      } catch (_) {
        return MockScoreadorEngineBridge();
      }
    }
    return MockScoreadorEngineBridge();
  }

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

class NativeScoreadorEngineBridge implements ScoreadorEngineBridge {
  NativeScoreadorEngineBridge() : _api = EngineNativeApi.load() {
    _session = _api.createSession();
    if (_session == null) {
      throw StateError('No se pudo crear la sesion nativa.');
    }
    _configData = TacticalConfig.defaults();
    applyConfig(_configData);
    setPlaybackRate(1.0);
  }

  static const double _minPlaybackRate = 0.25;
  static const double _maxPlaybackRate = 10.0;

  final EngineNativeApi _api;
  Pointer<EngineSessionNative>? _session;
  late TacticalConfig _configData;
  double _playbackRate = 1.0;

  @override
  bool get isNative => true;

  Pointer<EngineFrameNative> get _framePointer {
    final session = _session;
    if (session == null) {
      throw StateError('Sesion nativa cerrada.');
    }
    return _api.sessionFrame(session);
  }

  @override
  void applyConfig(TacticalConfig config) {
    _configData = config;
    final session = _session;
    if (session == null) {
      return;
    }
    final blockValues = _api.allocateBlockConfigs(
      values: config.blocks
          .map(
            (block) => BlockTacticalConfigNativeValues(
              lineHeight: block.lineHeight,
              compactness: block.compactness,
              pressureAggression: block.pressureAggression,
              offsideTrap: block.offsideTrap,
              coverage: block.coverage,
            ),
          )
          .toList(growable: false),
    );
    final nativeConfig = _api.allocateConfig(
      tempo: config.tempo,
      pressingIntensity: config.pressingIntensity,
      blockHeight: config.blockHeight,
      riskAppetite: config.riskAppetite,
      blocks: blockValues,
    );
    _api.setConfigSession(session, nativeConfig);
    calloc.free(nativeConfig);
    calloc.free(blockValues);
  }

  @override
  void setPlaybackRate(double rate) {
    final clamped = rate.clamp(_minPlaybackRate, _maxPlaybackRate).toDouble();
    _playbackRate = clamped;
    final session = _session;
    if (session == null) {
      return;
    }
    _api.setPlaybackRateSession(session, clamped);
  }

  @override
  void setFramePublishing(bool enabled) {
    final session = _session;
    if (session == null) {
      return;
    }
    _api.setFramePublishingSession(session, enabled ? 1 : 0);
  }

  @override
  void advance(Duration elapsed) {
    final session = _session;
    if (session == null) {
      return;
    }
    _api.advanceSession(session, elapsed.inMicroseconds / 1000000.0);
  }

  @override
  void step([int ticks = 1]) {
    final session = _session;
    if (session == null) {
      return;
    }
    _api.stepSession(session, ticks);
  }

  @override
  LaboratorySummaryView runLaboratory({
    required int matches,
    required int ticksPerMatch,
    required int seed,
  }) {
    final session = _session;
    if (session == null) {
      throw StateError('Sesion nativa cerrada.');
    }
    final summaryPtr = calloc<LaboratorySummaryNative>();
    try {
      final status = _api.runLaboratorySession(
        session,
        matches,
        ticksPerMatch,
        seed,
        summaryPtr,
      );
      if (status != 0) {
        throw StateError('La ejecucion del laboratorio fallo con codigo $status.');
      }
      final summary = _api.readLaboratorySummary(summaryPtr);
      return LaboratorySummaryView(
        matches: summary.matches,
        homeWins: summary.homeWins,
        awayWins: summary.awayWins,
        draws: summary.draws,
        homeWinRate: summary.homeWinRate,
        awayWinRate: summary.awayWinRate,
        drawRate: summary.drawRate,
        averageHomeXg: summary.averageHomeXg,
        averageAwayXg: summary.averageAwayXg,
        averageHomeLbs: summary.averageHomeLbs,
        averageAwayLbs: summary.averageAwayLbs,
        averageHomeSgm: summary.averageHomeSgm,
        averageAwaySgm: summary.averageAwaySgm,
        averageFatigue: summary.averageFatigue,
        averageTicks: summary.averageTicks,
      );
    } finally {
      calloc.free(summaryPtr);
    }
  }

  @override
  EngineFrameView readFrame() {
    return _mapFrame(_framePointer.ref);
  }

  @override
  void reset() {
    dispose();
    _session = _api.createSession();
    if (_session == null) {
      throw StateError('No se pudo reinicializar la sesion nativa.');
    }
    applyConfig(_configData);
    setPlaybackRate(_playbackRate);
  }

  @override
  void dispose() {
    final session = _session;
    if (session == null) {
      return;
    }
    _api.destroySession(session);
    _session = null;
  }
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
    _playbackRate = rate.clamp(0.25, 10.0).toDouble();
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
    for (var i = 0; i < ticks; i++) {
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
      final awayXg = (frame.awayXg + pressureFactor - matchNoise / 2).clamp(0.05, 4.5);
      final homeLbs = (frame.homeLbs + _config.blockHeight * 1.8 + random.nextDouble()).clamp(0.0, 12.0);
      final awayLbs = (frame.awayLbs + (1.0 - _config.blockHeight) * 1.4 + random.nextDouble()).clamp(0.0, 12.0);
      final homeSgm = (frame.homeSgm + _config.tempo * 0.8 + riskFactor).clamp(0.0, 12.0);
      final awaySgm = (frame.awaySgm + _config.pressingIntensity * 0.7 + riskFactor / 2).clamp(0.0, 12.0);
      final fatigue = (frame.averageFatigue + (ticksPerMatch / 240.0) * 0.08 + random.nextDouble() * 0.05).clamp(0.0, 1.0);
      final homeStrength = homeXg + homeLbs * 0.06 + homeSgm * 0.04 - fatigue * 0.3;
      final awayStrength = awayXg + awayLbs * 0.06 + awaySgm * 0.04 - fatigue * 0.3;

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
}

EngineFrameView _mapFrame(EngineFrameNative frame) {
  final players = List<PlayerSnapshot>.generate(kPlayerCount, (index) {
    final isHome = index < 11;
    return PlayerSnapshot(
      index: index,
      team: isHome ? EngineTeam.home : EngineTeam.away,
      x: frame.playerX[index],
      y: frame.playerY[index],
      speed: frame.playerSpeed[index],
      stamina: frame.playerStamina[index],
    );
  });

  final blocks = <BlockSnapshotView>[];
  final blockPointer = frame.blocks;
  if (blockPointer.address != 0) {
    for (var index = 0; index < kBlockCount; index++) {
      final block = blockPointer.elementAt(index).ref;
      blocks.add(
        BlockSnapshotView(
          team: block.team,
          kind: block.kind,
          x: block.x,
          y: block.y,
          pressure: block.pressure,
          fatigue: block.fatigue,
          travelled: block.travelled,
          passingFactor: block.passingFactor,
          interceptFactor: block.interceptFactor,
          lineHeight: block.lineHeight,
          compactness: block.compactness,
          pressureAggression: block.pressureAggression,
          offsideTrap: block.offsideTrap,
          coverage: block.coverage,
        ),
      );
    }
  }

  return EngineFrameView(
    tick: frame.tick,
    possessionTeam: frame.possessionTeam,
    possessor: frame.possessor,
    lastPassCompleted: frame.lastPassCompleted != 0,
    homeXg: frame.homeXg,
    awayXg: frame.awayXg,
    homeLbs: frame.homeLbs,
    awayLbs: frame.awayLbs,
    homeSgm: frame.homeSgm,
    awaySgm: frame.awaySgm,
    averageFatigue: frame.averageFatigue,
    lastPassAccuracy: frame.lastPassAccuracy,
    lastShotXg: frame.lastShotXg,
    ballX: frame.ballX,
    ballY: frame.ballY,
    ballVx: frame.ballVx,
    ballVy: frame.ballVy,
    ballSpin: frame.ballSpin,
    players: players,
    blocks: blocks,
  );
}

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
  final currentPlayers =
      players ??
      List<PlayerSnapshot>.generate(kPlayerCount, (index) {
        final isHome = index < 11;
        final slot = index % 11;
        final row = slot ~/ 4;
        final col = slot % 4;
        return PlayerSnapshot(
          index: index,
          team: isHome ? EngineTeam.home : EngineTeam.away,
          x: isHome ? 18 + col * 6.5 : 87 - col * 6.5,
          y: 8 + row * 16.0,
          speed: 4.0 + (slot % 3),
          stamina: 0.7 - (index * 0.01),
        );
      });
  final currentBlocks = blocks ?? _defaultBlocks(currentPlayers, config, tick);
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
    players: currentPlayers,
    blocks: currentBlocks,
  );
}

List<BlockSnapshotView> _defaultBlocks(
  List<PlayerSnapshot> players,
  TacticalConfig config,
  int tick,
) {
  final results = <BlockSnapshotView>[];
  for (var team = 0; team < kTeamsPerMatch; team++) {
    for (var kind = 0; kind < kBlocksPerTeam; kind++) {
      final relevant = players
          .where(
            (player) =>
                player.team.index == team && _playerKind(player.index) == kind,
          )
          .toList(growable: false);
      if (relevant.isEmpty) {
        results.add(
          BlockSnapshotView(
            team: team,
            kind: kind,
            x: team == 0 ? 30 : 75,
            y: 34,
            pressure: config.pressingIntensity * 0.5,
            fatigue: 0.1,
            travelled: 0,
            passingFactor: 0.7,
            interceptFactor: 0.7,
            lineHeight: config.blocks[kind].lineHeight,
            compactness: config.blocks[kind].compactness,
            pressureAggression: config.blocks[kind].pressureAggression,
            offsideTrap: config.blocks[kind].offsideTrap,
            coverage: config.blocks[kind].coverage,
          ),
        );
        continue;
      }
      final avgX =
          relevant.map((player) => player.x).reduce((a, b) => a + b) /
          relevant.length;
      final avgY =
          relevant.map((player) => player.y).reduce((a, b) => a + b) /
          relevant.length;
      final cfg = config.blocks[kind];
      results.add(
        BlockSnapshotView(
          team: team,
          kind: kind,
          x: avgX,
          y: avgY,
          pressure: 0.2 + cfg.pressureAggression * 0.4,
          fatigue: 0.1 + kind * 0.03,
          travelled: tick * 0.08,
          passingFactor: 0.72 - kind * 0.03,
          interceptFactor: 0.74 - kind * 0.02,
          lineHeight: cfg.lineHeight,
          compactness: cfg.compactness,
          pressureAggression: cfg.pressureAggression,
          offsideTrap: cfg.offsideTrap,
          coverage: cfg.coverage,
        ),
      );
    }
  }
  return results;
}

List<BlockSnapshotView> _rebuildBlocks(
  List<PlayerSnapshot> players,
  TacticalConfig config,
  int tick,
) {
  return _defaultBlocks(players, config, tick);
}

int _playerKind(int index) {
  final slot = index % 11;
  if (slot <= 3) {
    return 0;
  }
  if (slot <= 6) {
    return 1;
  }
  if (slot == 7) {
    return 2;
  }
  return 3;
}

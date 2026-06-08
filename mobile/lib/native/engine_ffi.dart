import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

const int kPlayerCount = 22;
const int kBlocksPerTeam = 4;
const int kTeamsPerMatch = 2;
const int kBlockCount = kBlocksPerTeam * kTeamsPerMatch;
const double kPitchLength = 105.0;
const double kPitchWidth = 68.0;

final class BlockTacticalConfigNative extends Struct {
  @Double()
  external double lineHeight;

  @Double()
  external double compactness;

  @Double()
  external double pressureAggression;

  @Double()
  external double offsideTrap;

  @Double()
  external double coverage;
}

final class TacticalConfigNative extends Struct {
  @Double()
  external double tempo;

  @Double()
  external double pressingIntensity;

  @Double()
  external double blockHeight;

  @Double()
  external double riskAppetite;

  external Pointer<BlockTacticalConfigNative> blocks;
}

final class BlockFrameNative extends Struct {
  @Int32()
  external int team;

  @Int32()
  external int kind;

  @Double()
  external double x;

  @Double()
  external double y;

  @Double()
  external double pressure;

  @Double()
  external double fatigue;

  @Double()
  external double travelled;

  @Double()
  external double passingFactor;

  @Double()
  external double interceptFactor;

  @Double()
  external double lineHeight;

  @Double()
  external double compactness;

  @Double()
  external double pressureAggression;

  @Double()
  external double offsideTrap;

  @Double()
  external double coverage;
}

final class EngineFrameNative extends Struct {
  @Int32()
  external int tick;

  @Int32()
  external int possessionTeam;

  @Int32()
  external int possessor;

  @Int32()
  external int lastPassCompleted;

  @Double()
  external double homeXg;

  @Double()
  external double awayXg;

  @Double()
  external double homeLbs;

  @Double()
  external double awayLbs;

  @Double()
  external double homeSgm;

  @Double()
  external double awaySgm;

  @Double()
  external double averageFatigue;

  @Double()
  external double lastPassAccuracy;

  @Double()
  external double lastShotXg;

  @Double()
  external double ballX;

  @Double()
  external double ballY;

  @Double()
  external double ballVx;

  @Double()
  external double ballVy;

  @Double()
  external double ballSpin;

  @Array(kPlayerCount)
  external Array<Double> playerX;

  @Array(kPlayerCount)
  external Array<Double> playerY;

  @Array(kPlayerCount)
  external Array<Double> playerSpeed;

  @Array(kPlayerCount)
  external Array<Double> playerStamina;

  external Pointer<BlockFrameNative> blocks;
}

final class EngineSessionNative extends Struct {
  @Uint64()
  external int handle;

  external Pointer<EngineFrameNative> frame;
}

final class LaboratorySummaryNative extends Struct {
  @Int32()
  external int matches;

  @Int32()
  external int homeWins;

  @Int32()
  external int awayWins;

  @Int32()
  external int draws;

  @Double()
  external double homeWinRate;

  @Double()
  external double awayWinRate;

  @Double()
  external double drawRate;

  @Double()
  external double averageHomeXg;

  @Double()
  external double averageAwayXg;

  @Double()
  external double averageHomeLbs;

  @Double()
  external double averageAwayLbs;

  @Double()
  external double averageHomeSgm;

  @Double()
  external double averageAwaySgm;

  @Double()
  external double averageFatigue;

  @Double()
  external double averageTicks;
}

typedef _CreateSessionNative = Pointer<EngineSessionNative> Function();
typedef _DestroySessionNative = Void Function(Pointer<EngineSessionNative>);
typedef _SetConfigSessionNative =
    Int32 Function(Pointer<EngineSessionNative>, Pointer<TacticalConfigNative>);
typedef _StepSessionNative =
    Int32 Function(Pointer<EngineSessionNative>, Int32);
typedef _SessionFrameNative =
    Pointer<EngineFrameNative> Function(Pointer<EngineSessionNative>);
typedef _SetPlaybackRateSessionNative =
    Int32 Function(Pointer<EngineSessionNative>, Double);
typedef _AdvanceSessionNative =
    Int32 Function(Pointer<EngineSessionNative>, Double);
typedef _SetFramePublishingSessionNative =
    Int32 Function(Pointer<EngineSessionNative>, Int32);
typedef _RunLaboratorySessionNative = Int32 Function(
  Pointer<EngineSessionNative>,
  Int32,
  Int32,
  Int64,
  Pointer<LaboratorySummaryNative>,
);

class EngineNativeApi {
  EngineNativeApi._(this._lib)
    : createSession = _lib
          .lookupFunction<
            _CreateSessionNative,
            Pointer<EngineSessionNative> Function()
          >('EngineCreateSession'),
      destroySession = _lib
          .lookupFunction<
            _DestroySessionNative,
            void Function(Pointer<EngineSessionNative>)
          >('EngineDestroySession'),
      setConfigSession = _lib
          .lookupFunction<
            _SetConfigSessionNative,
            int Function(
              Pointer<EngineSessionNative>,
              Pointer<TacticalConfigNative>,
            )
          >('EngineSetConfigSession'),
      stepSession = _lib
          .lookupFunction<
            _StepSessionNative,
            int Function(Pointer<EngineSessionNative>, int)
          >('EngineStepSession'),
      setPlaybackRateSession = _lib
          .lookupFunction<
            _SetPlaybackRateSessionNative,
            int Function(Pointer<EngineSessionNative>, double)
          >('EngineSetPlaybackRateSession'),
      advanceSession = _lib
          .lookupFunction<
            _AdvanceSessionNative,
            int Function(Pointer<EngineSessionNative>, double)
          >('EngineAdvanceSession'),
      setFramePublishingSession = _lib
          .lookupFunction<
            _SetFramePublishingSessionNative,
            int Function(Pointer<EngineSessionNative>, int)
          >('EngineSetFramePublishingSession'),
      runLaboratorySession = _lib
          .lookupFunction<
            _RunLaboratorySessionNative,
            int Function(
              Pointer<EngineSessionNative>,
              int,
              int,
              int,
              Pointer<LaboratorySummaryNative>,
            )
          >('EngineRunLaboratorySession'),
      sessionFrame = _lib
          .lookupFunction<
            _SessionFrameNative,
            Pointer<EngineFrameNative> Function(Pointer<EngineSessionNative>)
          >('EngineSessionFrame');

  final DynamicLibrary _lib;
  final Pointer<EngineSessionNative> Function() createSession;
  final void Function(Pointer<EngineSessionNative>) destroySession;
  final int Function(
    Pointer<EngineSessionNative>,
    Pointer<TacticalConfigNative>,
  )
  setConfigSession;
  final int Function(Pointer<EngineSessionNative>, int) stepSession;
  final int Function(Pointer<EngineSessionNative>, double)
  setPlaybackRateSession;
  final int Function(Pointer<EngineSessionNative>, double) advanceSession;
  final int Function(Pointer<EngineSessionNative>, int)
  setFramePublishingSession;
  final int Function(
    Pointer<EngineSessionNative>,
    int,
    int,
    int,
    Pointer<LaboratorySummaryNative>,
  )
  runLaboratorySession;
  final Pointer<EngineFrameNative> Function(Pointer<EngineSessionNative>)
  sessionFrame;

  factory EngineNativeApi.load() {
    if (!Platform.isAndroid) {
      throw UnsupportedError('Native engine is only loaded on Android.');
    }
    return EngineNativeApi._(DynamicLibrary.open('libengine.so'));
  }

  Pointer<BlockTacticalConfigNative> allocateBlockConfigs({
    required List<BlockTacticalConfigNativeValues> values,
  }) {
    final blocks = calloc<BlockTacticalConfigNative>(values.length);
    for (var i = 0; i < values.length; i++) {
      blocks.elementAt(i).ref
        ..lineHeight = values[i].lineHeight
        ..compactness = values[i].compactness
        ..pressureAggression = values[i].pressureAggression
        ..offsideTrap = values[i].offsideTrap
        ..coverage = values[i].coverage;
    }
    return blocks;
  }

  Pointer<TacticalConfigNative> allocateConfig({
    required double tempo,
    required double pressingIntensity,
    required double blockHeight,
    required double riskAppetite,
    required Pointer<BlockTacticalConfigNative> blocks,
  }) {
    final config = calloc<TacticalConfigNative>();
    config.ref
      ..tempo = tempo
      ..pressingIntensity = pressingIntensity
      ..blockHeight = blockHeight
      ..riskAppetite = riskAppetite
      ..blocks = blocks;
    return config;
  }

  LaboratorySummaryNative readLaboratorySummary(Pointer<LaboratorySummaryNative> ptr) {
    return ptr.ref;
  }
}

class BlockTacticalConfigNativeValues {
  const BlockTacticalConfigNativeValues({
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
}

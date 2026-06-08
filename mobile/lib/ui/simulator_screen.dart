import 'dart:math';
import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';

import '../domain/career.dart';
import '../domain/career_store.dart';
import '../domain/laboratory_export.dart';
import '../native/engine_bridge.dart';
import 'pitch_painter.dart';

class SimulatorScreen extends StatefulWidget {
  const SimulatorScreen({super.key});

  @override
  State<SimulatorScreen> createState() => _SimulatorScreenState();
}

class _SimulatorScreenState extends State<SimulatorScreen>
    with TickerProviderStateMixin {
  static const List<String> _blockTitles = <String>[
    'Defensa',
    'Volantes',
    'Enganches',
    'Atacantes',
  ];

  late final ScoreadorEngineBridge _bridge;
  late final Ticker _ticker;
  late final TabController _tabController;
  late final LocalCareerStore _careerStore;
  late final CareerProgressionService _progression;
  late final CareerExportService _exportService;

  EngineFrameView? _frame;
  EngineFrameView? _renderFrame;
  TacticalConfig _config = TacticalConfig.defaults();
  CareerProfile _careerProfile = CareerProfile.initial();
  AccessTier _accessTier = AccessTier.basic;
  LaboratorySummaryView? _labSummary;
  String? _lastExportPath;

  bool _running = true;
  bool _ready = false;
  bool _labMode = false;
  bool _isDerby = false;
  bool _runningBeforeLab = true;
  bool _primeTickAfterReset = false;

  double _playbackRate = 1.0;
  double _labMatches = 240;
  double _labTicks = 240;
  Duration _lastElapsed = Duration.zero;
  List<Offset> _trajectory = <Offset>[];

  @override
  void initState() {
    super.initState();
    _careerStore = LocalCareerStore();
    _progression = const CareerProgressionService();
    _exportService = CareerExportService();
    _careerProfile = _careerStore.load();
    _accessTier =
        _careerProfile.premiumUnlocked ? AccessTier.premium : AccessTier.basic;
    _bridge = ScoreadorEngineBridge.create();
    _tabController = TabController(length: _blockTitles.length, vsync: this);
    _bridge.applyConfig(_config);
    _bridge.setPlaybackRate(_playbackRate);
    _bridge.setFramePublishing(true);
    _frame = _bridge.readFrame();
    _renderFrame = _frame;
    _appendTrajectory(_frame);
    _ready = true;
    _ticker = createTicker(_onTick)..start();
  }

  void _onTick(Duration elapsed) {
    if (!_ready || !mounted) {
      return;
    }

    if (_primeTickAfterReset) {
      _lastElapsed = elapsed;
      _primeTickAfterReset = false;
      return;
    }

    final delta = elapsed - _lastElapsed;
    _lastElapsed = elapsed;

    if (!_running || delta <= Duration.zero) {
      return;
    }

    _bridge.advance(delta);
    final frame = _bridge.readFrame();
    setState(() {
      _frame = frame;
      _renderFrame = _renderFrame == null
          ? frame
          : _interpolateFrame(_renderFrame!, frame, _smoothingFactor(delta));
      _appendTrajectory(_renderFrame);
    });
  }

  @override
  void dispose() {
    _careerStore.save(_careerProfile);
    _ticker.dispose();
    _tabController.dispose();
    _bridge.dispose();
    super.dispose();
  }

  void _appendTrajectory(EngineFrameView? frame) {
    if (frame == null) {
      return;
    }
    final point = Offset(frame.ballX, frame.ballY);
    if (_trajectory.isNotEmpty &&
        (_trajectory.last - point).distanceSquared < 0.04) {
      return;
    }
    final next = List<Offset>.from(_trajectory)..add(point);
    _trajectory = next.length > 220 ? next.sublist(next.length - 220) : next;
  }

  double _smoothingFactor(Duration delta) {
    final milliseconds = delta.inMicroseconds / 1000.0;
    if (milliseconds <= 0) {
      return 1.0;
    }
    return (1 - exp(-milliseconds / 70.0)).clamp(0.12, 0.42).toDouble();
  }

  EngineFrameView _interpolateFrame(
    EngineFrameView from,
    EngineFrameView to,
    double t,
  ) {
    PlayerSnapshot lerpPlayer(PlayerSnapshot a, PlayerSnapshot b) {
      return PlayerSnapshot(
        index: b.index,
        team: b.team,
        x: lerpDouble(a.x, b.x, t) ?? b.x,
        y: lerpDouble(a.y, b.y, t) ?? b.y,
        speed: lerpDouble(a.speed, b.speed, t) ?? b.speed,
        stamina: lerpDouble(a.stamina, b.stamina, t) ?? b.stamina,
      );
    }

    BlockSnapshotView lerpBlock(
      BlockSnapshotView a,
      BlockSnapshotView b,
    ) {
      return BlockSnapshotView(
        team: b.team,
        kind: b.kind,
        x: lerpDouble(a.x, b.x, t) ?? b.x,
        y: lerpDouble(a.y, b.y, t) ?? b.y,
        pressure: lerpDouble(a.pressure, b.pressure, t) ?? b.pressure,
        fatigue: lerpDouble(a.fatigue, b.fatigue, t) ?? b.fatigue,
        travelled: lerpDouble(a.travelled, b.travelled, t) ?? b.travelled,
        passingFactor:
            lerpDouble(a.passingFactor, b.passingFactor, t) ?? b.passingFactor,
        interceptFactor: lerpDouble(a.interceptFactor, b.interceptFactor, t) ??
            b.interceptFactor,
        lineHeight: lerpDouble(a.lineHeight, b.lineHeight, t) ?? b.lineHeight,
        compactness:
            lerpDouble(a.compactness, b.compactness, t) ?? b.compactness,
        pressureAggression: lerpDouble(
              a.pressureAggression,
              b.pressureAggression,
              t,
            ) ??
            b.pressureAggression,
        offsideTrap:
            lerpDouble(a.offsideTrap, b.offsideTrap, t) ?? b.offsideTrap,
        coverage: lerpDouble(a.coverage, b.coverage, t) ?? b.coverage,
      );
    }

    return EngineFrameView(
      tick: to.tick,
      possessionTeam: to.possessionTeam,
      possessor: to.possessor,
      lastPassCompleted: to.lastPassCompleted,
      homeXg: to.homeXg,
      awayXg: to.awayXg,
      homeLbs: to.homeLbs,
      awayLbs: to.awayLbs,
      homeSgm: to.homeSgm,
      awaySgm: to.awaySgm,
      averageFatigue: to.averageFatigue,
      lastPassAccuracy: to.lastPassAccuracy,
      lastShotXg: to.lastShotXg,
      ballX: lerpDouble(from.ballX, to.ballX, t) ?? to.ballX,
      ballY: lerpDouble(from.ballY, to.ballY, t) ?? to.ballY,
      ballVx: lerpDouble(from.ballVx, to.ballVx, t) ?? to.ballVx,
      ballVy: lerpDouble(from.ballVy, to.ballVy, t) ?? to.ballVy,
      ballSpin: lerpDouble(from.ballSpin, to.ballSpin, t) ?? to.ballSpin,
      players: List<PlayerSnapshot>.generate(
        to.players.length,
        (index) => lerpPlayer(from.players[index], to.players[index]),
        growable: false,
      ),
      blocks: List<BlockSnapshotView>.generate(
        to.blocks.length,
        (index) => lerpBlock(from.blocks[index], to.blocks[index]),
        growable: false,
      ),
    );
  }

  void _applyConfig(TacticalConfig next) {
    setState(() {
      _config = next;
    });
    _bridge.applyConfig(next);
    _frame = _bridge.readFrame();
    _renderFrame = _frame;
  }

  void _updateBlock(int index, BlockTacticalConfig updated) {
    final blocks = List<BlockTacticalConfig>.from(_config.blocks);
    blocks[index] = updated;
    _applyConfig(_config.copyWith(blocks: blocks));
  }

  double get _effectivePlaybackMax =>
      FeatureAccessPolicy.maxPlaybackRate(_accessTier);

  void _setPlaybackRate(double rate) {
    final clamped = rate.clamp(1.0, _effectivePlaybackMax).toDouble();
    setState(() {
      _playbackRate = clamped;
    });
    _bridge.setPlaybackRate(clamped);
  }

  void _setAccessTier(AccessTier tier) {
    if (_accessTier == tier) {
      return;
    }
    if (tier == AccessTier.premium && !_careerProfile.premiumUnlocked) {
      _showMessage('Desbloquea Premium local primero.');
      return;
    }

    final wasLabMode = _labMode;
    final nextPlaybackRate = _playbackRate.clamp(
      1.0,
      FeatureAccessPolicy.maxPlaybackRate(tier),
    );

    setState(() {
      _accessTier = tier;
      _playbackRate = nextPlaybackRate;
      if (tier == AccessTier.basic) {
        _labMode = false;
        _running = true;
      }
    });

    _bridge.setPlaybackRate(nextPlaybackRate);
    if (tier == AccessTier.basic && wasLabMode) {
      _bridge.setFramePublishing(true);
    }
  }

  void _unlockPremium() {
    if (_careerProfile.premiumUnlocked) {
      _setAccessTier(AccessTier.premium);
      return;
    }

    final updated = _progression.unlockPremium(_careerProfile);
    setState(() {
      _careerProfile = updated;
      _accessTier = AccessTier.premium;
    });
    _careerStore.save(_careerProfile);
  }

  void _setLabMode(bool enabled) {
    if (!FeatureAccessPolicy.canUseLaboratory(_accessTier)) {
      return;
    }

    if (enabled == _labMode) {
      return;
    }

    if (enabled) {
      setState(() {
        _runningBeforeLab = _running;
        _running = false;
        _labMode = true;
        _labSummary = null;
      });
      _bridge.setFramePublishing(false);
      return;
    }

    setState(() {
      _labMode = false;
      _running = _runningBeforeLab;
    });
    _bridge.setFramePublishing(true);
  }

  void _toggleRunning() {
    if (_labMode) {
      return;
    }
    setState(() {
      _running = !_running;
    });
  }

  void _registerCurrentMatch() {
    final frame = _frame;
    if (frame == null) {
      return;
    }

    final homeGoals = max(
      0,
      (frame.homeXg * 1.9 + frame.homeLbs * 0.08 + frame.homeSgm * 0.05)
          .round(),
    );
    final awayGoals = max(
      0,
      (frame.awayXg * 1.9 + frame.awayLbs * 0.08 + frame.awaySgm * 0.05)
          .round(),
    );
    final opponent = _isDerby ? 'Derbi de la ciudad' : 'Rival de jornada';

    final updated = _progression.registerMatch(
      _careerProfile,
      MatchRecord(
        opponent: opponent,
        derby: _isDerby,
        homeGoals: homeGoals,
        awayGoals: awayGoals,
        homeXg: frame.homeXg,
        awayXg: frame.awayXg,
        homeLbs: frame.homeLbs,
        awayLbs: frame.awayLbs,
        averageFatigue: frame.averageFatigue,
        playedAt: DateTime.now(),
      ),
    );

    setState(() {
      _careerProfile = updated;
    });
    _careerStore.save(_careerProfile);
  }

  void _runLaboratoryBatch() {
    if (!FeatureAccessPolicy.canUseLaboratory(_accessTier)) {
      return;
    }

    final summary = _bridge.runLaboratory(
      matches: _labMatches.round(),
      ticksPerMatch: _labTicks.round(),
      seed: DateTime.now().microsecondsSinceEpoch & 0x7fffffff,
    );

    final updated = _progression.registerLaboratoryBatch(
      _careerProfile,
      'Laboratorio invisible premium',
      summary.matches,
      homeWinRate: summary.homeWinRate,
      awayWinRate: summary.awayWinRate,
      averageHomeXg: summary.averageHomeXg,
      averageAwayXg: summary.averageAwayXg,
      averageHomeLbs: summary.averageHomeLbs,
      averageAwayLbs: summary.averageAwayLbs,
      averageHomeSgm: summary.averageHomeSgm,
      averageAwaySgm: summary.averageAwaySgm,
      averageFatigue: summary.averageFatigue,
      averageTicks: summary.averageTicks,
    );

    setState(() {
      _labSummary = summary;
      _careerProfile = updated;
    });
    _careerStore.save(_careerProfile);
  }

  void _exportLaboratoryReport() {
    final file = _exportService.exportLaboratoryReport(
      profile: _careerProfile,
      accessTier: _accessTier,
      labMode: _labMode,
      laboratorySummary: _labSummary?.toJson(),
    );
    setState(() {
      _lastExportPath = file.path;
    });
    _showMessage('Reporte exportado en ${file.path}');
  }

  void _reset() {
    final restorePublishing = _labMode;
    setState(() {
      _bridge.reset();
      _bridge.applyConfig(_config);
      _bridge.setPlaybackRate(_playbackRate);
      _bridge.setFramePublishing(true);
      _frame = _bridge.readFrame();
      _renderFrame = _frame;
      _trajectory = <Offset>[];
      _lastElapsed = Duration.zero;
      _primeTickAfterReset = true;
      _running = true;
      _labMode = false;
      _labSummary = null;
    });
    if (restorePublishing) {
      _bridge.setFramePublishing(true);
    }
    _appendTrajectory(_frame);
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  String _playbackLabel(double rate) => '${rate.toStringAsFixed(1)}x';

  @override
  Widget build(BuildContext context) {
    final frame = _frame;
    final renderFrame = _renderFrame ?? frame;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Scoreador'),
        actions: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Center(
              child: Text(
                _bridge.isNative ? 'FFI Android' : 'Mock local',
                style: Theme.of(context).textTheme.labelMedium,
              ),
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: frame == null
            ? const Center(child: CircularProgressIndicator())
            : LayoutBuilder(
                builder: (context, constraints) {
                  final isWide = constraints.maxWidth >= 1100;
                  final effectiveFrame = renderFrame ?? frame;
                  final field = _buildField(effectiveFrame);
                  final panel = _buildPanel(frame);

                  if (isWide) {
                    return Row(
                      children: [
                        Expanded(flex: 5, child: field),
                        Expanded(flex: 4, child: panel),
                      ],
                    );
                  }

                  return ListView(
                    padding: const EdgeInsets.all(16),
                    children: [
                      SizedBox(height: 420, child: field),
                      const SizedBox(height: 16),
                      panel,
                    ],
                  );
                },
              ),
      ),
    );
  }

  Widget _buildField(EngineFrameView frame) {
    return Container(
      margin: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.35),
            blurRadius: 24,
            offset: const Offset(0, 12),
          ),
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          Positioned.fill(
            child: CustomPaint(
              painter: PitchPainter(frame: frame, trajectory: _trajectory),
            ),
          ),
          Positioned(
            left: 16,
            top: 16,
            child: _Chip(label: 'Tick ${frame.tick}'),
          ),
          Positioned(
            right: 16,
            top: 16,
            child: _Chip(
              label: frame.homePossession
                  ? 'Posesion local'
                  : 'Posesion visitante',
            ),
          ),
          Positioned(
            left: 16,
            bottom: 16,
            child: _Chip(
              label:
                  'Balon (${frame.ballX.toStringAsFixed(1)}, ${frame.ballY.toStringAsFixed(1)})',
            ),
          ),
          if (_labMode)
            Positioned.fill(
              child: Container(
                color: Colors.black.withValues(alpha: 0.45),
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.science, size: 48),
                      const SizedBox(height: 12),
                      Text(
                        'Modo Laboratorio activo',
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Se suspendio la publicacion de coordenadas a Flutter.',
                        style: Theme.of(context).textTheme.bodyMedium,
                        textAlign: TextAlign.center,
                      ),
                    ],
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildPanel(EngineFrameView frame) {
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _SectionCard(
            title: 'Acceso',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Version activa: ${_accessTier.label}',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    ChoiceChip(
                      label: const Text('Versión Básica'),
                      selected: _accessTier == AccessTier.basic,
                      onSelected: (_) => _setAccessTier(AccessTier.basic),
                    ),
                    ChoiceChip(
                      label: const Text('Versión Premium'),
                      selected: _accessTier == AccessTier.premium,
                      onSelected: _careerProfile.premiumUnlocked
                          ? (_) => _setAccessTier(AccessTier.premium)
                          : null,
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Text(
                  'Basico: visualizacion + 5x + analitica simple. Premium: analitica espacial y modo laboratorio.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                const SizedBox(height: 12),
                if (_careerProfile.premiumUnlocked)
                  Text(
                    'Premium local desbloqueado desde ${_careerProfile.premiumUnlockedAt != null ? _careerProfile.premiumUnlockedAt!.toLocal().toIso8601String() : 'la sesion actual'}.',
                    style: Theme.of(context).textTheme.bodySmall,
                  )
                else
                  FilledButton.icon(
                    onPressed: _unlockPremium,
                    icon: const Icon(Icons.workspace_premium),
                    label: const Text('Desbloquear Premium local'),
                  ),
                const SizedBox(height: 12),
                SwitchListTile.adaptive(
                  value: _labMode,
                  onChanged: FeatureAccessPolicy.canUseLaboratory(_accessTier)
                      ? _setLabMode
                      : null,
                  title: const Text('Modo Laboratorio'),
                  subtitle: Text(
                    FeatureAccessPolicy.canUseLaboratory(_accessTier)
                        ? 'Procesamiento invisible sin publicar coordenadas.'
                        : 'Disponible solo en la Version Premium.',
                  ),
                  contentPadding: EdgeInsets.zero,
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          _SectionCard(
            title: 'Carrera y retencion',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _careerProfile.title.label,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 6),
                Text(
                  _careerProfile.latestStory,
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _careerProfile.badges
                      .map(
                        (badge) => Tooltip(
                          message: badge.summary,
                          child: InputChip(
                            label: Text(badge.kind.label),
                            avatar: const Icon(Icons.verified, size: 16),
                            onPressed: () {},
                          ),
                        ),
                      )
                      .toList(growable: false),
                ),
                if (_careerProfile.badges.isEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(
                      'Aun no hay insignias desbloqueadas.',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  children: [
                    FilledButton.icon(
                      onPressed: _registerCurrentMatch,
                      icon: const Icon(Icons.bookmark_add),
                      label: const Text('Registrar partido'),
                    ),
                    OutlinedButton.icon(
                      onPressed: _runLaboratoryBatch,
                      icon: const Icon(Icons.science),
                      label: const Text('Ejecutar laboratorio'),
                    ),
                    OutlinedButton.icon(
                      onPressed: _exportLaboratoryReport,
                      icon: const Icon(Icons.download),
                      label: const Text('Exportar reporte'),
                    ),
                  ],
                ),
                if (_lastExportPath != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: SelectableText(
                      _lastExportPath!,
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
                const SizedBox(height: 12),
                SwitchListTile.adaptive(
                  value: _isDerby,
                  onChanged: (value) => setState(() {
                    _isDerby = value;
                  }),
                  title: const Text('Marcar como derbi'),
                  subtitle: const Text(
                    'Activa insignias y narrativa contextual para partidos locales.',
                  ),
                  contentPadding: EdgeInsets.zero,
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          _SectionCard(
            title: _FeatureAccessPolicyText.analyticsTitle(
              _accessTier,
            ),
            child: Column(
              children: [
                _MetricRow(
                  label: 'Tick',
                  value: frame.tick.toString(),
                ),
                _MetricRow(
                  label: 'Posesion',
                  value: frame.homePossession ? 'Local' : 'Visitante',
                ),
                _MetricRow(
                  label: 'Ultimo pase',
                  value: frame.lastPassCompleted ? 'Correcto' : 'Fallido',
                ),
                _MetricRow(
                  label: 'Fatiga media',
                  value: frame.averageFatigue.toStringAsFixed(3),
                ),
                _MetricRow(
                  label: 'Precision ultimo pase',
                  value: frame.lastPassAccuracy.toStringAsFixed(3),
                ),
                if (FeatureAccessPolicy.canSeeAdvancedAnalytics(_accessTier)) ...[
                  _MetricRow(
                    label: 'xG local',
                    value: frame.homeXg.toStringAsFixed(3),
                  ),
                  _MetricRow(
                    label: 'xG visitante',
                    value: frame.awayXg.toStringAsFixed(3),
                  ),
                  _MetricRow(
                    label: 'LBS local',
                    value: frame.homeLbs.toStringAsFixed(2),
                  ),
                  _MetricRow(
                    label: 'SGM local',
                    value: frame.homeSgm.toStringAsFixed(2),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: 16),
          _SectionCard(
            title: 'Control de tiempo',
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _SliderControl(
                  label: 'Velocidad de tiempo',
                  value: _playbackRate,
                  min: 1.0,
                  max: _effectivePlaybackMax,
                  divisions: ((_effectivePlaybackMax - 1.0) / 0.5).round(),
                  valueFormatter: _playbackLabel,
                  onChanged: _setPlaybackRate,
                ),
                const SizedBox(height: 8),
                Text(
                  'Limite activo: ${_effectivePlaybackMax.toStringAsFixed(1)}x.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          _SectionCard(
            title: 'Bloques tacticos',
            child: Material(
              color: Colors.white.withValues(alpha: 0.03),
              borderRadius: BorderRadius.circular(16),
              child: Column(
                children: [
                  TabBar(
                    controller: _tabController,
                    isScrollable: false,
                    tabs: _blockTitles
                        .map((title) => Tab(text: title))
                        .toList(growable: false),
                  ),
                  SizedBox(
                    height: 420,
                    child: TabBarView(
                      controller: _tabController,
                      children: List.generate(
                        _blockTitles.length,
                        (index) => _BlockTacticalEditor(
                          title: _blockTitles[index],
                          block: _config.blocks[index],
                          onChanged: (updated) => _updateBlock(index, updated),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          _SectionCard(
            title: 'Estado del motor',
            child: Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                FilledButton.icon(
                  onPressed: _labMode ? null : _toggleRunning,
                  icon: Icon(_running ? Icons.pause : Icons.play_arrow),
                  label: Text(_running ? 'Pausar' : 'Reanudar'),
                ),
                OutlinedButton.icon(
                  onPressed: _labMode
                      ? null
                      : () {
                          _bridge.step();
                          setState(() {
                            _frame = _bridge.readFrame();
                            _appendTrajectory(_frame);
                          });
                        },
                  icon: const Icon(Icons.skip_next),
                  label: const Text('Un tick'),
                ),
                OutlinedButton.icon(
                  onPressed: _reset,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Reiniciar'),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          if (_labSummary != null)
            _SectionCard(
              title: 'Laboratorio invisible',
              child: Column(
                children: [
                  _MetricRow(
                    label: 'Partidos procesados',
                    value: _labSummary!.matches.toString(),
                  ),
                  _MetricRow(
                    label: 'Victorias locales',
                    value:
                        '${_labSummary!.homeWins} (${(_labSummary!.homeWinRate * 100).toStringAsFixed(1)}%)',
                  ),
                  _MetricRow(
                    label: 'Victorias visitantes',
                    value:
                        '${_labSummary!.awayWins} (${(_labSummary!.awayWinRate * 100).toStringAsFixed(1)}%)',
                  ),
                  _MetricRow(
                    label: 'Empates',
                    value:
                        '${_labSummary!.draws} (${(_labSummary!.drawRate * 100).toStringAsFixed(1)}%)',
                  ),
                  _MetricRow(
                    label: 'xG medio',
                    value:
                        '${_labSummary!.averageHomeXg.toStringAsFixed(2)} / ${_labSummary!.averageAwayXg.toStringAsFixed(2)}',
                  ),
                  _MetricRow(
                    label: 'LBS medio',
                    value:
                        '${_labSummary!.averageHomeLbs.toStringAsFixed(2)} / ${_labSummary!.averageAwayLbs.toStringAsFixed(2)}',
                  ),
                  _MetricRow(
                    label: 'SGM medio',
                    value:
                        '${_labSummary!.averageHomeSgm.toStringAsFixed(2)} / ${_labSummary!.averageAwaySgm.toStringAsFixed(2)}',
                  ),
                  _MetricRow(
                    label: 'Fatiga media',
                    value: _labSummary!.averageFatigue.toStringAsFixed(3),
                  ),
                ],
              ),
            ),
          const SizedBox(height: 8),
          Text(
            _labMode
                ? 'Coordenadas suspendidas para la UI. El laboratorio se ejecuta sin pintar frames.'
                : 'La vista continua leyendo memoria nativa y pintando el partido en tiempo real.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ],
      ),
    );
  }
}

class _FeatureAccessPolicyText {
  const _FeatureAccessPolicyText._();

  static String analyticsTitle(AccessTier tier) =>
      FeatureAccessPolicy.canSeeAdvancedAnalytics(tier)
          ? 'Analitica espacial'
          : 'Analitica simple';
}

class _BlockTacticalEditor extends StatelessWidget {
  const _BlockTacticalEditor({
    required this.title,
    required this.block,
    required this.onChanged,
  });

  final String title;
  final BlockTacticalConfig block;
  final ValueChanged<BlockTacticalConfig> onChanged;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(title, style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: 8),
        Text(
          'Cinco parametros por bloque. Los cambios se envian al nucleo nativo en caliente.',
          style: Theme.of(context).textTheme.bodySmall,
        ),
        const SizedBox(height: 16),
        _SliderControl(
          label: 'Altura de linea',
          value: block.lineHeight,
          onChanged: (value) => onChanged(block.copyWith(lineHeight: value)),
        ),
        _SliderControl(
          label: 'Compactacion horizontal',
          value: block.compactness,
          onChanged: (value) => onChanged(block.copyWith(compactness: value)),
        ),
        _SliderControl(
          label: 'Agresividad de presion',
          value: block.pressureAggression,
          onChanged: (value) =>
              onChanged(block.copyWith(pressureAggression: value)),
        ),
        _SliderControl(
          label: 'Trampa del fuera de juego',
          value: block.offsideTrap,
          onChanged: (value) => onChanged(block.copyWith(offsideTrap: value)),
        ),
        _SliderControl(
          label: 'Cobertura',
          value: block.coverage,
          onChanged: (value) => onChanged(block.copyWith(coverage: value)),
        ),
      ],
    );
  }
}

class _SectionCard extends StatelessWidget {
  const _SectionCard({
    required this.title,
    required this.child,
  });

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.03),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 12),
            child,
          ],
        ),
      ),
    );
  }
}

class _MetricRow extends StatelessWidget {
  const _MetricRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Flexible(
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ),
          const SizedBox(width: 16),
          Text(value, style: Theme.of(context).textTheme.bodyLarge),
        ],
      ),
    );
  }
}

class _SliderControl extends StatelessWidget {
  const _SliderControl({
    required this.label,
    required this.value,
    required this.onChanged,
    this.min = 0.0,
    this.max = 1.0,
    this.divisions,
    this.valueFormatter,
  });

  final String label;
  final double value;
  final double min;
  final double max;
  final int? divisions;
  final String Function(double value)? valueFormatter;
  final ValueChanged<double> onChanged;

  @override
  Widget build(BuildContext context) {
    final clamped = value.clamp(min, max).toDouble();
    String formatValue(double input) =>
        valueFormatter?.call(input) ?? input.toStringAsFixed(2);
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('$label ${formatValue(clamped)}'),
          Slider(
            min: min,
            max: max,
            divisions: divisions,
            value: clamped,
            onChanged: onChanged,
          ),
        ],
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.black.withValues(alpha: 0.4),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: Colors.white.withValues(alpha: 0.15)),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        child: Text(label),
      ),
    );
  }
}

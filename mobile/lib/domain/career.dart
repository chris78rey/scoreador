import 'dart:convert';

enum AccessTier { basic, premium }

extension AccessTierPresentation on AccessTier {
  String get label {
    switch (this) {
      case AccessTier.basic:
        return 'Versión Básica';
      case AccessTier.premium:
        return 'Versión Premium';
    }
  }
}

class FeatureAccessPolicy {
  const FeatureAccessPolicy._();

  static bool canUseLaboratory(AccessTier tier) => tier == AccessTier.premium;

  static bool canSeeAdvancedAnalytics(AccessTier tier) =>
      tier == AccessTier.premium;

  static double maxPlaybackRate(AccessTier tier) =>
      tier == AccessTier.premium ? 10.0 : 5.0;
}

enum CareerTitle {
  analystNovice('Analista Novato'),
  tacticalObserver('Observador Tactico'),
  spatialArchitect('Arquitecto Espacial'),
  labStrategist('Estratega del Laboratorio'),
  labLegend('Leyenda del Laboratorio');

  const CareerTitle(this.label);

  final String label;
}

enum BadgeKind {
  lowFatigueWin('Victoria con fatiga baja'),
  winningStreak('Racha de victorias'),
  derbyWinner('Ganador de derbi'),
  xgDominance('Dominio xG'),
  spaceMaster('Maestria espacial'),
  labRun('Batalla de laboratorio');

  const BadgeKind(this.label);

  final String label;
}

class CareerBadge {
  const CareerBadge({
    required this.kind,
    required this.unlockedAt,
    required this.summary,
  });

  final BadgeKind kind;
  final DateTime unlockedAt;
  final String summary;

  Map<String, Object?> toJson() => <String, Object?>{
    'kind': kind.name,
    'unlockedAt': unlockedAt.toIso8601String(),
    'summary': summary,
  };

  factory CareerBadge.fromJson(Map<String, dynamic> json) {
    return CareerBadge(
      kind: BadgeKind.values.byName(json['kind'] as String),
      unlockedAt: DateTime.parse(json['unlockedAt'] as String),
      summary: json['summary'] as String? ?? '',
    );
  }
}

class MatchRecord {
  const MatchRecord({
    required this.opponent,
    required this.derby,
    required this.homeGoals,
    required this.awayGoals,
    required this.homeXg,
    required this.awayXg,
    required this.homeLbs,
    required this.awayLbs,
    required this.averageFatigue,
    required this.playedAt,
    this.laboratoryRun = false,
  });

  final String opponent;
  final bool derby;
  final int homeGoals;
  final int awayGoals;
  final double homeXg;
  final double awayXg;
  final double homeLbs;
  final double awayLbs;
  final double averageFatigue;
  final DateTime playedAt;
  final bool laboratoryRun;

  bool get homeWin => homeGoals > awayGoals;

  Map<String, Object?> toJson() => <String, Object?>{
    'opponent': opponent,
    'derby': derby,
    'homeGoals': homeGoals,
    'awayGoals': awayGoals,
    'homeXg': homeXg,
    'awayXg': awayXg,
    'homeLbs': homeLbs,
    'awayLbs': awayLbs,
    'averageFatigue': averageFatigue,
    'playedAt': playedAt.toIso8601String(),
    'laboratoryRun': laboratoryRun,
  };

  factory MatchRecord.fromJson(Map<String, dynamic> json) {
    return MatchRecord(
      opponent: json['opponent'] as String? ?? 'Rival',
      derby: json['derby'] as bool? ?? false,
      homeGoals: (json['homeGoals'] as num? ?? 0).toInt(),
      awayGoals: (json['awayGoals'] as num? ?? 0).toInt(),
      homeXg: (json['homeXg'] as num? ?? 0).toDouble(),
      awayXg: (json['awayXg'] as num? ?? 0).toDouble(),
      homeLbs: (json['homeLbs'] as num? ?? 0).toDouble(),
      awayLbs: (json['awayLbs'] as num? ?? 0).toDouble(),
      averageFatigue: (json['averageFatigue'] as num? ?? 0).toDouble(),
      playedAt: DateTime.parse(
        json['playedAt'] as String? ?? DateTime.now().toIso8601String(),
      ),
      laboratoryRun: json['laboratoryRun'] as bool? ?? false,
    );
  }
}

class CareerProfile {
  const CareerProfile({
    required this.title,
    required this.totalMatches,
    required this.winStreak,
    required this.bestWinStreak,
    required this.labRuns,
    required this.premiumUnlocked,
    required this.premiumUnlockedAt,
    required this.badges,
    required this.matches,
    required this.latestStory,
    required this.updatedAt,
  });

  factory CareerProfile.initial() {
    return CareerProfile(
      title: CareerTitle.analystNovice,
      totalMatches: 0,
      winStreak: 0,
      bestWinStreak: 0,
      labRuns: 0,
      premiumUnlocked: false,
      premiumUnlockedAt: null,
      badges: const <CareerBadge>[],
      matches: const <MatchRecord>[],
      latestStory: 'Comienza el laboratorio.',
      updatedAt: DateTime.now(),
    );
  }

  final CareerTitle title;
  final int totalMatches;
  final int winStreak;
  final int bestWinStreak;
  final int labRuns;
  final bool premiumUnlocked;
  final DateTime? premiumUnlockedAt;
  final List<CareerBadge> badges;
  final List<MatchRecord> matches;
  final String latestStory;
  final DateTime updatedAt;

  Map<String, Object?> toJson() => <String, Object?>{
    'title': title.name,
    'totalMatches': totalMatches,
    'winStreak': winStreak,
    'bestWinStreak': bestWinStreak,
    'labRuns': labRuns,
    'premiumUnlocked': premiumUnlocked,
    'premiumUnlockedAt': premiumUnlockedAt?.toIso8601String(),
    'badges': badges.map((badge) => badge.toJson()).toList(growable: false),
    'matches': matches.map((match) => match.toJson()).toList(growable: false),
    'latestStory': latestStory,
    'updatedAt': updatedAt.toIso8601String(),
  };

  factory CareerProfile.fromJson(Map<String, dynamic> json) {
    return CareerProfile(
      title: CareerTitle.values.byName(
        json['title'] as String? ?? CareerTitle.analystNovice.name,
      ),
      totalMatches: (json['totalMatches'] as num? ?? 0).toInt(),
      winStreak: (json['winStreak'] as num? ?? 0).toInt(),
      bestWinStreak: (json['bestWinStreak'] as num? ?? 0).toInt(),
      labRuns: (json['labRuns'] as num? ?? 0).toInt(),
      premiumUnlocked: json['premiumUnlocked'] as bool? ?? false,
      premiumUnlockedAt: json['premiumUnlockedAt'] == null
          ? null
          : DateTime.tryParse(json['premiumUnlockedAt'] as String) ??
              DateTime.now(),
      badges: (json['badges'] as List<dynamic>? ?? const <dynamic>[])
          .map((item) => CareerBadge.fromJson(item as Map<String, dynamic>))
          .toList(growable: false),
      matches: (json['matches'] as List<dynamic>? ?? const <dynamic>[])
          .map((item) => MatchRecord.fromJson(item as Map<String, dynamic>))
          .toList(growable: false),
      latestStory: json['latestStory'] as String? ?? 'Comienza el laboratorio.',
      updatedAt: DateTime.parse(
        json['updatedAt'] as String? ?? DateTime.now().toIso8601String(),
      ),
    );
  }

  String toPrettyJson() => const JsonEncoder.withIndent('  ').convert(toJson());

  CareerProfile copyWith({
    CareerTitle? title,
    int? totalMatches,
    int? winStreak,
    int? bestWinStreak,
    int? labRuns,
    bool? premiumUnlocked,
    DateTime? premiumUnlockedAt,
    List<CareerBadge>? badges,
    List<MatchRecord>? matches,
    String? latestStory,
    DateTime? updatedAt,
  }) {
    return CareerProfile(
      title: title ?? this.title,
      totalMatches: totalMatches ?? this.totalMatches,
      winStreak: winStreak ?? this.winStreak,
      bestWinStreak: bestWinStreak ?? this.bestWinStreak,
      labRuns: labRuns ?? this.labRuns,
      premiumUnlocked: premiumUnlocked ?? this.premiumUnlocked,
      premiumUnlockedAt: premiumUnlockedAt ?? this.premiumUnlockedAt,
      badges: badges ?? this.badges,
      matches: matches ?? this.matches,
      latestStory: latestStory ?? this.latestStory,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}

class CareerProgressionService {
  const CareerProgressionService();

  CareerProfile registerMatch(CareerProfile current, MatchRecord match) {
    final history = List<MatchRecord>.from(current.matches)..add(match);
    var winStreak = current.winStreak;
    var bestWinStreak = current.bestWinStreak;
    if (match.homeWin) {
      winStreak += 1;
      if (winStreak > bestWinStreak) {
        bestWinStreak = winStreak;
      }
    } else {
      winStreak = 0;
    }

    final badges = List<CareerBadge>.from(current.badges);
    _maybeAddBadge(
      badges,
      BadgeKind.lowFatigueWin,
      match.homeWin && match.averageFatigue <= 0.25,
      'Victoria con fatiga media ${match.averageFatigue.toStringAsFixed(2)}.',
    );
    _maybeAddBadge(
      badges,
      BadgeKind.winningStreak,
      winStreak >= 3,
      'Racha activa de $winStreak victorias.',
    );
    _maybeAddBadge(
      badges,
      BadgeKind.derbyWinner,
      match.derby && match.homeWin,
      'Derbi controlado ante ${match.opponent}.',
    );
    _maybeAddBadge(
      badges,
      BadgeKind.xgDominance,
      match.homeXg > match.awayXg && match.homeWin,
      'Ventaja xG (${match.homeXg.toStringAsFixed(2)} vs ${match.awayXg.toStringAsFixed(2)}).',
    );
    _maybeAddBadge(
      badges,
      BadgeKind.spaceMaster,
      match.homeLbs >= match.awayLbs && match.homeXg >= match.awayXg,
      'Dominio espacial y de lineas.',
    );

    final title = _resolveTitle(
      totalMatches: current.totalMatches + 1,
      bestWinStreak: bestWinStreak,
      labRuns: current.labRuns,
      badgeCount: badges.length,
    );

    return current.copyWith(
      title: title,
      totalMatches: current.totalMatches + 1,
      winStreak: winStreak,
      bestWinStreak: bestWinStreak,
      badges: badges,
      matches: history,
      latestStory: _buildStory(
        title: title,
        match: match,
        winStreak: winStreak,
        bestWinStreak: bestWinStreak,
        badgeCount: badges.length,
      ),
      updatedAt: DateTime.now(),
    );
  }

  CareerProfile registerLaboratoryBatch(
    CareerProfile current,
    String label,
    int matches,
    {
    required double homeWinRate,
    required double awayWinRate,
    required double averageHomeXg,
    required double averageAwayXg,
    required double averageHomeLbs,
    required double averageAwayLbs,
    required double averageHomeSgm,
    required double averageAwaySgm,
    required double averageFatigue,
    required double averageTicks,
  }
  ) {
    final badges = List<CareerBadge>.from(current.badges);
    _maybeAddBadge(
      badges,
      BadgeKind.labRun,
      true,
      'Lote de laboratorio ejecutado: $matches partidos.',
    );

    final labRecord = MatchRecord(
      opponent: label,
      derby: false,
      homeGoals: (averageHomeXg * 4).round(),
      awayGoals: (averageAwayXg * 4).round(),
      homeXg: averageHomeXg,
      awayXg: averageAwayXg,
      homeLbs: averageHomeLbs,
      awayLbs: averageAwayLbs,
      averageFatigue: averageFatigue,
      playedAt: DateTime.now(),
      laboratoryRun: true,
    );

    final history = List<MatchRecord>.from(current.matches)..add(labRecord);
    final labRuns = current.labRuns + 1;
    final title = _resolveTitle(
      totalMatches: current.totalMatches,
      bestWinStreak: current.bestWinStreak,
      labRuns: labRuns,
      badgeCount: badges.length,
    );

    return current.copyWith(
      title: title,
      labRuns: labRuns,
      badges: badges,
      matches: history,
      latestStory: _buildLabStory(
        title: title,
        label: label,
        matches: matches,
        homeWinRate: homeWinRate,
        awayWinRate: awayWinRate,
        averageHomeXg: averageHomeXg,
        averageAwayXg: averageAwayXg,
        averageHomeLbs: averageHomeLbs,
        averageAwayLbs: averageAwayLbs,
        averageHomeSgm: averageHomeSgm,
        averageAwaySgm: averageAwaySgm,
        averageFatigue: averageFatigue,
        averageTicks: averageTicks,
        badgeCount: badges.length,
      ),
      updatedAt: DateTime.now(),
    );
  }

  CareerProfile unlockPremium(
    CareerProfile current, {
    DateTime? unlockedAt,
  }) {
    final at = unlockedAt ?? DateTime.now();
    return current.copyWith(
      premiumUnlocked: true,
      premiumUnlockedAt: at,
      latestStory:
          'Premium local activado. El laboratorio invisible ya puede exportar y analizar sin limites visuales.',
      updatedAt: at,
    );
  }

  CareerTitle _resolveTitle({
    required int totalMatches,
    required int bestWinStreak,
    required int labRuns,
    required int badgeCount,
  }) {
    final activity = totalMatches + labRuns;
    if (labRuns >= 6 || badgeCount >= 7 || bestWinStreak >= 5) {
      return CareerTitle.labLegend;
    }
    if (labRuns >= 3 || badgeCount >= 5 || bestWinStreak >= 4) {
      return CareerTitle.labStrategist;
    }
    if (badgeCount >= 3 || bestWinStreak >= 3 || activity >= 8) {
      return CareerTitle.spatialArchitect;
    }
    if (activity >= 2 || badgeCount >= 1) {
      return CareerTitle.tacticalObserver;
    }
    return CareerTitle.analystNovice;
  }

  String _buildStory({
    required CareerTitle title,
    required MatchRecord match,
    required int winStreak,
    required int bestWinStreak,
    required int badgeCount,
  }) {
    final result = match.homeWin ? 'victoria' : 'derrota';
    final derbyText = match.derby ? ' en derbi' : '';
    final streakText = winStreak > 1 ? 'Racha actual de $winStreak.' : '';
    final fatigueText =
        'Fatiga media ${match.averageFatigue.toStringAsFixed(2)}.';
    final badgeText = badgeCount > 0 ? '$badgeCount insignias activas.' : '';
    return [
      'Partido con $result$derbyText ante ${match.opponent}.',
      fatigueText,
      streakText,
      badgeText,
      'Titulo: ${title.label}. Mejor racha: $bestWinStreak.',
    ].where((part) => part.isNotEmpty).join(' ');
  }

  String _buildLabStory({
    required CareerTitle title,
    required String label,
    required int matches,
    required double homeWinRate,
    required double awayWinRate,
    required double averageHomeXg,
    required double averageAwayXg,
    required double averageHomeLbs,
    required double averageAwayLbs,
    required double averageHomeSgm,
    required double averageAwaySgm,
    required double averageFatigue,
    required double averageTicks,
    required int badgeCount,
  }) {
    return [
      'Laboratorio "$label" ejecutado con $matches partidos.',
      'Tasa de victoria local ${(homeWinRate * 100).toStringAsFixed(1)}%.',
      'xG medio ${averageHomeXg.toStringAsFixed(2)} vs ${averageAwayXg.toStringAsFixed(2)}.',
      'LBS medio ${averageHomeLbs.toStringAsFixed(2)} vs ${averageAwayLbs.toStringAsFixed(2)}.',
      'SGM medio ${averageHomeSgm.toStringAsFixed(2)} vs ${averageAwaySgm.toStringAsFixed(2)}.',
      'Fatiga media ${averageFatigue.toStringAsFixed(2)} en ${averageTicks.toStringAsFixed(0)} ticks.',
      'Victoria visitante ${(awayWinRate * 100).toStringAsFixed(1)}%.',
      'Titulo actual ${title.label}.',
      '$badgeCount insignias activas.',
    ].join(' ');
  }

  void _maybeAddBadge(
    List<CareerBadge> badges,
    BadgeKind kind,
    bool enabled,
    String summary,
  ) {
    if (!enabled || badges.any((badge) => badge.kind == kind)) {
      return;
    }
    badges.add(
      CareerBadge(
        kind: kind,
        unlockedAt: DateTime.now(),
        summary: summary,
      ),
    );
  }
}

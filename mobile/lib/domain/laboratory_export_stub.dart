import 'dart:convert';

import 'career.dart';

class LaboratoryReportFile {
  LaboratoryReportFile(this.path, this.contents);

  final String path;
  final String contents;
}

class CareerExportService {
  CareerExportService({Object? baseDirectory});

  static final Map<String, String> _reports = <String, String>{};

  LaboratoryReportFile exportLaboratoryReport({
    required CareerProfile profile,
    required AccessTier accessTier,
    required bool labMode,
    Map<String, Object?>? laboratorySummary,
  }) {
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final path = 'memory://scoreador/exports/lab_report_$timestamp.json';
    final payload = <String, Object?>{
      'generatedAt': DateTime.now().toIso8601String(),
      'accessTier': accessTier.name,
      'labMode': labMode,
      'profile': profile.toJson(),
      'laboratorySummary': laboratorySummary,
    };
    final contents = const JsonEncoder.withIndent('  ').convert(payload);
    _reports[path] = contents;
    return LaboratoryReportFile(path, contents);
  }

  static String? readReport(String path) => _reports[path];
}

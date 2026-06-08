import 'dart:convert';
import 'dart:io';

import 'career.dart';

class CareerExportService {
  CareerExportService({Directory? baseDirectory})
      : _baseDirectory = baseDirectory ??
            Directory(
              '${Directory.systemTemp.path}${Platform.pathSeparator}scoreador',
            );

  final Directory _baseDirectory;

  Directory get _exportDirectory => Directory(
        '${_baseDirectory.path}${Platform.pathSeparator}exports',
      );

  File exportLaboratoryReport({
    required CareerProfile profile,
    required AccessTier accessTier,
    required bool labMode,
    Map<String, Object?>? laboratorySummary,
  }) {
    if (!_exportDirectory.existsSync()) {
      _exportDirectory.createSync(recursive: true);
    }

    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final file = File(
      '${_exportDirectory.path}${Platform.pathSeparator}lab_report_$timestamp.json',
    );

    final payload = <String, Object?>{
      'generatedAt': DateTime.now().toIso8601String(),
      'accessTier': accessTier.name,
      'labMode': labMode,
      'profile': profile.toJson(),
      'laboratorySummary': laboratorySummary,
    };

    file.writeAsStringSync(const JsonEncoder.withIndent('  ').convert(payload));
    return file;
  }
}

import 'dart:convert';
import 'dart:io';

import 'career.dart';

class LocalCareerStore {
  LocalCareerStore({Directory? baseDirectory})
      : _baseDirectory = baseDirectory ??
            Directory('${Directory.systemTemp.path}${Platform.pathSeparator}scoreador');

  final Directory _baseDirectory;

  File get _file => File(
        '${_baseDirectory.path}${Platform.pathSeparator}career_profile.json',
      );

  CareerProfile load() {
    try {
      if (!_file.existsSync()) {
        return CareerProfile.initial();
      }
      final raw = _file.readAsStringSync();
      if (raw.trim().isEmpty) {
        return CareerProfile.initial();
      }
      final decoded = jsonDecode(raw);
      if (decoded is! Map<String, dynamic>) {
        return CareerProfile.initial();
      }
      return CareerProfile.fromJson(decoded);
    } catch (_) {
      return CareerProfile.initial();
    }
  }

  void save(CareerProfile profile) {
    try {
      if (!_baseDirectory.existsSync()) {
        _baseDirectory.createSync(recursive: true);
      }
      _file.writeAsStringSync(profile.toPrettyJson());
    } catch (_) {
      // Persistencia local opcional: si falla, la UI sigue operando.
    }
  }

  void clear() {
    try {
      if (_file.existsSync()) {
        _file.deleteSync();
      }
    } catch (_) {
      // Ignorado a proposito.
    }
  }
}

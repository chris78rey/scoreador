import 'career.dart';

class LocalCareerStore {
  LocalCareerStore({Object? baseDirectory});

  static CareerProfile? _cachedProfile;

  CareerProfile load() => _cachedProfile ?? CareerProfile.initial();

  void save(CareerProfile profile) {
    _cachedProfile = profile;
  }

  void clear() {
    _cachedProfile = null;
  }
}

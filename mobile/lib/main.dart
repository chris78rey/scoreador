import 'package:flutter/material.dart';

import 'ui/simulator_screen.dart';

void main() {
  runApp(const ScoreadorApp());
}

class ScoreadorApp extends StatelessWidget {
  const ScoreadorApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Scoreador',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2E7D32),
          brightness: Brightness.dark,
        ),
        scaffoldBackgroundColor: const Color(0xFF08120C),
        useMaterial3: true,
      ),
      home: const SimulatorScreen(),
    );
  }
}

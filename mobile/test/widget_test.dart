import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:scoreador/main.dart';

void main() {
  testWidgets('La pantalla principal expone acceso y laboratorio', (
    tester,
  ) async {
    await tester.pumpWidget(const ScoreadorApp());
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('Scoreador'), findsOneWidget);
    expect(find.text('Acceso'), findsOneWidget);
    expect(find.text('Modo Laboratorio'), findsOneWidget);
    expect(find.text('Versión Básica'), findsWidgets);
    expect(find.text('Desbloquear Premium local'), findsOneWidget);
  });

  testWidgets('La vista muestra panel de analitica y control de tiempo', (
    tester,
  ) async {
    await tester.pumpWidget(const ScoreadorApp());
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.textContaining('Analitica'), findsWidgets);
    expect(find.text('Control de tiempo'), findsOneWidget);
    expect(find.text('Bloques tacticos'), findsOneWidget);
    expect(find.text('Exportar reporte'), findsOneWidget);
  });
}

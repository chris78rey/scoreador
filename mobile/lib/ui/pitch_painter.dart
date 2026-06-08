import 'dart:ui';

import 'package:flutter/material.dart';

import '../native/engine_bridge.dart';
import '../native/engine_geometry.dart' as geom;

class PitchPainter extends CustomPainter {
  const PitchPainter({required this.frame, required this.trajectory});

  final EngineFrameView frame;
  final List<Offset> trajectory;

  @override
  void paint(Canvas canvas, Size size) {
    final pitchPaint = Paint()..color = const Color(0xFF0E5A2E);
    final stripePaint = Paint()..color = const Color(0xFF0C4F27);
    final linePaint = Paint()
      ..color = Colors.white.withValues(alpha: 0.9)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.4;
    final pressureFill = Paint()..style = PaintingStyle.fill;
    final pressureStroke = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.2;

    final pitch = RRect.fromRectAndRadius(
      Offset.zero & size,
      const Radius.circular(18),
    );
    canvas.drawRRect(pitch, pitchPaint);

    for (var i = 0; i < 6; i++) {
      final top = size.height * (i / 6);
      canvas.drawRect(
        Rect.fromLTWH(0, top, size.width, size.height / 12),
        Paint()
          ..color = stripePaint.color.withValues(alpha: i.isEven ? 0.12 : 0.05),
      );
    }

    final centerX = size.width / 2;
    canvas.drawLine(
      Offset(centerX, 0),
      Offset(centerX, size.height),
      linePaint,
    );
    canvas.drawCircle(
      Offset(centerX, size.height / 2),
      size.shortestSide * 0.12,
      linePaint,
    );
    canvas.drawRect(Offset.zero & size, linePaint);

    final trajectoryPath = Path();
    for (var i = 0; i < trajectory.length; i++) {
      final point = _toCanvasOffset(trajectory[i].dx, trajectory[i].dy, size);
      if (i == 0) {
        trajectoryPath.moveTo(point.dx, point.dy);
      } else {
        trajectoryPath.lineTo(point.dx, point.dy);
      }
    }
    if (trajectory.length > 1) {
      canvas.drawPath(
        trajectoryPath,
        Paint()
          ..color = const Color(0xFFFFF59D).withValues(alpha: 0.7)
          ..style = PaintingStyle.stroke
          ..strokeWidth = 2.3
          ..strokeCap = StrokeCap.round
          ..strokeJoin = StrokeJoin.round,
      );
    }

    for (final block in frame.blocks) {
      final isHome = block.isHome;
      final baseColor = isHome
          ? const Color(0xFF80DEEA)
          : const Color(0xFFFFAB91);
      final offset = _toCanvasOffset(block.x, block.y, size);
      final zoneRadius =
          22.0 + (block.pressure * 38.0) + (block.coverage * 16.0);
      pressureFill.color = baseColor.withValues(alpha: 0.08);
      pressureStroke.color = baseColor.withValues(alpha: 0.28);
      canvas.drawCircle(offset, zoneRadius, pressureFill);
      canvas.drawCircle(offset, zoneRadius, pressureStroke);
      canvas.drawCircle(
        offset,
        4.2,
        Paint()..color = baseColor.withValues(alpha: 0.95),
      );
    }

    for (final player in frame.players) {
      final isHome = player.team == EngineTeam.home;
      final color = isHome ? const Color(0xFF69F0AE) : const Color(0xFFFFA726);
      final offset = _toCanvasOffset(player.x, player.y, size);
      canvas.drawCircle(offset, 6.2, Paint()..color = color);
      canvas.drawCircle(
        offset,
        6.2,
        Paint()
          ..color = Colors.black.withValues(alpha: 0.28)
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1,
      );
    }

    final ballOffset = _toCanvasOffset(frame.ballX, frame.ballY, size);
    canvas.drawCircle(
      ballOffset,
      4.6,
      Paint()..color = const Color(0xFFF5F5F5),
    );
    canvas.drawCircle(
      ballOffset,
      4.6,
      Paint()
        ..color = Colors.black.withValues(alpha: 0.35)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 1,
    );
  }

  Offset _toCanvasOffset(double x, double y, Size size) {
    return Offset(
      (x / geom.kPitchLength) * size.width,
      (y / geom.kPitchWidth) * size.height,
    );
  }

  @override
  bool shouldRepaint(covariant PitchPainter oldDelegate) {
    return oldDelegate.frame.tick != frame.tick ||
        oldDelegate.frame.ballX != frame.ballX ||
        oldDelegate.frame.ballY != frame.ballY ||
        oldDelegate.frame.blocks != frame.blocks ||
        oldDelegate.frame.players != frame.players ||
        oldDelegate.trajectory != trajectory;
  }
}

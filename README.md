# Scoreador

Arquitectura base para un simulador tactico desacoplado entre Flutter/Dart y un motor nativo en Go.

## Estructura

- `mobile/`: app Flutter con la interfaz reactiva, sliders de configuracion y render 2D.
- `engine/`: motor de simulacion en Go, preparado para compilarse como libreria compartida con FFI.

## Objetivo tecnico

La UI no calcula la simulacion. Solo:

- captura parametros tacticos,
- renderiza el campo y los jugadores,
- lee el snapshot actual del motor.

El motor Go se encarga de:

- ticks de simulacion,
- movimiento basico de jugadores y balon,
- fatiga,
- estimacion simple de posesion y xG.

## Fase 2: puente en memoria

Ya existe un wrapper CGO en Go que expone una `EngineSession` con memoria nativa continua para:

- coordenadas de 22 jugadores,
- posicion y velocidad del balon,
- métricas del tick actual,
- acceso directo desde Dart FFI sin JSON.

La ruta nativa en Flutter lee directamente `session.frame` desde RAM y convierte ese bloque en un `EngineFrameView` para renderizar cada frame.

### Build nativo Android

Desde `engine/cmd/engine`:

```bash
GOOS=android GOARCH=arm64 go build -buildmode=c-shared -o libengine.so main.go
```

En Android, la UI usa ese `libengine.so`. En pruebas locales fuera de Android se usa un mock para no depender del binario nativo.

## Siguiente paso recomendado

1. Compilar el motor Go como `libengine` para Android/iOS.
2. Copiar la libreria nativa al paquete Flutter.
3. Enlazar el build nativo con los targets Android e iOS.
4. Sustituir la heuristica de simulacion por tu modelo tactico real.

# Simulador Montecarlo de Torneos

Motor Go para simular torneos de futbol usando Montecarlo y distribucion de Poisson.

El sistema es reutilizable para Mundial, Copa America, Champions, LigaPro u otros torneos. Solo cambian los archivos de entrada: configuracion JSON, tabla lambda CSV y partidos CSV.

## Entrada

- `config.json`: reglas del torneo.
- `matches.csv`: partidos de fase de grupos.
- `lambda.csv`: tabla generica de conversion de tiros + motivacion a lambda.

## Ejecucion

```bash
go run ./cmd/tournament-sim -config examples/demo/config.json -matches examples/demo/matches.csv -lambda examples/demo/lambda.csv -outdir out -seed 42
```

## Salida

- `summary.csv`
- `summary.json`

## Notas

- El sistema usa la fase de grupos para clasificar equipos.
- La fase eliminatoria se construye de forma dinamica a partir de los clasificados.
- Si hay empate en knockout, el desempate usa penales simulados por defecto o un desempate aleatorio configurable.

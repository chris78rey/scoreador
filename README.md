# Simulador Montecarlo de Torneos

Motor Go para simular torneos de futbol usando Montecarlo y distribucion de Poisson.

El sistema es reutilizable para Mundial, Copa America, Champions, LigaPro u otros torneos. Solo cambian los archivos de entrada: configuracion JSON, tabla lambda CSV y partidos CSV.

El motor carga y simula solo la primera fase. La continuidad del torneo se arma por pestañas en Excel para cada fase.

## Entrada

- `config.json`: reglas del torneo.
- `matches.csv`: partidos de fase de grupos.
- `lambda.csv`: tabla generica de conversion de tiros + motivacion a lambda.

## Ejecucion

Modo generico:

```bash
go run ./cmd/tournament-sim -config examples/demo/config.json -matches examples/demo/matches.csv -lambda examples/demo/lambda.csv -outdir out -seed 42
```

Preset Mundial 2026:

```bash
go run ./cmd/tournament-sim -preset worldcup2026 -outdir out -seed 42
```

GUI local para partido unico con SQLite:

```bash
go run ./cmd/tournament-gui -addr :8081 -db out/single_matches.db -lambda examples/demo/lambda.csv
```

Luego abre `http://localhost:8081` en el navegador.

Plantilla Excel para cargar un grupo completo y calcular la tabla de posiciones desde marcadores manuales:

```bash
go run ./cmd/tournament-sim -group-template -group-name "Grupo A" -group-teams "Ecuador,Alemania,Polonia,Corea" -outdir out
```

Esto genera un archivo como `out/grupo_grupo_a_template.xlsx` con hojas para:

- `Partidos`
- `Tabla Base`
- `Posiciones`
- `Resumen`

## Salida

- `summary.csv`
- `summary.json`
- `summary.xlsx`

La GUI guarda cada corrida en SQLite en `out/single_matches.db` por defecto.

En el preset `worldcup2026`, el Excel incluye:

- `Formato 2026`
- `Grupos`
- `Grupo A` a `Grupo L`
- `Clasificados`
- `Dieciseisavos`
- `Octavos`
- `Cuartos`
- `Semis`
- `Final`

## Notas

- El sistema usa la fase de grupos para clasificar equipos.
- La fase eliminatoria queda preparada como plantilla por fases en Excel.
- Si hay empate en las hojas de fase eliminatoria, el desempate se completa manualmente o con formulas en Excel.

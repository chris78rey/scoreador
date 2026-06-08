package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"scoreador/internal/loader"
	"scoreador/internal/report"
	"scoreador/internal/sim"
)

func main() {
	var (
		configPath  = flag.String("config", "", "Ruta al archivo JSON de configuracion")
		matchesPath = flag.String("matches", "", "Ruta al archivo CSV de partidos")
		lambdaPath  = flag.String("lambda", "", "Ruta al archivo CSV de tabla lambda")
		outDir      = flag.String("outdir", "out", "Directorio de salida")
		seedFlag    = flag.Int64("seed", 0, "Semilla opcional para reproducibilidad")
	)
	flag.Parse()

	if *configPath == "" || *matchesPath == "" || *lambdaPath == "" {
		fmt.Fprintln(os.Stderr, "Uso: go run ./cmd/tournament-sim -config config.json -matches matches.csv -lambda lambda.csv -outdir out [-seed 123]")
		os.Exit(1)
	}

	cfg, err := loader.LoadConfig(*configPath)
	if err != nil {
		fatal(err)
	}
	if *seedFlag != 0 {
		cfg.Seed = *seedFlag
	}
	cfg.ApplyDefaults()
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	matches, err := loader.LoadMatches(*matchesPath)
	if err != nil {
		fatal(err)
	}

	lambdaRules, err := loader.LoadLambdaRules(*lambdaPath)
	if err != nil {
		fatal(err)
	}

	summary, err := sim.RunMonteCarlo(cfg, matches, lambdaRules)
	if err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal(err)
	}

	csvPath := filepath.Join(*outDir, "summary.csv")
	jsonPath := filepath.Join(*outDir, "summary.json")

	if err := report.WriteCSV(csvPath, summary); err != nil {
		fatal(err)
	}
	if err := report.WriteJSON(jsonPath, summary); err != nil {
		fatal(err)
	}

	fmt.Printf("Simulacion completada: %s\n", summary.Name)
	fmt.Printf("CSV: %s\n", csvPath)
	fmt.Printf("JSON: %s\n", jsonPath)
	fmt.Printf("Semilla: %d\n", cfg.Seed)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

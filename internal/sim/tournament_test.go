package sim

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scoreador/internal/loader"
	"scoreador/internal/model"
	"scoreador/internal/poisson"
	"scoreador/internal/report"
)

const sampleConfig = `{
  "name": "Torneo Demo",
  "simulations": 250,
  "groups": 2,
  "teams_per_group": 4,
  "qualified_per_group": 2,
  "best_thirds": 0,
  "knockout": true,
  "knockout_tiebreaker": "penalties",
  "seed": 42
}`

const sampleLambda = `tiros_min,tiros_max,motivacion,lambda
0,1,baja,0.2
0,1,media,0.4
0,1,alta,0.7
2,3,baja,0.5
2,3,media,0.9
2,3,alta,1.3
4,5,baja,0.8
4,5,media,1.4
4,5,alta,1.9
6,7,baja,1.1
6,7,media,1.8
6,7,alta,2.4
8,9,baja,1.5
8,9,media,2.3
8,9,alta,3.0
10,99,baja,2.0
10,99,media,2.8
10,99,alta,3.6
`

const sampleMatches = `match_id,stage,grupo,equipo_a,equipo_b,tiros_a,tiros_b,motivacion_a,motivacion_b
1,group,A,Equipo A,Equipo B,7,10,alta,media
2,group,A,Equipo C,Equipo D,8,3,media,baja
3,group,A,Equipo A,Equipo C,6,6,media,media
4,group,A,Equipo B,Equipo D,5,4,baja,media
5,group,A,Equipo A,Equipo D,9,2,alta,baja
6,group,A,Equipo B,Equipo C,4,7,media,alta
7,group,B,Equipo E,Equipo F,7,6,alta,media
8,group,B,Equipo G,Equipo H,8,3,media,baja
9,group,B,Equipo E,Equipo G,6,6,media,media
10,group,B,Equipo F,Equipo H,5,4,baja,media
11,group,B,Equipo E,Equipo H,9,2,alta,baja
12,group,B,Equipo F,Equipo G,4,7,media,alta
`

func TestMonteCarloPipeline(t *testing.T) {
	dir := t.TempDir()

	configPath := writeTempFile(t, dir, "config.json", sampleConfig)
	lambdaPath := writeTempFile(t, dir, "lambda.csv", sampleLambda)
	matchesPath := writeTempFile(t, dir, "matches.csv", sampleMatches)

	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	matches, err := loader.LoadMatches(matchesPath)
	if err != nil {
		t.Fatalf("load matches: %v", err)
	}
	rules, err := loader.LoadLambdaRules(lambdaPath)
	if err != nil {
		t.Fatalf("load lambda: %v", err)
	}

	summary, err := RunMonteCarlo(cfg, matches, rules)
	if err != nil {
		t.Fatalf("run montecarlo: %v", err)
	}
	if summary.Name != "Torneo Demo" {
		t.Fatalf("name mismatch: %s", summary.Name)
	}
	if len(summary.Teams) != 8 {
		t.Fatalf("expected 8 teams, got %d", len(summary.Teams))
	}

	totalChampions := 0
	for _, team := range summary.Teams {
		totalChampions += team.Campeon
		if team.Simulations != 250 {
			t.Fatalf("unexpected simulation count for %s: %d", team.Team, team.Simulations)
		}
	}
	if totalChampions != 250 {
		t.Fatalf("champion counts should sum to simulations, got %d", totalChampions)
	}

	csvPath := filepath.Join(dir, "summary.csv")
	jsonPath := filepath.Join(dir, "summary.json")
	if err := report.WriteCSV(csvPath, summary); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if err := report.WriteJSON(jsonPath, summary); err != nil {
		t.Fatalf("write json: %v", err)
	}

	csvData, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(csvData), "porcentaje_campeon") {
		t.Fatalf("csv header missing champion column")
	}

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded model.TournamentSummary
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if decoded.Simulations != 250 {
		t.Fatalf("json simulations mismatch: %d", decoded.Simulations)
	}
}

func TestLambdaLookupAndPoissonDeterminism(t *testing.T) {
	rules, err := loader.LoadLambdaRules(writeTempFile(t, t.TempDir(), "lambda.csv", sampleLambda))
	if err != nil {
		t.Fatalf("load lambda: %v", err)
	}
	lambda, err := lookupLambda(rules, 7, model.MotivationHigh)
	if err != nil {
		t.Fatalf("lookup lambda: %v", err)
	}
	if lambda != 2.4 {
		t.Fatalf("expected 2.4, got %v", lambda)
	}
	if got := runPoissonOnce(42, 0); got != 0 {
		t.Fatalf("lambda 0 should yield 0, got %d", got)
	}
}

func runPoissonOnce(seed int64, lambda float64) int {
	rng := randSource(seed)
	return poisson.Sample(rng, lambda)
}

func randSource(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

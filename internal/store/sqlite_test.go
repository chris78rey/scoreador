package store

import (
	"os"
	"path/filepath"
	"testing"

	"scoreador/internal/loader"
	"scoreador/internal/model"
	"scoreador/internal/sim"
)

const testLambda = `tiros_min,tiros_max,motivacion,lambda
0,99,baja,0.5
0,99,media,1.0
0,99,alta,1.5
`

func TestSQLiteStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "runs.db")
	lambdaPath := filepath.Join(dir, "lambda.csv")
	if err := os.WriteFile(lambdaPath, []byte(testLambda), 0o644); err != nil {
		t.Fatalf("write lambda: %v", err)
	}

	rules, err := loader.LoadLambdaRules(lambdaPath)
	if err != nil {
		t.Fatalf("load lambda: %v", err)
	}

	db, err := OpenSQLite(dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	}()

	input := sim.SingleMatchInput{
		TeamA:       "Equipo A",
		TeamB:       "Equipo B",
		ShotsA:      10,
		ShotsB:      4,
		MotivationA: model.MotivationHigh,
		MotivationB: model.MotivationMedium,
		Tiebreaker:  "penalties",
	}
	summary, err := sim.RunSingleMatchSeries(42, 10, rules, input)
	if err != nil {
		t.Fatalf("run series: %v", err)
	}

	record, err := db.SaveSingleMatchRun(42, lambdaPath, input, summary)
	if err != nil {
		t.Fatalf("save run: %v", err)
	}
	if record.ID == 0 {
		t.Fatalf("expected inserted id")
	}

	runs, err := db.ListSingleMatchRuns(10)
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].TeamA != input.TeamA || runs[0].TeamB != input.TeamB {
		t.Fatalf("unexpected run data: %+v", runs[0])
	}
	if runs[0].Simulations != 10 {
		t.Fatalf("unexpected simulations: %d", runs[0].Simulations)
	}
}

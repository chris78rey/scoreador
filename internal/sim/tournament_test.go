package sim

import (
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"

	"scoreador/internal/loader"
	"scoreador/internal/model"
	"scoreador/internal/poisson"
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

func TestLookupLambdaUsesClosestMotivationWhenExactRuleIsMissing(t *testing.T) {
	rules := []model.LambdaRule{
		{ShotsMin: 0, ShotsMax: 10, Motivation: model.MotivationLow, Lambda: 0.5},
		{ShotsMin: 0, ShotsMax: 10, Motivation: model.MotivationMedium, Lambda: 1.5},
		{ShotsMin: 0, ShotsMax: 10, Motivation: model.MotivationHigh, Lambda: 2.5},
	}

	lambda, err := lookupLambda(rules, 7, model.Motivation(7))
	if err != nil {
		t.Fatalf("lookupLambda returned error: %v", err)
	}
	if lambda != 1.5 {
		t.Fatalf("lookupLambda(7, 7) = %v, want %v", lambda, 1.5)
	}
}

func assertFormula(t *testing.T, wb *excelize.File, sheet, cell, want string) {
	t.Helper()
	got, err := wb.GetCellFormula(sheet, cell)
	if err != nil {
		t.Fatalf("get formula %s!%s: %v", sheet, cell, err)
	}
	if got != want {
		t.Fatalf("formula mismatch %s!%s: got %q want %q", sheet, cell, got, want)
	}
}

func assertCellStyleSame(t *testing.T, wb *excelize.File, sheet, cellA, cellB string) {
	t.Helper()
	styleA := mustCellStyle(t, wb, sheet, cellA)
	styleB := mustCellStyle(t, wb, sheet, cellB)
	if styleA != styleB {
		t.Fatalf("style mismatch %s!%s vs %s!%s: %d != %d", sheet, cellA, sheet, cellB, styleA, styleB)
	}
}

func assertCellStyleDiff(t *testing.T, wb *excelize.File, sheet, cellA, cellB string) {
	t.Helper()
	styleA := mustCellStyle(t, wb, sheet, cellA)
	styleB := mustCellStyle(t, wb, sheet, cellB)
	if styleA == styleB {
		t.Fatalf("style unexpectedly equal %s!%s and %s!%s: %d", sheet, cellA, sheet, cellB, styleA)
	}
}

func assertColWidth(t *testing.T, wb *excelize.File, sheet, col string, want float64) {
	t.Helper()
	got, err := wb.GetColWidth(sheet, col)
	if err != nil {
		t.Fatalf("get col width %s!%s: %v", sheet, col, err)
	}
	if got != want {
		t.Fatalf("col width mismatch %s!%s: got %v want %v", sheet, col, got, want)
	}
}

func mustCellStyle(t *testing.T, wb *excelize.File, sheet, cell string) int {
	t.Helper()
	style, err := wb.GetCellStyle(sheet, cell)
	if err != nil {
		t.Fatalf("get cell style %s!%s: %v", sheet, cell, err)
	}
	return style
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

func TestSingleMatchDeterminism(t *testing.T) {
	rules, err := loader.LoadLambdaRules(writeTempFile(t, t.TempDir(), "lambda.csv", sampleLambda))
	if err != nil {
		t.Fatalf("load lambda: %v", err)
	}

	input := SingleMatchInput{
		TeamA:       "Equipo A",
		TeamB:       "Equipo B",
		ShotsA:      7,
		ShotsB:      10,
		MotivationA: model.MotivationHigh,
		MotivationB: model.MotivationMedium,
		Tiebreaker:  "penalties",
	}

	first, err := RunSingleMatch(42, rules, input)
	if err != nil {
		t.Fatalf("run single match: %v", err)
	}
	second, err := RunSingleMatch(42, rules, input)
	if err != nil {
		t.Fatalf("run single match again: %v", err)
	}
	if first != second {
		t.Fatalf("single match should be deterministic: %+v vs %+v", first, second)
	}
	if first.Winner != input.TeamA && first.Winner != input.TeamB {
		t.Fatalf("unexpected winner: %s", first.Winner)
	}
	if first.DecidedBy == "" {
		t.Fatalf("decided by should not be empty")
	}
}

func TestSingleMatchSeriesDeterminism(t *testing.T) {
	rules, err := loader.LoadLambdaRules(writeTempFile(t, t.TempDir(), "lambda.csv", sampleLambda))
	if err != nil {
		t.Fatalf("load lambda: %v", err)
	}

	input := SingleMatchInput{
		TeamA:       "Equipo A",
		TeamB:       "Equipo B",
		ShotsA:      10,
		ShotsB:      4,
		MotivationA: model.MotivationHigh,
		MotivationB: model.MotivationMedium,
		Tiebreaker:  "penalties",
	}

	first, err := RunSingleMatchSeries(42, 6000, rules, input)
	if err != nil {
		t.Fatalf("run single match series: %v", err)
	}
	second, err := RunSingleMatchSeries(42, 6000, rules, input)
	if err != nil {
		t.Fatalf("run single match series again: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("single match series should be deterministic: %+v vs %+v", first, second)
	}
	if first.Simulations != 6000 {
		t.Fatalf("unexpected simulation count: %d", first.Simulations)
	}
	if first.WinsA+first.WinsB != first.Simulations {
		t.Fatalf("wins should sum to simulations: %+v", first)
	}
	if first.Regulation+first.Penalties+first.RandomTie != first.Simulations {
		t.Fatalf("decision modes should sum to simulations: %+v", first)
	}
	if first.Regulation+first.RegulationDraws != first.Simulations {
		t.Fatalf("regular time decisions and draws should sum to simulations: %+v", first)
	}
	if first.MostRepeatedCount <= 0 || first.MostRepeatedCount > first.Simulations {
		t.Fatalf("unexpected most repeated count: %+v", first)
	}
	if first.MostRepeatedPercent <= 0 {
		t.Fatalf("unexpected most repeated percent: %+v", first)
	}
	if len(first.TopScores) == 0 {
		t.Fatalf("top scores should not be empty: %+v", first)
	}
	if first.TopScores[0].Count != first.MostRepeatedCount {
		t.Fatalf("top score should match most repeated count: %+v", first)
	}
	if first.ScoreCounts == nil {
		t.Fatalf("score counts should not be nil: %+v", first)
	}
	if got := first.ScoreCounts[[2]int{first.MostRepeatedGoalsA, first.MostRepeatedGoalsB}]; got != first.MostRepeatedCount {
		t.Fatalf("most repeated score count mismatch: got %d want %d", got, first.MostRepeatedCount)
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

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

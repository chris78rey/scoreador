package report

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestGroupDemoWorkbook(t *testing.T) {
	dir := t.TempDir()
	teams := []string{"Ecuador", "Alemania", "Polonia", "Corea"}
	results := DefaultGroupDemoResults(teams)
	path := GroupDemoPath(dir, "Grupo A")

	if filepath.Base(path) != "grupo_grupo_a_demo.xlsx" {
		t.Fatalf("unexpected file name: %s", filepath.Base(path))
	}

	if err := WriteGroupDemoXLSX(path, "Grupo A", teams, results); err != nil {
		t.Fatalf("write demo xlsx: %v", err)
	}

	wb, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer func() {
		if err := wb.Close(); err != nil {
			t.Fatalf("close xlsx: %v", err)
		}
	}()

	expectedSheets := []string{"Instrucciones", "Partidos", "Tabla Base", "Posiciones", "Resumen", "Salida"}
	gotSheets := wb.GetSheetList()
	if len(gotSheets) != len(expectedSheets) {
		t.Fatalf("unexpected sheet count: %v", gotSheets)
	}
	for i, sheet := range expectedSheets {
		if gotSheets[i] != sheet {
			t.Fatalf("unexpected sheet order: %v", gotSheets)
		}
	}

	valueChecks := map[string]string{
		"Partidos!C5":  "1",
		"Partidos!D5":  "2",
		"Partidos!C10": "1",
		"Partidos!D10": "1",
		"Salida!B5":    "Alemania",
		"Salida!J5":    "9",
		"Salida!B6":    "Corea",
		"Salida!J6":    "4",
		"Salida!B7":    "Ecuador",
		"Salida!J7":    "3",
		"Salida!B8":    "Polonia",
		"Salida!J8":    "1",
	}
	for cell, expected := range valueChecks {
		sheet, ref := splitSheetCell(t, cell)
		if got := mustGetCellValue(t, wb, sheet, ref); got != expected {
			t.Fatalf("unexpected value at %s: got %q want %q", cell, got, expected)
		}
	}
}

func TestDefaultGroupDemoResults(t *testing.T) {
	teams := []string{"Ecuador", "Alemania", "Polonia", "Corea"}
	results := DefaultGroupDemoResults(teams)
	if len(results) != 6 {
		t.Fatalf("unexpected result count: %d", len(results))
	}

	standings := GroupDemoStandings("Grupo A", results)
	if len(standings) != 4 {
		t.Fatalf("unexpected standings count: %d", len(standings))
	}

	wantOrder := []string{"Alemania", "Corea", "Ecuador", "Polonia"}
	wantPoints := []int{9, 4, 3, 1}
	for i, standing := range standings {
		if standing.Team != wantOrder[i] {
			t.Fatalf("unexpected team at %d: got %s want %s", i, standing.Team, wantOrder[i])
		}
		if standing.Points != wantPoints[i] {
			t.Fatalf("unexpected points for %s: got %d want %d", standing.Team, standing.Points, wantPoints[i])
		}
	}
}

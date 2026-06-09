package report

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestWriteGroupTemplateXLSX(t *testing.T) {
	dir := t.TempDir()
	path := GroupTemplatePath(dir, "Grupo A")
	teams := []string{"Ecuador", "Alemania", "Polonia", "Corea"}

	if err := WriteGroupTemplateXLSX(path, "Grupo A", teams); err != nil {
		t.Fatalf("write template: %v", err)
	}

	if filepath.Base(path) != "grupo_grupo_a_template.xlsx" {
		t.Fatalf("unexpected file name: %s", filepath.Base(path))
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

	expectedSheets := []string{"Instrucciones", "Partidos", "Tabla Base", "Posiciones", "Resumen"}
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
		"Partidos!B5": "Ecuador",
		"Partidos!E5": "Alemania",
		"Partidos!B6": "Ecuador",
		"Partidos!E6": "Polonia",
		"Resumen!B5":  "6",
	}
	for cell, expected := range valueChecks {
		sheet, ref := splitSheetCell(t, cell)
		if got := mustGetCellValue(t, wb, sheet, ref); got != expected {
			t.Fatalf("unexpected value at %s: got %q want %q", cell, got, expected)
		}
	}

	formulaChecks := map[string]string{
		"Tabla Base!B5": `=SUMPRODUCT(('Partidos'!$B$5:$B$10=A5)*('Partidos'!$C$5:$C$10<>"")*('Partidos'!$D$5:$D$10<>""))+SUMPRODUCT(('Partidos'!$E$5:$E$10=A5)*('Partidos'!$C$5:$C$10<>"")*('Partidos'!$D$5:$D$10<>""))`,
		"Tabla Base!I5": `=C5*3+D5`,
		"Posiciones!B5": `=INDEX('Tabla Base'!$A$5:$A$8,MATCH(LARGE('Tabla Base'!$J$5:$J$8,1),'Tabla Base'!$J$5:$J$8,0))`,
		"Resumen!B8":    `=IF(B6=0,0,B7/B6)`,
	}
	for cell, expected := range formulaChecks {
		sheet, ref := splitSheetCell(t, cell)
		got, err := wb.GetCellFormula(sheet, ref)
		if err != nil {
			t.Fatalf("get formula %s: %v", cell, err)
		}
		if got != expected {
			t.Fatalf("unexpected formula at %s: got %q want %q", cell, got, expected)
		}
	}
}

func mustGetCellValue(t *testing.T, wb *excelize.File, sheet, cell string) string {
	t.Helper()
	value, err := wb.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("get cell %s!%s: %v", sheet, cell, err)
	}
	return value
}

func splitSheetCell(t *testing.T, ref string) (string, string) {
	t.Helper()
	parts := strings.SplitN(ref, "!", 2)
	if len(parts) != 2 {
		t.Fatalf("invalid sheet cell ref: %s", ref)
	}
	return parts[0], parts[1]
}

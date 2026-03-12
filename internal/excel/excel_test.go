package excel_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KimHansenCubris/gh-godo/internal/excel"
)

func tmpPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "workload.xlsx")
}

func TestOpenOrCreate_NewFile(t *testing.T) {
	path := tmpPath(t)
	f, err := excel.OpenOrCreate(path)
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	f.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created on disk")
	}
}

func TestAppendAndReadRows(t *testing.T) {
	path := tmpPath(t)

	row := excel.Row{
		Project:   "Alpha",
		SwTB:      "TB1",
		Version:   "1.0",
		Week:      "12",
		Type:      "Dev",
		Resources: "3",
		Year:      "2026",
		Quarter:   "Q1",
	}

	if err := excel.AppendRow(path, row); err != nil {
		t.Fatalf("AppendRow: %v", err)
	}

	rows, err := excel.ReadRows(path)
	if err != nil {
		t.Fatalf("ReadRows: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	got := rows[0]
	if got.Project != row.Project {
		t.Errorf("Project: want %q got %q", row.Project, got.Project)
	}
	if got.Resources != row.Resources {
		t.Errorf("Resources: want %q got %q", row.Resources, got.Resources)
	}
}

func TestUpdateRow(t *testing.T) {
	path := tmpPath(t)

	original := excel.Row{Project: "Beta", SwTB: "TB2", Version: "2.0", Week: "5", Type: "Test", Resources: "1", Year: "2026", Quarter: "Q2"}
	if err := excel.AppendRow(path, original); err != nil {
		t.Fatalf("AppendRow: %v", err)
	}

	rows, _ := excel.ReadRows(path)
	updated := rows[0]
	updated.Project = "BetaUpdated"
	updated.Resources = "5"

	if err := excel.UpdateRow(path, updated); err != nil {
		t.Fatalf("UpdateRow: %v", err)
	}

	rows, _ = excel.ReadRows(path)
	if rows[0].Project != "BetaUpdated" {
		t.Errorf("expected updated project, got %q", rows[0].Project)
	}
	if rows[0].Resources != "5" {
		t.Errorf("expected resources=5, got %q", rows[0].Resources)
	}
}

func TestDeleteRow(t *testing.T) {
	path := tmpPath(t)

	for _, name := range []string{"Row1", "Row2", "Row3"} {
		if err := excel.AppendRow(path, excel.Row{Project: name}); err != nil {
			t.Fatalf("AppendRow: %v", err)
		}
	}

	rows, _ := excel.ReadRows(path)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Delete the middle row (id from ReadRows for second entry)
	if err := excel.DeleteRow(path, rows[1].ID); err != nil {
		t.Fatalf("DeleteRow: %v", err)
	}

	rows, _ = excel.ReadRows(path)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows after delete, got %d", len(rows))
	}
	if rows[0].Project != "Row1" || rows[1].Project != "Row3" {
		t.Errorf("unexpected projects after delete: %v %v", rows[0].Project, rows[1].Project)
	}
}

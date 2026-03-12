// Package excel provides read/write helpers for the workload capacity sheet.
package excel

import (
	"fmt"
	"path/filepath"

	excelize "github.com/xuri/excelize/v2"
)

const sheetName = "Workload"

// Row represents a single workload capacity entry.
type Row struct {
	ID         int    `json:"id"`
	Project    string `json:"project"`
	SwTB       string `json:"swtb"`
	Version    string `json:"version"`
	Week       string `json:"week"`
	Type       string `json:"type"`
	Resources  string `json:"resources"`
	Year       string `json:"year"`
	Quarter    string `json:"quarter"`
}

var headers = []string{"Project", "SwTB", "Version", "Week", "Type", "#Ressources", "Year", "Quarter"}

// OpenOrCreate opens an existing workbook or creates a new one with the
// Workload sheet and headers if the file does not exist yet.
func OpenOrCreate(path string) (*excelize.File, error) {
	path = filepath.Clean(path)
	f, err := excelize.OpenFile(path)
	if err != nil {
		// File does not exist – create a new workbook.
		f = excelize.NewFile()
		idx, err := f.NewSheet(sheetName)
		if err != nil {
			return nil, fmt.Errorf("create sheet: %w", err)
		}
		f.SetActiveSheet(idx)
		// Remove the default "Sheet1" that excelize always creates.
		if err := f.DeleteSheet("Sheet1"); err != nil {
			// Not fatal – Sheet1 may not exist on some versions.
			_ = err
		}
		if err := writeHeaders(f); err != nil {
			return nil, err
		}
		if err := f.SaveAs(path); err != nil {
			return nil, fmt.Errorf("save new workbook: %w", err)
		}
	} else {
		// Ensure the Workload sheet exists.
		idx, _ := f.GetSheetIndex(sheetName)
		if idx == -1 {
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, fmt.Errorf("create sheet: %w", err)
			}
			if err := writeHeaders(f); err != nil {
				return nil, err
			}
		}
	}
	return f, nil
}

func writeHeaders(f *excelize.File) error {
	for i, h := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return fmt.Errorf("header coordinate: %w", err)
		}
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return fmt.Errorf("set header: %w", err)
		}
	}
	return nil
}

// ReadRows returns all data rows from the Workload sheet (excluding the header row).
func ReadRows(path string) ([]Row, error) {
	path = filepath.Clean(path)
	f, err := OpenOrCreate(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	var result []Row
	for i, r := range rows {
		if i == 0 {
			continue // skip header
		}
		row := Row{ID: i} // 1-based sheet row index = i+1, use i as ID
		if len(r) > 0 {
			row.Project = r[0]
		}
		if len(r) > 1 {
			row.SwTB = r[1]
		}
		if len(r) > 2 {
			row.Version = r[2]
		}
		if len(r) > 3 {
			row.Week = r[3]
		}
		if len(r) > 4 {
			row.Type = r[4]
		}
		if len(r) > 5 {
			row.Resources = r[5]
		}
		if len(r) > 6 {
			row.Year = r[6]
		}
		if len(r) > 7 {
			row.Quarter = r[7]
		}
		result = append(result, row)
	}
	return result, nil
}

// AppendRow adds a new row to the Workload sheet and saves the file.
func AppendRow(path string, row Row) error {
	path = filepath.Clean(path)
	f, err := OpenOrCreate(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("get rows: %w", err)
	}
	nextRow := len(rows) + 1

	values := []string{row.Project, row.SwTB, row.Version, row.Week, row.Type, row.Resources, row.Year, row.Quarter}
	for col, val := range values {
		cell, err := excelize.CoordinatesToCellName(col+1, nextRow)
		if err != nil {
			return fmt.Errorf("cell coordinate: %w", err)
		}
		if err := f.SetCellValue(sheetName, cell, val); err != nil {
			return fmt.Errorf("set cell value: %w", err)
		}
	}

	return f.Save()
}

// UpdateRow replaces the data in sheet row id+1 (ID is the 0-based row index
// returned by ReadRows, so +1 converts it to 1-based Excel row numbering,
// which naturally skips the header at row 1).
func UpdateRow(path string, row Row) error {
	path = filepath.Clean(path)
	f, err := OpenOrCreate(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sheetRow := row.ID + 1 // ID is 0-based (rows slice index); +1 gives 1-based sheet row (header is row 1)
	values := []string{row.Project, row.SwTB, row.Version, row.Week, row.Type, row.Resources, row.Year, row.Quarter}
	for col, val := range values {
		cell, err := excelize.CoordinatesToCellName(col+1, sheetRow)
		if err != nil {
			return fmt.Errorf("cell coordinate: %w", err)
		}
		if err := f.SetCellValue(sheetName, cell, val); err != nil {
			return fmt.Errorf("set cell value: %w", err)
		}
	}

	return f.Save()
}

// DeleteRow removes the row at the given id (1-based row index from ReadRows).
func DeleteRow(path string, id int) error {
	path = filepath.Clean(path)
	f, err := OpenOrCreate(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sheetRow := id + 1
	if err := f.RemoveRow(sheetName, sheetRow); err != nil {
		return fmt.Errorf("remove row: %w", err)
	}
	return f.Save()
}

// SheetName returns the name of the managed sheet.
func SheetName() string { return sheetName }

// Headers returns the column headers.
func Headers() []string { return headers }

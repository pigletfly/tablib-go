package tablib

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

func init() {
	RegisterExporter(FormatXLSX, ExporterFunc(exportXLSX))
	RegisterImporter(FormatXLSX, ImporterFunc(importXLSX))
	RegisterDatabookExporter(FormatXLSX, DatabookExporterFunc(exportDatabookXLSX))
}

func exportXLSX(ds *Dataset, w io.Writer) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := ds.Title()
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	// Rename default sheet
	f.SetSheetName("Sheet1", sheetName)

	if err := writeDatasetToSheet(f, sheetName, ds); err != nil {
		return err
	}

	return f.Write(w)
}

func writeDatasetToSheet(f *excelize.File, sheetName string, ds *Dataset) error {
	rowNum := 1

	// Write headers
	if len(ds.headers) > 0 {
		for col, header := range ds.headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
			if err := f.SetCellValue(sheetName, cell, header); err != nil {
				return err
			}
		}
		rowNum++
	}

	// Write data rows
	for _, row := range ds.data {
		for col, value := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				return err
			}
		}
		rowNum++
	}

	return nil
}

func importXLSX(r io.Reader) (*Dataset, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return NewDataset(nil), nil
	}

	return readSheetToDataset(f, sheets[0])
}

func readSheetToDataset(f *excelize.File, sheetName string) (*Dataset, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		ds := NewDataset(nil)
		ds.SetTitle(sheetName)
		return ds, nil
	}

	// First row as headers
	headers := rows[0]
	ds := NewDataset(headers)
	ds.SetTitle(sheetName)

	// Remaining rows as data
	for _, row := range rows[1:] {
		// Pad row if necessary
		for len(row) < len(headers) {
			row = append(row, "")
		}

		dataRow := make([]any, len(headers))
		for i := 0; i < len(headers); i++ {
			if i < len(row) {
				dataRow[i] = row[i]
			} else {
				dataRow[i] = ""
			}
		}
		if err := ds.Append(dataRow); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

// ImportXLSX imports a Dataset from an XLSX file, optionally specifying a sheet name.
func ImportXLSX(r io.Reader, sheetName string) (*Dataset, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return NewDataset(nil), nil
		}
		sheetName = sheets[0]
	}

	return readSheetToDataset(f, sheetName)
}

// ImportXLSXDatabook imports all sheets from an XLSX file into a Databook.
func ImportXLSXDatabook(r io.Reader) (*Databook, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db := NewDatabook()
	for _, sheetName := range f.GetSheetList() {
		ds, err := readSheetToDataset(f, sheetName)
		if err != nil {
			return nil, err
		}
		db.AddSheet(ds)
	}

	return db, nil
}

func exportDatabookXLSX(db *Databook, w io.Writer) error {
	f := excelize.NewFile()
	defer f.Close()

	// Remove default sheet if we have sheets to add
	if db.Size() > 0 {
		f.DeleteSheet("Sheet1")
	}

	for i, ds := range db.sheets {
		sheetName := ds.Title()
		if sheetName == "" {
			sheetName = fmt.Sprintf("Sheet%d", i+1)
		}

		// Create sheet
		if _, err := f.NewSheet(sheetName); err != nil {
			return err
		}

		if err := writeDatasetToSheet(f, sheetName, ds); err != nil {
			return err
		}
	}

	return f.Write(w)
}

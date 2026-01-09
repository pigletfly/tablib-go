package tablib

import (
	"encoding/csv"
	"fmt"
	"io"
)

func init() {
	RegisterExporter(FormatCSV, ExporterFunc(exportCSV))
	RegisterImporter(FormatCSV, ImporterFunc(importCSV))
	RegisterExporter(FormatTSV, ExporterFunc(exportTSV))
	RegisterImporter(FormatTSV, ImporterFunc(importTSV))
}

// CSVOptions configures CSV export behavior.
type CSVOptions struct {
	Delimiter   rune
	WriteHeader bool
}

// DefaultCSVOptions returns the default CSV options.
func DefaultCSVOptions() CSVOptions {
	return CSVOptions{
		Delimiter:   ',',
		WriteHeader: true,
	}
}

func exportCSV(ds *Dataset, w io.Writer) error {
	return exportCSVWithOptions(ds, w, DefaultCSVOptions())
}

func exportTSV(ds *Dataset, w io.Writer) error {
	opts := DefaultCSVOptions()
	opts.Delimiter = '\t'
	return exportCSVWithOptions(ds, w, opts)
}

func exportCSVWithOptions(ds *Dataset, w io.Writer, opts CSVOptions) error {
	writer := csv.NewWriter(w)
	writer.Comma = opts.Delimiter

	// Write headers
	if opts.WriteHeader && len(ds.headers) > 0 {
		if err := writer.Write(ds.headers); err != nil {
			return err
		}
	}

	// Write data rows
	for _, row := range ds.data {
		record := make([]string, len(row))
		for i, v := range row {
			record[i] = fmt.Sprintf("%v", v)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// ExportCSV exports the Dataset to CSV format with custom options.
func (ds *Dataset) ExportCSV(w io.Writer, opts CSVOptions) error {
	return exportCSVWithOptions(ds, w, opts)
}

func importCSV(r io.Reader) (*Dataset, error) {
	return importCSVWithOptions(r, ',', true)
}

func importTSV(r io.Reader) (*Dataset, error) {
	return importCSVWithOptions(r, '\t', true)
}

func importCSVWithOptions(r io.Reader, delimiter rune, hasHeaders bool) (*Dataset, error) {
	reader := csv.NewReader(r)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return NewDataset(nil), nil
	}

	var headers []string
	var dataStart int

	if hasHeaders {
		headers = records[0]
		dataStart = 1
	} else {
		dataStart = 0
	}

	ds := NewDataset(headers)

	for _, record := range records[dataStart:] {
		row := make([]any, len(record))
		for i, v := range record {
			row[i] = v
		}
		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

// ImportCSV imports a Dataset from CSV with custom options.
func ImportCSV(r io.Reader, delimiter rune, hasHeaders bool) (*Dataset, error) {
	return importCSVWithOptions(r, delimiter, hasHeaders)
}

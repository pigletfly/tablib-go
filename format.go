package tablib

import (
	"io"
	"strings"
)

// Format represents a data format identifier.
type Format string

const (
	FormatCSV      Format = "csv"
	FormatTSV      Format = "tsv"
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
	FormatXLSX     Format = "xlsx"
	FormatHTML     Format = "html"
	FormatMarkdown Format = "markdown"
	FormatLatex    Format = "latex"
	FormatSQL      Format = "sql"
	FormatRST      Format = "rst"      // reStructuredText
	FormatJira     Format = "jira"     // Jira Wiki markup
	FormatCLI      Format = "cli"      // ASCII table for CLI
	FormatDBF      Format = "dbf"      // dBase format
	FormatODS      Format = "ods"      // OpenDocument Spreadsheet
	FormatXLS      Format = "xls"      // Legacy Excel format
)

// Exporter is the interface for exporting a Dataset to a specific format.
type Exporter interface {
	Export(ds *Dataset, w io.Writer) error
}

// Importer is the interface for importing data into a Dataset from a specific format.
type Importer interface {
	Import(r io.Reader) (*Dataset, error)
}

// ExporterFunc is an adapter to allow ordinary functions to be used as Exporters.
type ExporterFunc func(ds *Dataset, w io.Writer) error

func (f ExporterFunc) Export(ds *Dataset, w io.Writer) error {
	return f(ds, w)
}

// ImporterFunc is an adapter to allow ordinary functions to be used as Importers.
type ImporterFunc func(r io.Reader) (*Dataset, error)

func (f ImporterFunc) Import(r io.Reader) (*Dataset, error) {
	return f(r)
}

// DatabookExporter is the interface for exporting a Databook to a specific format.
type DatabookExporter interface {
	ExportDatabook(db *Databook, w io.Writer) error
}

// DatabookExporterFunc is an adapter for Databook exporters.
type DatabookExporterFunc func(db *Databook, w io.Writer) error

func (f DatabookExporterFunc) ExportDatabook(db *Databook, w io.Writer) error {
	return f(db, w)
}

var (
	exporters         = make(map[Format]Exporter)
	importers         = make(map[Format]Importer)
	databookExporters = make(map[Format]DatabookExporter)
)

// RegisterExporter registers an exporter for a format.
func RegisterExporter(format Format, exporter Exporter) {
	exporters[format] = exporter
}

// RegisterImporter registers an importer for a format.
func RegisterImporter(format Format, importer Importer) {
	importers[format] = importer
}

// RegisterDatabookExporter registers a Databook exporter for a format.
func RegisterDatabookExporter(format Format, exporter DatabookExporter) {
	databookExporters[format] = exporter
}

// Export exports the Dataset to the specified format.
func (ds *Dataset) Export(format Format, w io.Writer) error {
	exporter, ok := exporters[format]
	if !ok {
		return ErrUnsupportedFormat
	}
	return exporter.Export(ds, w)
}

// ExportString exports the Dataset to the specified format and returns a string.
func (ds *Dataset) ExportString(format Format) (string, error) {
	var buf strings.Builder
	if err := ds.Export(format, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Import imports data from the specified format into a new Dataset.
func Import(format Format, r io.Reader) (*Dataset, error) {
	importer, ok := importers[format]
	if !ok {
		return nil, ErrUnsupportedFormat
	}
	return importer.Import(r)
}

// ImportString imports data from a string in the specified format.
func ImportString(format Format, data string) (*Dataset, error) {
	return Import(format, strings.NewReader(data))
}

// Export exports the Databook to the specified format.
func (db *Databook) Export(format Format, w io.Writer) error {
	exporter, ok := databookExporters[format]
	if !ok {
		return ErrUnsupportedFormat
	}
	return exporter.ExportDatabook(db, w)
}

// ExportString exports the Databook to the specified format and returns a string.
func (db *Databook) ExportString(format Format) (string, error) {
	var buf strings.Builder
	if err := db.Export(format, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SupportedExportFormats returns all registered export formats.
func SupportedExportFormats() []Format {
	formats := make([]Format, 0, len(exporters))
	for f := range exporters {
		formats = append(formats, f)
	}
	return formats
}

// SupportedImportFormats returns all registered import formats.
func SupportedImportFormats() []Format {
	formats := make([]Format, 0, len(importers))
	for f := range importers {
		formats = append(formats, f)
	}
	return formats
}

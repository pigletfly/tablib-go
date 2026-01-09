package tablib

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatSQL, ExporterFunc(exportSQL))
}

// SQLOptions configures SQL export behavior.
type SQLOptions struct {
	TableName string
}

func exportSQL(ds *Dataset, w io.Writer) error {
	tableName := ds.Title()
	if tableName == "" {
		tableName = "export_table"
	}
	return exportSQLWithOptions(ds, w, SQLOptions{TableName: tableName})
}

func exportSQLWithOptions(ds *Dataset, w io.Writer, opts SQLOptions) error {
	if len(ds.headers) == 0 {
		return ErrHeadersRequired
	}

	if len(ds.data) == 0 {
		return nil
	}

	var sb strings.Builder

	// Quote column names
	columns := make([]string, len(ds.headers))
	for i, h := range ds.headers {
		columns[i] = fmt.Sprintf(`"%s"`, h)
	}
	columnList := strings.Join(columns, ", ")

	// Generate INSERT statements
	for _, row := range ds.data {
		values := make([]string, len(row))
		for i, v := range row {
			values[i] = sqlValue(v)
		}
		valueList := strings.Join(values, ", ")

		sb.WriteString(fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s);\n",
			opts.TableName, columnList, valueList))
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

// ExportSQL exports the Dataset to SQL INSERT statements with custom options.
func (ds *Dataset) ExportSQL(w io.Writer, opts SQLOptions) error {
	return exportSQLWithOptions(ds, w, opts)
}

// sqlValue converts a value to its SQL literal representation.
func sqlValue(v any) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case string:
		// Escape single quotes by doubling them
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	default:
		escaped := strings.ReplaceAll(fmt.Sprintf("%v", val), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

package tablib

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatJira, ExporterFunc(exportJira))
}

// exportJira exports the Dataset to Jira Wiki markup table format.
func exportJira(ds *Dataset, w io.Writer) error {
	if ds.Width() == 0 {
		return nil
	}

	var sb strings.Builder

	// Write headers (Jira uses || for header cells)
	if len(ds.headers) > 0 {
		sb.WriteString("||")
		for _, h := range ds.headers {
			sb.WriteString(escapeJira(h))
			sb.WriteString("||")
		}
		sb.WriteString("\n")
	}

	// Write data rows (Jira uses | for regular cells)
	for rowIdx, row := range ds.data {
		// Check for separator before this row
		if sep, ok := ds.GetSeparator(rowIdx); ok {
			// Jira doesn't have native separators, use a spanning row with emphasis
			sb.WriteString("|")
			sb.WriteString(fmt.Sprintf("*%s*", escapeJira(sep.Text)))
			sb.WriteString("|\n")
		}

		sb.WriteString("|")
		for _, v := range row {
			sb.WriteString(escapeJira(fmt.Sprintf("%v", v)))
			sb.WriteString("|")
		}
		sb.WriteString("\n")
	}

	// Check for separator after the last row
	if sep, ok := ds.GetSeparator(len(ds.data)); ok {
		sb.WriteString("|")
		sb.WriteString(fmt.Sprintf("*%s*", escapeJira(sep.Text)))
		sb.WriteString("|\n")
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

// escapeJira escapes special characters for Jira Wiki markup.
func escapeJira(s string) string {
	// Escape special Jira characters
	replacer := strings.NewReplacer(
		"|", "\\|",
		"[", "\\[",
		"]", "\\]",
		"{", "\\{",
		"}", "\\}",
		"*", "\\*",
		"_", "\\_",
		"-", "\\-",
		"+", "\\+",
		"^", "\\^",
		"~", "\\~",
	)
	return replacer.Replace(s)
}

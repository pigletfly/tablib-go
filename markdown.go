package tablib

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatMarkdown, ExporterFunc(exportMarkdown))
}

func exportMarkdown(ds *Dataset, w io.Writer) error {
	if ds.Width() == 0 {
		return nil
	}

	var sb strings.Builder

	// Calculate column widths
	widths := make([]int, ds.Width())
	for i, h := range ds.headers {
		if len(h) > widths[i] {
			widths[i] = len(h)
		}
	}
	for _, row := range ds.data {
		for i, v := range row {
			s := fmt.Sprintf("%v", v)
			if len(s) > widths[i] {
				widths[i] = len(s)
			}
		}
	}

	// Ensure minimum width of 3 for separator
	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
	}

	// Write headers
	if len(ds.headers) > 0 {
		sb.WriteString("|")
		for i, h := range ds.headers {
			sb.WriteString(fmt.Sprintf(" %-*s |", widths[i], h))
		}
		sb.WriteString("\n")

		// Write separator
		sb.WriteString("|")
		for _, w := range widths {
			sb.WriteString(fmt.Sprintf(" %s |", strings.Repeat("-", w)))
		}
		sb.WriteString("\n")
	}

	// Write data rows
	for _, row := range ds.data {
		sb.WriteString("|")
		for i, v := range row {
			sb.WriteString(fmt.Sprintf(" %-*v |", widths[i], v))
		}
		sb.WriteString("\n")
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

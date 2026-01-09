package tablib

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatRST, ExporterFunc(exportRST))
}

// exportRST exports the Dataset to reStructuredText grid table format.
func exportRST(ds *Dataset, w io.Writer) error {
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

	// Ensure minimum width of 1
	for i := range widths {
		if widths[i] < 1 {
			widths[i] = 1
		}
	}

	// Helper function to write a separator line
	writeSeparator := func(char string) {
		sb.WriteString("+")
		for _, w := range widths {
			sb.WriteString(strings.Repeat(char, w+2))
			sb.WriteString("+")
		}
		sb.WriteString("\n")
	}

	// Write top border
	writeSeparator("-")

	// Write headers
	if len(ds.headers) > 0 {
		sb.WriteString("|")
		for i, h := range ds.headers {
			sb.WriteString(fmt.Sprintf(" %-*s |", widths[i], h))
		}
		sb.WriteString("\n")
		writeSeparator("=")
	}

	// Write data rows
	for rowIdx, row := range ds.data {
		// Check for separator before this row
		if sep, ok := ds.GetSeparator(rowIdx); ok {
			// Write separator row
			sb.WriteString("|")
			totalWidth := 0
			for _, w := range widths {
				totalWidth += w + 3 // +3 for " | "
			}
			totalWidth-- // Remove last extra space
			text := sep.Text
			if len(text) > totalWidth-2 {
				text = text[:totalWidth-2]
			}
			sb.WriteString(fmt.Sprintf(" %-*s |", totalWidth-2, text))
			sb.WriteString("\n")
			writeSeparator("-")
		}

		sb.WriteString("|")
		for i, v := range row {
			sb.WriteString(fmt.Sprintf(" %-*v |", widths[i], v))
		}
		sb.WriteString("\n")
		writeSeparator("-")
	}

	// Check for separator after the last row
	if sep, ok := ds.GetSeparator(len(ds.data)); ok {
		sb.WriteString("|")
		totalWidth := 0
		for _, w := range widths {
			totalWidth += w + 3
		}
		totalWidth--
		text := sep.Text
		if len(text) > totalWidth-2 {
			text = text[:totalWidth-2]
		}
		sb.WriteString(fmt.Sprintf(" %-*s |", totalWidth-2, text))
		sb.WriteString("\n")
		writeSeparator("-")
	}

	_, err := w.Write([]byte(sb.String()))
	return err
}

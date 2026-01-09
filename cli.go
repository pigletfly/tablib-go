package tablib

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatCLI, ExporterFunc(exportCLI))
}

// CLIOptions holds options for CLI export.
type CLIOptions struct {
	// Border style: "single" (default), "double", "ascii", "none"
	BorderStyle string
}

// DefaultCLIOptions returns default CLI export options.
func DefaultCLIOptions() CLIOptions {
	return CLIOptions{
		BorderStyle: "single",
	}
}

// ExportCLI exports the Dataset to CLI ASCII table format with options.
func (ds *Dataset) ExportCLI(w io.Writer, opts CLIOptions) error {
	return exportCLIWithOptions(ds, w, opts)
}

// exportCLI exports the Dataset using default CLI options.
func exportCLI(ds *Dataset, w io.Writer) error {
	return exportCLIWithOptions(ds, w, DefaultCLIOptions())
}

// getBorderChars returns border characters based on style.
func getBorderChars(style string) (topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical, cross, topT, bottomT, leftT, rightT string) {
	switch style {
	case "double":
		return "╔", "╗", "╚", "╝", "═", "║", "╬", "╦", "╩", "╠", "╣"
	case "ascii":
		return "+", "+", "+", "+", "-", "|", "+", "+", "+", "+", "+"
	case "none":
		return "", "", "", "", "", " ", "", "", "", "", ""
	default: // "single"
		return "┌", "┐", "└", "┘", "─", "│", "┼", "┬", "┴", "├", "┤"
	}
}

func exportCLIWithOptions(ds *Dataset, w io.Writer, opts CLIOptions) error {
	if ds.Width() == 0 {
		return nil
	}

	var sb strings.Builder

	// Get border characters
	topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical, cross, topT, bottomT, leftT, rightT := getBorderChars(opts.BorderStyle)

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
	writeTopBorder := func() {
		if opts.BorderStyle == "none" {
			return
		}
		sb.WriteString(topLeft)
		for i, w := range widths {
			sb.WriteString(strings.Repeat(horizontal, w+2))
			if i < len(widths)-1 {
				sb.WriteString(topT)
			}
		}
		sb.WriteString(topRight)
		sb.WriteString("\n")
	}

	writeBottomBorder := func() {
		if opts.BorderStyle == "none" {
			return
		}
		sb.WriteString(bottomLeft)
		for i, w := range widths {
			sb.WriteString(strings.Repeat(horizontal, w+2))
			if i < len(widths)-1 {
				sb.WriteString(bottomT)
			}
		}
		sb.WriteString(bottomRight)
		sb.WriteString("\n")
	}

	writeMiddleBorder := func() {
		if opts.BorderStyle == "none" {
			return
		}
		sb.WriteString(leftT)
		for i, w := range widths {
			sb.WriteString(strings.Repeat(horizontal, w+2))
			if i < len(widths)-1 {
				sb.WriteString(cross)
			}
		}
		sb.WriteString(rightT)
		sb.WriteString("\n")
	}

	// Write top border
	writeTopBorder()

	// Write headers
	if len(ds.headers) > 0 {
		sb.WriteString(vertical)
		for i, h := range ds.headers {
			sb.WriteString(fmt.Sprintf(" %-*s ", widths[i], h))
			sb.WriteString(vertical)
		}
		sb.WriteString("\n")
		writeMiddleBorder()
	}

	// Write data rows
	for rowIdx, row := range ds.data {
		// Check for separator before this row
		if sep, ok := ds.GetSeparator(rowIdx); ok {
			if opts.BorderStyle != "none" {
				sb.WriteString(vertical)
			}
			totalWidth := 0
			for _, w := range widths {
				totalWidth += w + 3 // +3 for " | "
			}
			totalWidth-- // Remove last extra space
			text := sep.Text
			if len(text) > totalWidth-2 {
				text = text[:totalWidth-2]
			}
			sb.WriteString(fmt.Sprintf(" %-*s ", totalWidth-2, text))
			if opts.BorderStyle != "none" {
				sb.WriteString(vertical)
			}
			sb.WriteString("\n")
			writeMiddleBorder()
		}

		sb.WriteString(vertical)
		for i, v := range row {
			sb.WriteString(fmt.Sprintf(" %-*v ", widths[i], v))
			sb.WriteString(vertical)
		}
		sb.WriteString("\n")
	}

	// Check for separator after the last row
	if sep, ok := ds.GetSeparator(len(ds.data)); ok {
		writeMiddleBorder()
		sb.WriteString(vertical)
		totalWidth := 0
		for _, w := range widths {
			totalWidth += w + 3
		}
		totalWidth--
		text := sep.Text
		if len(text) > totalWidth-2 {
			text = text[:totalWidth-2]
		}
		sb.WriteString(fmt.Sprintf(" %-*s ", totalWidth-2, text))
		sb.WriteString(vertical)
		sb.WriteString("\n")
	}

	// Write bottom border
	writeBottomBorder()

	_, err := w.Write([]byte(sb.String()))
	return err
}

package tablib

import (
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatLatex, ExporterFunc(exportLatex))
}

func exportLatex(ds *Dataset, w io.Writer) error {
	if ds.Width() == 0 {
		return nil
	}

	var sb strings.Builder

	// Begin tabular environment
	cols := strings.Repeat("l", ds.Width())
	sb.WriteString(fmt.Sprintf("\\begin{tabular}{%s}\n", cols))
	sb.WriteString("\\hline\n")

	// Write headers
	if len(ds.headers) > 0 {
		escaped := make([]string, len(ds.headers))
		for i, h := range ds.headers {
			escaped[i] = escapeLatex(h)
		}
		sb.WriteString(strings.Join(escaped, " & "))
		sb.WriteString(" \\\\\n")
		sb.WriteString("\\hline\n")
	}

	// Write data rows
	for _, row := range ds.data {
		escaped := make([]string, len(row))
		for i, v := range row {
			escaped[i] = escapeLatex(fmt.Sprintf("%v", v))
		}
		sb.WriteString(strings.Join(escaped, " & "))
		sb.WriteString(" \\\\\n")
	}

	sb.WriteString("\\hline\n")
	sb.WriteString("\\end{tabular}")

	_, err := w.Write([]byte(sb.String()))
	return err
}

// escapeLatex escapes special LaTeX characters.
func escapeLatex(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"&", "\\&",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	return replacer.Replace(s)
}

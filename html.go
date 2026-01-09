package tablib

import (
	"fmt"
	"html"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatHTML, ExporterFunc(exportHTML))
}

func exportHTML(ds *Dataset, w io.Writer) error {
	var sb strings.Builder

	sb.WriteString("<table>\n")

	// Write headers
	if len(ds.headers) > 0 {
		sb.WriteString("  <thead>\n    <tr>\n")
		for _, h := range ds.headers {
			sb.WriteString(fmt.Sprintf("      <th>%s</th>\n", html.EscapeString(h)))
		}
		sb.WriteString("    </tr>\n  </thead>\n")
	}

	// Write body
	sb.WriteString("  <tbody>\n")
	for _, row := range ds.data {
		sb.WriteString("    <tr>\n")
		for _, v := range row {
			sb.WriteString(fmt.Sprintf("      <td>%s</td>\n", html.EscapeString(fmt.Sprintf("%v", v))))
		}
		sb.WriteString("    </tr>\n")
	}
	sb.WriteString("  </tbody>\n")

	sb.WriteString("</table>")

	_, err := w.Write([]byte(sb.String()))
	return err
}

// HTMLOptions configures HTML export behavior.
type HTMLOptions struct {
	TableClass string
	TableID    string
}

// ExportHTML exports the Dataset to HTML with custom options.
func (ds *Dataset) ExportHTML(w io.Writer, opts HTMLOptions) error {
	var sb strings.Builder

	tableAttrs := ""
	if opts.TableID != "" {
		tableAttrs += fmt.Sprintf(` id="%s"`, html.EscapeString(opts.TableID))
	}
	if opts.TableClass != "" {
		tableAttrs += fmt.Sprintf(` class="%s"`, html.EscapeString(opts.TableClass))
	}

	sb.WriteString(fmt.Sprintf("<table%s>\n", tableAttrs))

	if len(ds.headers) > 0 {
		sb.WriteString("  <thead>\n    <tr>\n")
		for _, h := range ds.headers {
			sb.WriteString(fmt.Sprintf("      <th>%s</th>\n", html.EscapeString(h)))
		}
		sb.WriteString("    </tr>\n  </thead>\n")
	}

	sb.WriteString("  <tbody>\n")
	for _, row := range ds.data {
		sb.WriteString("    <tr>\n")
		for _, v := range row {
			sb.WriteString(fmt.Sprintf("      <td>%s</td>\n", html.EscapeString(fmt.Sprintf("%v", v))))
		}
		sb.WriteString("    </tr>\n")
	}
	sb.WriteString("  </tbody>\n")

	sb.WriteString("</table>")

	_, err := w.Write([]byte(sb.String()))
	return err
}

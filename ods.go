package tablib

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatODS, ExporterFunc(exportODS))
	RegisterDatabookExporter(FormatODS, DatabookExporterFunc(exportODSDatabook))
}

// ODS XML structures
type odsDocument struct {
	XMLName       xml.Name      `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 document-content"`
	Version       string        `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 version,attr"`
	OfficeNS      string        `xml:"xmlns:office,attr"`
	TextNS        string        `xml:"xmlns:text,attr"`
	TableNS       string        `xml:"xmlns:table,attr"`
	StyleNS       string        `xml:"xmlns:style,attr"`
	FoNS          string        `xml:"xmlns:fo,attr"`
	AutoStyles    odsAutoStyles `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 automatic-styles"`
	Body          odsBody       `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 body"`
}

type odsAutoStyles struct {
	Styles []odsStyle `xml:"urn:oasis:names:tc:opendocument:xmlns:style:1.0 style"`
}

type odsStyle struct {
	Name       string              `xml:"urn:oasis:names:tc:opendocument:xmlns:style:1.0 name,attr"`
	Family     string              `xml:"urn:oasis:names:tc:opendocument:xmlns:style:1.0 family,attr"`
	Properties *odsTextProperties  `xml:"urn:oasis:names:tc:opendocument:xmlns:style:1.0 text-properties,omitempty"`
}

type odsTextProperties struct {
	FontWeight string `xml:"urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0 font-weight,attr,omitempty"`
}

type odsBody struct {
	Spreadsheet odsSpreadsheet `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 spreadsheet"`
}

type odsSpreadsheet struct {
	Tables []odsTable `xml:"urn:oasis:names:tc:opendocument:xmlns:table:1.0 table"`
}

type odsTable struct {
	Name string    `xml:"urn:oasis:names:tc:opendocument:xmlns:table:1.0 name,attr"`
	Rows []odsRow  `xml:"urn:oasis:names:tc:opendocument:xmlns:table:1.0 table-row"`
}

type odsRow struct {
	Cells []odsCell `xml:"urn:oasis:names:tc:opendocument:xmlns:table:1.0 table-cell"`
}

type odsCell struct {
	ValueType string  `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 value-type,attr,omitempty"`
	Value     string  `xml:"urn:oasis:names:tc:opendocument:xmlns:office:1.0 value,attr,omitempty"`
	StyleName string  `xml:"urn:oasis:names:tc:opendocument:xmlns:table:1.0 style-name,attr,omitempty"`
	Text      *odsText `xml:"urn:oasis:names:tc:opendocument:xmlns:text:1.0 p,omitempty"`
}

type odsText struct {
	Content string `xml:",chardata"`
}

func exportODS(ds *Dataset, w io.Writer) error {
	return exportODSSheets(w, []*Dataset{ds})
}

func exportODSDatabook(db *Databook, w io.Writer) error {
	return exportODSSheets(w, db.sheets)
}

func exportODSSheets(w io.Writer, sheets []*Dataset) error {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Write mimetype (must be first and uncompressed)
	mimetypeHeader := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // No compression
	}
	mimetypeWriter, err := zipWriter.CreateHeader(mimetypeHeader)
	if err != nil {
		return err
	}
	_, err = mimetypeWriter.Write([]byte("application/vnd.oasis.opendocument.spreadsheet"))
	if err != nil {
		return err
	}

	// Create manifest
	manifest := `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2">
  <manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/>
  <manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/>
  <manifest:file-entry manifest:full-path="styles.xml" manifest:media-type="text/xml"/>
</manifest:manifest>`

	manifestWriter, err := zipWriter.Create("META-INF/manifest.xml")
	if err != nil {
		return err
	}
	_, err = manifestWriter.Write([]byte(manifest))
	if err != nil {
		return err
	}

	// Create styles.xml
	styles := `<?xml version="1.0" encoding="UTF-8"?>
<office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" office:version="1.2">
</office:document-styles>`

	stylesWriter, err := zipWriter.Create("styles.xml")
	if err != nil {
		return err
	}
	_, err = stylesWriter.Write([]byte(styles))
	if err != nil {
		return err
	}

	// Create content.xml
	doc := odsDocument{
		Version:  "1.2",
		OfficeNS: "urn:oasis:names:tc:opendocument:xmlns:office:1.0",
		TextNS:   "urn:oasis:names:tc:opendocument:xmlns:text:1.0",
		TableNS:  "urn:oasis:names:tc:opendocument:xmlns:table:1.0",
		StyleNS:  "urn:oasis:names:tc:opendocument:xmlns:style:1.0",
		FoNS:     "urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0",
		AutoStyles: odsAutoStyles{
			Styles: []odsStyle{
				{
					Name:   "bold",
					Family: "table-cell",
					Properties: &odsTextProperties{
						FontWeight: "bold",
					},
				},
			},
		},
	}

	tables := make([]odsTable, 0, len(sheets))
	for _, ds := range sheets {
		table := odsTable{
			Name: ds.title,
		}
		if table.Name == "" {
			table.Name = "Sheet"
		}

		// Add header row
		if len(ds.headers) > 0 {
			headerRow := odsRow{
				Cells: make([]odsCell, len(ds.headers)),
			}
			for i, h := range ds.headers {
				headerRow.Cells[i] = odsCell{
					ValueType: "string",
					StyleName: "bold",
					Text:      &odsText{Content: h},
				}
			}
			table.Rows = append(table.Rows, headerRow)
		}

		// Add data rows
		for _, row := range ds.data {
			dataRow := odsRow{
				Cells: make([]odsCell, len(row)),
			}
			for i, v := range row {
				cell := odsCell{}
				switch val := v.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					cell.ValueType = "float"
					cell.Value = fmt.Sprintf("%v", val)
					cell.Text = &odsText{Content: fmt.Sprintf("%v", val)}
				case float32, float64:
					cell.ValueType = "float"
					cell.Value = fmt.Sprintf("%v", val)
					cell.Text = &odsText{Content: fmt.Sprintf("%v", val)}
				case bool:
					cell.ValueType = "boolean"
					cell.Value = fmt.Sprintf("%v", val)
					cell.Text = &odsText{Content: fmt.Sprintf("%v", val)}
				default:
					cell.ValueType = "string"
					cell.Text = &odsText{Content: fmt.Sprintf("%v", val)}
				}
				dataRow.Cells[i] = cell
			}
			table.Rows = append(table.Rows, dataRow)
		}

		tables = append(tables, table)
	}

	doc.Body.Spreadsheet.Tables = tables

	contentWriter, err := zipWriter.Create("content.xml")
	if err != nil {
		return err
	}
	_, err = contentWriter.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	encoder := xml.NewEncoder(contentWriter)
	encoder.Indent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		return err
	}

	if err := zipWriter.Close(); err != nil {
		return err
	}

	_, err = w.Write(buf.Bytes())
	return err
}

// ImportODS imports data from an ODS file.
func ImportODS(r io.ReaderAt, size int64, sheetName string) (*Dataset, error) {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}

	// Find and parse content.xml
	var contentFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "content.xml" {
			contentFile = f
			break
		}
	}
	if contentFile == nil {
		return nil, fmt.Errorf("content.xml not found in ODS file")
	}

	rc, err := contentFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	// Parse the XML
	type simpleCell struct {
		ValueType string `xml:"value-type,attr"`
		Value     string `xml:"value,attr"`
		Text      string `xml:"p"`
	}
	type simpleRow struct {
		Cells []simpleCell `xml:"table-cell"`
	}
	type simpleTable struct {
		Name string      `xml:"name,attr"`
		Rows []simpleRow `xml:"table-row"`
	}
	type simpleSpreadsheet struct {
		Tables []simpleTable `xml:"table"`
	}
	type simpleBody struct {
		Spreadsheet simpleSpreadsheet `xml:"spreadsheet"`
	}
	type simpleDoc struct {
		Body simpleBody `xml:"body"`
	}

	var doc simpleDoc
	decoder := xml.NewDecoder(rc)
	if err := decoder.Decode(&doc); err != nil {
		return nil, err
	}

	// Find the requested sheet
	var targetTable *simpleTable
	for i := range doc.Body.Spreadsheet.Tables {
		t := &doc.Body.Spreadsheet.Tables[i]
		if sheetName == "" || t.Name == sheetName {
			targetTable = t
			break
		}
	}
	if targetTable == nil {
		return nil, fmt.Errorf("sheet '%s' not found", sheetName)
	}

	// Convert to Dataset
	if len(targetTable.Rows) == 0 {
		return NewDataset(nil), nil
	}

	// First row as headers
	var headers []string
	if len(targetTable.Rows) > 0 {
		for _, cell := range targetTable.Rows[0].Cells {
			text := strings.TrimSpace(cell.Text)
			if text == "" {
				text = cell.Value
			}
			headers = append(headers, text)
		}
	}

	ds := NewDataset(headers)
	ds.SetTitle(targetTable.Name)

	// Remaining rows as data
	for i := 1; i < len(targetTable.Rows); i++ {
		row := make([]any, len(headers))
		for j, cell := range targetTable.Rows[i].Cells {
			if j >= len(headers) {
				break
			}
			text := strings.TrimSpace(cell.Text)
			if text == "" {
				text = cell.Value
			}
			row[j] = text
		}
		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

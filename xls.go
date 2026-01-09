package tablib

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func init() {
	RegisterExporter(FormatXLS, ExporterFunc(exportXLS))
	RegisterDatabookExporter(FormatXLS, DatabookExporterFunc(exportXLSDatabook))
}

// XLS export uses Microsoft Spreadsheet XML format which can be opened by Excel.
// Note: This is not the binary BIFF format, but a compatible XML format.

type xlsWorkbook struct {
	XMLName    xml.Name       `xml:"Workbook"`
	XMLNS      string         `xml:"xmlns,attr"`
	XMLNSO     string         `xml:"xmlns:o,attr"`
	XMLNSX     string         `xml:"xmlns:x,attr"`
	XMLNSS     string         `xml:"xmlns:ss,attr"`
	XMLNSHTML  string         `xml:"xmlns:html,attr"`
	Styles     xlsStyles      `xml:"Styles"`
	Worksheets []xlsWorksheet `xml:"Worksheet"`
}

type xlsStyles struct {
	Styles []xlsStyle `xml:"Style"`
}

type xlsStyle struct {
	ID   string  `xml:"ss:ID,attr"`
	Font *xlsFont `xml:"Font,omitempty"`
}

type xlsFont struct {
	Bold int `xml:"ss:Bold,attr,omitempty"`
}

type xlsWorksheet struct {
	Name  string   `xml:"ss:Name,attr"`
	Table xlsTable `xml:"Table"`
}

type xlsTable struct {
	Rows []xlsRow `xml:"Row"`
}

type xlsRow struct {
	Cells []xlsCell `xml:"Cell"`
}

type xlsCell struct {
	StyleID string  `xml:"ss:StyleID,attr,omitempty"`
	Data    xlsData `xml:"Data"`
}

type xlsData struct {
	Type  string `xml:"ss:Type,attr"`
	Value string `xml:",chardata"`
}

func exportXLS(ds *Dataset, w io.Writer) error {
	return exportXLSSheets(w, []*Dataset{ds})
}

func exportXLSDatabook(db *Databook, w io.Writer) error {
	return exportXLSSheets(w, db.sheets)
}

func exportXLSSheets(w io.Writer, sheets []*Dataset) error {
	workbook := xlsWorkbook{
		XMLNS:     "urn:schemas-microsoft-com:office:spreadsheet",
		XMLNSO:    "urn:schemas-microsoft-com:office:office",
		XMLNSX:    "urn:schemas-microsoft-com:office:excel",
		XMLNSS:    "urn:schemas-microsoft-com:office:spreadsheet",
		XMLNSHTML: "http://www.w3.org/TR/REC-html40",
		Styles: xlsStyles{
			Styles: []xlsStyle{
				{ID: "Default"},
				{ID: "Header", Font: &xlsFont{Bold: 1}},
			},
		},
	}

	for _, ds := range sheets {
		worksheet := xlsWorksheet{
			Name: ds.title,
		}
		if worksheet.Name == "" {
			worksheet.Name = "Sheet"
		}

		// Add header row
		if len(ds.headers) > 0 {
			headerRow := xlsRow{
				Cells: make([]xlsCell, len(ds.headers)),
			}
			for i, h := range ds.headers {
				headerRow.Cells[i] = xlsCell{
					StyleID: "Header",
					Data: xlsData{
						Type:  "String",
						Value: h,
					},
				}
			}
			worksheet.Table.Rows = append(worksheet.Table.Rows, headerRow)
		}

		// Add data rows
		for _, row := range ds.data {
			dataRow := xlsRow{
				Cells: make([]xlsCell, len(row)),
			}
			for i, v := range row {
				cell := xlsCell{}
				switch val := v.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					cell.Data = xlsData{Type: "Number", Value: fmt.Sprintf("%v", val)}
				case float32, float64:
					cell.Data = xlsData{Type: "Number", Value: fmt.Sprintf("%v", val)}
				case bool:
					boolVal := "0"
					if val {
						boolVal = "1"
					}
					cell.Data = xlsData{Type: "Boolean", Value: boolVal}
				default:
					cell.Data = xlsData{Type: "String", Value: fmt.Sprintf("%v", val)}
				}
				dataRow.Cells[i] = cell
			}
			worksheet.Table.Rows = append(worksheet.Table.Rows, dataRow)
		}

		workbook.Worksheets = append(workbook.Worksheets, worksheet)
	}

	// Write XML declaration
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}

	// Write processing instruction for Excel
	if _, err := w.Write([]byte("<?mso-application progid=\"Excel.Sheet\"?>\n")); err != nil {
		return err
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	return encoder.Encode(workbook)
}

// ImportXLS imports data from an XLS file.
// Note: This only supports the XML Spreadsheet format, not the binary BIFF format.
func ImportXLS(r io.Reader, sheetName string) (*Dataset, error) {
	// Parse the XML
	var workbook xlsWorkbook
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&workbook); err != nil {
		return nil, fmt.Errorf("failed to parse XLS XML: %w", err)
	}

	// Find the requested sheet
	var targetSheet *xlsWorksheet
	for i := range workbook.Worksheets {
		ws := &workbook.Worksheets[i]
		if sheetName == "" || ws.Name == sheetName {
			targetSheet = ws
			break
		}
	}
	if targetSheet == nil {
		return nil, fmt.Errorf("sheet '%s' not found", sheetName)
	}

	// Convert to Dataset
	if len(targetSheet.Table.Rows) == 0 {
		return NewDataset(nil), nil
	}

	// First row as headers
	var headers []string
	if len(targetSheet.Table.Rows) > 0 {
		for _, cell := range targetSheet.Table.Rows[0].Cells {
			headers = append(headers, strings.TrimSpace(cell.Data.Value))
		}
	}

	ds := NewDataset(headers)
	ds.SetTitle(targetSheet.Name)

	// Remaining rows as data
	for i := 1; i < len(targetSheet.Table.Rows); i++ {
		row := make([]any, len(headers))
		for j, cell := range targetSheet.Table.Rows[i].Cells {
			if j >= len(headers) {
				break
			}
			row[j] = strings.TrimSpace(cell.Data.Value)
		}
		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

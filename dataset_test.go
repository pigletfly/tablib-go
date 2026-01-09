package tablib

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewDataset(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	ds := NewDataset(headers)

	if ds.Height() != 0 {
		t.Errorf("expected height 0, got %d", ds.Height())
	}
	if ds.Width() != 3 {
		t.Errorf("expected width 3, got %d", ds.Width())
	}

	got := ds.Headers()
	for i, h := range headers {
		if got[i] != h {
			t.Errorf("expected header %s, got %s", h, got[i])
		}
	}
}

func TestDatasetAppend(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})

	err := ds.Append([]any{"Alice", 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Height() != 1 {
		t.Errorf("expected height 1, got %d", ds.Height())
	}

	row, err := ds.Row(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if row[0] != "Alice" || row[1] != 30 {
		t.Errorf("unexpected row values: %v", row)
	}
}

func TestDatasetAppendInvalidDimensions(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})

	err := ds.Append([]any{"Alice", 30, "Extra"})
	if err != ErrInvalidDimensions {
		t.Errorf("expected ErrInvalidDimensions, got %v", err)
	}
}

func TestDatasetInsert(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})
	ds.Append([]any{"Charlie"})

	err := ds.Insert(1, []any{"Bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Height() != 3 {
		t.Errorf("expected height 3, got %d", ds.Height())
	}

	row, _ := ds.Row(1)
	if row[0] != "Bob" {
		t.Errorf("expected Bob, got %v", row[0])
	}
}

func TestDatasetPop(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})
	ds.Append([]any{"Bob"})

	row, err := ds.Pop(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if row[0] != "Alice" {
		t.Errorf("expected Alice, got %v", row[0])
	}

	if ds.Height() != 1 {
		t.Errorf("expected height 1, got %d", ds.Height())
	}
}

func TestDatasetLpopRpop(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})
	ds.Append([]any{"Bob"})
	ds.Append([]any{"Charlie"})

	row, _ := ds.Lpop()
	if row[0] != "Alice" {
		t.Errorf("expected Alice, got %v", row[0])
	}

	row, _ = ds.Rpop()
	if row[0] != "Charlie" {
		t.Errorf("expected Charlie, got %v", row[0])
	}

	if ds.Height() != 1 {
		t.Errorf("expected height 1, got %d", ds.Height())
	}
}

func TestDatasetColumn(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	col, err := ds.Column(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if col[0] != 30 || col[1] != 25 {
		t.Errorf("unexpected column values: %v", col)
	}
}

func TestDatasetColumnByHeader(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	col, err := ds.ColumnByHeader("Age")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if col[0] != 30 || col[1] != 25 {
		t.Errorf("unexpected column values: %v", col)
	}
}

func TestDatasetAppendCol(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})
	ds.Append([]any{"Bob"})

	err := ds.AppendCol("Age", []any{30, 25})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Width() != 2 {
		t.Errorf("expected width 2, got %d", ds.Width())
	}

	val, _ := ds.Get(0, 1)
	if val != 30 {
		t.Errorf("expected 30, got %v", val)
	}
}

func TestDatasetDeleteCol(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age", "City"})
	ds.Append([]any{"Alice", 30, "NYC"})

	err := ds.DeleteCol(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Width() != 2 {
		t.Errorf("expected width 2, got %d", ds.Width())
	}

	headers := ds.Headers()
	if headers[0] != "Name" || headers[1] != "City" {
		t.Errorf("unexpected headers: %v", headers)
	}
}

func TestDatasetSort(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Charlie", 35})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	sorted, err := ds.Sort(1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	row, _ := sorted.Row(0)
	if row[1] != 25 {
		t.Errorf("expected 25, got %v", row[1])
	}

	row, _ = sorted.Row(2)
	if row[1] != 35 {
		t.Errorf("expected 35, got %v", row[1])
	}
}

func TestDatasetFilter(t *testing.T) {
	ds := NewDataset([]string{"Name", "Role"})
	ds.AppendTagged([]any{"Alice", "Admin"}, []string{"admin"})
	ds.AppendTagged([]any{"Bob", "User"}, []string{"user"})
	ds.AppendTagged([]any{"Charlie", "Admin"}, []string{"admin"})

	filtered := ds.Filter("admin")

	if filtered.Height() != 2 {
		t.Errorf("expected height 2, got %d", filtered.Height())
	}
}

func TestDatasetStackRows(t *testing.T) {
	ds1 := NewDataset([]string{"Name"})
	ds1.Append([]any{"Alice"})

	ds2 := NewDataset([]string{"Name"})
	ds2.Append([]any{"Bob"})

	stacked, err := ds1.StackRows(ds2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stacked.Height() != 2 {
		t.Errorf("expected height 2, got %d", stacked.Height())
	}
}

func TestDatasetStackCols(t *testing.T) {
	ds1 := NewDataset([]string{"Name"})
	ds1.Append([]any{"Alice"})

	ds2 := NewDataset([]string{"Age"})
	ds2.Append([]any{30})

	stacked, err := ds1.StackCols(ds2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stacked.Width() != 2 {
		t.Errorf("expected width 2, got %d", stacked.Width())
	}
}

func TestDatasetSubset(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age", "City"})
	ds.Append([]any{"Alice", 30, "NYC"})

	subset, err := ds.Subset([]string{"Name", "City"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if subset.Width() != 2 {
		t.Errorf("expected width 2, got %d", subset.Width())
	}

	headers := subset.Headers()
	if headers[0] != "Name" || headers[1] != "City" {
		t.Errorf("unexpected headers: %v", headers)
	}
}

func TestDatasetRemoveDuplicates(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})
	ds.Append([]any{"Bob"})
	ds.Append([]any{"Alice"})

	unique := ds.RemoveDuplicates()

	if unique.Height() != 2 {
		t.Errorf("expected height 2, got %d", unique.Height())
	}
}

func TestDatasetDict(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})

	dict, err := ds.Dict()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dict[0]["Name"] != "Alice" || dict[0]["Age"] != 30 {
		t.Errorf("unexpected dict: %v", dict[0])
	}
}

func TestDatasetDynamicColumn(t *testing.T) {
	ds := NewDataset([]string{"FirstName", "LastName"})
	ds.Append([]any{"Alice", "Smith"})

	ds.AddDynamicColumn("FullName", func(row []any) any {
		return row[0].(string) + " " + row[1].(string)
	})

	dict, _ := ds.Dict()
	if dict[0]["FullName"] != "Alice Smith" {
		t.Errorf("expected 'Alice Smith', got %v", dict[0]["FullName"])
	}
}

func TestDatabook(t *testing.T) {
	db := NewDatabook()

	ds1 := NewDataset([]string{"Name"})
	ds1.SetTitle("Sheet1")
	ds1.Append([]any{"Alice"})

	ds2 := NewDataset([]string{"City"})
	ds2.SetTitle("Sheet2")
	ds2.Append([]any{"NYC"})

	db.AddSheet(ds1)
	db.AddSheet(ds2)

	if db.Size() != 2 {
		t.Errorf("expected size 2, got %d", db.Size())
	}

	sheet, err := db.SheetByTitle("Sheet1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sheet.Title() != "Sheet1" {
		t.Errorf("expected Sheet1, got %s", sheet.Title())
	}

	db.Wipe()
	if db.Size() != 0 {
		t.Errorf("expected size 0, got %d", db.Size())
	}
}

func TestExportCSV(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	var buf bytes.Buffer
	err := ds.Export(FormatCSV, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Name,Age\nAlice,30\nBob,25\n"
	if buf.String() != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, buf.String())
	}
}

func TestImportCSV(t *testing.T) {
	csv := "Name,Age\nAlice,30\nBob,25"

	ds, err := ImportString(FormatCSV, csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Height() != 2 {
		t.Errorf("expected height 2, got %d", ds.Height())
	}

	headers := ds.Headers()
	if headers[0] != "Name" || headers[1] != "Age" {
		t.Errorf("unexpected headers: %v", headers)
	}
}

func TestExportJSON(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})

	var buf bytes.Buffer
	err := ds.Export(FormatJSON, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "Alice") {
		t.Errorf("expected JSON to contain Alice, got:\n%s", buf.String())
	}
}

func TestImportJSON(t *testing.T) {
	jsonData := `[{"Name": "Alice", "Age": 30}, {"Name": "Bob", "Age": 25}]`

	ds, err := ImportString(FormatJSON, jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Height() != 2 {
		t.Errorf("expected height 2, got %d", ds.Height())
	}
}

func TestExportHTML(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})

	var buf bytes.Buffer
	err := ds.Export(FormatHTML, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "<table>") {
		t.Errorf("expected HTML table, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "<th>Name</th>") {
		t.Errorf("expected header, got:\n%s", buf.String())
	}
}

func TestExportMarkdown(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})

	var buf bytes.Buffer
	err := ds.Export(FormatMarkdown, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "| Name") {
		t.Errorf("expected markdown table, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "---") {
		t.Errorf("expected separator, got:\n%s", buf.String())
	}
}

func TestExportSQL(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.SetTitle("users")
	ds.Append([]any{"Alice", 30})

	var buf bytes.Buffer
	err := ds.Export(FormatSQL, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "INSERT INTO") {
		t.Errorf("expected SQL INSERT, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "users") {
		t.Errorf("expected table name, got:\n%s", buf.String())
	}
}

func TestExportLatex(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})

	var buf bytes.Buffer
	err := ds.Export(FormatLatex, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "\\begin{tabular}") {
		t.Errorf("expected LaTeX tabular, got:\n%s", buf.String())
	}
}

func TestLpushRpush(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Bob"})

	err := ds.Lpush([]any{"Alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = ds.Rpush([]any{"Charlie"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.Height() != 3 {
		t.Errorf("expected height 3, got %d", ds.Height())
	}

	row, _ := ds.Row(0)
	if row[0] != "Alice" {
		t.Errorf("expected Alice at index 0, got %v", row[0])
	}

	row, _ = ds.Row(2)
	if row[0] != "Charlie" {
		t.Errorf("expected Charlie at index 2, got %v", row[0])
	}
}

func TestLpushColRpushCol(t *testing.T) {
	ds := NewDataset([]string{"Age"})
	ds.Append([]any{30})
	ds.Append([]any{25})

	err := ds.LpushCol("Name", []any{"Alice", "Bob"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = ds.RpushCol("City", []any{"NYC", "LA"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headers := ds.Headers()
	if headers[0] != "Name" {
		t.Errorf("expected Name at index 0, got %s", headers[0])
	}
	if headers[2] != "City" {
		t.Errorf("expected City at index 2, got %s", headers[2])
	}
}

func TestSeparators(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.InsertSeparator(1, "Section 2")
	ds.Append([]any{"Bob", 25})
	ds.AppendSeparator("End")

	if !ds.HasSeparator(1) {
		t.Error("expected separator at index 1")
	}

	sep, ok := ds.GetSeparator(1)
	if !ok || sep.Text != "Section 2" {
		t.Errorf("expected separator 'Section 2', got %v", sep)
	}

	seps := ds.Separators()
	if len(seps) != 2 {
		t.Errorf("expected 2 separators, got %d", len(seps))
	}
}

func TestFormatters(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"alice"})

	// Add formatter to uppercase strings
	ds.AddFormatter(func(value any) any {
		if s, ok := value.(string); ok {
			return strings.ToUpper(s)
		}
		return value
	})

	result := ds.ApplyFormatters("hello")
	if result != "HELLO" {
		t.Errorf("expected HELLO, got %v", result)
	}
}

func TestExportRST(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	var buf bytes.Buffer
	err := ds.Export(FormatRST, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "+---") {
		t.Errorf("expected RST table border, got:\n%s", output)
	}
	if !strings.Contains(output, "| Name") {
		t.Errorf("expected Name header, got:\n%s", output)
	}
	if !strings.Contains(output, "===") {
		t.Errorf("expected header separator, got:\n%s", output)
	}
}

func TestExportJira(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	var buf bytes.Buffer
	err := ds.Export(FormatJira, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "||Name||") {
		t.Errorf("expected Jira header format, got:\n%s", output)
	}
	if !strings.Contains(output, "|Alice|") {
		t.Errorf("expected Jira cell format, got:\n%s", output)
	}
}

func TestExportCLI(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	var buf bytes.Buffer
	err := ds.Export(FormatCLI, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "┌") {
		t.Errorf("expected CLI table corner, got:\n%s", output)
	}
	if !strings.Contains(output, "│ Name") {
		t.Errorf("expected CLI table cell, got:\n%s", output)
	}
}

func TestExportCLIWithOptions(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", 30})

	var buf bytes.Buffer
	opts := CLIOptions{BorderStyle: "ascii"}
	err := ds.ExportCLI(&buf, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "+") {
		t.Errorf("expected ASCII border, got:\n%s", output)
	}
}

func TestExportDBF(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", "30"})
	ds.Append([]any{"Bob", "25"})

	var buf bytes.Buffer
	err := ds.Export(FormatDBF, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// DBF files have specific structure
	data := buf.Bytes()
	if len(data) < 32 {
		t.Error("DBF file too small")
	}
	// Check version byte
	if data[0] != 0x03 {
		t.Errorf("expected DBF version 0x03, got 0x%02x", data[0])
	}
}

func TestImportDBF(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.Append([]any{"Alice", "30"})
	ds.Append([]any{"Bob", "25"})

	// Export to DBF
	var buf bytes.Buffer
	err := ds.Export(FormatDBF, &buf)
	if err != nil {
		t.Fatalf("export error: %v", err)
	}

	// Import back
	imported, err := Import(FormatDBF, bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("import error: %v", err)
	}

	if imported.Height() != 2 {
		t.Errorf("expected height 2, got %d", imported.Height())
	}

	headers := imported.Headers()
	if headers[0] != "NAME" { // DBF uppercases headers
		t.Errorf("expected NAME header, got %s", headers[0])
	}
}

func TestExportODS(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.SetTitle("TestSheet")
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	var buf bytes.Buffer
	err := ds.Export(FormatODS, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ODS is a ZIP file
	data := buf.Bytes()
	if len(data) < 4 {
		t.Error("ODS file too small")
	}
	// Check ZIP magic number
	if data[0] != 0x50 || data[1] != 0x4B {
		t.Error("expected ZIP file signature")
	}
}

func TestExportXLS(t *testing.T) {
	ds := NewDataset([]string{"Name", "Age"})
	ds.SetTitle("TestSheet")
	ds.Append([]any{"Alice", 30})
	ds.Append([]any{"Bob", 25})

	var buf bytes.Buffer
	err := ds.Export(FormatXLS, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<?xml") {
		t.Errorf("expected XML header, got:\n%s", output[:100])
	}
	if !strings.Contains(output, "Workbook") {
		t.Errorf("expected Workbook element, got:\n%s", output[:200])
	}
}

func TestCopyWithFormattersAndSeparators(t *testing.T) {
	ds := NewDataset([]string{"Name"})
	ds.Append([]any{"Alice"})
	ds.AddFormatter(func(v any) any { return v })
	ds.InsertSeparator(0, "Start")

	copied := ds.Copy()

	if !copied.HasSeparator(0) {
		t.Error("expected separator to be copied")
	}

	// Formatters should be copied
	if len(copied.formatters) != 1 {
		t.Errorf("expected 1 formatter, got %d", len(copied.formatters))
	}
}


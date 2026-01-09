# tablib-go

A Go implementation of the Python [tablib](https://tablib.readthedocs.io/) library. Provides a clean and elegant API for working with tabular data, supporting multiple import/export formats and rich data manipulation features.

## Features

- **Clean API** - Idiomatic Go design, easy to use
- **Multiple Formats** - CSV, TSV, JSON, YAML, XLSX, XLS, ODS, DBF, HTML, Markdown, LaTeX, SQL, RST, Jira, CLI
- **Rich Data Operations** - Sort, filter, deduplicate, transpose, merge, and more
- **Dynamic Columns** - Compute column values via functions
- **Tag-based Filtering** - Add tags to rows and filter by tags
- **Separators** - Add separator rows for visual grouping in exports
- **Formatters** - Apply custom formatting functions to cell values during export
- **Multi-sheet Support** - Databook manages multiple Datasets

## Installation

```bash
go get github.com/pigletfly/tablib-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"

    "tablib-go"
)

func main() {
    // Create a Dataset
    ds := tablib.NewDataset([]string{"Name", "Age", "City"})

    // Add rows
    ds.Append([]any{"Alice", 30, "New York"})
    ds.Append([]any{"Bob", 25, "Los Angeles"})
    ds.Append([]any{"Charlie", 35, "Chicago"})

    // Export to CSV
    ds.Export(tablib.FormatCSV, os.Stdout)
    // Output:
    // Name,Age,City
    // Alice,30,New York
    // Bob,25,Los Angeles
    // Charlie,35,Chicago

    // Export to JSON string
    jsonStr, _ := ds.ExportString(tablib.FormatJSON)
    fmt.Println(jsonStr)
}
```

## Core Concepts

### Dataset

Dataset is the core data structure representing a two-dimensional table.

```go
// Create an empty Dataset with headers
ds := tablib.NewDataset([]string{"Column1", "Column2"})

// Create a Dataset with initial data
ds, err := tablib.NewDatasetWithData(
    []string{"Name", "Age"},
    [][]any{
        {"Alice", 30},
        {"Bob", 25},
    },
)
```

### Databook

Databook manages multiple Datasets, similar to an Excel workbook with multiple sheets.

```go
db := tablib.NewDatabook()

sheet1 := tablib.NewDataset([]string{"Name"})
sheet1.SetTitle("Users")
sheet1.Append([]any{"Alice"})

sheet2 := tablib.NewDataset([]string{"Product"})
sheet2.SetTitle("Products")
sheet2.Append([]any{"Laptop"})

db.AddSheet(sheet1)
db.AddSheet(sheet2)

// Export to multi-sheet Excel file
file, _ := os.Create("workbook.xlsx")
db.Export(tablib.FormatXLSX, file)
```

## Data Operations

### Row Operations

```go
ds := tablib.NewDataset([]string{"Name", "Age"})

// Append a row
ds.Append([]any{"Alice", 30})

// Prepend a row (add at beginning)
ds.Lpush([]any{"Bob", 25})

// Append a row (alias for Append)
ds.Rpush([]any{"Charlie", 35})

// Insert a row at index
ds.Insert(1, []any{"David", 28})

// Get a row
row, _ := ds.Row(0)

// Remove and return a row
row, _ = ds.Pop(0)      // Remove by index
row, _ = ds.Lpop()      // Remove first row
row, _ = ds.Rpop()      // Remove last row
```

### Column Operations

```go
ds := tablib.NewDataset([]string{"Name"})
ds.Append([]any{"Alice"})
ds.Append([]any{"Bob"})

// Append a column at the end
ds.AppendCol("Age", []any{30, 25})

// Prepend a column at the beginning
ds.LpushCol("ID", []any{1, 2})

// Append a column (alias for AppendCol)
ds.RpushCol("City", []any{"NYC", "LA"})

// Insert a column at index
ds.InsertCol(2, "Country", []any{"USA", "USA"})

// Get column data
col, _ := ds.Column(0)
col, _ = ds.ColumnByHeader("Name")

// Delete a column
ds.DeleteCol(1)
ds.DeleteColByHeader("Age")
```

### Cell Operations

```go
// Get cell value
value, _ := ds.Get(0, 1)  // row=0, col=1

// Set cell value
ds.Set(0, 1, "new value")
```

### Sorting

```go
ds := tablib.NewDataset([]string{"Name", "Age"})
ds.Append([]any{"Charlie", 35})
ds.Append([]any{"Alice", 30})
ds.Append([]any{"Bob", 25})

// Sort by column index
sorted, _ := ds.Sort(1, false)  // Sort by Age ascending

// Sort by column header
sorted, _ = ds.SortByHeader("Age", true)  // Sort by Age descending
```

### Filtering with Tags

```go
ds := tablib.NewDataset([]string{"Name", "Role"})

// Add rows with tags
ds.AppendTagged([]any{"Alice", "Admin"}, []string{"admin", "active"})
ds.AppendTagged([]any{"Bob", "User"}, []string{"user", "active"})
ds.AppendTagged([]any{"Charlie", "Admin"}, []string{"admin", "inactive"})

// Filter by tag
admins := ds.Filter("admin")       // Returns all admins
activeUsers := ds.Filter("active") // Returns all active users
```

### Subset

```go
ds := tablib.NewDataset([]string{"Name", "Age", "City", "Country"})
ds.Append([]any{"Alice", 30, "NYC", "USA"})

// Select specific columns
subset, _ := ds.Subset([]string{"Name", "City"})
```

### Stacking

```go
ds1 := tablib.NewDataset([]string{"Name"})
ds1.Append([]any{"Alice"})

ds2 := tablib.NewDataset([]string{"Name"})
ds2.Append([]any{"Bob"})

// Stack rows (vertical concatenation)
stacked, _ := ds1.StackRows(ds2)

// Stack columns (horizontal concatenation)
ds3 := tablib.NewDataset([]string{"Age"})
ds3.Append([]any{30})

stacked, _ = ds1.StackCols(ds3)
```

### Transpose

```go
ds := tablib.NewDataset([]string{"Name", "Age"})
ds.Append([]any{"Alice", 30})
ds.Append([]any{"Bob", 25})

transposed := ds.Transpose()
```

### Remove Duplicates

```go
ds := tablib.NewDataset([]string{"Name"})
ds.Append([]any{"Alice"})
ds.Append([]any{"Bob"})
ds.Append([]any{"Alice"})  // Duplicate

unique := ds.RemoveDuplicates()  // Only Alice and Bob remain
```

### Dynamic Columns

Dynamic columns are virtual columns computed via functions, not stored in the dataset.

```go
ds := tablib.NewDataset([]string{"FirstName", "LastName", "Salary"})
ds.Append([]any{"Alice", "Smith", 50000})
ds.Append([]any{"Bob", "Johnson", 60000})

// Add dynamic column: full name
ds.AddDynamicColumn("FullName", func(row []any) any {
    return row[0].(string) + " " + row[1].(string)
})

// Add dynamic column: net salary
ds.AddDynamicColumn("NetSalary", func(row []any) any {
    salary := row[2].(int)
    return float64(salary) * 0.8
})

// Dynamic columns appear in Dict() output
records, _ := ds.Dict()
fmt.Println(records[0]["FullName"])   // "Alice Smith"
fmt.Println(records[0]["NetSalary"])  // 40000
```

### Separators

Separators allow you to add visual dividers between rows in supported export formats.

```go
ds := tablib.NewDataset([]string{"Name", "Department"})
ds.Append([]any{"Alice", "Engineering"})
ds.Append([]any{"Bob", "Engineering"})

// Add a separator before the next rows
ds.InsertSeparator(2, "--- Marketing Team ---")

ds.Append([]any{"Charlie", "Marketing"})
ds.Append([]any{"David", "Marketing"})

// Add a separator at the end
ds.AppendSeparator("--- End of List ---")

// Check for separators
if ds.HasSeparator(2) {
    sep, _ := ds.GetSeparator(2)
    fmt.Println(sep.Text)  // "--- Marketing Team ---"
}
```

### Formatters

Formatters are functions applied to cell values during export.

```go
ds := tablib.NewDataset([]string{"Name", "Salary"})
ds.Append([]any{"Alice", 50000})

// Add a formatter to format numbers as currency
ds.AddFormatter(func(value any) any {
    if num, ok := value.(int); ok {
        return fmt.Sprintf("$%d", num)
    }
    return value
})

// Apply formatters manually
result := ds.ApplyFormatters(50000)  // "$50000"
```

## Format Support

### Export Formats

| Format | Constant | Description |
|--------|----------|-------------|
| CSV | `FormatCSV` | Comma-separated values |
| TSV | `FormatTSV` | Tab-separated values |
| JSON | `FormatJSON` | Array of objects (with headers) or array of arrays |
| YAML | `FormatYAML` | Same structure as JSON |
| XLSX | `FormatXLSX` | Microsoft Excel format |
| XLS | `FormatXLS` | Microsoft Excel XML format (compatible with Excel) |
| ODS | `FormatODS` | OpenDocument Spreadsheet |
| DBF | `FormatDBF` | dBase format |
| HTML | `FormatHTML` | HTML table |
| Markdown | `FormatMarkdown` | Markdown table |
| LaTeX | `FormatLatex` | LaTeX tabular environment |
| SQL | `FormatSQL` | INSERT statements |
| RST | `FormatRST` | reStructuredText grid table |
| Jira | `FormatJira` | Jira Wiki markup table |
| CLI | `FormatCLI` | ASCII table for command line |

### Import Formats

| Format | Import Supported |
|--------|------------------|
| CSV/TSV | ✅ |
| JSON | ✅ |
| YAML | ✅ |
| XLSX | ✅ |
| DBF | ✅ |
| ODS | ✅ (via ImportODS) |
| XLS | ✅ (XML format via ImportXLS) |

### Export Examples

```go
ds := tablib.NewDataset([]string{"Name", "Age"})
ds.Append([]any{"Alice", 30})

// Export to io.Writer
var buf bytes.Buffer
ds.Export(tablib.FormatJSON, &buf)

// Export to string
str, _ := ds.ExportString(tablib.FormatJSON)

// Export to file
file, _ := os.Create("data.xlsx")
defer file.Close()
ds.Export(tablib.FormatXLSX, file)
```

### Import Examples

```go
// Import from string
csvData := `Name,Age
Alice,30
Bob,25`
ds, _ := tablib.ImportString(tablib.FormatCSV, csvData)

// Import from io.Reader
file, _ := os.Open("data.json")
ds, _ = tablib.Import(tablib.FormatJSON, file)

// Import Excel with specific sheet
file, _ = os.Open("workbook.xlsx")
ds, _ = tablib.ImportXLSX(file, "Sheet1")

// Import all Excel sheets into Databook
file, _ = os.Open("workbook.xlsx")
db, _ := tablib.ImportXLSXDatabook(file)
```

### Format Options

Some formats support custom options:

```go
// CSV with custom delimiter
opts := tablib.CSVOptions{
    Delimiter:   ';',
    WriteHeader: true,
}
ds.ExportCSV(writer, opts)

// Import CSV with custom options
ds, _ := tablib.ImportCSV(reader, ';', true)

// HTML with custom attributes
htmlOpts := tablib.HTMLOptions{
    TableClass: "data-table",
    TableID:    "users",
}
ds.ExportHTML(writer, htmlOpts)

// SQL with custom table name
sqlOpts := tablib.SQLOptions{
    TableName: "users",
}
ds.ExportSQL(writer, sqlOpts)

// CLI with custom border style
cliOpts := tablib.CLIOptions{
    BorderStyle: "double",  // "single", "double", "ascii", "none"
}
ds.ExportCLI(writer, cliOpts)
```

## Output Format Examples

### JSON

```json
[
  {"Name": "Alice", "Age": 30},
  {"Name": "Bob", "Age": 25}
]
```

### Markdown

```markdown
| Name  | Age |
| ----- | --- |
| Alice | 30  |
| Bob   | 25  |
```

### HTML

```html
<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Age</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Alice</td>
      <td>30</td>
    </tr>
  </tbody>
</table>
```

### SQL

```sql
INSERT INTO "users" ("Name", "Age") VALUES ('Alice', 30);
INSERT INTO "users" ("Name", "Age") VALUES ('Bob', 25);
```

### LaTeX

```latex
\begin{tabular}{ll}
\hline
Name & Age \\
\hline
Alice & 30 \\
Bob & 25 \\
\hline
\end{tabular}
```

### RST (reStructuredText)

```
+-------+-----+
| Name  | Age |
+=======+=====+
| Alice | 30  |
+-------+-----+
| Bob   | 25  |
+-------+-----+
```

### Jira

```
||Name||Age||
|Alice|30|
|Bob|25|
```

### CLI

```
┌───────┬─────┐
│ Name  │ Age │
├───────┼─────┤
│ Alice │ 30  │
│ Bob   │ 25  │
└───────┴─────┘
```

## Error Handling

tablib-go defines the following error types:

| Error | Description |
|-------|-------------|
| `ErrInvalidDimensions` | Row/column size doesn't match dataset dimensions |
| `ErrHeadersRequired` | Operation requires headers but none are set |
| `ErrInvalidRowIndex` | Row index out of bounds |
| `ErrInvalidColumnIndex` | Column index out of bounds |
| `ErrColumnNotFound` | Specified column header not found |
| `ErrUnsupportedFormat` | Unsupported format |
| `ErrEmptyDataset` | Dataset is empty |
| `ErrInvalidData` | Invalid data format |

```go
ds := tablib.NewDataset([]string{"Name", "Age"})

err := ds.Append([]any{"Alice", 30, "Extra"})
if err == tablib.ErrInvalidDimensions {
    fmt.Println("Row size doesn't match headers")
}
```

## API Reference

### Dataset

| Method | Description |
|--------|-------------|
| `NewDataset(headers)` | Create a new Dataset |
| `NewDatasetWithData(headers, data)` | Create a Dataset with initial data |
| `Headers()` | Get headers |
| `SetHeaders(headers)` | Set headers |
| `Title()` / `SetTitle(title)` | Get/set title |
| `Height()` | Number of rows |
| `Width()` | Number of columns |
| `Append(row, tags...)` | Append a row |
| `AppendTagged(row, tags)` | Append a row with tags |
| `Lpush(row, tags...)` | Prepend a row |
| `Rpush(row, tags...)` | Append a row (alias for Append) |
| `Insert(index, row, tags...)` | Insert a row |
| `Pop(index)` | Remove and return row at index |
| `Lpop()` / `Rpop()` | Remove first/last row |
| `Row(index)` | Get row by index |
| `Column(index)` | Get column by index |
| `ColumnByHeader(header)` | Get column by header |
| `AppendCol(header, col)` | Append a column |
| `LpushCol(header, col)` | Prepend a column |
| `RpushCol(header, col)` | Append a column (alias for AppendCol) |
| `InsertCol(index, header, col)` | Insert a column |
| `DeleteCol(index)` | Delete column by index |
| `DeleteColByHeader(header)` | Delete column by header |
| `Get(row, col)` | Get cell value |
| `Set(row, col, value)` | Set cell value |
| `Filter(tag)` | Filter rows by tag |
| `Sort(colIndex, reverse)` | Sort by column index |
| `SortByHeader(header, reverse)` | Sort by column header |
| `Transpose()` | Transpose rows and columns |
| `StackRows(other)` | Stack datasets vertically |
| `StackCols(other)` | Stack datasets horizontally |
| `Subset(headers)` | Select column subset |
| `RemoveDuplicates()` | Remove duplicate rows |
| `Copy()` | Deep copy |
| `Dict()` | Convert to slice of maps |
| `Records()` | Convert to 2D slice |
| `Wipe()` | Clear all data |
| `AddDynamicColumn(header, fn)` | Add dynamic column |
| `AddFormatter(fn)` | Add a formatter function |
| `ApplyFormatters(value)` | Apply all formatters to a value |
| `InsertSeparator(index, text)` | Insert separator before row |
| `AppendSeparator(text)` | Append separator at end |
| `HasSeparator(index)` | Check if separator exists |
| `GetSeparator(index)` | Get separator at index |
| `Separators()` | Get all separators |
| `Export(format, writer)` | Export to writer |
| `ExportString(format)` | Export to string |

### Databook

| Method | Description |
|--------|-------------|
| `NewDatabook()` | Create a new Databook |
| `AddSheet(ds)` | Add a Dataset |
| `Sheet(index)` | Get sheet by index |
| `SheetByTitle(title)` | Get sheet by title |
| `Sheets()` | Get all sheets |
| `Size()` | Number of sheets |
| `RemoveSheet(index)` | Remove sheet by index |
| `Wipe()` | Remove all sheets |
| `Export(format, writer)` | Export to writer |
| `ExportString(format)` | Export to string |

### Import Functions

| Function | Description |
|----------|-------------|
| `Import(format, reader)` | Import from Reader |
| `ImportString(format, data)` | Import from string |
| `ImportCSV(reader, delimiter, hasHeaders)` | Import CSV with options |
| `ImportXLSX(reader, sheetName)` | Import Excel sheet |
| `ImportXLSXDatabook(reader)` | Import Excel as Databook |
| `ImportYAML(data)` | Import YAML data |
| `ImportODS(reader, size, sheetName)` | Import ODS sheet |
| `ImportXLS(reader, sheetName)` | Import XLS (XML format) |

## Dependencies

- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) - YAML support
- [github.com/xuri/excelize/v2](https://github.com/xuri/excelize) - Excel support

## License

MIT License

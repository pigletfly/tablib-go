package tablib

import (
	"cmp"
	"fmt"
	"slices"
)

// DynamicColumn represents a function that computes a column value based on a row.
type DynamicColumn func(row []any) any

// Formatter represents a function that formats cell values during export.
type Formatter func(value any) any

// Separator represents a separator row in the dataset.
type Separator struct {
	Text string
}

// Dataset is the primary data structure for tabular data.
type Dataset struct {
	headers     []string
	data        [][]any
	tags        [][]string // tags for each row
	title       string     // optional title for the dataset
	dynamicCols map[string]DynamicColumn
	formatters  []Formatter
	separators  map[int]Separator // row index -> separator (separator appears before the row)
}

// NewDataset creates a new empty Dataset.
func NewDataset(headers []string) *Dataset {
	h := make([]string, len(headers))
	copy(h, headers)
	return &Dataset{
		headers:     h,
		data:        make([][]any, 0),
		tags:        make([][]string, 0),
		dynamicCols: make(map[string]DynamicColumn),
		formatters:  make([]Formatter, 0),
		separators:  make(map[int]Separator),
	}
}

// NewDatasetWithData creates a Dataset with initial data.
func NewDatasetWithData(headers []string, data [][]any) (*Dataset, error) {
	ds := NewDataset(headers)
	for _, row := range data {
		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}
	return ds, nil
}

// Headers returns the headers of the dataset.
func (ds *Dataset) Headers() []string {
	if ds.headers == nil {
		return nil
	}
	h := make([]string, len(ds.headers))
	copy(h, ds.headers)
	return h
}

// SetHeaders sets the headers of the dataset.
func (ds *Dataset) SetHeaders(headers []string) error {
	if len(ds.data) > 0 && len(headers) != ds.Width() {
		return ErrInvalidDimensions
	}
	ds.headers = make([]string, len(headers))
	copy(ds.headers, headers)
	return nil
}

// Title returns the title of the dataset.
func (ds *Dataset) Title() string {
	return ds.title
}

// SetTitle sets the title of the dataset.
func (ds *Dataset) SetTitle(title string) {
	ds.title = title
}

// Height returns the number of rows in the dataset.
func (ds *Dataset) Height() int {
	return len(ds.data)
}

// Width returns the number of columns in the dataset.
func (ds *Dataset) Width() int {
	if len(ds.headers) > 0 {
		return len(ds.headers)
	}
	if len(ds.data) > 0 {
		return len(ds.data[0])
	}
	return 0
}

// Append adds a row to the dataset.
func (ds *Dataset) Append(row []any, rowTags ...string) error {
	if ds.Width() > 0 && len(row) != ds.Width() {
		return ErrInvalidDimensions
	}
	r := make([]any, len(row))
	copy(r, row)
	ds.data = append(ds.data, r)

	t := make([]string, len(rowTags))
	copy(t, rowTags)
	ds.tags = append(ds.tags, t)
	return nil
}

// AppendTagged adds a row with tags to the dataset.
func (ds *Dataset) AppendTagged(row []any, tags []string) error {
	return ds.Append(row, tags...)
}

// Insert inserts a row at the specified index.
func (ds *Dataset) Insert(index int, row []any, rowTags ...string) error {
	if index < 0 || index > len(ds.data) {
		return ErrInvalidRowIndex
	}
	if ds.Width() > 0 && len(row) != ds.Width() {
		return ErrInvalidDimensions
	}

	r := make([]any, len(row))
	copy(r, row)
	ds.data = slices.Insert(ds.data, index, r)

	t := make([]string, len(rowTags))
	copy(t, rowTags)
	ds.tags = slices.Insert(ds.tags, index, t)
	return nil
}

// Pop removes and returns the row at the specified index.
func (ds *Dataset) Pop(index int) ([]any, error) {
	if index < 0 || index >= len(ds.data) {
		return nil, ErrInvalidRowIndex
	}
	row := ds.data[index]
	ds.data = slices.Delete(ds.data, index, index+1)
	ds.tags = slices.Delete(ds.tags, index, index+1)
	return row, nil
}

// Lpop removes and returns the first row.
func (ds *Dataset) Lpop() ([]any, error) {
	if len(ds.data) == 0 {
		return nil, ErrEmptyDataset
	}
	return ds.Pop(0)
}

// Rpop removes and returns the last row.
func (ds *Dataset) Rpop() ([]any, error) {
	if len(ds.data) == 0 {
		return nil, ErrEmptyDataset
	}
	return ds.Pop(len(ds.data) - 1)
}

// Lpush adds a row at the beginning of the dataset.
func (ds *Dataset) Lpush(row []any, rowTags ...string) error {
	return ds.Insert(0, row, rowTags...)
}

// Rpush is an alias for Append - adds a row at the end of the dataset.
func (ds *Dataset) Rpush(row []any, rowTags ...string) error {
	return ds.Append(row, rowTags...)
}

// LpushCol adds a column at the beginning of the dataset.
func (ds *Dataset) LpushCol(header string, col []any) error {
	return ds.InsertCol(0, header, col)
}

// RpushCol is an alias for AppendCol - adds a column at the end of the dataset.
func (ds *Dataset) RpushCol(header string, col []any) error {
	return ds.AppendCol(header, col)
}

// Row returns the row at the specified index.
func (ds *Dataset) Row(index int) ([]any, error) {
	if index < 0 || index >= len(ds.data) {
		return nil, ErrInvalidRowIndex
	}
	row := make([]any, len(ds.data[index]))
	copy(row, ds.data[index])
	return row, nil
}

// Column returns all values in the specified column by index.
func (ds *Dataset) Column(index int) ([]any, error) {
	if index < 0 || index >= ds.Width() {
		return nil, ErrInvalidColumnIndex
	}
	col := make([]any, len(ds.data))
	for i, row := range ds.data {
		col[i] = row[index]
	}
	return col, nil
}

// ColumnByHeader returns all values in the column with the specified header.
func (ds *Dataset) ColumnByHeader(header string) ([]any, error) {
	index := ds.headerIndex(header)
	if index == -1 {
		return nil, ErrColumnNotFound
	}
	return ds.Column(index)
}

// AppendCol adds a column to the dataset.
func (ds *Dataset) AppendCol(header string, col []any) error {
	if len(ds.data) > 0 && len(col) != len(ds.data) {
		return ErrInvalidDimensions
	}

	// If dataset is empty, initialize rows
	if len(ds.data) == 0 && len(col) > 0 {
		ds.data = make([][]any, len(col))
		ds.tags = make([][]string, len(col))
		for i := range ds.data {
			ds.data[i] = make([]any, 0)
			ds.tags[i] = make([]string, 0)
		}
	}

	ds.headers = append(ds.headers, header)
	for i := range ds.data {
		ds.data[i] = append(ds.data[i], col[i])
	}
	return nil
}

// InsertCol inserts a column at the specified index.
func (ds *Dataset) InsertCol(index int, header string, col []any) error {
	if index < 0 || index > ds.Width() {
		return ErrInvalidColumnIndex
	}
	if len(ds.data) > 0 && len(col) != len(ds.data) {
		return ErrInvalidDimensions
	}

	// If dataset is empty, initialize rows
	if len(ds.data) == 0 && len(col) > 0 {
		ds.data = make([][]any, len(col))
		ds.tags = make([][]string, len(col))
		for i := range ds.data {
			ds.data[i] = make([]any, 0)
			ds.tags[i] = make([]string, 0)
		}
	}

	ds.headers = slices.Insert(ds.headers, index, header)
	for i := range ds.data {
		ds.data[i] = slices.Insert(ds.data[i], index, col[i])
	}
	return nil
}

// DeleteCol removes the column at the specified index.
func (ds *Dataset) DeleteCol(index int) error {
	if index < 0 || index >= ds.Width() {
		return ErrInvalidColumnIndex
	}
	ds.headers = slices.Delete(ds.headers, index, index+1)
	for i := range ds.data {
		ds.data[i] = slices.Delete(ds.data[i], index, index+1)
	}
	return nil
}

// DeleteColByHeader removes the column with the specified header.
func (ds *Dataset) DeleteColByHeader(header string) error {
	index := ds.headerIndex(header)
	if index == -1 {
		return ErrColumnNotFound
	}
	return ds.DeleteCol(index)
}

// AddDynamicColumn adds a dynamic (computed) column to the dataset.
func (ds *Dataset) AddDynamicColumn(header string, fn DynamicColumn) {
	ds.dynamicCols[header] = fn
}

// AddFormatter adds a formatter function that will be applied to cell values during export.
func (ds *Dataset) AddFormatter(fn Formatter) {
	ds.formatters = append(ds.formatters, fn)
}

// ApplyFormatters applies all registered formatters to a value.
func (ds *Dataset) ApplyFormatters(value any) any {
	result := value
	for _, fn := range ds.formatters {
		result = fn(result)
	}
	return result
}

// InsertSeparator inserts a separator before the row at the specified index.
func (ds *Dataset) InsertSeparator(index int, text string) error {
	if index < 0 || index > len(ds.data) {
		return ErrInvalidRowIndex
	}
	ds.separators[index] = Separator{Text: text}
	return nil
}

// AppendSeparator adds a separator at the end of the dataset (after the last row).
func (ds *Dataset) AppendSeparator(text string) {
	ds.separators[len(ds.data)] = Separator{Text: text}
}

// HasSeparator checks if there is a separator before the specified row index.
func (ds *Dataset) HasSeparator(index int) bool {
	_, ok := ds.separators[index]
	return ok
}

// GetSeparator returns the separator before the specified row index.
func (ds *Dataset) GetSeparator(index int) (Separator, bool) {
	sep, ok := ds.separators[index]
	return sep, ok
}

// Separators returns a copy of all separators.
func (ds *Dataset) Separators() map[int]Separator {
	result := make(map[int]Separator, len(ds.separators))
	for k, v := range ds.separators {
		result[k] = v
	}
	return result
}

// Get returns a cell value by row and column index.
func (ds *Dataset) Get(row, col int) (any, error) {
	if row < 0 || row >= len(ds.data) {
		return nil, ErrInvalidRowIndex
	}
	if col < 0 || col >= ds.Width() {
		return nil, ErrInvalidColumnIndex
	}
	return ds.data[row][col], nil
}

// Set sets a cell value by row and column index.
func (ds *Dataset) Set(row, col int, value any) error {
	if row < 0 || row >= len(ds.data) {
		return ErrInvalidRowIndex
	}
	if col < 0 || col >= ds.Width() {
		return ErrInvalidColumnIndex
	}
	ds.data[row][col] = value
	return nil
}

// Filter returns a new Dataset containing only rows with the specified tag.
func (ds *Dataset) Filter(tag string) *Dataset {
	result := NewDataset(ds.headers)
	result.title = ds.title
	for k, v := range ds.dynamicCols {
		result.dynamicCols[k] = v
	}

	for i, row := range ds.data {
		if slices.Contains(ds.tags[i], tag) {
			r := make([]any, len(row))
			copy(r, row)
			result.data = append(result.data, r)
			t := make([]string, len(ds.tags[i]))
			copy(t, ds.tags[i])
			result.tags = append(result.tags, t)
		}
	}
	return result
}

// Sort returns a new Dataset sorted by the specified column.
func (ds *Dataset) Sort(colIndex int, reverse bool) (*Dataset, error) {
	if colIndex < 0 || colIndex >= ds.Width() {
		return nil, ErrInvalidColumnIndex
	}

	result := ds.Copy()
	indices := make([]int, len(result.data))
	for i := range indices {
		indices[i] = i
	}

	slices.SortFunc(indices, func(i, j int) int {
		a, b := result.data[i][colIndex], result.data[j][colIndex]
		c := compareAny(a, b)
		if reverse {
			return -c
		}
		return c
	})

	newData := make([][]any, len(result.data))
	newTags := make([][]string, len(result.tags))
	for i, idx := range indices {
		newData[i] = result.data[idx]
		newTags[i] = result.tags[idx]
	}
	result.data = newData
	result.tags = newTags
	return result, nil
}

// SortByHeader returns a new Dataset sorted by the specified header.
func (ds *Dataset) SortByHeader(header string, reverse bool) (*Dataset, error) {
	index := ds.headerIndex(header)
	if index == -1 {
		return nil, ErrColumnNotFound
	}
	return ds.Sort(index, reverse)
}

// Transpose returns a new Dataset with rows and columns swapped.
func (ds *Dataset) Transpose() *Dataset {
	if len(ds.data) == 0 {
		return NewDataset(nil)
	}

	width := ds.Width()
	height := ds.Height()

	var newHeaders []string
	if len(ds.headers) > 0 {
		// First column becomes headers
		newHeaders = make([]string, height)
		for i := 0; i < height; i++ {
			if v, ok := ds.data[i][0].(string); ok {
				newHeaders[i] = v
			} else {
				newHeaders[i] = ""
			}
		}
	}

	result := NewDataset(newHeaders)
	result.title = ds.title

	startCol := 0
	if len(ds.headers) > 0 {
		startCol = 1
	}

	for col := startCol; col < width; col++ {
		row := make([]any, height)
		for r := 0; r < height; r++ {
			row[r] = ds.data[r][col]
		}
		result.data = append(result.data, row)
		result.tags = append(result.tags, []string{})
	}

	return result
}

// StackRows stacks another dataset below this one.
func (ds *Dataset) StackRows(other *Dataset) (*Dataset, error) {
	if ds.Width() != other.Width() {
		return nil, ErrInvalidDimensions
	}

	result := ds.Copy()
	for i, row := range other.data {
		r := make([]any, len(row))
		copy(r, row)
		result.data = append(result.data, r)
		t := make([]string, len(other.tags[i]))
		copy(t, other.tags[i])
		result.tags = append(result.tags, t)
	}
	return result, nil
}

// StackCols stacks another dataset to the right of this one.
func (ds *Dataset) StackCols(other *Dataset) (*Dataset, error) {
	if ds.Height() != other.Height() {
		return nil, ErrInvalidDimensions
	}

	result := ds.Copy()
	result.headers = append(result.headers, other.headers...)
	for i := range result.data {
		result.data[i] = append(result.data[i], other.data[i]...)
	}
	return result, nil
}

// Subset returns a new Dataset with only the specified columns.
func (ds *Dataset) Subset(headers []string) (*Dataset, error) {
	indices := make([]int, len(headers))
	for i, h := range headers {
		idx := ds.headerIndex(h)
		if idx == -1 {
			return nil, ErrColumnNotFound
		}
		indices[i] = idx
	}

	result := NewDataset(headers)
	result.title = ds.title
	for i, row := range ds.data {
		newRow := make([]any, len(indices))
		for j, idx := range indices {
			newRow[j] = row[idx]
		}
		result.data = append(result.data, newRow)
		t := make([]string, len(ds.tags[i]))
		copy(t, ds.tags[i])
		result.tags = append(result.tags, t)
	}
	return result, nil
}

// RemoveDuplicates returns a new Dataset with duplicate rows removed.
func (ds *Dataset) RemoveDuplicates() *Dataset {
	result := NewDataset(ds.headers)
	result.title = ds.title
	for k, v := range ds.dynamicCols {
		result.dynamicCols[k] = v
	}

	seen := make(map[string]bool)
	for i, row := range ds.data {
		key := rowKey(row)
		if !seen[key] {
			seen[key] = true
			r := make([]any, len(row))
			copy(r, row)
			result.data = append(result.data, r)
			t := make([]string, len(ds.tags[i]))
			copy(t, ds.tags[i])
			result.tags = append(result.tags, t)
		}
	}
	return result
}

// Copy returns a deep copy of the dataset.
func (ds *Dataset) Copy() *Dataset {
	result := NewDataset(ds.headers)
	result.title = ds.title
	for k, v := range ds.dynamicCols {
		result.dynamicCols[k] = v
	}
	result.formatters = append(result.formatters, ds.formatters...)
	for k, v := range ds.separators {
		result.separators[k] = v
	}
	for i, row := range ds.data {
		r := make([]any, len(row))
		copy(r, row)
		result.data = append(result.data, r)
		t := make([]string, len(ds.tags[i]))
		copy(t, ds.tags[i])
		result.tags = append(result.tags, t)
	}
	return result
}

// Dict returns the dataset as a slice of maps (one map per row).
func (ds *Dataset) Dict() ([]map[string]any, error) {
	if len(ds.headers) == 0 {
		return nil, ErrHeadersRequired
	}

	result := make([]map[string]any, len(ds.data))
	for i, row := range ds.data {
		m := make(map[string]any)
		for j, h := range ds.headers {
			m[h] = row[j]
		}
		// Add dynamic columns
		for h, fn := range ds.dynamicCols {
			m[h] = fn(row)
		}
		result[i] = m
	}
	return result, nil
}

// Records returns all rows as a slice of slices.
func (ds *Dataset) Records() [][]any {
	result := make([][]any, len(ds.data))
	for i, row := range ds.data {
		r := make([]any, len(row))
		copy(r, row)
		// Add dynamic columns
		for _, fn := range ds.dynamicCols {
			r = append(r, fn(row))
		}
		result[i] = r
	}
	return result
}

// Wipe clears all data from the dataset.
func (ds *Dataset) Wipe() {
	ds.data = make([][]any, 0)
	ds.tags = make([][]string, 0)
}

// headerIndex returns the index of the header, or -1 if not found.
func (ds *Dataset) headerIndex(header string) int {
	for i, h := range ds.headers {
		if h == header {
			return i
		}
	}
	return -1
}

// rowKey generates a string key for a row for deduplication.
func rowKey(row []any) string {
	var s string
	for _, v := range row {
		s += fmt.Sprintf("%v\x00", v)
	}
	return s
}

// compareAny compares two values of any type.
func compareAny(a, b any) int {
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok {
			return cmp.Compare(va, vb)
		}
	case int64:
		if vb, ok := b.(int64); ok {
			return cmp.Compare(va, vb)
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return cmp.Compare(va, vb)
		}
	case string:
		if vb, ok := b.(string); ok {
			return cmp.Compare(va, vb)
		}
	}
	// Fallback to string comparison
	return cmp.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

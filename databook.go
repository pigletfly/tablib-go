package tablib

// Databook is a collection of Datasets, similar to a workbook with multiple sheets.
type Databook struct {
	sheets []*Dataset
}

// NewDatabook creates a new empty Databook.
func NewDatabook() *Databook {
	return &Databook{
		sheets: make([]*Dataset, 0),
	}
}

// AddSheet adds a Dataset to the Databook.
func (db *Databook) AddSheet(ds *Dataset) {
	db.sheets = append(db.sheets, ds)
}

// Sheet returns the Dataset at the specified index.
func (db *Databook) Sheet(index int) (*Dataset, error) {
	if index < 0 || index >= len(db.sheets) {
		return nil, ErrInvalidRowIndex
	}
	return db.sheets[index], nil
}

// SheetByTitle returns the first Dataset with the specified title.
func (db *Databook) SheetByTitle(title string) (*Dataset, error) {
	for _, ds := range db.sheets {
		if ds.Title() == title {
			return ds, nil
		}
	}
	return nil, ErrColumnNotFound
}

// Size returns the number of Datasets in the Databook.
func (db *Databook) Size() int {
	return len(db.sheets)
}

// Sheets returns all Datasets in the Databook.
func (db *Databook) Sheets() []*Dataset {
	result := make([]*Dataset, len(db.sheets))
	copy(result, db.sheets)
	return result
}

// RemoveSheet removes the Dataset at the specified index.
func (db *Databook) RemoveSheet(index int) error {
	if index < 0 || index >= len(db.sheets) {
		return ErrInvalidRowIndex
	}
	db.sheets = append(db.sheets[:index], db.sheets[index+1:]...)
	return nil
}

// Wipe removes all Datasets from the Databook.
func (db *Databook) Wipe() {
	db.sheets = make([]*Dataset, 0)
}

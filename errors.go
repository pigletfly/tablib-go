package tablib

import "errors"

var (
	// ErrInvalidDimensions is returned when the size of row/column doesn't match the dataset dimensions.
	ErrInvalidDimensions = errors.New("tablib: invalid dimensions")

	// ErrHeadersRequired is returned when an operation requires headers but none are set.
	ErrHeadersRequired = errors.New("tablib: headers required")

	// ErrInvalidRowIndex is returned when accessing a row with an out-of-bounds index.
	ErrInvalidRowIndex = errors.New("tablib: invalid row index")

	// ErrInvalidColumnIndex is returned when accessing a column with an out-of-bounds index.
	ErrInvalidColumnIndex = errors.New("tablib: invalid column index")

	// ErrColumnNotFound is returned when a column with the specified header is not found.
	ErrColumnNotFound = errors.New("tablib: column not found")

	// ErrUnsupportedFormat is returned when attempting to use an unsupported format.
	ErrUnsupportedFormat = errors.New("tablib: unsupported format")

	// ErrEmptyDataset is returned when an operation requires data but the dataset is empty.
	ErrEmptyDataset = errors.New("tablib: empty dataset")

	// ErrInvalidData is returned when the input data is malformed or invalid.
	ErrInvalidData = errors.New("tablib: invalid data")
)

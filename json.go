package tablib

import (
	"encoding/json"
	"io"
)

func init() {
	RegisterExporter(FormatJSON, ExporterFunc(exportJSON))
	RegisterImporter(FormatJSON, ImporterFunc(importJSON))
	RegisterDatabookExporter(FormatJSON, DatabookExporterFunc(exportDatabookJSON))
}

func exportJSON(ds *Dataset, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if len(ds.headers) > 0 {
		// Export as array of objects
		records, err := ds.Dict()
		if err != nil {
			return err
		}
		return encoder.Encode(records)
	}

	// Export as array of arrays
	return encoder.Encode(ds.Records())
}

func importJSON(r io.Reader) (*Dataset, error) {
	decoder := json.NewDecoder(r)

	// First, decode into a raw JSON value to determine the structure
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, err
	}

	// Try to decode as array of objects
	var objects []map[string]any
	if err := json.Unmarshal(raw, &objects); err == nil && len(objects) > 0 {
		return importJSONObjects(objects)
	}

	// Try to decode as array of arrays
	var arrays [][]any
	if err := json.Unmarshal(raw, &arrays); err == nil {
		return importJSONArrays(arrays)
	}

	return nil, ErrInvalidData
}

func importJSONObjects(objects []map[string]any) (*Dataset, error) {
	if len(objects) == 0 {
		return NewDataset(nil), nil
	}

	// Extract headers from the first object
	// Note: map iteration order is not guaranteed, so we collect all keys
	headerSet := make(map[string]bool)
	for _, obj := range objects {
		for k := range obj {
			headerSet[k] = true
		}
	}

	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}

	ds := NewDataset(headers)

	for _, obj := range objects {
		row := make([]any, len(headers))
		for i, h := range headers {
			row[i] = obj[h]
		}
		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func importJSONArrays(arrays [][]any) (*Dataset, error) {
	ds := NewDataset(nil)

	for _, arr := range arrays {
		row := make([]any, len(arr))
		copy(row, arr)
		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

func exportDatabookJSON(db *Databook, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	result := make([]map[string]any, 0, db.Size())
	for _, ds := range db.sheets {
		sheet := make(map[string]any)
		sheet["title"] = ds.Title()

		if len(ds.headers) > 0 {
			records, err := ds.Dict()
			if err != nil {
				return err
			}
			sheet["data"] = records
		} else {
			sheet["data"] = ds.Records()
		}
		result = append(result, sheet)
	}

	return encoder.Encode(result)
}

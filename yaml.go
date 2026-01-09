package tablib

import (
	"io"

	"gopkg.in/yaml.v3"
)

func init() {
	RegisterExporter(FormatYAML, ExporterFunc(exportYAML))
	RegisterImporter(FormatYAML, ImporterFunc(importYAML))
	RegisterDatabookExporter(FormatYAML, DatabookExporterFunc(exportDatabookYAML))
}

func exportYAML(ds *Dataset, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()

	if len(ds.headers) > 0 {
		records, err := ds.Dict()
		if err != nil {
			return err
		}
		return encoder.Encode(records)
	}

	return encoder.Encode(ds.Records())
}

func importYAML(r io.Reader) (*Dataset, error) {
	decoder := yaml.NewDecoder(r)

	// Try to decode as array of objects first
	var objects []map[string]any
	if err := decoder.Decode(&objects); err == nil && len(objects) > 0 {
		return importYAMLObjects(objects)
	}

	// Reset reader is not possible, so we need a different approach
	// Re-read and try as array of arrays
	return nil, ErrInvalidData
}

// ImportYAML imports a Dataset from YAML data.
func ImportYAML(data []byte) (*Dataset, error) {
	// Try to decode as array of objects
	var objects []map[string]any
	if err := yaml.Unmarshal(data, &objects); err == nil && len(objects) > 0 {
		return importYAMLObjects(objects)
	}

	// Try to decode as array of arrays
	var arrays [][]any
	if err := yaml.Unmarshal(data, &arrays); err == nil {
		return importYAMLArrays(arrays)
	}

	return nil, ErrInvalidData
}

func importYAMLObjects(objects []map[string]any) (*Dataset, error) {
	if len(objects) == 0 {
		return NewDataset(nil), nil
	}

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

func importYAMLArrays(arrays [][]any) (*Dataset, error) {
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

func exportDatabookYAML(db *Databook, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()

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

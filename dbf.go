package tablib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

func init() {
	RegisterExporter(FormatDBF, ExporterFunc(exportDBF))
	RegisterImporter(FormatDBF, ImporterFunc(importDBF))
}

// DBF file structure constants
const (
	dbfVersion        = 0x03 // dBASE III
	dbfHeaderTerminator = 0x0D
	dbfRecordDeleted    = 0x2A // '*'
	dbfRecordActive     = 0x20 // ' '
	dbfEOF              = 0x1A
)

// DBF field types
const (
	dbfFieldTypeChar    = 'C' // Character
	dbfFieldTypeNumber  = 'N' // Numeric
	dbfFieldTypeLogical = 'L' // Logical
	dbfFieldTypeDate    = 'D' // Date
	dbfFieldTypeFloat   = 'F' // Float
)

// dbfHeader represents the DBF file header
type dbfHeader struct {
	Version       byte
	Year          byte
	Month         byte
	Day           byte
	RecordCount   uint32
	HeaderSize    uint16
	RecordSize    uint16
	Reserved      [20]byte
}

// dbfFieldDescriptor represents a field descriptor in DBF
type dbfFieldDescriptor struct {
	Name          [11]byte
	Type          byte
	Reserved1     [4]byte
	Length        byte
	DecimalCount  byte
	Reserved2     [14]byte
}

func exportDBF(ds *Dataset, w io.Writer) error {
	if len(ds.headers) == 0 {
		return ErrHeadersRequired
	}

	// Calculate field descriptors
	fields := make([]dbfFieldDescriptor, len(ds.headers))
	fieldLengths := make([]int, len(ds.headers))

	// Determine field lengths by scanning all data
	for i, header := range ds.headers {
		// Start with header length
		maxLen := len(header)
		if maxLen > 254 {
			maxLen = 254
		}
		fieldLengths[i] = maxLen

		// Check all values
		for _, row := range ds.data {
			if i < len(row) {
				valLen := len(fmt.Sprintf("%v", row[i]))
				if valLen > fieldLengths[i] {
					fieldLengths[i] = valLen
				}
			}
		}

		// Ensure minimum length and maximum
		if fieldLengths[i] < 1 {
			fieldLengths[i] = 1
		}
		if fieldLengths[i] > 254 {
			fieldLengths[i] = 254
		}

		// Create field descriptor
		var fd dbfFieldDescriptor
		name := header
		if len(name) > 10 {
			name = name[:10]
		}
		copy(fd.Name[:], strings.ToUpper(name))
		fd.Type = dbfFieldTypeChar // All fields as character for simplicity
		fd.Length = byte(fieldLengths[i])
		fd.DecimalCount = 0
		fields[i] = fd
	}

	// Calculate record size (1 byte for deletion flag + sum of field lengths)
	recordSize := 1
	for _, l := range fieldLengths {
		recordSize += l
	}

	// Calculate header size (32 bytes header + 32 bytes per field + 1 byte terminator)
	headerSize := 32 + (32 * len(fields)) + 1

	// Create header
	now := time.Now()
	header := dbfHeader{
		Version:     dbfVersion,
		Year:        byte(now.Year() - 1900),
		Month:       byte(now.Month()),
		Day:         byte(now.Day()),
		RecordCount: uint32(len(ds.data)),
		HeaderSize:  uint16(headerSize),
		RecordSize:  uint16(recordSize),
	}

	var buf bytes.Buffer

	// Write header
	if err := binary.Write(&buf, binary.LittleEndian, &header); err != nil {
		return err
	}

	// Write field descriptors
	for _, fd := range fields {
		if err := binary.Write(&buf, binary.LittleEndian, &fd); err != nil {
			return err
		}
	}

	// Write header terminator
	buf.WriteByte(dbfHeaderTerminator)

	// Write records
	for _, row := range ds.data {
		// Write deletion flag (space = active)
		buf.WriteByte(dbfRecordActive)

		// Write field values
		for i, l := range fieldLengths {
			var val string
			if i < len(row) {
				val = fmt.Sprintf("%v", row[i])
			}
			// Pad or truncate to field length
			if len(val) > l {
				val = val[:l]
			}
			val = fmt.Sprintf("%-*s", l, val)
			buf.WriteString(val)
		}
	}

	// Write EOF marker
	buf.WriteByte(dbfEOF)

	_, err := w.Write(buf.Bytes())
	return err
}

func importDBF(r io.Reader) (*Dataset, error) {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) < 32 {
		return nil, ErrInvalidData
	}

	// Parse header
	var header dbfHeader
	headerReader := bytes.NewReader(data[:32])
	if err := binary.Read(headerReader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	// Calculate number of fields
	numFields := (int(header.HeaderSize) - 32 - 1) / 32
	if numFields < 0 || numFields > 1000 {
		return nil, ErrInvalidData
	}

	// Parse field descriptors
	fields := make([]dbfFieldDescriptor, numFields)
	for i := 0; i < numFields; i++ {
		offset := 32 + (i * 32)
		if offset+32 > len(data) {
			return nil, ErrInvalidData
		}
		fieldReader := bytes.NewReader(data[offset : offset+32])
		if err := binary.Read(fieldReader, binary.LittleEndian, &fields[i]); err != nil {
			return nil, err
		}
	}

	// Extract headers
	headers := make([]string, numFields)
	for i, f := range fields {
		// Find null terminator in name
		name := string(f.Name[:])
		if idx := strings.IndexByte(name, 0); idx >= 0 {
			name = name[:idx]
		}
		headers[i] = strings.TrimSpace(name)
	}

	ds := NewDataset(headers)

	// Parse records
	recordStart := int(header.HeaderSize)
	recordSize := int(header.RecordSize)

	for i := 0; i < int(header.RecordCount); i++ {
		offset := recordStart + (i * recordSize)
		if offset+recordSize > len(data) {
			break
		}

		recordData := data[offset : offset+recordSize]

		// Check deletion flag
		if recordData[0] == dbfRecordDeleted {
			continue
		}

		// Parse fields
		row := make([]any, numFields)
		fieldOffset := 1 // Skip deletion flag
		for j, f := range fields {
			fieldLen := int(f.Length)
			if fieldOffset+fieldLen > len(recordData) {
				break
			}
			value := string(recordData[fieldOffset : fieldOffset+fieldLen])
			row[j] = strings.TrimSpace(value)
			fieldOffset += fieldLen
		}

		if err := ds.Append(row); err != nil {
			return nil, err
		}
	}

	return ds, nil
}

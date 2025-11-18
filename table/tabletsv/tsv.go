package tabletsv

import (
	"encoding/csv"
	"fmt"
	"io"
)

type Table struct {
	Header []string
	Rows   [][]string
}

func NewTable(header []string, rows [][]string) (*Table, error) {
	for _, row := range rows {
		if len(row) != len(header) {
			return nil, fmt.Errorf("tabletsv.NewTable: row length mismatch")
		}
	}
	return &Table{Header: header, Rows: rows}, nil
}

func ParseTable(r io.Reader) (*Table, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("tabletsv.ParseTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("tabletsv.ParseTable: %w", err)
	}
	return &Table{Header: header, Rows: rows}, nil
}

func WriteTable(w io.Writer, table *Table) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header); err != nil {
		return fmt.Errorf("tabletsv.WriteTable: %w", err)
	}

	for _, row := range table.Rows {
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("tabletsv.WriteTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("tabletsv.WriteTable: %w", err)
	}
	return nil
}

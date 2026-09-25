package tabletsv

import (
	"encoding/csv"
	"fmt"
	"io"
)

const IDColumnHeader = "ID"

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

func RequireColumns(header []string, minColumns int) error {
	if len(header) < minColumns {
		return fmt.Errorf("tabletsv.RequireColumns: 表のヘッダは %d 列以上必要ですが %d 列でした: %v", minColumns, len(header), header)
	}
	return nil
}

func RequireUniqueFirstColumn(rows [][]string) error {
	seen := make(map[string]struct{}, len(rows))
	for i, row := range rows {
		if len(row) == 0 {
			return fmt.Errorf("tabletsv.RequireUniqueFirstColumn: %d 件目のレコードに ID 列がありません", i+1)
		}
		if _, dup := seen[row[0]]; dup {
			return fmt.Errorf("tabletsv.RequireUniqueFirstColumn: ID %q の行が重複しています（%d 件目のレコード）。1 つの ID に 1 行だけにしてください", row[0], i+1)
		}
		seen[row[0]] = struct{}{}
	}
	return nil
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

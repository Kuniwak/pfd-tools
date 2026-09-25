package pfdtsv

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/table/tabletsv"
)

func WriteAtomicProcessTable(w io.Writer, table *pfd.AtomicProcessTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("pfdtsv.WriteAtomicProcessTable: %w", err)
	}
	for _, row := range table.Rows {
		row := append([]string{string(row.ID), row.Description}, row.ExtraCells...)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("pfdtsv.WriteAtomicProcessTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("pfdtsv.WriteAtomicProcessTable: %w", err)
	}
	return nil
}

func ParseAtomicProcessTable(r io.Reader) (*pfd.AtomicProcessTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicProcessTable: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 2); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicProcessTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicProcessTable: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicProcessTable: %w", err)
	}
	rows2 := make([]*pfd.AtomicProcessRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &pfd.AtomicProcessRow{ID: pfd.AtomicProcessID(row[0]), Description: row[1], ExtraCells: row[2:]})
	}
	return &pfd.AtomicProcessTable{ExtraHeaders: header[2:], Rows: rows2}, nil
}

func WriteAtomicDeliverableTable(w io.Writer, table *pfd.AtomicDeliverableTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("pfdtsv.WriteAtomicDeliverableTable: %w", err)
	}
	for _, row := range table.Rows {
		row := append([]string{string(row.ID), row.Description}, row.ExtraCells...)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("pfdtsv.WriteAtomicDeliverableTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("pfdtsv.WriteAtomicDeliverableTable: %w", err)
	}
	return nil
}

func ParseAtomicDeliverableTable(r io.Reader) (*pfd.AtomicDeliverableTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicDeliverableTable: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 2); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicDeliverableTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicDeliverableTable: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseAtomicDeliverableTable: %w", err)
	}
	rows2 := make([]*pfd.AtomicDeliverableRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &pfd.AtomicDeliverableRow{ID: pfd.AtomicDeliverableID(row[0]), Description: row[1], ExtraCells: row[2:]})
	}
	return &pfd.AtomicDeliverableTable{ExtraHeaders: header[2:], Rows: rows2}, nil
}

func WriteCompositeProcessTable(w io.Writer, table *pfd.CompositeProcessTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("pfdtsv.WriteCompositeProcessTable: %w", err)
	}
	for _, row := range table.Rows {
		row := append([]string{string(row.ID), row.Description}, row.ExtraCells...)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("pfdtsv.WriteCompositeProcessTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("pfdtsv.WriteCompositeProcessTable: %w", err)
	}
	return nil
}

func ParseCompositeProcessTable(r io.Reader) (*pfd.CompositeProcessTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeProcessTable: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 2); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeProcessTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeProcessTable: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeProcessTable: %w", err)
	}
	rows2 := make([]*pfd.CompositeProcessRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &pfd.CompositeProcessRow{ID: pfd.CompositeProcessID(row[0]), Description: row[1], ExtraCells: row[2:]})
	}
	return &pfd.CompositeProcessTable{ExtraHeaders: header[2:], Rows: rows2}, nil
}

func WriteCompositeDeliverableTable(w io.Writer, table *pfd.CompositeDeliverableTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("pfdtsv.WriteCompositeDeliverableTable: %w", err)
	}
	for _, row := range table.Rows {

		row := append([]string{string(row.ID), row.Description, row.DeliverablesCell()}, row.ExtraCells...)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("pfdtsv.WriteCompositeDeliverableTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("pfdtsv.WriteCompositeDeliverableTable: %w", err)
	}
	return nil
}

func ParseCompositeDeliverableTable(r io.Reader) (*pfd.CompositeDeliverableTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeDeliverableTable: reading header: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 3); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeDeliverableTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeDeliverableTable: reading body: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseCompositeDeliverableTable: %w", err)
	}
	rows2 := make([]*pfd.CompositeDeliverableRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &pfd.CompositeDeliverableRow{ID: pfd.CompositeDeliverableID(row[0]), Description: row[1], Deliverables: pfd.ParseDeliverablesCell(row[2]), ExtraCells: row[3:]})
	}

	return &pfd.CompositeDeliverableTable{ExtraHeaders: header[3:], Rows: rows2}, nil
}

func ParseCompositeDeliverableTableOrEmpty(r io.Reader) (*pfd.CompositeDeliverableTable, error) {
	if r == nil {
		return &pfd.CompositeDeliverableTable{}, nil
	}
	return ParseCompositeDeliverableTable(r)
}

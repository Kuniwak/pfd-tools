package fsmtsv

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/table/tabletsv"
)

func WriteResourceTable(w io.Writer, table *fsmtable.ResourceTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("fsmtsv.WriteResourceTable: %w", err)
	}
	for _, row := range table.Rows {
		row := append([]string{string(row.ID), row.Description}, row.ExtraCells...)
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("fsmtsv.WriteResourceTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("fsmtsv.WriteResourceTable: %w", err)
	}
	return nil
}

func ParseResourceTable(r io.Reader) (*fsmtable.ResourceTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseResourceTable: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 2); err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseResourceTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseResourceTable: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseResourceTable: %w", err)
	}
	rows2 := make([]*fsmtable.ResourceTableRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &fsmtable.ResourceTableRow{ID: fsm.ResourceID(row[0]), Description: row[1], ExtraCells: row[2:]})
	}
	return &fsmtable.ResourceTable{ExtraHeaders: header[2:], Rows: rows2}, nil
}

func WriteMilestoneTable(w io.Writer, table *fsmtable.MilestoneTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("fsmtsv.WriteMilestoneTable: %w", err)
	}
	for _, row := range table.Rows {
		row := row.Row()
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("fsmtsv.WriteMilestoneTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("fsmtsv.WriteMilestoneTable: %w", err)
	}
	return nil
}

func ParseMilestoneTable(r io.Reader) (*fsmtable.MilestoneTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseMilestoneTable: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 4); err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseMilestoneTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseMilestoneTable: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseMilestoneTable: %w", err)
	}
	rows2 := make([]*fsmtable.MilestoneTableRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &fsmtable.MilestoneTableRow{MilestoneID: fsmmasterschedule.Milestone(row[0]), Description: row[1], GroupIDs: row[2], Successors: row[3], ExtraCells: row[4:]})
	}
	return &fsmtable.MilestoneTable{ExtraHeaders: header[4:], Rows: rows2}, nil
}

func ParseGroupTable(r io.Reader) (*fsmtable.GroupTable, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseGroupTable: %w", err)
	}
	if err := tabletsv.RequireColumns(header, 2); err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseGroupTable: %w", err)
	}
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseGroupTable: %w", err)
	}
	if err := tabletsv.RequireUniqueFirstColumn(rows); err != nil {
		return nil, fmt.Errorf("fsmtsv.ParseGroupTable: %w", err)
	}
	rows2 := make([]*fsmtable.GroupTableRow, 0, len(rows))
	for _, row := range rows {
		rows2 = append(rows2, &fsmtable.GroupTableRow{ID: fsmmasterschedule.Group(row[0]), Description: row[1], ExtraCells: row[2:]})
	}
	return &fsmtable.GroupTable{ExtraHeaders: header[2:], Rows: rows2}, nil
}

func WriteGroupTable(w io.Writer, table *fsmtable.GroupTable) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(table.Header()); err != nil {
		return fmt.Errorf("fsmtsv.WriteGroupTable: %w", err)
	}
	for _, row := range table.Rows {
		if err := csvWriter.Write(row.Row()); err != nil {
			return fmt.Errorf("fsmtsv.WriteGroupTable: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("fsmtsv.WriteGroupTable: %w", err)
	}
	return nil
}

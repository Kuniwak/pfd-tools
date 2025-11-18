package fsmhtml

import (
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/table/tablehtml"
)

func WriteResourceTable(w io.Writer, table *fsmtable.ResourceTable) error {
	t, err := ResourceTable(table)
	if err != nil {
		return fmt.Errorf("fsmhtml.WriteResourceTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func ResourceTable(table *fsmtable.ResourceTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, append([]string{string(row.ID), row.Description}, row.ExtraCells...))
	}

	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("fsmhtml.ResourceTable: %w", err)
	}
	return tbl, nil
}

func WriteMilestoneTable(w io.Writer, table *fsmtable.MilestoneTable) error {
	t, err := MilestoneTable(table)
	if err != nil {
		return fmt.Errorf("fsmhtml.WriteMilestoneTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func MilestoneTable(table *fsmtable.MilestoneTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, row.Row())
	}
	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("fsmhtml.MilestoneTable: %w", err)
	}
	return tbl, nil
}

func WriteGroupTable(w io.Writer, table *fsmtable.GroupTable) error {
	t, err := GroupTable(table)
	if err != nil {
		return fmt.Errorf("fsmhtml.WriteGroupTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func GroupTable(table *fsmtable.GroupTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, row.Row())
	}
	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("fsmhtml.GroupTable: %w", err)
	}
	return tbl, nil
}

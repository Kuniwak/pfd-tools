package pfdhtml

import (
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/table/tablehtml"
)

func WriteAtomicProcessTable(w io.Writer, table *pfd.AtomicProcessTable) error {
	t, err := AtomicProcessTable(table)
	if err != nil {
		return fmt.Errorf("pfdhtml.WriteAtomicProcessTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func AtomicProcessTable(table *pfd.AtomicProcessTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, append([]string{string(row.ID), row.Description}, row.ExtraCells...))
	}

	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("pfdhtml.AtomicProcessTable: %w", err)
	}
	return tbl, nil
}

func WriteAtomicDeliverableTable(w io.Writer, table *pfd.AtomicDeliverableTable) error {
	t, err := AtomicDeliverableTable(table)
	if err != nil {
		return fmt.Errorf("pfdhtml.WriteAtomicDeliverableTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func AtomicDeliverableTable(table *pfd.AtomicDeliverableTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, append([]string{string(row.ID), row.Description}, row.ExtraCells...))
	}
	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("pfdhtml.AtomicDeliverableTable: %w", err)
	}
	return tbl, nil
}

func WriteCompositeProcessTable(w io.Writer, table *pfd.CompositeProcessTable) error {
	t, err := CompositeProcessTable(table)
	if err != nil {
		return fmt.Errorf("pfdhtml.WriteCompositeProcessTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func CompositeProcessTable(table *pfd.CompositeProcessTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, append([]string{string(row.ID), row.Description}, row.ExtraCells...))
	}

	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("pfdhtml.CompositeProcessTable: %w", err)
	}
	return tbl, nil
}

func WriteCompositeDeliverableTable(w io.Writer, table *pfd.CompositeDeliverableTable) error {
	t, err := CompositeDeliverableTable(table)
	if err != nil {
		return fmt.Errorf("pfdhtml.WriteCompositeDeliverableTable: %w", err)
	}
	return tablehtml.RenderTable(w, t)
}

func CompositeDeliverableTable(table *pfd.CompositeDeliverableTable) (*tablehtml.Table, error) {
	rows := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		rows = append(rows, append([]string{string(row.ID), row.Description}, row.ExtraCells...))
	}
	tbl, err := tablehtml.NewTable(table.Header(), rows)
	if err != nil {
		return nil, fmt.Errorf("pfdhtml.CompositeDeliverableTable: %w", err)
	}
	return tbl, nil
}

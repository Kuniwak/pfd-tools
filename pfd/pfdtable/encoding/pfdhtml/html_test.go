package pfdhtml

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/table/tablehtml"
	"github.com/google/go-cmp/cmp"
)

func TestAtomicProcessTable(t *testing.T) {
	table := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{"Resources", "Min Est. Volume", "Max Est. Volume"},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "Process 1", ExtraCells: []string{"A", "1", "2"}},
			{ID: "P2", Description: "Process 2", ExtraCells: []string{"B", "3", "4"}},
		},
	}

	got, err := AtomicProcessTable(table)
	if err != nil {
		t.Fatal(err)
	}
	want := &tablehtml.Table{
		Header: []string{"ID", "Description", "Resources", "Min Est. Volume", "Max Est. Volume"},
		Rows: [][]string{
			{"P1", "Process 1", "A", "1", "2"},
			{"P2", "Process 2", "B", "3", "4"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestDeliverableTable(t *testing.T) {
	table := &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{"Location", "Available Date"},
		Rows: []*pfd.AtomicDeliverableRow{
			{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{"https://example.com/1", "2021-01-01"}},
			{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{"https://example.com/2", "2022-01-01"}},
		},
	}

	got, err := AtomicDeliverableTable(table)
	if err != nil {
		t.Fatal(err)
	}
	want := &tablehtml.Table{
		Header: []string{"ID", "Description", "Location", "Available Date"},
		Rows: [][]string{
			{"D1", "Deliverable 1", "https://example.com/1", "2021-01-01"},
			{"D2", "Deliverable 2", "https://example.com/2", "2022-01-01"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestCompositeProcessTable(t *testing.T) {
	table := &pfd.CompositeProcessTable{
		ExtraHeaders: []string{"Example"},
		Rows: []*pfd.CompositeProcessRow{
			{ID: "C1", Description: "Composite Process 1", ExtraCells: []string{"A"}},
			{ID: "C2", Description: "Composite Process 2", ExtraCells: []string{"B"}},
		},
	}

	got, err := CompositeProcessTable(table)
	if err != nil {
		t.Fatal(err)
	}
	want := &tablehtml.Table{
		Header: []string{"ID", "Description", "Example"},
		Rows: [][]string{
			{"C1", "Composite Process 1", "A"},
			{"C2", "Composite Process 2", "B"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Error(cmp.Diff(want, got))
	}
}

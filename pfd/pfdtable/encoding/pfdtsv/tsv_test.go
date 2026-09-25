package pfdtsv

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/google/go-cmp/cmp"
)

func TestCompositeDeliverableTableRoundTrip(t *testing.T) {
	testCases := map[string]struct {
		Table *pfd.CompositeDeliverableTable
	}{
		"header only": {
			Table: &pfd.CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: []*pfd.CompositeDeliverableRow{}},
		},

		"a composite deliverable with its breakdown": {
			Table: &pfd.CompositeDeliverableTable{
				ExtraHeaders: []string{},
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D0", Description: "設計書", Deliverables: []pfd.NodeID{"D1", "D2"}, ExtraCells: []string{}},
				},
			},
		},
		"extra columns are kept": {
			Table: &pfd.CompositeDeliverableTable{
				ExtraHeaders: []string{"備考"},
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D0", Description: "設計書", Deliverables: []pfd.NodeID{"D1"}, ExtraCells: []string{"あとで分割する"}},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			if err := WriteCompositeDeliverableTable(sb, tc.Table); err != nil {
				t.Fatalf("WriteCompositeDeliverableTable: %v", err)
			}
			actual, err := ParseCompositeDeliverableTable(strings.NewReader(sb.String()))
			if err != nil {
				t.Fatalf("ParseCompositeDeliverableTable: %v (written: %q)", err, sb.String())
			}
			if !reflect.DeepEqual(actual, tc.Table) {
				t.Error(cmp.Diff(tc.Table, actual))
			}
		})
	}
}

func TestParseTablesInvalid(t *testing.T) {
	testCases := map[string]struct {
		Parse func(io.Reader) error
		TSV   string
	}{

		"atomic process table with only the ID column": {
			Parse: func(r io.Reader) error { _, err := ParseAtomicProcessTable(r); return err },
			TSV:   "ID\nP1\n",
		},
		"atomic deliverable table with only the ID column": {
			Parse: func(r io.Reader) error { _, err := ParseAtomicDeliverableTable(r); return err },
			TSV:   "ID\nD1\n",
		},
		"composite process table with only the ID column": {
			Parse: func(r io.Reader) error { _, err := ParseCompositeProcessTable(r); return err },
			TSV:   "ID\nP1\n",
		},

		"composite deliverable table without the deliverables column": {
			Parse: func(r io.Reader) error { _, err := ParseCompositeDeliverableTable(r); return err },
			TSV:   "ID\tDescription\nD0\tcomposite\n",
		},

		"header only with too few columns": {
			Parse: func(r io.Reader) error { _, err := ParseAtomicProcessTable(r); return err },
			TSV:   "ID\n",
		},

		"row with more columns than the header": {
			Parse: func(r io.Reader) error { _, err := ParseAtomicProcessTable(r); return err },
			TSV:   "ID\tDescription\nP1\t実装\t余り\n",
		},

		"atomic process table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseAtomicProcessTable(r); return err },
			TSV:   "ID\tDescription\nP1\t実装\nP1\t別の実装\n",
		},
		"atomic deliverable table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseAtomicDeliverableTable(r); return err },
			TSV:   "ID\tDescription\nD1\t仕様\nD1\t別の仕様\n",
		},
		"composite process table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseCompositeProcessTable(r); return err },
			TSV:   "ID\tDescription\nP1\t開発\nP1\t別の開発\n",
		},
		"composite deliverable table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseCompositeDeliverableTable(r); return err },
			TSV:   "ID\tDescription\tDeliverables\nD0\t設計書\tD1\nD0\t設計書\tD2\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if err := tc.Parse(strings.NewReader(tc.TSV)); err == nil {
				t.Fatal("expected an error for the malformed table, got nil")
			}
		})
	}
}

func TestParseCompositeDeliverableTableOrEmpty(t *testing.T) {
	testCases := map[string]struct {
		Reader   io.Reader
		WantRows int
	}{
		"nil reader yields empty table": {
			Reader:   nil,
			WantRows: 0,
		},
		"header-only reader yields zero rows": {
			Reader:   strings.NewReader("ID\tDescription\tDeliverables\n"),
			WantRows: 0,
		},
		"reader with a row": {
			Reader:   strings.NewReader("ID\tDescription\tDeliverables\nD0\tcomposite\tD1,D2\n"),
			WantRows: 1,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseCompositeDeliverableTableOrEmpty(tc.Reader)
			if err != nil {
				t.Fatalf("ParseCompositeDeliverableTableOrEmpty: %v", err)
			}
			if got == nil {
				t.Fatal("expected a non-nil table")
			}
			if len(got.Rows) != tc.WantRows {
				t.Errorf("len(Rows) = %d, want %d", len(got.Rows), tc.WantRows)
			}
		})
	}
}

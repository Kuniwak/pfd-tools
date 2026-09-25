package fsmtsv

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/google/go-cmp/cmp"
)

func TestParseTablesInvalid(t *testing.T) {
	testCases := map[string]struct {
		Parse func(io.Reader) error
		TSV   string
	}{

		"resource table with only the ID column": {
			Parse: func(r io.Reader) error { _, err := ParseResourceTable(r); return err },
			TSV:   "ID\nR1\n",
		},
		"group table with only the ID column": {
			Parse: func(r io.Reader) error { _, err := ParseGroupTable(r); return err },
			TSV:   "ID\nG1\n",
		},

		"milestone table without the successors column": {
			Parse: func(r io.Reader) error { _, err := ParseMilestoneTable(r); return err },
			TSV:   "ID\tDescription\tGroups\nM1\tリリース\tG1\n",
		},

		"header only with too few columns": {
			Parse: func(r io.Reader) error { _, err := ParseResourceTable(r); return err },
			TSV:   "ID\n",
		},

		"resource table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseResourceTable(r); return err },
			TSV:   "ID\tDescription\nR1\t開発者\nR1\t別の開発者\n",
		},
		"group table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseGroupTable(r); return err },
			TSV:   "ID\tDescription\nG1\t基盤\nG1\t別の基盤\n",
		},
		"milestone table with duplicate ID rows": {
			Parse: func(r io.Reader) error { _, err := ParseMilestoneTable(r); return err },
			TSV:   "ID\tDescription\tGroups\tSuccessors\nM1\tリリース\tG1\t\nM1\t別のリリース\tG1\t\n",
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

func TestRoundTrip(t *testing.T) {
	t.Run("resource table", func(t *testing.T) {
		table := &fsmtable.ResourceTable{
			ExtraHeaders: []string{"備考"},
			Rows: []*fsmtable.ResourceTableRow{
				{ID: "R1", Description: "開発者", ExtraCells: []string{"2 人"}},
			},
		}
		sb := &strings.Builder{}
		if err := WriteResourceTable(sb, table); err != nil {
			t.Fatalf("WriteResourceTable: %v", err)
		}
		actual, err := ParseResourceTable(strings.NewReader(sb.String()))
		if err != nil {
			t.Fatalf("ParseResourceTable: %v (written: %q)", err, sb.String())
		}
		if !reflect.DeepEqual(actual, table) {
			t.Error(cmp.Diff(table, actual))
		}
	})

	t.Run("group table", func(t *testing.T) {
		table := &fsmtable.GroupTable{
			ExtraHeaders: []string{},
			Rows: []*fsmtable.GroupTableRow{
				{ID: "G1", Description: "基盤", ExtraCells: []string{}},
			},
		}
		sb := &strings.Builder{}
		if err := WriteGroupTable(sb, table); err != nil {
			t.Fatalf("WriteGroupTable: %v", err)
		}
		actual, err := ParseGroupTable(strings.NewReader(sb.String()))
		if err != nil {
			t.Fatalf("ParseGroupTable: %v (written: %q)", err, sb.String())
		}
		if !reflect.DeepEqual(actual, table) {
			t.Error(cmp.Diff(table, actual))
		}
	})

	t.Run("milestone table", func(t *testing.T) {
		table := &fsmtable.MilestoneTable{
			ExtraHeaders: []string{},
			Rows: []*fsmtable.MilestoneTableRow{
				{MilestoneID: "M1", Description: "リリース", GroupIDs: "G1", Successors: "M2", ExtraCells: []string{}},
			},
		}
		sb := &strings.Builder{}
		if err := WriteMilestoneTable(sb, table); err != nil {
			t.Fatalf("WriteMilestoneTable: %v", err)
		}
		actual, err := ParseMilestoneTable(strings.NewReader(sb.String()))
		if err != nil {
			t.Fatalf("ParseMilestoneTable: %v (written: %q)", err, sb.String())
		}
		if !reflect.DeepEqual(actual, table) {
			t.Error(cmp.Diff(table, actual))
		}
	})
}

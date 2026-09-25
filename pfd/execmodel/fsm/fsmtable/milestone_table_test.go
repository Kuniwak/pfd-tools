package fsmtable

import (
	"slices"
	"testing"

	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
)

func apTableWithMilestoneGroup(headers []string, rows [][]string) *pfd.AtomicProcessTable {
	apRows := make([]*pfd.AtomicProcessRow, 0, len(rows))
	for i, r := range rows {
		id := pfd.AtomicProcessID("P" + string(rune('1'+i)))
		apRows = append(apRows, &pfd.AtomicProcessRow{ID: id, Description: "", ExtraCells: r})
	}
	return &pfd.AtomicProcessTable{ExtraHeaders: headers, Rows: apRows}
}

func TestNewMilestoneTableByAtomicProcessTable(t *testing.T) {
	testCases := map[string]struct {
		AP   *pfd.AtomicProcessTable
		Want []*MilestoneTableRow
	}{
		"no milestone column": {
			AP:   apTableWithMilestoneGroup([]string{}, [][]string{}),
			Want: []*MilestoneTableRow{},
		},
		"milestone with group": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.BarColumnHeader, mastertsv.RowColumnHeader},
				[][]string{{"M1", "G1"}},
			),
			Want: []*MilestoneTableRow{
				{MilestoneID: "M1", GroupIDs: "G1", Description: "", Successors: "", ExtraCells: []string{}},
			},
		},
		"multiple processes same milestone different groups": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.BarColumnHeader, mastertsv.RowColumnHeader},
				[][]string{{"M1", "G1"}, {"M1", "G2"}},
			),
			Want: []*MilestoneTableRow{
				{MilestoneID: "M1", GroupIDs: "G1,G2", Description: "", Successors: "", ExtraCells: []string{}},
			},
		},
		"empty milestone cell is skipped": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.BarColumnHeader, mastertsv.RowColumnHeader},
				[][]string{{"", "G1"}},
			),
			Want: []*MilestoneTableRow{},
		},
		"milestone without group column": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.BarColumnHeader},
				[][]string{{"M1"}},
			),
			Want: []*MilestoneTableRow{
				{MilestoneID: "M1", GroupIDs: "", Description: "", Successors: "", ExtraCells: []string{}},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewMilestoneTableByAtomicProcessTable(tc.AP)
			if len(got.Rows) != len(tc.Want) {
				t.Fatalf("len(Rows) = %d, want %d", len(got.Rows), len(tc.Want))
			}
			slices.SortFunc(got.Rows, (*MilestoneTableRow).Compare)
			wantSorted := slices.Clone(tc.Want)
			slices.SortFunc(wantSorted, (*MilestoneTableRow).Compare)
			for i, row := range got.Rows {
				if row.MilestoneID != wantSorted[i].MilestoneID {
					t.Errorf("row[%d].MilestoneID = %q, want %q", i, row.MilestoneID, wantSorted[i].MilestoneID)
				}
			}
		})
	}
}

func TestRawMilestoneMap(t *testing.T) {
	testCases := map[string]struct {
		AP      *pfd.AtomicProcessTable
		Want    map[pfd.AtomicProcessID]string
		WantErr bool
	}{
		"missing milestone column": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.RowColumnHeader},
				[][]string{{"G1"}},
			),
			WantErr: true,
		},
		"with milestone column": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.BarColumnHeader, mastertsv.RowColumnHeader},
				[][]string{{"M1", "G1"}, {"M2", "G1"}},
			),
			Want: map[pfd.AtomicProcessID]string{"P1": "M1", "P2": "M2"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := RawMilestoneMap(tc.AP, DefaultBarColumnMatchFunc)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if len(got) != len(tc.Want) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tc.Want))
			}
			for ap, milestone := range tc.Want {
				if got[ap] != milestone {
					t.Errorf("got[%q] = %q, want %q", ap, got[ap], milestone)
				}
			}
		})
	}
}

func TestNewGroupTableByAtomicProcessTable(t *testing.T) {
	testCases := map[string]struct {
		AP   *pfd.AtomicProcessTable
		Want []fsmmasterschedule.Group
	}{
		"no group column": {
			AP:   apTableWithMilestoneGroup([]string{}, [][]string{}),
			Want: []fsmmasterschedule.Group{},
		},
		"single group": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.RowColumnHeader},
				[][]string{{"G1"}},
			),
			Want: []fsmmasterschedule.Group{"G1"},
		},
		"comma-separated groups in one cell": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.RowColumnHeader},
				[][]string{{"G1,G2"}},
			),
			Want: []fsmmasterschedule.Group{"G1", "G2"},
		},
		"empty group cell is skipped": {
			AP: apTableWithMilestoneGroup(
				[]string{mastertsv.RowColumnHeader},
				[][]string{{""}},
			),
			Want: []fsmmasterschedule.Group{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := NewGroupTableByAtomicProcessTable(tc.AP)
			if len(got.Rows) != len(tc.Want) {
				t.Fatalf("len(Rows) = %d, want %d", len(got.Rows), len(tc.Want))
			}
		})
	}
}

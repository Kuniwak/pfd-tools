package fsmtable

import (
	"slices"
	"testing"

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
				[]string{MilestoneColumnHeaderJa, GroupColumnHeaderJa},
				[][]string{{"M1", "G1"}},
			),
			Want: []*MilestoneTableRow{
				{MilestoneID: "M1", GroupIDs: "G1", Description: "", Successors: "", ExtraCells: []string{}},
			},
		},
		"multiple processes same milestone different groups": {
			AP: apTableWithMilestoneGroup(
				[]string{MilestoneColumnHeaderJa, GroupColumnHeaderJa},
				[][]string{{"M1", "G1"}, {"M1", "G2"}},
			),
			Want: []*MilestoneTableRow{
				{MilestoneID: "M1", GroupIDs: "G1,G2", Description: "", Successors: "", ExtraCells: []string{}},
			},
		},
		"empty milestone cell is skipped": {
			AP: apTableWithMilestoneGroup(
				[]string{MilestoneColumnHeaderJa, GroupColumnHeaderJa},
				[][]string{{"", "G1"}},
			),
			Want: []*MilestoneTableRow{},
		},
		"milestone without group column": {
			AP: apTableWithMilestoneGroup(
				[]string{MilestoneColumnHeaderJa},
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
				[]string{GroupColumnHeaderJa},
				[][]string{{"G1"}},
			),
			Want: []fsmmasterschedule.Group{"G1"},
		},
		"comma-separated groups in one cell": {
			AP: apTableWithMilestoneGroup(
				[]string{GroupColumnHeaderJa},
				[][]string{{"G1,G2"}},
			),
			Want: []fsmmasterschedule.Group{"G1", "G2"},
		},
		"empty group cell is skipped": {
			AP: apTableWithMilestoneGroup(
				[]string{GroupColumnHeaderJa},
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

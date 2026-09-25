package pfd

import (
	"reflect"
	"slices"
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
)

func TestCompositeDeliverableTableNodeIDMap(t *testing.T) {
	testCases := map[string]struct {
		Table    *CompositeDeliverableTable
		Expected map[NodeID][]NodeID
	}{

		"duplicate composite deliverable ID keeps the first": {
			Table: &CompositeDeliverableTable{
				ExtraHeaders: []string{},
				Rows: []*CompositeDeliverableRow{
					{ID: "D1", Description: "仕様", Deliverables: []NodeID{"D2"}},
					{ID: "D1", Description: "仕様", Deliverables: []NodeID{"D3"}},
				},
			},
			Expected: map[NodeID][]NodeID{"D1": {"D2"}},
		},

		"the same atomic deliverable listed twice is deduplicated": {
			Table: &CompositeDeliverableTable{
				ExtraHeaders: []string{},
				Rows: []*CompositeDeliverableRow{
					{ID: "D1", Description: "仕様", Deliverables: []NodeID{"D2", "D2"}},
				},
			},
			Expected: map[NodeID][]NodeID{"D1": {"D2"}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := tc.Table.NodeIDMap(testLogger())
			if len(actual) != len(tc.Expected) {
				t.Fatalf("expected %d entries, but got %d", len(tc.Expected), len(actual))
			}
			for id, expected := range tc.Expected {
				got, ok := actual[id]
				if !ok {
					t.Fatalf("expected an entry for %q, but got none", id)
				}
				if !slices.Equal(got.Slice(), expected) {
					t.Errorf("expected %v for %q, but got %v", expected, id, got.Slice())
				}
			}
		})
	}
}

func TestAtomicProcessTableIDs(t *testing.T) {
	testCases := map[string]struct {
		Table    *AtomicProcessTable
		Expected []AtomicProcessID
	}{
		"empty table": {
			Table:    &AtomicProcessTable{ExtraHeaders: []string{}, Rows: []*AtomicProcessRow{}},
			Expected: []AtomicProcessID{},
		},
		"keeps the row order": {
			Table: &AtomicProcessTable{
				ExtraHeaders: []string{},
				Rows: []*AtomicProcessRow{
					{ID: "P2", Description: "Process 2"},
					{ID: "P1", Description: "Process 1"},
				},
			},
			Expected: []AtomicProcessID{"P2", "P1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if actual := tc.Table.IDs(); !slices.Equal(actual, tc.Expected) {
				t.Errorf("expected %v, but got %v", tc.Expected, actual)
			}
		})
	}
}

func TestCompositeDeliverableTableRefresh(t *testing.T) {
	testCases := map[string]struct {
		Nodes    []*Node
		Rows     []*CompositeDeliverableRow
		Expected []string
	}{
		"keeps a row used only as a member": {
			Nodes: []*Node{{ID: "D6", Description: "D6", Type: NodeTypeCompositeDeliverable}},
			Rows: []*CompositeDeliverableRow{
				{ID: "D3", Description: "D3", Deliverables: []NodeID{"D1", "D2"}, ExtraCells: []string{}},
				{ID: "D6", Description: "D6", Deliverables: []NodeID{"D3", "D5"}, ExtraCells: []string{}},
			},
			Expected: []string{"D3", "D6"},
		},
		"removes chained rows unreachable from any drawn node": {
			Nodes: []*Node{{ID: "D6", Description: "D6", Type: NodeTypeCompositeDeliverable}},
			Rows: []*CompositeDeliverableRow{
				{ID: "D6", Description: "D6", Deliverables: []NodeID{"D1"}, ExtraCells: []string{}},
				{ID: "D8", Description: "D8", Deliverables: []NodeID{"D9"}, ExtraCells: []string{}},
				{ID: "D9", Description: "D9", Deliverables: []NodeID{"D2"}, ExtraCells: []string{}},
			},
			Expected: []string{"D6"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			p := &PFD{
				Nodes: sets.New((*Node).Compare, tc.Nodes...),
				Edges: sets.New((*Edge).Compare),
			}
			table := &CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: tc.Rows}

			table.Refresh(p, NewNodeMap(p.Nodes, testLogger()))

			got := make([]string, 0, len(table.Rows))
			for _, row := range table.Rows {
				got = append(got, string(row.ID))
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("rows = %v, want %v", got, tc.Expected)
			}
		})
	}
}

func TestCompositeDeliverableTableNestedIDs(t *testing.T) {
	testCases := map[string]struct {
		Rows     []*CompositeDeliverableRow
		Nodes    []*Node
		From     []CompositeDeliverableID
		Expected []CompositeDeliverableID
	}{
		"empty from": {
			Rows: []*CompositeDeliverableRow{
				{ID: "D6", Deliverables: []NodeID{"D3"}},
			},
			From:     []CompositeDeliverableID{},
			Expected: []CompositeDeliverableID{},
		},

		"transitive": {
			Rows: []*CompositeDeliverableRow{
				{ID: "D6", Deliverables: []NodeID{"D3", "D5"}},
				{ID: "D3", Deliverables: []NodeID{"D1", "D2"}},
			},
			From:     []CompositeDeliverableID{"D6"},
			Expected: []CompositeDeliverableID{"D3"},
		},

		"member drawn as an atomic deliverable is not nested": {
			Rows: []*CompositeDeliverableRow{
				{ID: "D6", Deliverables: []NodeID{"D3", "D5"}},
				{ID: "D3", Deliverables: []NodeID{"D1", "D2"}},
				{ID: "D5", Deliverables: []NodeID{"D1"}},
			},
			Nodes: []*Node{
				{ID: "D5", Type: NodeTypeAtomicDeliverable},
				{ID: "D6", Type: NodeTypeCompositeDeliverable},
			},
			From:     []CompositeDeliverableID{"D6"},
			Expected: []CompositeDeliverableID{"D3"},
		},
		"unreachable row is not included": {
			Rows: []*CompositeDeliverableRow{
				{ID: "D6", Deliverables: []NodeID{"D1"}},
				{ID: "D8", Deliverables: []NodeID{"D9"}},
				{ID: "D9", Deliverables: []NodeID{"D2"}},
			},
			From:     []CompositeDeliverableID{"D6"},
			Expected: []CompositeDeliverableID{},
		},
		"cyclic nesting terminates": {
			Rows: []*CompositeDeliverableRow{
				{ID: "D3", Deliverables: []NodeID{"D6"}},
				{ID: "D6", Deliverables: []NodeID{"D3"}},
			},
			From:     []CompositeDeliverableID{"D3"},
			Expected: []CompositeDeliverableID{"D3", "D6"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			table := &CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: tc.Rows}
			got := table.NestedIDs(sets.New(CompositeDeliverableID.Compare, tc.From...), sets.New((*Node).Compare, tc.Nodes...))
			if !reflect.DeepEqual(got.Slice(), tc.Expected) {
				t.Errorf("nested = %v, want %v", got.Slice(), tc.Expected)
			}
		})
	}
}

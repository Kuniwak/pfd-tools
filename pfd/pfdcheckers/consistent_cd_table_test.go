package pfdcheckers

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestConsistentCDTable(t *testing.T) {
	testCases := map[string]struct {
		PFD      *pfd.PFD
		CDTable  *pfd.CompositeDeliverableTable
		Expected []checkers.Problem
	}{
		"missing: composite deliverable node without a cd table row": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D0", Type: pfd.NodeTypeCompositeDeliverable, Description: "D0"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{},
			Expected: []checkers.Problem{
				checkers.NewProblem("missing-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D0"))...),
			},
		},
		"extra: cd table row without a composite deliverable node": {
			PFD: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{{ID: "D0", Description: "D0"}},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("extra-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D0"))...),
			},
		},

		"member only: cd table row referenced by another row": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D6", Type: pfd.NodeTypeCompositeDeliverable, Description: "D6"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D3", Description: "D3", Deliverables: []pfd.NodeID{"D1", "D2"}},
					{ID: "D6", Description: "D6", Deliverables: []pfd.NodeID{"D3", "D5"}},
				},
			},
			Expected: []checkers.Problem{},
		},

		"extra: row for an ID drawn as an atomic deliverable": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D9", Type: pfd.NodeTypeAtomicDeliverable, Description: "D9"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D8", Description: "D8", Deliverables: []pfd.NodeID{"D9"}},
					{ID: "D9", Description: "D9", Deliverables: []pfd.NodeID{"D1"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("extra-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D8"))...),
				checkers.NewProblem("extra-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D9"))...),
			},
		},

		"extra: row for a member drawn as an atomic deliverable": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D5", Type: pfd.NodeTypeAtomicDeliverable, Description: "D5"},
					&pfd.Node{ID: "D6", Type: pfd.NodeTypeCompositeDeliverable, Description: "D6"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D3", Description: "D3", Deliverables: []pfd.NodeID{"D1", "D2"}},
					{ID: "D5", Description: "D5", Deliverables: []pfd.NodeID{"D1"}},
					{ID: "D6", Description: "D6", Deliverables: []pfd.NodeID{"D3", "D5"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("extra-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D5"))...),
			},
		},

		"extra: chained rows unreachable from any drawn node": {
			PFD: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D8", Description: "D8", Deliverables: []pfd.NodeID{"D9"}},
					{ID: "D9", Description: "D9", Deliverables: []pfd.NodeID{"D1"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("extra-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D8"))...),
				checkers.NewProblem("extra-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D9"))...),
			},
		},
		"consistent: node and row match": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D0", Type: pfd.NodeTypeCompositeDeliverable, Description: "D0"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{{ID: "D0", Description: "D0"}},
			},
			Expected: []checkers.Problem{},
		},
		"nil table with composite deliverable node": {
			PFD: &pfd.PFD{
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{ID: "D0", Type: pfd.NodeTypeCompositeDeliverable, Description: "D0"},
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable: nil,
			Expected: []checkers.Problem{
				checkers.NewProblem("missing-cd-table", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D0"))...),
			},
		},
		"nil table without composite deliverable node": {
			PFD: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			CDTable:  nil,
			Expected: []checkers.Problem{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := pfdcommon.NewMemoized(tc.PFD, slog.New(slogtest.NewTestHandler(t)))
			ch := make(chan checkers.Problem)
			go func() {
				defer close(ch)
				tgt := pfdcommon.NewTarget(tc.PFD, nil, nil, nil, tc.CDTable, m)
				if err := ConsistentCDTable.Check(tgt, ch); err != nil {
					t.Errorf("ConsistentCDTable.Check: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("got %v, expected %v", got, tc.Expected)
			}
		})
	}
}

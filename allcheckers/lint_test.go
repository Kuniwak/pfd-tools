package allcheckers

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestCompositeDeliverableTableProblems(t *testing.T) {
	testCases := map[string]struct {
		Table    *pfd.CompositeDeliverableTable
		Expected []checkers.Problem
	}{

		"nil table": {
			Table:    nil,
			Expected: []checkers.Problem{},
		},
		"empty table": {
			Table:    &pfd.CompositeDeliverableTable{},
			Expected: []checkers.Problem{},
		},
		"nested composition without a cycle": {
			Table: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D3", Description: "D3", Deliverables: []pfd.NodeID{"D1", "D2"}},
					{ID: "D6", Description: "D6", Deliverables: []pfd.NodeID{"D3", "D5"}},
				},
			},
			Expected: []checkers.Problem{},
		},
		"cyclic nesting": {
			Table: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D3", Description: "D3", Deliverables: []pfd.NodeID{"D6"}},
					{ID: "D6", Description: "D6", Deliverables: []pfd.NodeID{"D3"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("acyclic-cd-comp", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D3", "D6"))...),
			},
		},
		"self cycle": {
			Table: &pfd.CompositeDeliverableTable{
				Rows: []*pfd.CompositeDeliverableRow{
					{ID: "D3", Description: "D3", Deliverables: []pfd.NodeID{"D3"}},
				},
			},
			Expected: []checkers.Problem{
				checkers.NewProblem("acyclic-cd-comp", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, "D3"))...),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := CompositeDeliverableTableProblems(tc.Table, slog.New(slogtest.NewTestHandler(t)))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Errorf("got %v, expected %v", got, tc.Expected)
			}
		})
	}
}

package pfd

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestRenumberPlan(t *testing.T) {
	testCases := map[string]struct {
		PFD           *PFD
		ExpectedOneOf []RenumberPlan
	}{
		"empty": {
			PFD: &PFD{
				Nodes: sets.New((*Node).Compare),
				Edges: sets.New((*Edge).Compare),
			},
			ExpectedOneOf: []RenumberPlan{{}},
		},
		"simplest": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Process", Type: NodeTypeAtomicProcess},
					&Node{ID: "Output deliverable", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "Input deliverable", Target: "Process"},
					&Edge{Source: "Process", Target: "Output deliverable"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"Output deliverable": {ID: "D1", Description: "Output deliverable", Type: NodeTypeAtomicDeliverable},
					"Input deliverable": {ID: "D2", Description: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					"Process":  {ID: "P1", Description: "Process", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"like Y shape": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "Input deliverable 1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Input deliverable 2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Process", Type: NodeTypeAtomicProcess},
					&Node{ID: "Output deliverable", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "Input deliverable 1", Target: "Process"},
					&Edge{Source: "Input deliverable 2", Target: "Process"},
					&Edge{Source: "Process", Target: "Output deliverable"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"Output deliverable":  {ID: "D1", Description: "Output deliverable", Type: NodeTypeAtomicDeliverable},
					"Input deliverable 1": {ID: "D2", Description: "Input deliverable 1", Type: NodeTypeAtomicDeliverable},
					"Input deliverable 2": {ID: "D3", Description: "Input deliverable 2", Type: NodeTypeAtomicDeliverable},
					"Process":   {ID: "P1", Description: "Process", Type: NodeTypeAtomicProcess},
				},
				{
					"Output deliverable":  {ID: "D1", Description: "Output deliverable", Type: NodeTypeAtomicDeliverable},
					"Input deliverable 2": {ID: "D2", Description: "Input deliverable 2", Type: NodeTypeAtomicDeliverable},
					"Input deliverable 1": {ID: "D3", Description: "Input deliverable 1", Type: NodeTypeAtomicDeliverable},
					"Process":   {ID: "P1", Description: "Process", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"like reversed Y shape": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Process", Type: NodeTypeAtomicProcess},
					&Node{ID: "Output deliverable 1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Output deliverable 2", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "Input deliverable", Target: "Process"},
					&Edge{Source: "Process", Target: "Output deliverable 1"},
					&Edge{Source: "Process", Target: "Output deliverable 2"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"Output deliverable 1": {ID: "D1", Description: "Output deliverable 1", Type: NodeTypeAtomicDeliverable},
					"Output deliverable 2": {ID: "D2", Description: "Output deliverable 2", Type: NodeTypeAtomicDeliverable},
					"Input deliverable":  {ID: "D3", Description: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					"Process":   {ID: "P1", Description: "Process", Type: NodeTypeAtomicProcess},
				},
				{
					"Output deliverable 2": {ID: "D1", Description: "Output deliverable 2", Type: NodeTypeAtomicDeliverable},
					"Output deliverable 1": {ID: "D2", Description: "Output deliverable 1", Type: NodeTypeAtomicDeliverable},
					"Input deliverable":  {ID: "D3", Description: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					"Process":   {ID: "P1", Description: "Process", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"sequential": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Process 1", Type: NodeTypeAtomicProcess},
					&Node{ID: "Intermediate deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Process 2", Type: NodeTypeAtomicProcess},
					&Node{ID: "Output deliverable", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "Input deliverable", Target: "Process 1"},
					&Edge{Source: "Process 1", Target: "Intermediate deliverable"},
					&Edge{Source: "Intermediate deliverable", Target: "Process 2"},
					&Edge{Source: "Process 2", Target: "Output deliverable"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"Output deliverable": {ID: "D1", Description: "Output deliverable", Type: NodeTypeAtomicDeliverable},
					"Input deliverable": {ID: "D2", Description: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					"Intermediate deliverable": {ID: "D3", Description: "Intermediate deliverable", Type: NodeTypeAtomicDeliverable},
					"Process 1": {ID: "P1", Description: "Process 1", Type: NodeTypeAtomicProcess},
					"Process 2": {ID: "P2", Description: "Process 2", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"composite process": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Intermediate deliverable 1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Intermediate deliverable 2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Intermediate deliverable 3", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Output deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Atomic process 1", Type: NodeTypeAtomicProcess},
					&Node{ID: "Atomic process 2", Type: NodeTypeAtomicProcess},
					&Node{ID: "Atomic process 3", Type: NodeTypeAtomicProcess},
					&Node{ID: "Atomic process 4", Type: NodeTypeAtomicProcess},
					&Node{ID: "Composite process 1", Type: NodeTypeCompositeProcess},
					&Node{ID: "Composite process 2", Type: NodeTypeCompositeProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "Input deliverable", Target: "Atomic process 1"},
					&Edge{Source: "Atomic process 1", Target: "Intermediate deliverable 1"},
					&Edge{Source: "Intermediate deliverable 1", Target: "Atomic process 2"},
					&Edge{Source: "Atomic process 2", Target: "Intermediate deliverable 2"},
					&Edge{Source: "Intermediate deliverable 2", Target: "Atomic process 3"},
					&Edge{Source: "Atomic process 3", Target: "Intermediate deliverable 3"},
					&Edge{Source: "Intermediate deliverable 3", Target: "Atomic process 4"},
					&Edge{Source: "Atomic process 4", Target: "Output deliverable"},

					&Edge{Source: "Input deliverable", Target: "Composite process 1"},
					&Edge{Source: "Composite process 1", Target: "Intermediate deliverable 2"},
					&Edge{Source: "Intermediate deliverable 2", Target: "Composite process 2"},
					&Edge{Source: "Composite process 2", Target: "Output deliverable"},
				),
				ProcessComposition: map[NodeID]*sets.Set[NodeID]{
					"Composite process 1": sets.New(NodeID.Compare, "Atomic process 1", "Atomic process 2"),
					"Composite process 2": sets.New(NodeID.Compare, "Atomic process 3", "Atomic process 4"),
				},
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"Output deliverable":   {ID: "D1", Description: "Output deliverable", Type: NodeTypeAtomicDeliverable},
					"Input deliverable":   {ID: "D2", Description: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					"Intermediate deliverable 1":  {ID: "D3", Description: "Intermediate deliverable 1", Type: NodeTypeAtomicDeliverable},
					"Intermediate deliverable 2":  {ID: "D4", Description: "Intermediate deliverable 2", Type: NodeTypeAtomicDeliverable},
					"Intermediate deliverable 3":  {ID: "D5", Description: "Intermediate deliverable 3", Type: NodeTypeAtomicDeliverable},
					"Atomic process 1": {ID: "P1", Description: "Atomic process 1", Type: NodeTypeAtomicProcess},
					"Atomic process 2": {ID: "P2", Description: "Atomic process 2", Type: NodeTypeAtomicProcess},
					"Atomic process 3": {ID: "P3", Description: "Atomic process 3", Type: NodeTypeAtomicProcess},
					"Atomic process 4": {ID: "P4", Description: "Atomic process 4", Type: NodeTypeAtomicProcess},
					"Composite process 1": {ID: "P5", Description: "Composite process 1", Type: NodeTypeCompositeProcess},
					"Composite process 2": {ID: "P6", Description: "Composite process 2", Type: NodeTypeCompositeProcess},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodeMap := NewNodeMap(testCase.PFD.Nodes, logger)
			graphExceptFB := testCase.PFD.GraphExceptFeedback(nodeMap, logger)
			errs := make([]Error, 0)

			plan, ok := NewRenumberPlan(testCase.PFD, graphExceptFB, nodeMap, &errs)
			if !ok {
				t.Fatalf("NewRenumberPlan: expected true, got false: %s", Errors(errs).Error())
			}
			for _, expected := range testCase.ExpectedOneOf {
				if reflect.DeepEqual(plan, expected) {
					ok = true
					break
				}
				t.Log(cmp.Diff(expected, plan))
			}
			if !ok {
				for _, expected := range testCase.ExpectedOneOf {
					t.Log(cmp.Diff(expected, plan))
				}
				t.Error("expected one of the plans")
			}
		})
	}
}

func TestRenumberPlanError(t *testing.T) {
	testCases := map[string]struct {
		PFD *PFD
	}{
		"loop": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "Input deliverable", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "Process", Type: NodeTypeAtomicProcess},
					&Node{ID: "Output deliverable", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "Input deliverable", Target: "Process"},
					&Edge{Source: "Process", Target: "Output deliverable"},
					&Edge{Source: "Output deliverable", Target: "Process"},
				),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodeMap := NewNodeMap(testCase.PFD.Nodes, logger)
			graphExceptFB := testCase.PFD.GraphExceptFeedback(nodeMap, logger)

			errs := make([]Error, 0)

			_, ok := NewRenumberPlan(testCase.PFD, graphExceptFB, nodeMap, &errs)
			if ok {
				t.Errorf("NewRenumberPlan: expected false, got true: %s", Errors(errs).Error())
			}
		})
	}
}

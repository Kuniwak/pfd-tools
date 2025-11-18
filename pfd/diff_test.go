package pfd

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestDiff(t *testing.T) {
	testCases := map[string]struct {
		PFD1     *PFD
		PFD2     *PFD
		Expected *DiffResult
	}{
		"empty": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare),
				Edges: sets.New((*Edge).Compare),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare),
				Edges: sets.New((*Edge).Compare),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"missing node": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
				Edges: sets.New((*Edge).Compare),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare),
				Edges: sets.New((*Edge).Compare),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"extra node": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare),
				Edges: sets.New((*Edge).Compare),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
				Edges: sets.New((*Edge).Compare),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"changed node": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A1"}),
				Edges: sets.New((*Edge).Compare),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A2"}),
				Edges: sets.New((*Edge).Compare),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare, &NodeDiffChange{
						ID:  "A",
						Old: sets.New((*NodeData).Compare, &NodeData{Description: "A1"}),
						New: sets.New((*NodeData).Compare, &NodeData{Description: "A2"}),
					}),
					SameNodes: sets.New((*Node).Compare),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"same nodes": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
				Edges: sets.New((*Edge).Compare),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
				Edges: sets.New((*Edge).Compare),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare, &Node{ID: "A", Description: "A"}),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"missing edge": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"extra edge": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"changed edge": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: false}),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare, &EdgeDiffChange{Source: "A", Target: "B", OldFeedbackFlags: sets.New(cmp2.CompareBool, true), NewFeedbackFlags: sets.New(cmp2.CompareBool, false)}),
					SameEdges:    sets.New((*Edge).Compare),
				},
			},
		},
		"same edges": {
			PFD1: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
			},
			PFD2: &PFD{
				Nodes: sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				Edges: sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
			},
			Expected: &DiffResult{
				NodeDiff: &NodeDiffResult{
					ExtraNodes:   sets.New((*Node).Compare),
					MissingNodes: sets.New((*Node).Compare),
					ChangedNodes: sets.New((*NodeDiffChange).Compare),
					SameNodes:    sets.New((*Node).Compare, &Node{ID: "A"}, &Node{ID: "B"}),
				},
				EdgeDiff: &EdgeDiffResult{
					ExtraEdges:   sets.New((*Edge).Compare),
					MissingEdges: sets.New((*Edge).Compare),
					ChangedEdges: sets.New((*EdgeDiffChange).Compare),
					SameEdges:    sets.New((*Edge).Compare, &Edge{Source: "A", Target: "B", IsFeedback: true}),
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := Diff(testCase.PFD1, testCase.PFD2)
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

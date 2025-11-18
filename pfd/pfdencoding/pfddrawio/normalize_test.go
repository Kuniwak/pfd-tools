package pfddrawio

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestNormalize(t *testing.T) {
	testCases := map[string]struct {
		File              []Diagram
		ExpectedPFD       *pfd.PFD
		ExpectedSourceMap *SourceMap
	}{
		"example": {
			File: exampleFile,
			ExpectedPFD: &pfd.PFD{
				Title: "Example",
				Nodes: sets.New(
					(*pfd.Node).Compare,
					&pfd.Node{
						ID:          "D1",
						Description: "Implementation",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D1",
						Description: "Implementation",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D2",
						Description: "Review\ncomments",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D3",
						Description: "Verification result",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D4",
						Description: "Specification",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "D4",
						Description: "Specification",
						Type:        pfd.NodeTypeAtomicDeliverable,
					},
					&pfd.Node{
						ID:          "P1",
						Description: "Implement",
						Type:        pfd.NodeTypeCompositeProcess,
					},
					&pfd.Node{
						ID:          "P2",
						Description: "Review",
						Type:        pfd.NodeTypeAtomicProcess,
					},
					&pfd.Node{
						ID:          "P3",
						Description: "Verify",
						Type:        pfd.NodeTypeAtomicProcess,
					},
					&pfd.Node{
						ID:          "P4",
						Description: "Implement",
						Type:        pfd.NodeTypeAtomicProcess,
					},
				),
				Edges: sets.New(
					(*pfd.Edge).Compare,
					&pfd.Edge{
						Source: "D1",
						Target: "P2",
					},
					&pfd.Edge{
						Source: "D1",
						Target: "P3",
					},
					&pfd.Edge{
						Source:     "D2",
						Target:     "P1",
						IsFeedback: true,
					},
					&pfd.Edge{
						Source:     "D3",
						Target:     "P1",
						IsFeedback: true,
					},
					&pfd.Edge{
						Source: "D4",
						Target: "P1",
					},
					&pfd.Edge{
						Source: "D4",
						Target: "P3",
					},
					&pfd.Edge{
						Source: "D4",
						Target: "P4",
					},
					&pfd.Edge{
						Source: "P1",
						Target: "D1",
					},
					&pfd.Edge{
						Source: "P2",
						Target: "D2",
					},
					&pfd.Edge{
						Source: "P3",
						Target: "D3",
					},
					&pfd.Edge{
						Source: "P4",
						Target: "D1",
					},
				),
				ProcessComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{
					"P1": sets.New(pfd.NodeID.Compare, "P4"),
				},
				DeliverableComposition: map[pfd.NodeID]*sets.Set[pfd.NodeID]{},
			},
			ExpectedSourceMap: exampleSourceMap,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			actual, srcMap, err := NormalizeDiagrams("Example", testCase.File, logger)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, testCase.ExpectedPFD) {
				t.Error(cmp.Diff(testCase.ExpectedPFD, actual))
			}
			if !reflect.DeepEqual(srcMap, testCase.ExpectedSourceMap) {
				t.Error(cmp.Diff(testCase.ExpectedSourceMap, srcMap))
			}
		})
	}
}

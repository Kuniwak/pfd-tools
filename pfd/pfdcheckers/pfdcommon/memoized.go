package pfdcommon

import (
	"log/slog"

	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

type Memoized struct {
	NodeMap           map[pfd.NodeID]*pfd.Node
	EdgeMap           map[pfd.NodeID]*sets.Set[pfd.NodeID]
	FBEdgeMap         map[pfd.NodeID]*sets.Set[pfd.NodeID]
	ReversedEdgeMap   map[pfd.NodeID]*sets.Set[pfd.NodeID]
	ReversedFBEdgeMap map[pfd.NodeID]*sets.Set[pfd.NodeID]
	GraphExceptFB     *graph.Graph
	GraphIncludingFB  *graph.Graph
}

func NewMemoized(p *pfd.PFD, logger *slog.Logger) Memoized {
	nodeMap := pfd.NewNodeMap(p.Nodes, logger)

	edgeMap, fbMap := pfd.NewEdgeMap(p.Edges)
	reversedEdgeMap, reversedFBEdgeMap := pfd.NewReversedEdgeMap(p.Edges)

	graphExceptFB := p.GraphExceptFeedback(nodeMap, logger)
	graphIncludingFB := p.GraphIncludingFeedback(nodeMap, logger)

	return Memoized{
		NodeMap:           nodeMap,
		EdgeMap:           edgeMap,
		FBEdgeMap:         fbMap,
		ReversedEdgeMap:   reversedEdgeMap,
		ReversedFBEdgeMap: reversedFBEdgeMap,
		GraphExceptFB:     graphExceptFB,
		GraphIncludingFB:  graphIncludingFB,
	}
}

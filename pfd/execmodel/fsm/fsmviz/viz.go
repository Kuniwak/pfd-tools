package fsmviz

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/dot"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

func Dot(isCompleted func(fsm.State) bool, g fsm.StateTransitionGraph, plans *sets.Set[*fsm.Plan], logger *slog.Logger) (*dot.Digraph, error) {
	res := &dot.Digraph{
		Name: "PFD FSM State Transition Diagram",
		GraphAttributes: sets.New(
			dot.CompareAttribute,
			&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
				First:  dot.AttributeNameCharset,
				Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "UTF-8"},
			},
		),
		NodeAttributes: sets.New(
			dot.CompareAttribute,
			&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
				First:  dot.AttributeNameShape,
				Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "box"},
			},
			&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
				First:  dot.AttributeNameStyle,
				Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "filled"},
			},
		),
		EdgeAttributes: sets.New(dot.CompareAttribute),
		Nodes:          make([]*dot.Node, 0, len(g.Nodes)),
		Edges:          make([]*dot.Edge, 0),
	}

	logger.Debug("Dot: starting...")

	sb := &strings.Builder{}

	completedState := sets.New(fsm.StateID.Compare)
	minTime := execmodel.Time(math.MaxFloat64)
	for nodeID, state := range g.Nodes {
		if !isCompleted(state) {
			continue
		}
		completedState.Add(fsm.StateID.Compare, nodeID)
		minTime = min(minTime, state.Time)
	}

	logger.Debug("Dot: completed state completed")

	bestPlanPaths := sets.New(fsm.State.Compare, g.Nodes[g.InitialState])
	for _, plan := range plans.Iter() {
		for _, tr := range plan.Transitions {
			bestPlanPaths.Add(fsm.State.Compare, tr.NextState)
		}
	}

	logger.Debug("Dot: best plan paths completed")

	for nodeID, state := range g.Nodes {
		node, err := DotNode(nodeID, state, completedState, bestPlanPaths, g.Edges, sb)
		if err != nil {
			return nil, err
		}
		res.Nodes = append(res.Nodes, node)
	}

	logger.Debug("Dot: nodes completed")

	for _, p1 := range sortedKeyValue(g.Edges, fsm.StateID.Compare) {
		for _, p2 := range sortedKeyValue(p1.Second, fsm.StateID.Compare) {
			for _, allocation := range p2.Second.Iter() {
				edge, err := DotEdge(fsm.StateTransitionGraphEdge{
					Source:     p1.First,
					Target:     p2.First,
					Allocation: allocation,
				}, bestPlanPaths, sb, g)
				if err != nil {
					return nil, err
				}
				res.Edges = append(res.Edges, edge)
			}
		}
	}

	logger.Debug("Dot: edges completed")

	return res, nil
}

func InitialState(g fsm.StateTransitionGraph) (fsm.StateID, fsm.State) {
	for nodeID, state := range g.Nodes {
		if state.Time == 0 {
			return nodeID, state
		}
	}
	panic("fsmviz.InitialState: initial state not found")
}

func DotNodeID(nodeID fsm.StateID) dot.NodeID {
	return dot.NodeID(fmt.Sprintf("S%d", nodeID))
}

func DotNode(stateID fsm.StateID, state fsm.State, completedState *sets.Set[fsm.StateID], bestPlanPaths *sets.Set[fsm.State], edgeMap map[fsm.StateID]map[fsm.StateID]*sets.Set[fsm.Allocation], sb *strings.Builder) (*dot.Node, error) {
	sb.Reset()
	if err := state.Write(sb); err != nil {
		return nil, err
	}
	stateLabel := sb.String()

	attrs := sets.New(
		dot.CompareAttribute,
		&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameLabel,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: stateLabel},
		},
	)

	isInBestPlan := bestPlanPaths.Contains(fsm.State.Compare, state)
	isCompleted := completedState.Contains(fsm.StateID.Compare, stateID)
	var isUnexpectedDeadlock bool
	if m, ok := edgeMap[stateID]; !ok || len(m) == 0 {
		isUnexpectedDeadlock = !isCompleted
	}

	fillColor := "white"
	fontColor := "black"
	color := "black"

	if isUnexpectedDeadlock {
		fillColor = "lightpink"
		fontColor = "red"
		color = "red"
	}

	if isCompleted {
		fillColor = "aliceblue"
		fontColor = "blue"
		color = "blue"
	}

	if isInBestPlan {
		fillColor = "lightblue"
	}

	attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
		First:  dot.AttributeNameFillColor,
		Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: fillColor},
	})
	attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
		First:  dot.AttributeNameFontColor,
		Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: fontColor},
	})
	attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
		First:  dot.AttributeNameColor,
		Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: color},
	})

	return &dot.Node{
		ID:         DotNodeID(stateID),
		Attributes: attrs,
	}, nil
}

func DotEdge(edge fsm.StateTransitionGraphEdge, bestPlanPaths *sets.Set[fsm.State], sb *strings.Builder, g fsm.StateTransitionGraph) (*dot.Edge, error) {
	sb.Reset()
	if err := edge.Allocation.Write(sb); err != nil {
		return nil, err
	}
	edgeLabel := sb.String()

	attrs := sets.New(
		dot.CompareAttribute,
		&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameLabel,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: edgeLabel},
		},
	)
	source, ok := g.Nodes[edge.Source]
	if !ok {
		return nil, fmt.Errorf("fsmviz.DotEdge: source not found")
	}
	target, ok := g.Nodes[edge.Target]
	if !ok {
		return nil, fmt.Errorf("fsmviz.DotEdge: target not found")
	}
	if bestPlanPaths.Contains(fsm.State.Compare, source) && bestPlanPaths.Contains(fsm.State.Compare, target) {
		attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameColor,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "blue"},
		})
		attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameFontColor,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "blue"},
		})
	}

	return &dot.Edge{
		Source:     DotNodeID(edge.Source),
		Target:     DotNodeID(edge.Target),
		Attributes: attrs,
	}, nil
}

func sortedKeyValue[K comparable, V any](m map[K]V, cmpKey func(K, K) int) []*pairs.Pair[K, V] {
	xs := make([]*pairs.Pair[K, V], 0, len(m))
	for k := range m {
		xs = append(xs, pairs.New(k, m[k]))
	}
	slices.SortFunc(xs, pairs.CompareFirst[K, V](cmpKey))
	return xs
}

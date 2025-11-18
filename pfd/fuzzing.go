package pfd

import (
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"pgregory.net/rapid"
)

// ValidPFD generates a valid PFD with approximately n elements (may vary slightly).
func AnyValidPFD(t *rapid.T, n int) *PFD {
	p := AnyValidPFDWithoutFeedback(t, n)

	edgeMap, _ := NewEdgeMap(p.Edges)
	paths := CollectAtomicProcess2DeliverablePaths(edgeMap, OnlyAtomicProcesses(p.Nodes))

	for _, path := range paths.Iter() {
		// NOTE: Add a small number of feedback edges.
		if rapid.IntRange(0, 4).Draw(t, fmt.Sprintf("path %v", path)) == 0 {
			p.Edges.Add((*Edge).Compare, &Edge{Source: path[len(path)-1], Target: path[0], IsFeedback: true})
		}
	}

	return p
}

func CollectAtomicProcess2DeliverablePaths(edgeMap map[NodeID]*sets.Set[NodeID], aps *sets.Set[NodeID]) *sets.Set[[]NodeID] {
	var dfs func(ap NodeID, path []NodeID, res *sets.Set[[]NodeID])
	dfs = func(ap NodeID, path []NodeID, res *sets.Set[[]NodeID]) {
		path = append(path, ap)

		dsts, ok := edgeMap[ap]
		if !ok {
			panic(fmt.Sprintf("AnyValidPFD: missing dsts in edgeMap: %q", ap))
		}

		for _, d := range dsts.Iter() {
			newPath := append(path, d)
			res.Add(cmp2.CompareSlice[[]NodeID](NodeID.Compare), newPath)

			m, ok := edgeMap[d]
			if !ok {
				continue
			}
			for _, ap2 := range m.Iter() {
				dfs(ap2, newPath, res)
			}
		}
	}

	ap2DPaths := sets.New(cmp2.CompareSlice[[]NodeID](NodeID.Compare))
	for _, ap := range aps.Iter() {
		dfs(ap, make([]NodeID, 0), ap2DPaths)
	}
	return ap2DPaths
}

// AnyValidPFDWithoutFeedback generates a valid PFD without feedback edges with approximately n elements (may vary slightly).
func AnyValidPFDWithoutFeedback(t *rapid.T, n int) *PFD {
	nodeMap := make(map[NodeID]*Node)
	nodes := sets.New((*Node).Compare)
	edges := sets.New((*Edge).Compare)

	// NOTE: Create up to 5 initial deliverables.
	initDelivsCount := rapid.IntRange(1, 5).Draw(t, "initDelivsCount")
	for i := 0; i < initDelivsCount; i++ {
		node := &Node{
			ID:          NewDeliverableID(nodes.Len()),
			Description: fmt.Sprintf("Initial deliverable %d", nodes.Len()),
			Type:        NodeTypeAtomicDeliverable,
		}
		nodes.Add((*Node).Compare, node)
		nodeMap[node.ID] = node
	}

	// NOTE: Create a DAG with alternating deliverables and atomic processes.
	//       Deliverables without output may remain at the end of the loop.
	for i := 0; i < n; i++ {
		isAP := rapid.Bool().Draw(t, fmt.Sprintf("apOrDeliv %d", i))
		var node *Node
		if isAP {
			ap := &Node{
				ID:          NewAtomicProcessID(nodes.Len()),
				Description: fmt.Sprintf("Atomic Process %d", nodes.Len()),
				Type:        NodeTypeAtomicProcess,
			}
			node = ap

			ds := OnlyAtomicDeliverables(nodes)

			// NOTE: Atomic processes must have at least one input deliverable.
			d := AnyDeliverableID(ds, t)
			edges.Add((*Edge).Compare, &Edge{Source: d, Target: ap.ID})

			// NOTE: Atomic processes may have two or more input deliverables.
			for j, d := range ds.Iter() {
				if rapid.Bool().Draw(t, fmt.Sprintf("edge %d %d", i, j)) {
					dNode, ok := nodeMap[d]
					if !ok {
						panic(fmt.Sprintf("AnyValidPFD: missing ap in nodeMap: %q", d))
					}
					edges.Add((*Edge).Compare, &Edge{
						Source: dNode.ID,
						Target: ap.ID,
					})
				}
			}
		} else {
			d := &Node{
				ID:          NewDeliverableID(nodes.Len()),
				Description: fmt.Sprintf("Deliverable %d", nodes.Len()),
				Type:        NodeTypeAtomicDeliverable,
			}
			node = d

			aps := OnlyAtomicProcesses(nodes)
			if aps.Len() == 0 {
				// NOTE: Deliverables may have no input.
				continue
			}

			if rapid.Bool().Draw(t, fmt.Sprintf("edge %d", i)) {
				// NOTE: If a deliverable has input, it has exactly one.
				ap := AnyAtomicProcessID(aps, t)
				edges.Add((*Edge).Compare, &Edge{Source: ap, Target: d.ID})
			}
		}

		nodeMap[node.ID] = node
		nodes.Add((*Node).Compare, node)
	}

	aps := OnlyAtomicProcesses(nodes)
	apsNotHaveOutput := aps.Clone()
	for _, ap := range aps.Iter() {
		for _, edge := range edges.Iter() {
			if edge.Source == ap {
				apsNotHaveOutput.Remove(NodeID.Compare, ap)
			}
		}
	}

	for _, ap := range apsNotHaveOutput.Iter() {
		// NOTE: Attach new deliverables to atomic processes that have no output.
		dNode := &Node{
			ID:          NewDeliverableID(nodes.Len()),
			Description: fmt.Sprintf("Deliverable %d", nodes.Len()),
			Type:        NodeTypeAtomicDeliverable,
		}
		nodeMap[dNode.ID] = dNode
		nodes.Add((*Node).Compare, dNode)
		edges.Add((*Edge).Compare, &Edge{Source: ap, Target: dNode.ID})
	}

	p := &PFD{
		Nodes: nodes,
		Edges: edges,
	}

	// NOTE: Connect weakly connected components to make the graph weakly connected.
	logger := slog.New(slogtest.NewRapidHandler(t))
	g := p.GraphExceptFeedback(nodeMap, logger)
	components := g.WeaklyConnectedComponents()
	baseComponent, ok := components.At(0)
	if !ok {
		panic(fmt.Sprintf("AnyValidPFDWithFeedback: missing component: %v", components))
	}
	pBaseComponent := sets.NewWithCapacity[*Node](baseComponent.Len())
	for _, n := range baseComponent.Iter() {
		node, ok := nodeMap[NodeID(n)]
		if !ok {
			panic(fmt.Sprintf("AnyValidPFDWithFeedback: missing node: %v", n))
		}
		pBaseComponent.Add((*Node).Compare, node)
	}
	pBaseAPs := OnlyAtomicProcesses(pBaseComponent)
	if pBaseAPs.Len() == 0 {
		// NOTE: Only the pattern with exactly one initial deliverable is possible.
		id, ok := baseComponent.At(0)
		if !ok {
			panic(fmt.Sprintf("AnyValidPFDWithFeedback: missing component: %v", components))
		}

		newP := &Node{
			ID:          NewAtomicProcessID(p.Nodes.Len()),
			Description: fmt.Sprintf("Atomic Process %d", p.Nodes.Len()),
			Type:        NodeTypeAtomicProcess,
		}
		p.Nodes.Add((*Node).Compare, newP)
		newD := &Node{
			ID:          NewDeliverableID(p.Nodes.Len()),
			Description: fmt.Sprintf("Deliverable %d", p.Nodes.Len()),
			Type:        NodeTypeAtomicDeliverable,
		}
		p.Nodes.Add((*Node).Compare, newD)
		p.Edges.Add((*Edge).Compare, &Edge{Source: NodeID(id), Target: newP.ID})
		p.Edges.Add((*Edge).Compare, &Edge{Source: newP.ID, Target: newD.ID})
		pBaseAPs.Add(NodeID.Compare, newP.ID)
	}

	minimals := g.Minimals()
	for i := 1; i < components.Len(); i++ {
		component, ok := components.At(i)
		if !ok {
			panic(fmt.Sprintf("AnyValidPFDWithFeedback: missing component: %v", components))
		}

		componentMinimals := component
		componentMinimals.Intersection(graph.Node.Compare, minimals)

		ids := sets.NewWithCapacity[NodeID](componentMinimals.Len())
		for _, n := range componentMinimals.Iter() {
			ids.Add(NodeID.Compare, NodeID(n))
		}

		// NOTE: Attach initial deliverables to appropriate atomic processes.
		id := AnyDeliverableID(ids, t)
		ap := AnyAtomicProcessID(pBaseAPs, t)
		p.Edges.Add((*Edge).Compare, &Edge{Source: ap, Target: id})
	}

	return p
}

func OnlyAtomicDeliverables(nodes *sets.Set[*Node]) *sets.Set[NodeID] {
	ds := sets.NewWithCapacity[NodeID](nodes.Len())
	for _, node := range nodes.Iter() {
		if node.Type == NodeTypeAtomicDeliverable {
			ds.Add(NodeID.Compare, node.ID)
		}
	}
	return ds
}

func AnyDeliverableID(ds *sets.Set[NodeID], t *rapid.T) NodeID {
	idx := rapid.IntRange(0, ds.Len()-1).Draw(t, fmt.Sprintf("AnyDeliverableID: %v", ds))
	d, ok := ds.At(idx)
	if !ok {
		panic(fmt.Sprintf("AnyDeliverableID: %v", ds))
	}
	return d
}

func OnlyAtomicProcesses(nodes *sets.Set[*Node]) *sets.Set[NodeID] {
	aps := sets.NewWithCapacity[NodeID](nodes.Len())
	for _, node := range nodes.Iter() {
		if node.Type == NodeTypeAtomicProcess {
			aps.Add(NodeID.Compare, node.ID)
		}
	}
	return aps
}

func AnyAtomicProcessID(aps *sets.Set[NodeID], t *rapid.T) NodeID {
	idx := rapid.IntRange(0, aps.Len()-1).Draw(t, fmt.Sprintf("AnyAtomicProcessID: %v", aps))
	d, ok := aps.At(idx)
	if !ok {
		panic(fmt.Sprintf("AnyAtomicProcessID: %v", aps))
	}
	return d
}

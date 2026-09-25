package pfdfuzz

import (
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"pgregory.net/rapid"
)

func AnyValidPFD(t *rapid.T, n int) *pfd.PFD {
	p := AnyValidPFDWithoutFeedback(t, n)

	edgeMap, _ := pfd.NewEdgeMap(p.Edges)
	paths := CollectAtomicProcess2DeliverablePaths(edgeMap, OnlyAtomicProcesses(p.Nodes))

	for _, path := range paths.Iter() {

		if rapid.IntRange(0, 4).Draw(t, fmt.Sprintf("path %v", path)) == 0 {
			p.Edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: path[len(path)-1], Target: path[0], IsFeedback: true})
		}
	}

	return p
}

func CollectAtomicProcess2DeliverablePaths(edgeMap map[pfd.NodeID]*sets.Set[pfd.NodeID], aps *sets.Set[pfd.NodeID]) *sets.Set[[]pfd.NodeID] {
	var dfs func(ap pfd.NodeID, path []pfd.NodeID, res *sets.Set[[]pfd.NodeID])
	dfs = func(ap pfd.NodeID, path []pfd.NodeID, res *sets.Set[[]pfd.NodeID]) {
		path = append(path, ap)

		dsts, ok := edgeMap[ap]
		if !ok {
			panic(fmt.Sprintf("AnyValidPFD: missing dsts in edgeMap: %q", ap))
		}

		for _, d := range dsts.Iter() {
			newPath := append(path, d)
			res.Add(cmp2.CompareSlice[[]pfd.NodeID](pfd.NodeID.Compare), newPath)

			m, ok := edgeMap[d]
			if !ok {
				continue
			}
			for _, ap2 := range m.Iter() {
				dfs(ap2, newPath, res)
			}
		}
	}

	ap2DPaths := sets.New(cmp2.CompareSlice[[]pfd.NodeID](pfd.NodeID.Compare))
	for _, ap := range aps.Iter() {
		dfs(ap, make([]pfd.NodeID, 0), ap2DPaths)
	}
	return ap2DPaths
}

func AnyValidPFDWithoutFeedback(t *rapid.T, n int) *pfd.PFD {
	nodeMap := make(map[pfd.NodeID]*pfd.Node)
	nodes := sets.New((*pfd.Node).Compare)
	edges := sets.New((*pfd.Edge).Compare)

	initDelivsCount := rapid.IntRange(1, 5).Draw(t, "initDelivsCount")
	for i := 0; i < initDelivsCount; i++ {
		node := &pfd.Node{
			ID:          pfd.NewDeliverableID(nodes.Len()),
			Description: fmt.Sprintf("Initial deliverable %d", nodes.Len()),
			Type:        pfd.NodeTypeAtomicDeliverable,
		}
		nodes.Add((*pfd.Node).Compare, node)
		nodeMap[node.ID] = node
	}

	for i := 0; i < n; i++ {
		isAP := rapid.Bool().Draw(t, fmt.Sprintf("apOrDeliv %d", i))
		var node *pfd.Node
		if isAP {
			ap := &pfd.Node{
				ID:          pfd.NewAtomicProcessID(nodes.Len()),
				Description: fmt.Sprintf("Atomic Process %d", nodes.Len()),
				Type:        pfd.NodeTypeAtomicProcess,
			}
			node = ap

			ds := OnlyAtomicDeliverables(nodes)

			d := AnyDeliverableID(ds, t)
			edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: d, Target: ap.ID})

			for j, d := range ds.Iter() {
				if rapid.Bool().Draw(t, fmt.Sprintf("edge %d %d", i, j)) {
					dNode, ok := nodeMap[d]
					if !ok {
						panic(fmt.Sprintf("AnyValidPFD: missing ap in nodeMap: %q", d))
					}
					edges.Add((*pfd.Edge).Compare, &pfd.Edge{
						Source: dNode.ID,
						Target: ap.ID,
					})
				}
			}
		} else {
			d := &pfd.Node{
				ID:          pfd.NewDeliverableID(nodes.Len()),
				Description: fmt.Sprintf("Deliverable %d", nodes.Len()),
				Type:        pfd.NodeTypeAtomicDeliverable,
			}
			node = d

			aps := OnlyAtomicProcesses(nodes)
			if aps.Len() == 0 {

				continue
			}

			if rapid.Bool().Draw(t, fmt.Sprintf("edge %d", i)) {

				ap := AnyAtomicProcessID(aps, t)
				edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: ap, Target: d.ID})
			}
		}

		nodeMap[node.ID] = node
		nodes.Add((*pfd.Node).Compare, node)
	}

	aps := OnlyAtomicProcesses(nodes)
	apsNotHaveOutput := aps.Clone()
	for _, ap := range aps.Iter() {
		for _, edge := range edges.Iter() {
			if edge.Source == ap {
				apsNotHaveOutput.Remove(pfd.NodeID.Compare, ap)
			}
		}
	}

	for _, ap := range apsNotHaveOutput.Iter() {

		dNode := &pfd.Node{
			ID:          pfd.NewDeliverableID(nodes.Len()),
			Description: fmt.Sprintf("Deliverable %d", nodes.Len()),
			Type:        pfd.NodeTypeAtomicDeliverable,
		}
		nodeMap[dNode.ID] = dNode
		nodes.Add((*pfd.Node).Compare, dNode)
		edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: ap, Target: dNode.ID})
	}

	p := &pfd.PFD{
		Nodes: nodes,
		Edges: edges,
	}

	logger := slog.New(slogtest.NewRapidHandler(t))
	g := p.GraphExceptFeedback(nodeMap, logger)
	components := g.WeaklyConnectedComponents()
	baseComponent, ok := components.At(0)
	if !ok {
		panic(fmt.Sprintf("AnyValidPFDWithoutFeedback: missing component: %v", components))
	}
	pBaseComponent := sets.NewWithCapacity[*pfd.Node](baseComponent.Len())
	for _, n := range baseComponent.Iter() {
		node, ok := nodeMap[pfd.NodeID(n)]
		if !ok {
			panic(fmt.Sprintf("AnyValidPFDWithoutFeedback: missing node: %v", n))
		}
		pBaseComponent.Add((*pfd.Node).Compare, node)
	}
	pBaseAPs := OnlyAtomicProcesses(pBaseComponent)
	if pBaseAPs.Len() == 0 {

		id, ok := baseComponent.At(0)
		if !ok {
			panic(fmt.Sprintf("AnyValidPFDWithoutFeedback: missing component: %v", components))
		}

		newP := &pfd.Node{
			ID:          pfd.NewAtomicProcessID(p.Nodes.Len()),
			Description: fmt.Sprintf("Atomic Process %d", p.Nodes.Len()),
			Type:        pfd.NodeTypeAtomicProcess,
		}
		p.Nodes.Add((*pfd.Node).Compare, newP)
		newD := &pfd.Node{
			ID:          pfd.NewDeliverableID(p.Nodes.Len()),
			Description: fmt.Sprintf("Deliverable %d", p.Nodes.Len()),
			Type:        pfd.NodeTypeAtomicDeliverable,
		}
		p.Nodes.Add((*pfd.Node).Compare, newD)
		p.Edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: pfd.NodeID(id), Target: newP.ID})
		p.Edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: newP.ID, Target: newD.ID})
		pBaseAPs.Add(pfd.NodeID.Compare, newP.ID)
	}

	minimals := g.Minimals()
	for i := 1; i < components.Len(); i++ {
		component, ok := components.At(i)
		if !ok {
			panic(fmt.Sprintf("AnyValidPFDWithoutFeedback: missing component: %v", components))
		}

		componentMinimals := component
		componentMinimals.Intersection(graph.Node.Compare, minimals)

		ids := sets.NewWithCapacity[pfd.NodeID](componentMinimals.Len())
		for _, n := range componentMinimals.Iter() {
			ids.Add(pfd.NodeID.Compare, pfd.NodeID(n))
		}

		id := AnyDeliverableID(ids, t)
		ap := AnyAtomicProcessID(pBaseAPs, t)
		p.Edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: ap, Target: id})
	}

	return p
}

func OnlyAtomicDeliverables(nodes *sets.Set[*pfd.Node]) *sets.Set[pfd.NodeID] {
	ds := sets.NewWithCapacity[pfd.NodeID](nodes.Len())
	for _, node := range nodes.Iter() {
		if node.Type == pfd.NodeTypeAtomicDeliverable {
			ds.Add(pfd.NodeID.Compare, node.ID)
		}
	}
	return ds
}

func AnyDeliverableID(ds *sets.Set[pfd.NodeID], t *rapid.T) pfd.NodeID {
	idx := rapid.IntRange(0, ds.Len()-1).Draw(t, fmt.Sprintf("AnyDeliverableID: %v", ds))
	d, ok := ds.At(idx)
	if !ok {
		panic(fmt.Sprintf("AnyDeliverableID: %v", ds))
	}
	return d
}

func OnlyAtomicProcesses(nodes *sets.Set[*pfd.Node]) *sets.Set[pfd.NodeID] {
	aps := sets.NewWithCapacity[pfd.NodeID](nodes.Len())
	for _, node := range nodes.Iter() {
		if node.Type == pfd.NodeTypeAtomicProcess {
			aps.Add(pfd.NodeID.Compare, node.ID)
		}
	}
	return aps
}

func AnyAtomicProcessID(aps *sets.Set[pfd.NodeID], t *rapid.T) pfd.NodeID {
	idx := rapid.IntRange(0, aps.Len()-1).Draw(t, fmt.Sprintf("AnyAtomicProcessID: %v", aps))
	d, ok := aps.At(idx)
	if !ok {
		panic(fmt.Sprintf("AnyAtomicProcessID: %v", aps))
	}
	return d
}

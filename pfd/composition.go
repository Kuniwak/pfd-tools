package pfd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/sets"
)

type CompositionCycleError struct {
	Kind string

	Cycle []NodeID
}

func (e *CompositionCycleError) Error() string {
	ids := make([]string, 0, len(e.Cycle)+1)
	for _, id := range e.Cycle {
		ids = append(ids, string(id))
	}
	if len(e.Cycle) > 0 {
		ids = append(ids, string(e.Cycle[0]))
	}
	return fmt.Sprintf("%sの入れ子に循環があります: %s", e.Kind, strings.Join(ids, " -> "))
}

func FlattenComposition(direct map[NodeID]*sets.Set[NodeID], kind string, isComposite func(NodeID) bool) (map[NodeID]*sets.Set[NodeID], error) {
	flat := make(map[NodeID]*sets.Set[NodeID], len(direct))
	path := make([]NodeID, 0, len(direct))

	var resolve func(comp NodeID) (*sets.Set[NodeID], error)
	resolve = func(comp NodeID) (*sets.Set[NodeID], error) {
		if cached, ok := flat[comp]; ok {
			return cached, nil
		}
		if i := slices.Index(path, comp); i >= 0 {
			return nil, &CompositionCycleError{Kind: kind, Cycle: slices.Clone(path[i:])}
		}
		path = append(path, comp)

		atomics := sets.New(NodeID.Compare)
		for _, member := range direct[comp].Iter() {
			if isComposite(member) {
				sub, err := resolve(member)
				if err != nil {
					return nil, err
				}
				atomics.Union(NodeID.Compare, sub)
				continue
			}
			atomics.Add(NodeID.Compare, member)
		}

		path = path[:len(path)-1]
		flat[comp] = atomics
		return atomics, nil
	}

	comps := make([]NodeID, 0, len(direct))
	for comp := range direct {
		comps = append(comps, comp)
	}
	slices.SortFunc(comps, NodeID.Compare)

	for _, comp := range comps {
		if _, err := resolve(comp); err != nil {
			return nil, err
		}
	}
	return flat, nil
}

func FlattenDeliverableComposition(direct map[NodeID]*sets.Set[NodeID], nodes *sets.Set[*Node]) (map[NodeID]*sets.Set[NodeID], error) {
	candidates := sets.NewWithCapacity[NodeID](len(direct))
	for id := range direct {
		candidates.Add(NodeID.Compare, id)
	}
	composites := CompositeDeliverableIDs(candidates, nodes)

	flat, err := FlattenComposition(direct, "複合成果物", func(member NodeID) bool {
		return composites.Contains(NodeID.Compare, member)
	})
	if err != nil {
		return nil, fmt.Errorf("pfd.FlattenDeliverableComposition: %w", err)
	}
	return flat, nil
}

func CompositeDeliverableIDs(candidates *sets.Set[NodeID], nodes *sets.Set[*Node]) *sets.Set[NodeID] {
	others := sets.NewWithCapacity[NodeID](nodes.Len())
	for _, node := range nodes.Iter() {
		if node.Type == NodeTypeCompositeDeliverable {
			continue
		}
		others.Add(NodeID.Compare, node.ID)
	}

	composites := candidates.Clone()
	composites.Difference(NodeID.Compare, others)
	return composites
}

func MustSetExpandedDeliverableComposition(p *PFD, direct map[NodeID]*sets.Set[NodeID]) *PFD {
	if err := p.SetExpandedDeliverableComposition(direct); err != nil {
		panic(fmt.Sprintf("pfd.MustSetExpandedDeliverableComposition: %s", err))
	}
	return p
}

func (p *PFD) SetExpandedDeliverableComposition(direct map[NodeID]*sets.Set[NodeID]) error {
	flat, err := FlattenDeliverableComposition(direct, p.Nodes)
	if err != nil {
		return fmt.Errorf("pfd.PFD.SetExpandedDeliverableComposition: %w", err)
	}
	p.DeliverableComposition = DrawnDeliverableComposition(flat, p.Nodes)

	newEdges := p.Edges.Clone()
	for _, edge := range p.Edges.Iter() {
		if ns, ok := p.DeliverableComposition[edge.Source]; ok {
			for _, newSource := range ns.Iter() {
				newEdges.Add((*Edge).Compare, &Edge{Source: newSource, Target: edge.Target, IsFeedback: edge.IsFeedback})
			}
		}

		if ns, ok := p.DeliverableComposition[edge.Target]; ok {
			for _, newTarget := range ns.Iter() {
				newEdges.Add((*Edge).Compare, &Edge{Source: edge.Source, Target: newTarget, IsFeedback: edge.IsFeedback})
			}
		}
	}
	p.Edges = newEdges

	return nil
}

func DrawnDeliverableComposition(composition map[NodeID]*sets.Set[NodeID], nodes *sets.Set[*Node]) map[NodeID]*sets.Set[NodeID] {
	drawn := make(map[NodeID]*sets.Set[NodeID], len(composition))
	for _, node := range nodes.Iter() {
		if node.Type != NodeTypeCompositeDeliverable {
			continue
		}
		if members, ok := composition[node.ID]; ok {
			drawn[node.ID] = members
		}
	}
	return drawn
}

func FlattenProcessComposition(
	direct map[NodeID]*sets.Set[NodeID],
	nodeMap map[NodeID]*Node,
) (map[NodeID]*sets.Set[NodeID], error) {
	flat, err := FlattenComposition(direct, "複合プロセス", func(member NodeID) bool {
		node, ok := nodeMap[member]
		return ok && node.Type == NodeTypeCompositeProcess
	})
	if err != nil {
		return nil, fmt.Errorf("pfd.FlattenProcessComposition: %w", err)
	}
	return flat, nil
}

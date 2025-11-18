package pfd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

type PFD struct {
	Title                  string                       `json:"title,omitempty"`
	Nodes                  *sets.Set[*Node]             `json:"nodes,omitempty"`
	Edges                  *sets.Set[*Edge]             `json:"edges,omitempty"`
	ProcessComposition     map[NodeID]*sets.Set[NodeID] `json:"process_composition,omitempty"`
	DeliverableComposition map[NodeID]*sets.Set[NodeID] `json:"deliverable_composition,omitempty"`
}

func NewPFD(
	title string,
	nodes *sets.Set[*Node],
	edges *sets.Set[*Edge],
	processComposition map[NodeID]*sets.Set[NodeID],
	deliverableComposition map[NodeID]*sets.Set[NodeID],
) *PFD {
	return &PFD{
		Title:                  title,
		Nodes:                  nodes,
		Edges:                  edges,
		ProcessComposition:     processComposition,
		DeliverableComposition: deliverableComposition,
	}
}

// GraphExceptFeedback returns a graph excluding feedback edges and compound processes.
func (p *PFD) GraphExceptFeedback(nodeMap map[NodeID]*Node, logger *slog.Logger) *graph.Graph {
	footprint := sets.NewWithCapacity[NodeID](p.Nodes.Len())

	nodes := sets.NewWithCapacity[graph.Node](p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type == NodeTypeCompositeProcess {
			continue
		}
		if footprint.Contains(NodeID.Compare, node.ID) {
			continue
		}
		footprint.Add(NodeID.Compare, node.ID)
		nodes.Add(graph.Node.Compare, graph.Node(node.ID))
	}

	edges := sets.NewWithCapacity[*pairs.Pair[graph.Node, graph.Node]](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.IsFeedback {
			continue
		}

		src, srcOk := nodeMap[edge.Source]
		if !srcOk {
			logger.Warn("GraphExceptFeedback: missing node", "source", edge.Source)
			continue
		}
		if src.Type == NodeTypeCompositeProcess {
			continue
		}

		dst, dstOk := nodeMap[edge.Target]
		if !dstOk {
			logger.Warn("GraphExceptFeedback: missing node", "target", edge.Target)
			continue
		}
		if dst.Type == NodeTypeCompositeProcess {
			continue
		}

		edges.Add(pairs.Compare(graph.Node.Compare, graph.Node.Compare), pairs.New(graph.Node(edge.Source), graph.Node(edge.Target)))
	}

	return &graph.Graph{
		Nodes: nodes,
		Edges: edges,
	}
}

func (p *PFD) GraphIncludingFeedback(nodeMap map[NodeID]*Node, logger *slog.Logger) *graph.Graph {
	g := p.GraphExceptFeedback(nodeMap, logger)

	for _, edge := range p.Edges.Iter() {
		if !edge.IsFeedback {
			continue
		}

		src, srcOk := nodeMap[edge.Source]
		if !srcOk {
			logger.Warn("GraphIncludingFeedback: missing node", "source", edge.Source)
			continue
		}
		if src.Type == NodeTypeCompositeProcess {
			continue
		}

		dst, dstOk := nodeMap[edge.Target]
		if !dstOk {
			logger.Warn("GraphIncludingFeedback: missing node", "target", edge.Target)
			continue
		}
		if dst.Type == NodeTypeCompositeProcess {
			continue
		}

		g.Edges.Add(pairs.Compare(graph.Node.Compare, graph.Node.Compare), pairs.New(graph.Node(edge.Source), graph.Node(edge.Target)))
	}
	return g
}

func (p *PFD) InputsIncludingFeedback(n NodeID) *sets.Set[NodeID] {
	inputs := sets.NewWithCapacity[NodeID](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.Target == n {
			inputs.Add(NodeID.Compare, edge.Source)
		}
	}
	return inputs
}

func (p *PFD) InputsExceptFeedback(n NodeID) *sets.Set[NodeID] {
	inputs := sets.NewWithCapacity[NodeID](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.IsFeedback {
			continue
		}
		if edge.Target == n {
			inputs.Add(NodeID.Compare, edge.Source)
		}
	}
	return inputs
}

func (p *PFD) InputsOnlyFeedback(n NodeID) *sets.Set[NodeID] {
	inputs := sets.NewWithCapacity[NodeID](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.IsFeedback && edge.Target == n {
			inputs.Add(NodeID.Compare, edge.Source)
		}
	}
	return inputs
}

func (p *PFD) OutputsIncludingFeedback(n NodeID) *sets.Set[NodeID] {
	outputs := sets.NewWithCapacity[NodeID](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.Source == n {
			outputs.Add(NodeID.Compare, edge.Target)
		}
	}
	return outputs
}

func (p *PFD) OutputsExceptFeedback(n NodeID) *sets.Set[NodeID] {
	outputs := sets.NewWithCapacity[NodeID](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.IsFeedback {
			continue
		}
		if edge.Source == n {
			outputs.Add(NodeID.Compare, edge.Target)
		}
	}
	return outputs
}

func (p *PFD) OutputsOnlyFeedback(n NodeID) *sets.Set[NodeID] {
	outputs := sets.NewWithCapacity[NodeID](p.Edges.Len())
	for _, edge := range p.Edges.Iter() {
		if edge.IsFeedback && edge.Source == n {
			outputs.Add(NodeID.Compare, edge.Target)
		}
	}
	return outputs
}

func (p *PFD) InitialAtomicDeliverables(nodeMap map[NodeID]*Node, logger *slog.Logger) *sets.Set[NodeID] {
	initials := sets.NewWithCapacity[NodeID](p.Nodes.Len())
	minimals := p.GraphExceptFeedback(nodeMap, logger).Minimals()
	for _, minimal := range minimals.Iter() {
		node, ok := nodeMap[NodeID(minimal)]
		if !ok {
			panic(fmt.Sprintf("pfd.PFD.InitialAtomicDeliverables: missing node: %q", minimal))
		}
		if node.Type != NodeTypeAtomicDeliverable {
			continue
		}
		initials.Add(NodeID.Compare, node.ID)
	}
	return initials
}

func (p *PFD) FinalAtomicDeliverables(nodeMap map[NodeID]*Node, logger *slog.Logger) *sets.Set[NodeID] {
	finals := sets.NewWithCapacity[NodeID](p.Nodes.Len())
	maximals := p.GraphExceptFeedback(nodeMap, logger).Maximals()
	for _, maximal := range maximals.Iter() {
		node, ok := nodeMap[NodeID(maximal)]
		if !ok {
			panic(fmt.Sprintf("pfd.PFD.FinalAtomicDeliverables: missing node: %q", maximal))
		}
		if node.Type != NodeTypeAtomicDeliverable {
			continue
		}
		finals.Add(NodeID.Compare, node.ID)
	}
	return finals
}

func (p *PFD) AtomicProcesses() *sets.Set[NodeID] {
	aps := sets.NewWithCapacity[NodeID](p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type == NodeTypeAtomicProcess {
			aps.Add(NodeID.Compare, node.ID)
		}
	}
	return aps
}

type NodeType string

const (
	NodeTypeAtomicProcess        NodeType = "PROCESS"
	NodeTypeCompositeProcess     NodeType = "COMPOSITE_PROCESS"
	NodeTypeAtomicDeliverable    NodeType = "DELIVERABLE"
	NodeTypeCompositeDeliverable NodeType = "COMPOSITE_DELIVERABLE"
)

func (t NodeType) IsProcess() bool {
	return t == NodeTypeAtomicProcess || t == NodeTypeCompositeProcess
}

func (t NodeType) IsDeliverable() bool {
	return t == NodeTypeAtomicDeliverable || t == NodeTypeCompositeDeliverable
}

func (t NodeType) Compare(other NodeType) int {
	return strings.Compare(string(t), string(other))
}

// Node is a PFD element. Description may be empty with the description stored in ID instead.
type Node struct {
	ID          NodeID   `json:"id"`
	Description string   `json:"desc"`
	Type        NodeType `json:"type"`
}

// HasID returns true if the element has an ID, otherwise returns false.
func (n *Node) HasID() bool {
	// NOTE: Due to parsing constraints, when there is no ID, the description is stored in ID. In that case, the description is empty, so this is used for determination.
	return n.Description != ""
}

// HasDescription returns true if the element has a description, otherwise returns false.
func (n *Node) HasDescription() bool {
	return n.HasID()
}

func (n *Node) Clone() *Node {
	return &Node{
		ID:          n.ID,
		Description: n.Description,
		Type:        n.Type,
	}
}

func (n *Node) Compare(b *Node) int {
	cmp1 := n.ID.Compare(b.ID)
	if cmp1 != 0 {
		return cmp1
	}
	return strings.Compare(n.Description, b.Description)
}

func (n *Node) Data() *NodeData {
	return &NodeData{
		Description: n.Description,
		Type:        n.Type,
	}
}

type NodeData struct {
	Description string   `json:"desc"`
	Type        NodeType `json:"type"`
}

func (n *NodeData) Compare(other *NodeData) int {
	c := strings.Compare(n.Description, other.Description)
	if c != 0 {
		return c
	}
	return n.Type.Compare(other.Type)
}

func NewNodeMap(nodes *sets.Set[*Node], logger *slog.Logger) map[NodeID]*Node {
	nodeMap := make(map[NodeID]*Node)
	for _, node := range nodes.Iter() {
		if _, ok := nodeMap[node.ID]; ok {
			logger.Warn("NewNodeMap: duplicate node ID", "id", node.ID)
			continue
		}
		nodeMap[node.ID] = node
	}
	return nodeMap
}

type Edge struct {
	Source     NodeID `json:"source"`
	Target     NodeID `json:"target"`
	IsFeedback bool   `json:"feedback,omitempty"`
}

func (e *Edge) Clone() *Edge {
	return &Edge{
		Source:     e.Source,
		Target:     e.Target,
		IsFeedback: e.IsFeedback,
	}
}

func (e *Edge) Compare(b *Edge) int {
	cmp1 := e.Source.Compare(b.Source)
	if cmp1 != 0 {
		return cmp1
	}
	cmp2 := e.Target.Compare(b.Target)
	if cmp2 != 0 {
		return cmp2
	}
	if e.IsFeedback {
		if !b.IsFeedback {
			return -1
		}
	} else {
		if b.IsFeedback {
			return 1
		}
	}
	return 0
}

func NewEdgeMap(edges *sets.Set[*Edge]) (map[NodeID]*sets.Set[NodeID], map[NodeID]*sets.Set[NodeID]) {
	edgeMap := make(map[NodeID]*sets.Set[NodeID])
	fbMap := make(map[NodeID]*sets.Set[NodeID])
	for _, edge := range edges.Iter() {
		if edge.IsFeedback {
			if fbm, ok := fbMap[edge.Source]; ok {
				fbm.Add(NodeID.Compare, edge.Target)
			} else {
				fbMap[edge.Source] = sets.New(NodeID.Compare, edge.Target)
			}
		} else {
			if em, ok := edgeMap[edge.Source]; ok {
				em.Add(NodeID.Compare, edge.Target)
			} else {
				edgeMap[edge.Source] = sets.New(NodeID.Compare, edge.Target)
			}
		}
	}
	return edgeMap, fbMap
}

func NewReversedEdgeMap(edges *sets.Set[*Edge]) (map[NodeID]*sets.Set[NodeID], map[NodeID]*sets.Set[NodeID]) {
	reversedEdgeMap := make(map[NodeID]*sets.Set[NodeID])
	reversedFBEdgeMap := make(map[NodeID]*sets.Set[NodeID])
	for _, edge := range edges.Iter() {
		if edge.IsFeedback {
			if fbem, ok := reversedFBEdgeMap[edge.Target]; ok {
				fbem.Add(NodeID.Compare, edge.Source)
			} else {
				reversedFBEdgeMap[edge.Target] = sets.New(NodeID.Compare, edge.Source)
			}
		} else {
			if rem, ok := reversedEdgeMap[edge.Target]; ok {
				rem.Add(NodeID.Compare, edge.Source)
			} else {
				reversedEdgeMap[edge.Target] = sets.New(NodeID.Compare, edge.Source)
			}
		}
	}
	return reversedEdgeMap, reversedFBEdgeMap
}

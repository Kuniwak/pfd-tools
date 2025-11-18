package pfd

import (
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/sets"
)

type DiffElementKind string

const (
	DiffElementKindAdd    DiffElementKind = "add"
	DiffElementKindRemove DiffElementKind = "remove"
	DiffElementKindChange DiffElementKind = "change"
)

type EdgeDiffResult struct {
	ExtraEdges   *sets.Set[*Edge]           `json:"extraEdges,omitempty"`
	MissingEdges *sets.Set[*Edge]           `json:"missingEdges,omitempty"`
	ChangedEdges *sets.Set[*EdgeDiffChange] `json:"changedEdges,omitempty"`
	SameEdges    *sets.Set[*Edge]           `json:"sameEdges,omitempty"`
}

func (e EdgeDiffResult) IsEmpty() bool {
	return e.ExtraEdges.Len() == 0 && e.MissingEdges.Len() == 0 && e.ChangedEdges.Len() == 0
}

type EdgeDiffChange struct {
	Source           NodeID          `json:"source"`
	Target           NodeID          `json:"target"`
	OldFeedbackFlags *sets.Set[bool] `json:"oldFeedbackFlags,omitempty"`
	NewFeedbackFlags *sets.Set[bool] `json:"newFeedbackFlags,omitempty"`
}

func (e *EdgeDiffChange) Compare(other *EdgeDiffChange) int {
	c := sets.Compare(cmp2.CompareBool)(e.OldFeedbackFlags, other.OldFeedbackFlags)
	if c != 0 {
		return c
	}
	return sets.Compare(cmp2.CompareBool)(e.NewFeedbackFlags, other.NewFeedbackFlags)
}

type NodeDiffResult struct {
	ExtraNodes   *sets.Set[*Node]           `json:"extraNodes,omitempty"`
	MissingNodes *sets.Set[*Node]           `json:"missingNodes,omitempty"`
	ChangedNodes *sets.Set[*NodeDiffChange] `json:"changedNodes,omitempty"`
	SameNodes    *sets.Set[*Node]           `json:"sameNodes,omitempty"`
}

func (n NodeDiffResult) IsEmpty() bool {
	return n.ExtraNodes.Len() == 0 && n.MissingNodes.Len() == 0 && n.ChangedNodes.Len() == 0
}

type NodeDiffChange struct {
	ID  NodeID               `json:"id"`
	Old *sets.Set[*NodeData] `json:"old,omitempty"`
	New *sets.Set[*NodeData] `json:"new,omitempty"`
}

func (n *NodeDiffChange) Compare(other *NodeDiffChange) int {
	c := n.ID.Compare(other.ID)
	if c != 0 {
		return c
	}
	c = sets.Compare((*NodeData).Compare)(n.Old, other.Old)
	if c != 0 {
		return c
	}
	return sets.Compare((*NodeData).Compare)(n.New, other.New)
}

type DiffResult struct {
	NodeDiff *NodeDiffResult `json:"nodeDiff,omitempty"`
	EdgeDiff *EdgeDiffResult `json:"edgeDiff,omitempty"`
}

func (d DiffResult) IsEmpty() bool {
	return d.NodeDiff.IsEmpty() && d.EdgeDiff.IsEmpty()
}

func (d *DiffResult) Write(w io.Writer, showSame bool) error {
	const sep = ": "
	const extraPrefix = "+ "
	const missingPrefix = "- "
	const samePrefix = "  "
	const feedbackEdge = " - - > "
	const normalEdge = " ----> "
	const openAtomicProcess = "("
	const closeAtomicProcess = ")"
	const openAtomicDeliverable = "["
	const closeAtomicDeliverable = "]"
	const openCompositeDeliverable = "[["
	const closeCompositeDeliverable = "]]"
	const openCompositeProcess = "(("
	const closeCompositeProcess = "))"
	const newLine = "\n"

	for _, node := range d.NodeDiff.ExtraNodes.Iter() {
		io.WriteString(w, extraPrefix)
		io.WriteString(w, string(node.ID))
		io.WriteString(w, sep)
		io.WriteString(w, node.Description)
		io.WriteString(w, newLine)
	}
	for _, node := range d.NodeDiff.MissingNodes.Iter() {
		io.WriteString(w, missingPrefix)
		io.WriteString(w, string(node.ID))
		io.WriteString(w, sep)
		io.WriteString(w, node.Description)
		io.WriteString(w, newLine)
	}
	for _, node := range d.NodeDiff.ChangedNodes.Iter() {
		for _, old := range node.Old.Iter() {
			io.WriteString(w, missingPrefix)
			switch old.Type {
			case NodeTypeAtomicProcess:
				io.WriteString(w, openAtomicProcess)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeAtomicProcess)

			case NodeTypeCompositeProcess:
				io.WriteString(w, openCompositeProcess)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeCompositeProcess)

			case NodeTypeAtomicDeliverable:
				io.WriteString(w, openAtomicDeliverable)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeAtomicDeliverable)

			case NodeTypeCompositeDeliverable:
				io.WriteString(w, openCompositeDeliverable)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeCompositeDeliverable)

			default:
				panic(fmt.Sprintf("pfd.DiffResult.Write: unknown node type: %s", old.Type))
			}
			io.WriteString(w, sep)
			io.WriteString(w, old.Description)
			io.WriteString(w, newLine)
		}
		for _, new := range node.New.Iter() {
			io.WriteString(w, extraPrefix)
			switch new.Type {
			case NodeTypeAtomicProcess:
				io.WriteString(w, openAtomicProcess)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeAtomicProcess)

			case NodeTypeCompositeProcess:
				io.WriteString(w, openCompositeProcess)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeCompositeProcess)

			case NodeTypeAtomicDeliverable:
				io.WriteString(w, openAtomicDeliverable)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeAtomicDeliverable)

			case NodeTypeCompositeDeliverable:
				io.WriteString(w, openCompositeDeliverable)
				io.WriteString(w, string(node.ID))
				io.WriteString(w, closeCompositeDeliverable)

			default:
				panic(fmt.Sprintf("pfd.DiffResult.Write: unknown node type: %s", new.Type))
			}
			io.WriteString(w, sep)
			io.WriteString(w, new.Description)
			io.WriteString(w, newLine)
		}
	}

	if showSame {
		for _, node := range d.NodeDiff.SameNodes.Iter() {
			io.WriteString(w, samePrefix)
			io.WriteString(w, string(node.ID))
			io.WriteString(w, sep)
			io.WriteString(w, node.Description)
			io.WriteString(w, newLine)
		}
	}

	for _, edge := range d.EdgeDiff.ExtraEdges.Iter() {
		io.WriteString(w, extraPrefix)
		io.WriteString(w, string(edge.Source))
		if edge.IsFeedback {
			io.WriteString(w, feedbackEdge)
		} else {
			io.WriteString(w, normalEdge)
		}
		io.WriteString(w, string(edge.Target))
		io.WriteString(w, newLine)
	}

	for _, edge := range d.EdgeDiff.MissingEdges.Iter() {
		io.WriteString(w, missingPrefix)
		io.WriteString(w, string(edge.Source))
		if edge.IsFeedback {
			io.WriteString(w, feedbackEdge)
		} else {
			io.WriteString(w, normalEdge)
		}
		io.WriteString(w, string(edge.Target))
		io.WriteString(w, newLine)
	}

	for _, edge := range d.EdgeDiff.ChangedEdges.Iter() {
		for _, old := range edge.OldFeedbackFlags.Iter() {
			io.WriteString(w, missingPrefix)
			io.WriteString(w, string(edge.Source))
			if old {
				io.WriteString(w, feedbackEdge)
			} else {
				io.WriteString(w, normalEdge)
			}
			io.WriteString(w, string(edge.Target))
			io.WriteString(w, newLine)
		}
		for _, new := range edge.NewFeedbackFlags.Iter() {
			io.WriteString(w, extraPrefix)
			io.WriteString(w, string(edge.Source))
			if new {
				io.WriteString(w, feedbackEdge)
			} else {
				io.WriteString(w, normalEdge)
			}
			io.WriteString(w, string(edge.Target))
			io.WriteString(w, newLine)
		}
	}

	if showSame {
		for _, edge := range d.EdgeDiff.SameEdges.Iter() {
			io.WriteString(w, samePrefix)
			io.WriteString(w, string(edge.Source))
			io.WriteString(w, sep)
			io.WriteString(w, string(edge.Target))
			io.WriteString(w, newLine)
		}
	}

	return nil
}

func NodeDiff(ns1, ns2 *sets.Set[*Node]) *NodeDiffResult {
	pADescsMap := make(map[NodeID]*sets.Set[*NodeData])
	pAIDs := sets.NewWithCapacity[NodeID](ns1.Len())
	for _, node := range ns1.Iter() {
		pAIDs.Add(NodeID.Compare, node.ID)
		if _, ok := pADescsMap[node.ID]; ok {
			pADescsMap[node.ID].Add((*NodeData).Compare, node.Data())
		} else {
			pADescsMap[node.ID] = sets.New((*NodeData).Compare, node.Data())
		}
	}
	pBDescsMap := make(map[NodeID]*sets.Set[*NodeData])
	pBIDs := sets.NewWithCapacity[NodeID](ns2.Len())
	for _, node := range ns2.Iter() {
		pBIDs.Add(NodeID.Compare, node.ID)
		if _, ok := pBDescsMap[node.ID]; ok {
			pBDescsMap[node.ID].Add((*NodeData).Compare, node.Data())
		} else {
			pBDescsMap[node.ID] = sets.New((*NodeData).Compare, node.Data())
		}
	}

	missingIDs := pAIDs.Clone()
	missingIDs.Difference(NodeID.Compare, pBIDs)
	extraIDs := pBIDs.Clone()
	extraIDs.Difference(NodeID.Compare, pAIDs)
	sameIDs := pAIDs.Clone()
	sameIDs.Intersection(NodeID.Compare, pBIDs)

	missingNodes := sets.NewWithCapacity[*Node](missingIDs.Len())
	for _, node := range ns1.Iter() {
		if missingIDs.Contains(NodeID.Compare, node.ID) {
			missingNodes.Add((*Node).Compare, node)
		}
	}
	extraNodes := sets.NewWithCapacity[*Node](extraIDs.Len())
	for _, node := range ns2.Iter() {
		if extraIDs.Contains(NodeID.Compare, node.ID) {
			extraNodes.Add((*Node).Compare, node)
		}
	}
	changedNodes := sets.NewWithCapacity[*NodeDiffChange](sameIDs.Len())
	sameNodes := sets.NewWithCapacity[*Node](sameIDs.Len())
	for _, node := range ns1.Iter() {
		if sameIDs.Contains(NodeID.Compare, node.ID) {
			if sets.IsEqual((*NodeData).Compare, pADescsMap[node.ID], pBDescsMap[node.ID]) {
				sameNodes.Add((*Node).Compare, node)
			} else {
				changedNodes.Add((*NodeDiffChange).Compare, &NodeDiffChange{
					ID:  node.ID,
					Old: pADescsMap[node.ID],
					New: pBDescsMap[node.ID],
				})
			}
		}
	}
	return &NodeDiffResult{
		ExtraNodes:   extraNodes,
		MissingNodes: missingNodes,
		ChangedNodes: changedNodes,
		SameNodes:    sameNodes,
	}
}

type justEdge struct {
	source NodeID
	target NodeID
}

func makeJustEdge(e *Edge) justEdge {
	return justEdge{source: e.Source, target: e.Target}
}

func (e justEdge) Compare(other justEdge) int {
	cmp1 := e.source.Compare(other.source)
	if cmp1 != 0 {
		return cmp1
	}
	return e.target.Compare(other.target)
}

func EdgeDiff(es1, es2 *sets.Set[*Edge]) *EdgeDiffResult {
	eAs := sets.NewWithCapacity[justEdge](es1.Len())
	eAm := make(map[justEdge]*sets.Set[bool])
	for _, edge := range es1.Iter() {
		e := makeJustEdge(edge)
		eAs.Add(justEdge.Compare, e)

		if _, ok := eAm[e]; ok {
			eAm[e].Add(cmp2.CompareBool, edge.IsFeedback)
		} else {
			eAm[e] = sets.New(cmp2.CompareBool, edge.IsFeedback)
		}
	}

	eBs := sets.NewWithCapacity[justEdge](es2.Len())
	eBm := make(map[justEdge]*sets.Set[bool])
	for _, edge := range es2.Iter() {
		e := makeJustEdge(edge)
		eBs.Add(justEdge.Compare, e)

		if _, ok := eBm[e]; ok {
			eBm[e].Add(cmp2.CompareBool, edge.IsFeedback)
		} else {
			eBm[e] = sets.New(cmp2.CompareBool, edge.IsFeedback)
		}
	}

	missingJustEdges := eAs.Clone()
	missingJustEdges.Difference(justEdge.Compare, eBs)
	extraJustEdges := eBs.Clone()
	extraJustEdges.Difference(justEdge.Compare, eAs)
	sameJustEdges := eAs.Clone()
	sameJustEdges.Intersection(justEdge.Compare, eBs)

	missingEdges := sets.NewWithCapacity[*Edge](missingJustEdges.Len())
	for _, edge := range es1.Iter() {
		if !missingJustEdges.Contains(justEdge.Compare, makeJustEdge(edge)) {
			continue
		}
		missingEdges.Add((*Edge).Compare, edge)
	}

	extraEdges := sets.NewWithCapacity[*Edge](extraJustEdges.Len())
	for _, edge := range es2.Iter() {
		if !extraJustEdges.Contains(justEdge.Compare, makeJustEdge(edge)) {
			continue
		}
		extraEdges.Add((*Edge).Compare, edge)
	}

	sameEdges := sets.NewWithCapacity[*Edge](sameJustEdges.Len())
	changedEdges := sets.NewWithCapacity[*EdgeDiffChange](sameEdges.Len())
	for _, edge := range es1.Iter() {
		if sameJustEdges.Contains(justEdge.Compare, makeJustEdge(edge)) {
			if sets.IsEqual(cmp2.CompareBool, eAm[makeJustEdge(edge)], eBm[makeJustEdge(edge)]) {
				sameEdges.Add((*Edge).Compare, edge)
			} else {
				changedEdges.Add((*EdgeDiffChange).Compare, &EdgeDiffChange{
					Source:           edge.Source,
					Target:           edge.Target,
					OldFeedbackFlags: eAm[makeJustEdge(edge)],
					NewFeedbackFlags: eBm[makeJustEdge(edge)],
				})
			}
		}
	}
	return &EdgeDiffResult{
		ExtraEdges:   extraEdges,
		MissingEdges: missingEdges,
		ChangedEdges: changedEdges,
		SameEdges:    sameEdges,
	}
}

func Diff(pA, pB *PFD) *DiffResult {
	return &DiffResult{
		NodeDiff: NodeDiff(pA.Nodes, pB.Nodes),
		EdgeDiff: EdgeDiff(pA.Edges, pB.Edges),
	}
}

package pfd

import (
	"fmt"
	"slices"

	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

// RenumberPlan is a dictionary from old IDs to renumbered elements.
type RenumberPlan map[NodeID]*Node

// RenumberPlanAcc is the intermediate state of a plan to renumber PFD nodes.
type RenumberPlanAcc struct {
	// ProcessNumberNext is the next process ID number to be assigned.
	ProcessNumberNext int

	// DeliverableNumberNext is the next deliverable ID number to be assigned.
	DeliverableNumberNext int

	// Plan being constructed.
	Plan RenumberPlan
}

func NewRenumberPlanAcc(p *PFD) *RenumberPlanAcc {
	return &RenumberPlanAcc{
		ProcessNumberNext:     GetMaxProcessID(p) + 1,
		DeliverableNumberNext: GetMaxDeliverableID(p) + 1,
		Plan:                  make(RenumberPlan),
	}
}

// Visit assigns a number to the element if it's not numbered. If the element is already numbered, it maintains it.
// The numbering ID is incremented each time a number is assigned.
func (r *RenumberPlanAcc) Visit(nodeID NodeID, nodeMap map[NodeID]*Node) {
	node, ok := nodeMap[nodeID]
	if !ok {
		panic(fmt.Sprintf("pfd.RenumberPlanAcc#Visit: missing node: %q", nodeID))
	}

	if node.Type.IsProcess() {
		r.VisitProcess(nodeID, nodeMap)
		return
	}

	if node.Type.IsDeliverable() {
		r.VisitDeliverable(nodeID, nodeMap)
		return
	}
}

// VisitProcess assigns a number to the process if it's not numbered. If the atomic process is already numbered, it maintains it.
// The numbering ID is incremented each time a number is assigned.
func (r *RenumberPlanAcc) VisitProcess(nodeID NodeID, nodeMap map[NodeID]*Node) {
	node, ok := nodeMap[nodeID]
	if !ok {
		panic(fmt.Sprintf("pfd.RenumberPlanAcc#VisitProcess: missing node: %q", nodeID))
	}

	if !node.Type.IsProcess() {
		panic(fmt.Sprintf("pfd.RenumberPlanAcc#VisitProcess: not a process: %q", nodeID))
	}

	newNode := node.Clone()

	if node.HasID() {
		// NOTE: If there is an ID, maintain it as is.
		r.Plan[nodeID] = newNode
		return
	}

	newNode.ID = NewAtomicProcessID(r.ProcessNumberNext)
	r.ProcessNumberNext++
	// NOTE: If there is no ID, the description is in the ID, so restore it.
	newNode.Description = string(nodeID)

	r.Plan[nodeID] = newNode
}

// VisitDeliverable assigns a number to the deliverable if it's not numbered.
// The numbering ID is incremented each time a number is assigned.
func (r *RenumberPlanAcc) VisitDeliverable(nodeID NodeID, nodeMap map[NodeID]*Node) {
	node, ok := nodeMap[nodeID]
	if !ok {
		panic(fmt.Sprintf("pfd.RenumberPlanAcc#VisitDeliverable: missing node: %q", nodeID))
	}

	if !node.Type.IsDeliverable() {
		panic(fmt.Sprintf("pfd.RenumberPlanAcc#VisitDeliverable: not a deliverable: %q", nodeID))
	}

	newNode := node.Clone()

	if node.HasID() {
		// NOTE: If there is an ID, maintain it as is.
		r.Plan[nodeID] = newNode
		return
	}

	newNode.ID = NewDeliverableID(r.DeliverableNumberNext)
	r.DeliverableNumberNext++
	// NOTE: If there is no ID, the description is in the ID, so restore it.
	newNode.Description = string(nodeID)

	r.Plan[nodeID] = newNode
}

// NewRenumberPlan returns a plan to renumber PFD nodes. The renumbering satisfies all of the following rules:
//   - RULE1: If there is an ID, the previous ID is retained
//   - RULE2: If the existing ID of renumbered deliverables is in D%d format, the ID of the final deliverable is lexically before non-final deliverables
//   - RULE3: If the existing ID of renumbered deliverables is in D%d format and deliverable B is reachable from non-final deliverable A, then A's ID is lexically before B's ID
//   - RULE4: If the existing ID of renumbered atomic processes is in P%d format and atomic process B is reachable from atomic process A, then A's ID is lexically before B's ID
//   - RULE5: The ID of renumbered atomic processes is lexically before the ID of renumbered compound processes
//   - RULE6: If the minimum ID of atomic processes contained in renumbered compound process A is smaller than that of compound process B, then A's ID is lexically before B's ID
//   - RULE7: Returns an error if the graph contains cycles
func NewRenumberPlan(p *PFD, graphExceptFB *graph.Graph, nodeMap map[NodeID]*Node, errs *[]Error) (RenumberPlan, bool) {
	cycles := graphExceptFB.Cycles()
	if cycles.Len() > 0 {
		for _, cycle := range cycles.Iter() {
			locs := make([]Location, 0, len(cycle))
			for i := 0; i < len(cycle)-1; i++ {
				locs = append(locs, NewEdgeLocation(NodeID(cycle[i]), NodeID(cycle[i+1])))
			}
			*errs = append(*errs, Error{Locations: locs, Wrapped: fmt.Errorf("pfd.NewRenumberPlan: cyclic graph")})
		}
		return nil, false
	}

	// NOTE: Find final deliverables.
	maximals := graphExceptFB.Maximals()
	finals := sets.NewWithCapacity[NodeID](maximals.Len())
	for _, maximal := range maximals.Iter() {
		node, ok := nodeMap[NodeID(maximal)]
		if !ok {
			panic(fmt.Sprintf("pfd.NewRenumberPlan: missing maximal: %q", maximal))
		}
		if node.Type != NodeTypeAtomicDeliverable {
			continue
		}
		finals.Add(NodeID.Compare, node.ID)
	}

	acc := NewRenumberPlanAcc(p)

	// NOTE: Number final deliverables first.
	for _, final := range finals.Iter() {
		acc.VisitDeliverable(final, nodeMap)
	}

	// NOTE: Get order excluding final deliverables.
	order := graphExceptFB.TopologicalSort()
	orderExceptFinals := make([]NodeID, 0, len(order))
	for _, nodeID := range order {
		pfdNodeID := NodeID(nodeID)

		if finals.Contains(NodeID.Compare, pfdNodeID) {
			// NOTE: Exclude final deliverables.
			continue
		}

		orderExceptFinals = append(orderExceptFinals, pfdNodeID)
	}

	// NOTE: Number the order excluding final deliverables.
	for _, pfdNodeID := range orderExceptFinals {
		acc.Visit(pfdNodeID, nodeMap)
	}

	// NOTE: Number compound processes in order of smallest minimum ID of contained atomic processes.
	nodeIDWithMinIDs := make([]*pairs.Pair[NodeID, NodeID], 0, len(p.ProcessComposition))
	for nodeID, nodeIDs := range p.ProcessComposition {
		var minID NodeID
		if nodeIDs.Len() > 0 {
			nodesIDs2 := nodeIDs.Slice()
			minID = nodesIDs2[0]
		} else {
			// NOTE: If empty, place it earlier in the order.
			minID = NewAtomicProcessID(0)
		}
		nodeIDWithMinIDs = append(nodeIDWithMinIDs, pairs.New(nodeID, minID))
	}

	slices.SortFunc(nodeIDWithMinIDs, pairs.CompareSecond[NodeID](NodeID.Compare))

	for _, pair := range nodeIDWithMinIDs {
		acc.VisitProcess(pair.First, nodeMap)
	}

	if len(*errs) > 0 {
		return nil, false
	}
	return acc.Plan, true
}

// GetMaxProcessID returns the maximum ID among IDs numbered in P%d format.
// Returns 0 if none are numbered.
func GetMaxProcessID(p *PFD) int {
	maxNum := 0
	for _, node := range p.Nodes.Iter() {
		if node.Type.IsProcess() {
			i, err := ParseProcessID(node.ID)
			if err != nil {
				continue
			}
			if i > maxNum {
				maxNum = i
			}
		}
	}
	return maxNum
}

// GetMaxDeliverableID returns the maximum ID among IDs numbered in D%d format.
// Returns 0 if none are numbered.
func GetMaxDeliverableID(p *PFD) int {
	maxNum := 0
	for _, node := range p.Nodes.Iter() {
		if node.Type.IsDeliverable() {
			i, err := ParseDeliverableID(node.ID)
			if err != nil {
				continue
			}
			if i > maxNum {
				maxNum = i
			}
		}
	}
	return maxNum
}

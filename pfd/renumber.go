package pfd

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

type RenumberPlan map[NodeID]*Node

func (p RenumberPlan) NewlyNumbered() RenumberPlan {
	newly := make(RenumberPlan, len(p))
	for key, node := range p {
		if key == node.ID {
			continue
		}
		newly[key] = node
	}
	return newly
}

func MergeRenumberPlans(a, b RenumberPlan) RenumberPlan {
	merged := make(RenumberPlan, len(a)+len(b))
	for key, node := range a {
		merged[key] = node
	}
	for key, node := range b {
		merged[key] = node
	}
	return merged
}

type RenumberBase struct {
	Plan RenumberPlan

	MaxProcessNumber     int
	MaxDeliverableNumber int
}

func FormatMaxIDs(maxProcessNumber, maxDeliverableNumber int) string {
	return fmt.Sprintf("%s\t%s", NewAtomicProcessID(maxProcessNumber), NewDeliverableID(maxDeliverableNumber))
}

func ParseMaxIDs(s string) (int, int, error) {
	maxProcessNumber, maxDeliverableNumber := 0, 0
	for _, field := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	}) {
		id := NodeID(field)

		if num, err := ParseProcessID(id); err == nil {
			maxProcessNumber = max(maxProcessNumber, num)
			continue
		}
		if num, err := ParseDeliverableID(id); err == nil {
			maxDeliverableNumber = max(maxDeliverableNumber, num)
			continue
		}

		return 0, 0, fmt.Errorf("pfd.ParseMaxIDs: invalid node ID: %q", field)
	}
	return maxProcessNumber, maxDeliverableNumber, nil
}

type RenumberPlanAcc struct {
	ProcessNumberNext int

	DeliverableNumberNext int

	Base RenumberBase

	Plan RenumberPlan
}

func NewRenumberPlanAccWithBase(p *PFD, base RenumberBase) *RenumberPlanAcc {
	return &RenumberPlanAcc{
		ProcessNumberNext:     max(GetMaxProcessID(p), GetMaxProcessIDInPlan(base.Plan), base.MaxProcessNumber) + 1,
		DeliverableNumberNext: max(GetMaxDeliverableID(p), GetMaxDeliverableIDInPlan(base.Plan), base.MaxDeliverableNumber) + 1,
		Base:                  base,
		Plan:                  make(RenumberPlan),
	}
}

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

		r.Plan[nodeID] = newNode
		return
	}

	if based, ok := r.Base.Plan[nodeID]; ok {

		newNode.ID = based.ID
		newNode.Description = nodeID.Label()
		r.Plan[nodeID] = newNode
		return
	}

	newNode.ID = NewAtomicProcessID(r.ProcessNumberNext)
	r.ProcessNumberNext++

	newNode.Description = nodeID.Label()

	r.Plan[nodeID] = newNode
}

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

		r.Plan[nodeID] = newNode
		return
	}

	if based, ok := r.Base.Plan[nodeID]; ok {

		newNode.ID = based.ID
		newNode.Description = nodeID.Label()
		r.Plan[nodeID] = newNode
		return
	}

	newNode.ID = NewDeliverableID(r.DeliverableNumberNext)
	r.DeliverableNumberNext++

	newNode.Description = nodeID.Label()

	r.Plan[nodeID] = newNode
}

func NewRenumberPlan(p *PFD, graphExceptFB *graph.Graph, nodeMap map[NodeID]*Node, errs *[]Error) (RenumberPlan, bool) {
	return NewRenumberPlanWithBase(p, graphExceptFB, nodeMap, RenumberBase{}, errs)
}

func NewRenumberPlanWithBase(p *PFD, graphExceptFB *graph.Graph, nodeMap map[NodeID]*Node, base RenumberBase, errs *[]Error) (RenumberPlan, bool) {
	cycles := graphExceptFB.Cycles()
	if cycles.Len() > 0 {
		for _, cycle := range cycles.Iter() {
			locs := make([]Location, 0, len(cycle))
			for i := 0; i < len(cycle)-1; i++ {
				locs = append(locs, NewEdgeLocation(NodeID(cycle[i]), NodeID(cycle[i+1])))
			}
			*errs = append(*errs, Error{Locations: locs, Wrapped: fmt.Errorf("pfd.NewRenumberPlanWithBase: cyclic graph")})
		}
		return nil, false
	}

	maximals := graphExceptFB.Maximals()
	finals := sets.NewWithCapacity[NodeID](maximals.Len())
	for _, maximal := range maximals.Iter() {
		node, ok := nodeMap[NodeID(maximal)]
		if !ok {
			panic(fmt.Sprintf("pfd.NewRenumberPlanWithBase: missing maximal: %q", maximal))
		}
		if node.Type != NodeTypeAtomicDeliverable {
			continue
		}
		finals.Add(NodeID.Compare, node.ID)
	}

	acc := NewRenumberPlanAccWithBase(p, base)

	for _, final := range finals.Iter() {
		acc.VisitDeliverable(final, nodeMap)
	}

	order := graphExceptFB.TopologicalSort()
	orderExceptFinals := make([]NodeID, 0, len(order))
	for _, nodeID := range order {
		pfdNodeID := NodeID(nodeID)

		if finals.Contains(NodeID.Compare, pfdNodeID) {

			continue
		}

		orderExceptFinals = append(orderExceptFinals, pfdNodeID)
	}

	for _, pfdNodeID := range orderExceptFinals {
		acc.Visit(pfdNodeID, nodeMap)
	}

	nodeIDWithMinIDs := make([]*pairs.Pair[NodeID, NodeID], 0, len(p.ProcessComposition))
	for nodeID, nodeIDs := range p.ProcessComposition {

		minID, ok := nodeIDs.At(0)
		if !ok {

			minID = NewAtomicProcessID(0)
		}
		nodeIDWithMinIDs = append(nodeIDWithMinIDs, pairs.New(nodeID, minID))
	}

	slices.SortFunc(nodeIDWithMinIDs, func(a, b *pairs.Pair[NodeID, NodeID]) int {
		if c := NodeID.Compare(a.Second, b.Second); c != 0 {
			return c
		}
		return NodeID.Compare(a.First, b.First)
	})

	for _, pair := range nodeIDWithMinIDs {
		acc.VisitProcess(pair.First, nodeMap)
	}

	for nodeID := range base.Plan {
		node, ok := acc.Plan[nodeID]
		if !ok {
			continue
		}

		if _, ok := nodeMap[node.ID]; ok {
			*errs = append(*errs, Error{
				Locations: []Location{NewNodeLocation(nodeID), NewNodeLocation(node.ID)},
				Wrapped:   fmt.Errorf("pfd.NewRenumberPlanWithBase: the base plan assigns the ID %q to %q, but the ID is already used in this PFD", node.ID, nodeID),
			})
			continue
		}

		if node.Type.IsProcess() {
			if _, err := ParseProcessID(node.ID); err != nil {
				*errs = append(*errs, Error{
					Locations: []Location{NewNodeLocation(nodeID)},
					Wrapped:   fmt.Errorf("pfd.NewRenumberPlanWithBase: the base plan assigns a non-process ID %q to a process %q", node.ID, nodeID),
				})
			}
			continue
		}
		if node.Type.IsDeliverable() {
			if _, err := ParseDeliverableID(node.ID); err != nil {
				*errs = append(*errs, Error{
					Locations: []Location{NewNodeLocation(nodeID)},
					Wrapped:   fmt.Errorf("pfd.NewRenumberPlanWithBase: the base plan assigns a non-deliverable ID %q to a deliverable %q", node.ID, nodeID),
				})
			}
		}
	}

	if len(*errs) > 0 {
		return nil, false
	}
	return acc.Plan, true
}

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

func GetMaxProcessIDInPlan(plan RenumberPlan) int {
	maxNum := 0
	for _, node := range plan {
		i, err := ParseProcessID(node.ID)
		if err != nil {
			continue
		}
		if i > maxNum {
			maxNum = i
		}
	}
	return maxNum
}

func GetMaxDeliverableIDInPlan(plan RenumberPlan) int {
	maxNum := 0
	for _, node := range plan {
		i, err := ParseDeliverableID(node.ID)
		if err != nil {
			continue
		}
		if i > maxNum {
			maxNum = i
		}
	}
	return maxNum
}

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

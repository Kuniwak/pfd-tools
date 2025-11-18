package pfd

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

type AtomicDeliverableID string

func (id AtomicDeliverableID) Compare(other AtomicDeliverableID) int {
	if len(id) < len(other) {
		return -1
	}
	if len(id) > len(other) {
		return 1
	}
	return strings.Compare(string(id), string(other))
}

func AtomicDeliverableIDFromNodeID(id NodeID, nodeMap map[NodeID]*Node) AtomicDeliverableID {
	node, ok := nodeMap[id]
	if !ok {
		panic(fmt.Sprintf("pfd.AtomicDeliverableIDFromNodeID: missing node: %q", id))
	}
	if node.Type != NodeTypeAtomicDeliverable {
		panic(fmt.Sprintf("pfd.AtomicDeliverableIDFromNodeID: non-deliverable node: %q", id))
	}
	return AtomicDeliverableID(node.ID)
}

type CompositeDeliverableID string

func (id CompositeDeliverableID) Compare(other CompositeDeliverableID) int {
	if len(id) < len(other) {
		return -1
	}
	if len(id) > len(other) {
		return 1
	}
	return strings.Compare(string(id), string(other))
}

func CompositeDeliverableIDFromNodeID(id NodeID, nodeMap map[NodeID]*Node) CompositeDeliverableID {
	node, ok := nodeMap[id]
	if !ok {
		panic(fmt.Sprintf("pfd.CompositeDeliverableIDFromNodeID: missing node: %q", id))
	}
	if node.Type != NodeTypeCompositeDeliverable {
		panic(fmt.Sprintf("pfd.CompositeDeliverableIDFromNodeID: non-composite deliverable node: %q", id))
	}
	return CompositeDeliverableID(node.ID)
}

type AtomicProcessID string

func (id AtomicProcessID) Compare(other AtomicProcessID) int {
	if len(id) < len(other) {
		return -1
	}
	if len(id) > len(other) {
		return 1
	}
	return strings.Compare(string(id), string(other))
}

func AtomicProcessIDFromNodeID(id NodeID, nodeMap map[NodeID]*Node) AtomicProcessID {
	node, ok := nodeMap[id]
	if !ok {
		panic(fmt.Sprintf("pfd.AtomicProcessIDFromNodeID: missing node: %q", id))
	}
	if node.Type != NodeTypeAtomicProcess {
		panic(fmt.Sprintf("pfd.AtomicProcessIDFromNodeID: non-atomic process node: %q", id))
	}
	return AtomicProcessID(node.ID)
}

type CompositeProcessID string

func (id CompositeProcessID) Compare(other CompositeProcessID) int {
	if len(id) < len(other) {
		return -1
	}
	if len(id) > len(other) {
		return 1
	}
	return strings.Compare(string(id), string(other))
}

func CompositeProcessIDFromNodeID(id NodeID, nodeMap map[NodeID]*Node) CompositeProcessID {
	node, ok := nodeMap[id]
	if !ok {
		panic(fmt.Sprintf("pfd.CompositeProcessIDFromNodeID: missing node: %q", id))
	}
	if node.Type != NodeTypeCompositeProcess {
		panic(fmt.Sprintf("pfd.CompositeProcessIDFromNodeID: non-composite process node: %q", id))
	}
	return CompositeProcessID(node.ID)
}

type RelationTriple struct {
	Inputs         *sets.Set[AtomicDeliverableID] `json:"inputs"`
	FeedbackInputs *sets.Set[AtomicDeliverableID] `json:"feedback_inputs"`
	Outputs        *sets.Set[AtomicDeliverableID] `json:"outputs"`
}

func (r *RelationTriple) Compare(other *RelationTriple) int {
	c := sets.Compare(AtomicDeliverableID.Compare)(r.Inputs, other.Inputs)
	if c != 0 {
		return c
	}
	c = sets.Compare(AtomicDeliverableID.Compare)(r.Outputs, other.Outputs)
	if c != 0 {
		return c
	}
	return sets.Compare(AtomicDeliverableID.Compare)(r.FeedbackInputs, other.FeedbackInputs)
}

type ValidPFD struct {
	AtomicProcesses                    *sets.Set[AtomicProcessID]                                `json:"atomic_processes"`
	AtomicProcessDescriptionMap        map[AtomicProcessID]string                                `json:"atomic_process_description_map"`
	AtomicDeliverables                 *sets.Set[AtomicDeliverableID]                            `json:"atomic_deliverables"`
	AtomicDeliverableDescriptionMap    map[AtomicDeliverableID]string                            `json:"atomic_deliverable_description_map"`
	Relations                          map[AtomicProcessID]*RelationTriple                       `json:"relations"`
	ProcessComposition                 map[CompositeProcessID]*sets.Set[AtomicProcessID]         `json:"process_composition"`
	CompositeProcessDescriptionMap     map[CompositeProcessID]string                             `json:"composite_process_description_map"`
	DeliverableComposition             map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID] `json:"deliverable_composition"`
	CompositeDeliverableDescriptionMap map[CompositeDeliverableID]string                         `json:"composite_deliverable_description_map"`

	Memoized *SafePFDMemoized `json:"-"`
}

type SafePFDMemoized struct {
	SourceAtomicProcess                   map[AtomicDeliverableID]AtomicProcessID
	NotFeedbackDestinationAtomicProcesses map[AtomicDeliverableID]*sets.Set[AtomicProcessID]
	FeedbackDestinationAtomicProcesses    map[AtomicDeliverableID]*sets.Set[AtomicProcessID]
	EitherDestinationAtomicProcesses      map[AtomicDeliverableID]*sets.Set[AtomicProcessID]
	InputsIncludingFeedback               map[AtomicProcessID]*sets.Set[AtomicDeliverableID]
	InitialDeliverables                   *sets.Set[AtomicDeliverableID]
	FeedbackSourceDeliverables            *sets.Set[AtomicDeliverableID]
}

func (m *SafePFDMemoized) Clone() *SafePFDMemoized {
	return &SafePFDMemoized{
		SourceAtomicProcess:                   maps.Clone(m.SourceAtomicProcess),
		NotFeedbackDestinationAtomicProcesses: maps.Clone(m.NotFeedbackDestinationAtomicProcesses),
		FeedbackDestinationAtomicProcesses:    maps.Clone(m.FeedbackDestinationAtomicProcesses),
		EitherDestinationAtomicProcesses:      maps.Clone(m.EitherDestinationAtomicProcesses),
		InputsIncludingFeedback:               maps.Clone(m.InputsIncludingFeedback),
		InitialDeliverables:                   m.InitialDeliverables.Clone(),
		FeedbackSourceDeliverables:            m.FeedbackSourceDeliverables.Clone(),
	}
}

func NewSafePFD(
	atomicProcesses map[AtomicProcessID]string,
	deliverables map[AtomicDeliverableID]string,
	relations map[AtomicProcessID]*RelationTriple,
	processComposition map[CompositeProcessID]*pairs.Pair[string, *sets.Set[AtomicProcessID]],
	deliverableComposition map[CompositeDeliverableID]*pairs.Pair[string, *sets.Set[AtomicDeliverableID]],
) *ValidPFD {
	apDescMap := make(map[AtomicProcessID]string)
	aps := sets.New(AtomicProcessID.Compare)

	inputsIncludingFeedback := make(map[AtomicProcessID]*sets.Set[AtomicDeliverableID])
	for ap, desc := range atomicProcesses {
		inputsIncludingFeedback[ap] = sets.New(AtomicDeliverableID.Compare)
		apDescMap[ap] = desc
		aps.Add(AtomicProcessID.Compare, ap)
	}

	ds := sets.New(AtomicDeliverableID.Compare)
	dDescMap := make(map[AtomicDeliverableID]string)
	notFeedbackDestinationAtomicProcesses := make(map[AtomicDeliverableID]*sets.Set[AtomicProcessID])
	feedbackDestinationAtomicProcesses := make(map[AtomicDeliverableID]*sets.Set[AtomicProcessID])
	eitherDestinationAtomicProcesses := make(map[AtomicDeliverableID]*sets.Set[AtomicProcessID])
	for d, desc := range deliverables {
		notFeedbackDestinationAtomicProcesses[d] = sets.New(AtomicProcessID.Compare)
		feedbackDestinationAtomicProcesses[d] = sets.New(AtomicProcessID.Compare)
		eitherDestinationAtomicProcesses[d] = sets.New(AtomicProcessID.Compare)
		dDescMap[d] = desc
		ds.Add(AtomicDeliverableID.Compare, d)
	}

	ids := ds.Clone()
	for _, rel := range relations {
		ids.Difference(AtomicDeliverableID.Compare, rel.Outputs)
	}

	sourceAtomicProcess := make(map[AtomicDeliverableID]AtomicProcessID)
	for ap, rel := range relations {
		for _, d := range rel.Inputs.Iter() {
			notFeedbackDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
			eitherDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
			inputsIncludingFeedback[ap].Add(AtomicDeliverableID.Compare, d)
		}
		for _, d := range rel.FeedbackInputs.Iter() {
			feedbackDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
			eitherDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
			inputsIncludingFeedback[ap].Add(AtomicDeliverableID.Compare, d)
		}
		for _, d := range rel.Outputs.Iter() {
			if ap2, ok := sourceAtomicProcess[d]; ok {
				if ap2 != ap {
					panic(fmt.Sprintf("pfd.NewSafePFD: deliverable: %q is connected to multiple atomic processes: %q and %q", d, ap, ap2))
				}
			} else {
				sourceAtomicProcess[d] = ap
			}
		}
	}

	pComp := make(map[CompositeProcessID]*sets.Set[AtomicProcessID])
	cpDescMap := make(map[CompositeProcessID]string)
	cps := sets.New(CompositeProcessID.Compare)
	for cp, pair := range processComposition {
		cps.Add(CompositeProcessID.Compare, cp)
		cpDescMap[cp] = pair.First
		pComp[cp] = pair.Second
	}

	dComp := make(map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID])
	cdDescMap := make(map[CompositeDeliverableID]string)
	cds := sets.New(CompositeDeliverableID.Compare)
	for cd, pair := range deliverableComposition {
		cds.Add(CompositeDeliverableID.Compare, cd)
		cdDescMap[cd] = pair.First
		dComp[cd] = pair.Second
	}

	feedbackSourceDeliverables := sets.New(AtomicDeliverableID.Compare)
	for _, d := range ds.Iter() {
		if feedbackDestinationAtomicProcesses[d].Len() == 0 {
			continue
		}
		feedbackSourceDeliverables.Add(AtomicDeliverableID.Compare, d)
	}

	return &ValidPFD{
		AtomicProcesses:                    aps,
		AtomicProcessDescriptionMap:        apDescMap,
		AtomicDeliverables:                 ds,
		AtomicDeliverableDescriptionMap:    dDescMap,
		Relations:                          relations,
		ProcessComposition:                 pComp,
		CompositeProcessDescriptionMap:     cpDescMap,
		DeliverableComposition:             dComp,
		CompositeDeliverableDescriptionMap: cdDescMap,
		Memoized: &SafePFDMemoized{
			SourceAtomicProcess:                   sourceAtomicProcess,
			NotFeedbackDestinationAtomicProcesses: notFeedbackDestinationAtomicProcesses,
			FeedbackDestinationAtomicProcesses:    feedbackDestinationAtomicProcesses,
			EitherDestinationAtomicProcesses:      eitherDestinationAtomicProcesses,
			InputsIncludingFeedback:               inputsIncludingFeedback,
			InitialDeliverables:                   ids,
			FeedbackSourceDeliverables:            feedbackSourceDeliverables,
		},
	}
}

func NewSafePFDByUnsafePFD(p *PFD) (*ValidPFD, error) {
	logger := slog.New(slog.DiscardHandler)
	nodeMap := NewNodeMap(p.Nodes, logger)

	aps := sets.NewWithCapacity[AtomicProcessID](p.Nodes.Len())
	apDescMap := make(map[AtomicProcessID]string)
	ds := sets.NewWithCapacity[AtomicDeliverableID](p.Nodes.Len())
	dDescMap := make(map[AtomicDeliverableID]string)
	cps := sets.NewWithCapacity[CompositeProcessID](p.Nodes.Len())
	cds := sets.NewWithCapacity[CompositeDeliverableID](p.Nodes.Len())
	cpDescMap := make(map[CompositeProcessID]string)
	cdDescMap := make(map[CompositeDeliverableID]string)
	relations := make(map[AtomicProcessID]*RelationTriple)

	memoized := &SafePFDMemoized{
		SourceAtomicProcess:                   make(map[AtomicDeliverableID]AtomicProcessID),
		NotFeedbackDestinationAtomicProcesses: make(map[AtomicDeliverableID]*sets.Set[AtomicProcessID]),
		FeedbackDestinationAtomicProcesses:    make(map[AtomicDeliverableID]*sets.Set[AtomicProcessID]),
		EitherDestinationAtomicProcesses:      make(map[AtomicDeliverableID]*sets.Set[AtomicProcessID]),
		InputsIncludingFeedback:               make(map[AtomicProcessID]*sets.Set[AtomicDeliverableID]),
		FeedbackSourceDeliverables:            sets.New(AtomicDeliverableID.Compare),
	}

	for _, node := range p.Nodes.Iter() {
		switch node.Type {
		case NodeTypeAtomicProcess:
			ap := AtomicProcessIDFromNodeID(node.ID, nodeMap)
			aps.Add(AtomicProcessID.Compare, ap)
			apDescMap[ap] = node.Description

			relations[ap] = &RelationTriple{
				Inputs:         sets.New(AtomicDeliverableID.Compare),
				FeedbackInputs: sets.New(AtomicDeliverableID.Compare),
				Outputs:        sets.New(AtomicDeliverableID.Compare),
			}
			memoized.InputsIncludingFeedback[ap] = sets.New(AtomicDeliverableID.Compare)

		case NodeTypeAtomicDeliverable:
			d := AtomicDeliverableIDFromNodeID(node.ID, nodeMap)

			ds.Add(AtomicDeliverableID.Compare, d)
			dDescMap[d] = node.Description
			memoized.NotFeedbackDestinationAtomicProcesses[d] = sets.New(AtomicProcessID.Compare)
			memoized.FeedbackDestinationAtomicProcesses[d] = sets.New(AtomicProcessID.Compare)
			memoized.EitherDestinationAtomicProcesses[d] = sets.New(AtomicProcessID.Compare)

		case NodeTypeCompositeProcess:
			cp := CompositeProcessIDFromNodeID(node.ID, nodeMap)
			cps.Add(CompositeProcessID.Compare, cp)
			cpDescMap[cp] = node.Description

		case NodeTypeCompositeDeliverable:
			cd := CompositeDeliverableIDFromNodeID(node.ID, nodeMap)
			cds.Add(CompositeDeliverableID.Compare, cd)
			cdDescMap[cd] = node.Description

		default:
			panic(fmt.Sprintf("pfd.NewSafePFD: unknown node type: %q", node.Type))
		}
	}

	pComp := make(map[CompositeProcessID]*sets.Set[AtomicProcessID])
	for n, ns := range p.ProcessComposition {
		node, ok := nodeMap[n]
		if !ok {
			return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", n)
		}
		if node.Type != NodeTypeCompositeProcess {
			return nil, fmt.Errorf("pfd.NewSafePFD: non-composite process node: %q", n)
		}

		cp := CompositeProcessIDFromNodeID(n, nodeMap)
		aps := sets.NewWithCapacity[AtomicProcessID](ns.Len())
		for _, n := range ns.Iter() {
			node, ok := nodeMap[n]
			if !ok {
				return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", n)
			}
			if node.Type != NodeTypeAtomicProcess {
				return nil, fmt.Errorf("pfd.NewSafePFD: non-atomic process node: %q", n)
			}
			ap := AtomicProcessIDFromNodeID(n, nodeMap)
			aps.Add(AtomicProcessID.Compare, ap)
		}
		pComp[cp] = aps
		cpDescMap[cp] = node.Description
	}

	dComp := make(map[CompositeDeliverableID]*sets.Set[AtomicDeliverableID])
	for n, ns := range p.DeliverableComposition {
		node, ok := nodeMap[n]
		if !ok {
			return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", n)
		}
		if node.Type != NodeTypeCompositeDeliverable {
			return nil, fmt.Errorf("pfd.NewSafePFD: non-composite deliverable node: %q", n)
		}

		cd := CompositeDeliverableIDFromNodeID(n, nodeMap)
		ads := sets.NewWithCapacity[AtomicDeliverableID](ns.Len())
		for _, n := range ns.Iter() {
			node, ok := nodeMap[n]
			if !ok {
				return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", n)
			}
			if node.Type != NodeTypeAtomicDeliverable {
				return nil, fmt.Errorf("pfd.NewSafePFD: non-atomic deliverable node: %q", n)
			}
			ad := AtomicDeliverableIDFromNodeID(n, nodeMap)
			ads.Add(AtomicDeliverableID.Compare, ad)
		}
		dComp[cd] = ads
		cdDescMap[cd] = node.Description
	}

	for _, edge := range p.Edges.Iter() {
		node1, ok := nodeMap[edge.Source]
		if !ok {
			return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", edge.Source)
		}
		node2, ok := nodeMap[edge.Target]
		if !ok {
			return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", edge.Target)
		}

		switch node1.Type {
		case NodeTypeCompositeProcess:
			if !node2.Type.IsDeliverable() {
				return nil, fmt.Errorf("pfd.NewSafePFD: composite process node: %q is connected to non-deliverable node: %q", edge.Target, edge.Source)
			}
			continue

		case NodeTypeAtomicProcess:
			ap := AtomicProcessIDFromNodeID(node1.ID, nodeMap)
			switch node2.Type {
			case NodeTypeAtomicDeliverable:
				d := AtomicDeliverableIDFromNodeID(node2.ID, nodeMap)
				if edge.IsFeedback {
					return nil, fmt.Errorf("pfd.NewSafePFD: deliverable: %q is connected to atomic process: %q with feedback edge", d, ap)
				}

				relations[ap].Outputs.Add(AtomicDeliverableID.Compare, d)

				if ap2, ok := memoized.SourceAtomicProcess[d]; ok {
					if ap2 != ap {
						return nil, fmt.Errorf("pfd.NewSafePFD: deliverable: %q is connected to multiple atomic processes: %q and %q", d, ap, ap2)
					}
				} else {
					memoized.SourceAtomicProcess[d] = ap
				}

			case NodeTypeAtomicProcess:
				return nil, fmt.Errorf("pfd.NewSafePFD: atomic process node: %q is connected to another atomic process: %q", edge.Target, edge.Source)

			case NodeTypeCompositeProcess:
				return nil, fmt.Errorf("pfd.NewSafePFD: atomic process node: %q is connected to composite process: %q", edge.Target, edge.Source)

			case NodeTypeCompositeDeliverable:
				return nil, fmt.Errorf("pfd.NewSafePFD: atomic process node: %q is connected to composite deliverable: %q", edge.Target, edge.Source)

			default:
				panic(fmt.Sprintf("pfd.NewSafePFD: unknown node type: %q", node1.Type))
			}

		case NodeTypeAtomicDeliverable:
			d := AtomicDeliverableIDFromNodeID(node1.ID, nodeMap)
			switch node2.Type {
			case NodeTypeCompositeProcess:
				// Do nothing.
				continue

			case NodeTypeAtomicProcess:
				ap := AtomicProcessIDFromNodeID(node2.ID, nodeMap)

				if edge.IsFeedback {
					relations[ap].FeedbackInputs.Add(AtomicDeliverableID.Compare, d)
					memoized.FeedbackDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
					memoized.EitherDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
					memoized.InputsIncludingFeedback[ap].Add(AtomicDeliverableID.Compare, d)
					memoized.FeedbackSourceDeliverables.Add(AtomicDeliverableID.Compare, d)
				} else {
					relations[ap].Inputs.Add(AtomicDeliverableID.Compare, d)
					memoized.NotFeedbackDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
					memoized.EitherDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
					memoized.InputsIncludingFeedback[ap].Add(AtomicDeliverableID.Compare, d)
				}

			case NodeTypeAtomicDeliverable, NodeTypeCompositeDeliverable:
				return nil, fmt.Errorf("pfd.NewSafePFD: deliverable node: %q is connected to another deliverable: %q", edge.Target, edge.Source)

			default:
				return nil, fmt.Errorf("pfd.NewSafePFD: unknown node type: %q", node2.Type)
			}

		case NodeTypeCompositeDeliverable:
			cd := CompositeDeliverableIDFromNodeID(node1.ID, nodeMap)
			switch node2.Type {
			case NodeTypeAtomicProcess:
				ap := AtomicProcessIDFromNodeID(node2.ID, nodeMap)

				ds, ok := dComp[cd]
				if !ok {
					panic(fmt.Sprintf("pfd.NewSafePFD: missing composite deliverable: %q", cd))
				}

				if edge.IsFeedback {
					for _, d := range ds.Iter() {
						relations[ap].FeedbackInputs.Add(AtomicDeliverableID.Compare, d)
						memoized.FeedbackDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
						memoized.EitherDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
						memoized.InputsIncludingFeedback[ap].Add(AtomicDeliverableID.Compare, d)
						memoized.FeedbackSourceDeliverables.Add(AtomicDeliverableID.Compare, d)
					}
				} else {
					for _, d := range ds.Iter() {
						relations[ap].Inputs.Add(AtomicDeliverableID.Compare, d)
						memoized.NotFeedbackDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
						memoized.EitherDestinationAtomicProcesses[d].Add(AtomicProcessID.Compare, ap)
						memoized.InputsIncludingFeedback[ap].Add(AtomicDeliverableID.Compare, d)
					}
				}

			case NodeTypeAtomicDeliverable, NodeTypeCompositeDeliverable:
				return nil, fmt.Errorf("pfd.NewSafePFD: composite deliverable node: %q is connected to non-process node: %q", edge.Target, edge.Source)

			case NodeTypeCompositeProcess:
				// Do nothing.
				continue
			}

		default:
			return nil, fmt.Errorf("pfd.NewSafePFD: unknown node type: %q", node1.Type)
		}
	}

	for n1, n2s := range p.ProcessComposition {
		node1, ok := nodeMap[n1]
		if !ok {
			return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", n1)
		}
		if node1.Type != NodeTypeCompositeProcess {
			return nil, fmt.Errorf("pfd.NewSafePFD: non-composite process node: %q", n1)
		}

		cp := CompositeProcessIDFromNodeID(n1, nodeMap)
		for _, n2 := range n2s.Iter() {
			node2, ok := nodeMap[n2]
			if !ok {
				return nil, fmt.Errorf("pfd.NewSafePFD: missing node: %q", n2)
			}
			if node2.Type != NodeTypeAtomicProcess {
				return nil, fmt.Errorf("pfd.NewSafePFD: non-atomic process node: %q", n2)
			}

			ap := AtomicProcessIDFromNodeID(n2, nodeMap)
			pComp[cp].Add(AtomicProcessID.Compare, ap)
		}
	}

	initialDeliverables := sets.NewWithCapacity[AtomicDeliverableID](ds.Len())
	for _, d := range p.InitialAtomicDeliverables(nodeMap, logger).Iter() {
		initialDeliverables.Add(AtomicDeliverableID.Compare, AtomicDeliverableIDFromNodeID(d, nodeMap))
	}
	memoized.InitialDeliverables = initialDeliverables

	return &ValidPFD{
		AtomicProcesses:                    aps,
		AtomicProcessDescriptionMap:        apDescMap,
		AtomicDeliverables:                 ds,
		AtomicDeliverableDescriptionMap:    dDescMap,
		Relations:                          relations,
		ProcessComposition:                 pComp,
		CompositeProcessDescriptionMap:     cpDescMap,
		DeliverableComposition:             dComp,
		CompositeDeliverableDescriptionMap: cdDescMap,
		Memoized:                           memoized,
	}, nil
}

func (s *ValidPFD) Clone() *ValidPFD {
	return &ValidPFD{
		AtomicProcesses:                    s.AtomicProcesses.Clone(),
		AtomicProcessDescriptionMap:        maps.Clone(s.AtomicProcessDescriptionMap),
		AtomicDeliverables:                 s.AtomicDeliverables.Clone(),
		AtomicDeliverableDescriptionMap:    maps.Clone(s.AtomicDeliverableDescriptionMap),
		Relations:                          maps.Clone(s.Relations),
		ProcessComposition:                 maps.Clone(s.ProcessComposition),
		CompositeProcessDescriptionMap:     maps.Clone(s.CompositeProcessDescriptionMap),
		DeliverableComposition:             maps.Clone(s.DeliverableComposition),
		CompositeDeliverableDescriptionMap: maps.Clone(s.CompositeDeliverableDescriptionMap),
		Memoized:                           s.Memoized.Clone(),
	}
}

func (s *ValidPFD) Compare(other *ValidPFD) int {
	c := sets.Compare(AtomicProcessID.Compare)(s.AtomicProcesses, other.AtomicProcesses)
	if c != 0 {
		return c
	}
	c = sets.Compare(AtomicDeliverableID.Compare)(s.AtomicDeliverables, other.AtomicDeliverables)
	if c != 0 {
		return c
	}
	return cmp2.CompareMap(s.Relations, other.Relations, AtomicProcessID.Compare, (*RelationTriple).Compare)
}

func (s *ValidPFD) InputDeliverablesIncludingFeedback(ap AtomicProcessID) *sets.Set[AtomicDeliverableID] {
	inputs, ok := s.Memoized.InputsIncludingFeedback[ap]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.InputDeliverablesIncludingFeedback: missing inputs including feedback for atomic process: %q", ap))
	}
	return inputs
}

func (s *ValidPFD) InputDeliverablesExceptFeedback(ap AtomicProcessID) *sets.Set[AtomicDeliverableID] {
	rel, ok := s.Relations[ap]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.InputDeliverablesExceptFeedback: missing inputs including feedback for atomic process: %q", ap))
	}
	return rel.Inputs
}

func (s *ValidPFD) InputDeliverablesOnlyFeedback(ap AtomicProcessID) *sets.Set[AtomicDeliverableID] {
	rel, ok := s.Relations[ap]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.InputDeliverablesOnlyFeedback: missing inputs including feedback for atomic process: %q", ap))
	}
	return rel.FeedbackInputs
}

func (s *ValidPFD) OutputDeliverables(ap AtomicProcessID) *sets.Set[AtomicDeliverableID] {
	rel, ok := s.Relations[ap]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.OutputDeliverables: missing outputs for atomic process: %q", ap))
	}
	return rel.Outputs
}

func (s *ValidPFD) SourceAtomicProcess(d AtomicDeliverableID) (AtomicProcessID, bool) {
	ap, ok := s.Memoized.SourceAtomicProcess[d]
	return ap, ok
}

func (s *ValidPFD) EitherFeedbackOrNotDestinationAtomicProcesses(d AtomicDeliverableID) *sets.Set[AtomicProcessID] {
	aps, ok := s.Memoized.EitherDestinationAtomicProcesses[d]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.EitherFeedbackOrNotDestinationAtomicProcesses: missing either feedback or not destination atomic processes for deliverable: %q", d))
	}
	return aps
}

func (s *ValidPFD) FeedbackDestinationAtomicProcesses(d AtomicDeliverableID) *sets.Set[AtomicProcessID] {
	aps, ok := s.Memoized.FeedbackDestinationAtomicProcesses[d]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.FeedbackDestinationAtomicProcesses: missing feedback or not destination atomic processes for deliverable %q", d))
	}
	return aps
}

func (s *ValidPFD) NotFeedbackDestinationAtomicProcesses(d AtomicDeliverableID) *sets.Set[AtomicProcessID] {
	aps, ok := s.Memoized.NotFeedbackDestinationAtomicProcesses[d]
	if !ok {
		panic(fmt.Sprintf("pfd.ValidPFD.NotFeedbackDestinationAtomicProcesses: missing destination atomic processes for deliverable: %q", d))
	}
	return aps
}

func (s *ValidPFD) FeedbackSourceDeliverables() *sets.Set[AtomicDeliverableID] {
	return s.Memoized.FeedbackSourceDeliverables
}

func (s *ValidPFD) CollectReachableAtomicProcessesExceptFeedback(ap AtomicProcessID, res *sets.Set[AtomicProcessID], logger *slog.Logger) {
	res.Add(AtomicProcessID.Compare, ap)

	for _, d := range s.OutputDeliverables(ap).Iter() {
		for _, ap := range s.NotFeedbackDestinationAtomicProcesses(d).Iter() {
			if res.Contains(AtomicProcessID.Compare, ap) {
				continue
			}
			res.Add(AtomicProcessID.Compare, ap)
			s.CollectReachableAtomicProcessesExceptFeedback(ap, res, logger)
		}
	}
}

func (s *ValidPFD) CollectReachableAtomicProcessesIncludingFeedback(ap AtomicProcessID, res *sets.Set[AtomicProcessID], logger *slog.Logger) {
	res.Add(AtomicProcessID.Compare, ap)

	for _, d := range s.OutputDeliverables(ap).Iter() {
		for _, ap := range s.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
			if res.Contains(AtomicProcessID.Compare, ap) {
				continue
			}
			res.Add(AtomicProcessID.Compare, ap)
			s.CollectReachableAtomicProcessesIncludingFeedback(ap, res, logger)
		}
	}
}

func (s *ValidPFD) CollectReachableDeliverablesExceptFeedback(ap AtomicProcessID, res *sets.Set[AtomicDeliverableID], logger *slog.Logger) {
	for _, d := range s.OutputDeliverables(ap).Iter() {
		if res.Contains(AtomicDeliverableID.Compare, d) {
			continue
		}

		res.Add(AtomicDeliverableID.Compare, d)

		for _, ap2 := range s.NotFeedbackDestinationAtomicProcesses(d).Iter() {
			s.CollectReachableDeliverablesExceptFeedback(ap2, res, logger)
		}
	}
}

func (s *ValidPFD) CollectBackwardReachableDeliverablesExceptFeedback(ap AtomicProcessID, res *sets.Set[AtomicDeliverableID], logger *slog.Logger) {
	for _, d := range s.InputDeliverablesExceptFeedback(ap).Iter() {
		if res.Contains(AtomicDeliverableID.Compare, d) {
			continue
		}

		res.Add(AtomicDeliverableID.Compare, d)

		ap2, ok := s.SourceAtomicProcess(d)
		if !ok {
			continue
		}
		s.CollectBackwardReachableDeliverablesExceptFeedback(ap2, res, logger)
	}
}

func (s *ValidPFD) CollectBackwardReachableAtomicProcessesExceptFeedback(ap AtomicProcessID, res *sets.Set[AtomicProcessID], logger *slog.Logger) {
	res.Add(AtomicProcessID.Compare, ap)

	for _, d := range s.InputDeliverablesExceptFeedback(ap).Iter() {
		ap2, ok := s.SourceAtomicProcess(d)
		if !ok {
			continue
		}
		s.CollectBackwardReachableAtomicProcessesExceptFeedback(ap2, res, logger)
	}
}

func (s *ValidPFD) CollectPaths(src AtomicProcessID, dst AtomicProcessID, res *sets.Set[[]AtomicProcessID], logger *slog.Logger) {
	s.collectPaths(src, dst, []AtomicProcessID{}, res, logger)
}

func (s *ValidPFD) collectPaths(src AtomicProcessID, dst AtomicProcessID, path []AtomicProcessID, res *sets.Set[[]AtomicProcessID], logger *slog.Logger) {
	if src == dst {
		newPath := slices.Clone(append(path, dst))
		res.Add(cmp2.CompareSlice[[]AtomicProcessID](AtomicProcessID.Compare), newPath)
		return
	}

	for _, d := range s.OutputDeliverables(src).Iter() {
		for _, ap := range s.NotFeedbackDestinationAtomicProcesses(d).Iter() {
			s.collectPaths(ap, dst, append(slices.Clone(path), src), res, logger)
		}
	}
}

func (s *ValidPFD) InitialDeliverables() *sets.Set[AtomicDeliverableID] {
	return s.Memoized.InitialDeliverables
}

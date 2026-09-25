package pfd

import (
	"github.com/Kuniwak/pfd-tools/sets"
)

var PresetsAll = map[string]*PFD{
	"smallest":                                PresetSmallest,
	"sequential":                              PresetSequential,
	"counterclockwise_rotated_y_shape":        PresetCounterclockwiseRotatedYShape,
	"clockwise_rotated_y_shape":               PresetClockwiseRotatedYShape,
	"bigger_counterclockwise_rotated_y_shape": PresetBiggerCounterclockwiseRotatedYShape,
	"bigger_clockwise_rotated_y_shape":        PresetBiggerClockwiseRotatedYShape,
	"smallest_loop":                           PresetSmallestLoop,
	"nested_loop":                             PresetNestedLoop,
	"butterfly_loop":                          PresetButterflyLoop,
	"cross_loop":                              PresetCrossLoop,
	"composite_deliverable":                   PresetCompositeDeliverable,
}

var PresetSmallest = &PFD{
	Title: "Smallest",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "P1", Target: "D2"},
	),
}

var PresetSequential = &PFD{
	Title: "Sequential",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P2", Target: "D3"},
	),
}

var PresetCounterclockwiseRotatedYShape = &PFD{
	Title: "CounterclockwiseRotatedYShape",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P1"},
		&Edge{Source: "P1", Target: "D3"},
	),
}

var PresetClockwiseRotatedYShape = &PFD{
	Title: "ClockwiseRotatedYShape",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P1", Target: "D3"},
	),
}

var PresetBiggerCounterclockwiseRotatedYShape = &PFD{
	Title: "BiggerCounterclockwiseRotatedYShape",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D5", Description: "D5", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "D3", Target: "P3"},
		&Edge{Source: "D4", Target: "P3"},
		&Edge{Source: "P1", Target: "D3"},
		&Edge{Source: "P2", Target: "D4"},
		&Edge{Source: "P3", Target: "D5"},
	),
}

var PresetBiggerClockwiseRotatedYShape = &PFD{
	Title: "BiggerClockwiseRotatedYShape",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "D2", Target: "P3"},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P2", Target: "D3"},
		&Edge{Source: "P3", Target: "D4"},
	),
}

var PresetSmallestLoop = &PFD{
	Title: "SmallestLoop",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "D2", Target: "P1", IsFeedback: true},
	),
}

var PresetNestedLoop = &PFD{
	Title: "NestedLoop",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "D3", Target: "P3"},
		&Edge{Source: "D3", Target: "P2", IsFeedback: true},
		&Edge{Source: "D4", Target: "P1", IsFeedback: true},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P2", Target: "D3"},
		&Edge{Source: "P3", Target: "D4"},
	),
}

var PresetButterflyLoop = &PFD{
	Title: "ButterflyLoop",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D5", Description: "D5", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "D3", Target: "P3"},
		&Edge{Source: "D4", Target: "P1", IsFeedback: true},
		&Edge{Source: "D5", Target: "P1", IsFeedback: true},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P1", Target: "D3"},
		&Edge{Source: "P2", Target: "D4"},
		&Edge{Source: "P3", Target: "D5"},
	),
}

var PresetCrossLoop = &PFD{
	Title: "CrossLoop",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "D3", Target: "P1", IsFeedback: true},
		&Edge{Source: "D3", Target: "P3"},
		&Edge{Source: "D4", Target: "P2", IsFeedback: true},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P2", Target: "D3"},
		&Edge{Source: "P3", Target: "D4"},
	),
}

var PresetWaitLoopEnd = &PFD{
	Title: "WaitLoopEnd",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D2", Target: "P2"},
		&Edge{Source: "D2", Target: "P1", IsFeedback: true},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P2", Target: "D3"},
	),
}

var PresetTripleBranch = &PFD{
	Title: "TripleBranch",
	Nodes: sets.New(
		(*Node).Compare,
		&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "D4", Description: "D4", Type: NodeTypeAtomicDeliverable},
		&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
		&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		&Node{ID: "P3", Description: "P3", Type: NodeTypeAtomicProcess},
	),
	Edges: sets.New(
		(*Edge).Compare,
		&Edge{Source: "D1", Target: "P1"},
		&Edge{Source: "D1", Target: "P2"},
		&Edge{Source: "D1", Target: "P3"},
		&Edge{Source: "P1", Target: "D2"},
		&Edge{Source: "P2", Target: "D3"},
		&Edge{Source: "P3", Target: "D4"},
	),
}

var PresetCompositeDeliverable = MustSetExpandedDeliverableComposition(
	&PFD{
		Title: "Composite deliverable",
		Nodes: sets.New(
			(*Node).Compare,
			&Node{ID: "D1", Description: "D1", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "D2", Description: "D2", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "D3", Description: "D3", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "D4", Description: "D4", Type: NodeTypeCompositeDeliverable},
			&Node{ID: "D5", Description: "D5", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "P1", Description: "P1", Type: NodeTypeAtomicProcess},
			&Node{ID: "P2", Description: "P2", Type: NodeTypeAtomicProcess},
		),
		Edges: sets.New(
			(*Edge).Compare,
			&Edge{Source: "D1", Target: "P1"},
			&Edge{Source: "P1", Target: "D2"},
			&Edge{Source: "P1", Target: "D3"},
			&Edge{Source: "D4", Target: "P2"},
			&Edge{Source: "P2", Target: "D5"},
		),
	},
	map[NodeID]*sets.Set[NodeID]{
		"D4": sets.New(NodeID.Compare, "D2", "D3"),
	},
)

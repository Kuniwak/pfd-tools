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
}

// PresetSmallest is a PFD like:
//
//	[D1] -> (P1) -> [D2]
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

// PresetSequential is a PFD like:
//
//	[D1] -> (P1) -> [D2] -> (P2) -> [D3]
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

// PresetCounterclockwiseRotatedYShape is a PFD like:
//
//	[D1] -> (P1) -> [D3]
//	       ^
//	      /
//	[D2] +
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

// PresetClockwiseRotatedYShape is a PFD like:
//
//	[D1] -> (P1) -> [D2]
//	           \
//	            +-> [D3]
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

// PresetBiggerCounterclockwiseRotatedYShape is a PFD like:
//
//	[D1] -> (P1) -> [D3] -> (P3) -> [D5]
//	                        ^
//	                       /
//	[D2] -> (P2) -> [D4] -+
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

// PresetBiggerClockwiseRotatedYShape is a PFD like:
//
//	[D1] -> (P1) -> [D2] -> (P2) -> [D3]
//	                   \
//	                    +-> (P3) -> [D4]
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

// PresetSmallestLoop is a PFD like:
//
//	.
//	            - - -
//	           V     \
//	[D1] -> (P1) -> [D2]
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

// PresetNestedLoop is a PFD like:
//
//		[D1] -> (P1) -> [D2] -> (P2) -> [D3] -> (P3) -> [D4]
//	            ^               ^       /              /
//	             \               - - - -              /
//	              - - - - - - - - - - - - - - - - - -
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

// PresetButterflyLoop is a PFD like:
//
//	.         - - - - - - - - - - - - -
//	         /                          \
//	        /    +--> [D2] -> (P2) -> [D4]
//	       V    /
//	[D1] -> (P1)
//	       ^    \
//	        \    +--> [D3] -> (P3) -> [D5]
//	         \                          /
//	          - - - - - - - - - - - - -
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

// PresetCrossLoop is a PFD like:
//
//	.                          + - - - - - - - +
//	         - - - - - - - - - - +              \
//	        V                V    \              \
//	[D1] -> (P1) -> [D2] -> (P2) -> [D3] -> (P3) -> [D4]
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

// PresetWaitLoopEnd is a PFD like:
//
//	 .             - - -
//	              V     \
//		[D1] -> (P1) -> [D2] -> (P2) -> [D3]
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

// PresetTripleBranch is a PFD like:
//
//	[D1] ---> (P1) ---> [D2]
//	    \
//	     +---> (P2) ---> [D3]
//	      \
//	       +---> (P3) ---> [D4]
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

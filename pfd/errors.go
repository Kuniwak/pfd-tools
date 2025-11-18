package pfd

import (
	"strings"
)

type Error struct {
	Locations []Location
	Wrapped   error
}

// Location represents the location where an error occurred.
type Location struct {
	// IsNode is true if the error is related to a node element. false otherwise.
	IsNode bool
	// If IsNode is true, this is the ID of the node element, otherwise undefined.
	NodeID NodeID

	// If IsNode is false, the error is related to an edge. true otherwise.
	// If IsNode is false, this is the ID of the source node of the edge (including feedback edges), otherwise undefined.
	EdgeSourceID NodeID
	// If IsNode is false, this is the ID of the target node of the edge (including feedback edges), otherwise undefined.
	EdgeTargetID NodeID
}

func NewNodeLocation(nodeID NodeID) Location {
	return Location{NodeID: nodeID, IsNode: true}
}

func NewEdgeLocation(edgeSourceID, edgeTargetID NodeID) Location {
	return Location{EdgeSourceID: edgeSourceID, EdgeTargetID: edgeTargetID, IsNode: false}
}

// CompareLocation compares Locations. The ordering satisfies the following conditions:
// 1. Node-related errors are smaller than edge-related errors
// 2. If both are node-related errors, the error with the smaller node ID is smaller
// 3. If both are edge-related errors and the edge sources are different, the error with the smaller source ID is smaller
// 4. If both are edge-related errors and the edge sources are the same, the error with the smaller target ID is smaller
func CompareLocation(a, b Location) int {
	if a.IsNode {
		if !b.IsNode {
			return -1
		}
		return NodeID.Compare(a.NodeID, b.NodeID)
	}

	if b.IsNode {
		return 1
	}
	c := NodeID.Compare(a.EdgeSourceID, b.EdgeSourceID)
	if c != 0 {
		return c
	}
	return NodeID.Compare(a.EdgeTargetID, b.EdgeTargetID)
}

func (l Location) Write(sb *strings.Builder) {
	if l.IsNode {
		sb.WriteString(string(l.NodeID))
		return
	}
	sb.WriteString(string(l.EdgeSourceID))
	sb.WriteString(" -> ")
	sb.WriteString(string(l.EdgeTargetID))
}

func (e Error) Error() string {
	sb := strings.Builder{}
	sb.WriteString("pfd.Error: ")
	sb.WriteString(e.Wrapped.Error())
	sb.WriteString(": [")

	for i, loc := range e.Locations {
		if i > 0 {
			sb.WriteString(", ")
		}
		loc.Write(&sb)
	}
	sb.WriteString("]")

	return sb.String()
}

func (e Error) Unwrap() error {
	return e.Wrapped
}

type Errors []Error

func (e Errors) Error() string {
	sb := strings.Builder{}
	for i, err := range e {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

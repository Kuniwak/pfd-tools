package pfd

import (
	"strings"
)

type Error struct {
	Locations []Location
	Wrapped   error
}

type Location struct {
	IsNode bool

	NodeID NodeID

	EdgeSourceID NodeID

	EdgeTargetID NodeID
}

func NewNodeLocation(nodeID NodeID) Location {
	return Location{NodeID: nodeID, IsNode: true}
}

func NewEdgeLocation(edgeSourceID, edgeTargetID NodeID) Location {
	return Location{EdgeSourceID: edgeSourceID, EdgeTargetID: edgeTargetID, IsNode: false}
}

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

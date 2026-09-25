package pfd

import "strings"

const UnnumberedProcessOpen = "("
const UnnumberedProcessClose = ")"
const UnnumberedDeliverableOpen = "["
const UnnumberedDeliverableClose = "]"

func NewUnnumberedNodeID(t NodeType, label string) NodeID {
	if label == "" {
		return ""
	}
	if t.IsProcess() {
		return NodeID(UnnumberedProcessOpen + label + UnnumberedProcessClose)
	}
	return NodeID(UnnumberedDeliverableOpen + label + UnnumberedDeliverableClose)
}

func (id NodeID) Unnumbered() (string, NodeType, bool) {
	if label, ok := UnwrapNodeID(id, UnnumberedProcessOpen, UnnumberedProcessClose); ok {
		return label, NodeTypeAtomicProcess, true
	}
	if label, ok := UnwrapNodeID(id, UnnumberedDeliverableOpen, UnnumberedDeliverableClose); ok {
		return label, NodeTypeAtomicDeliverable, true
	}
	return "", "", false
}

func (id NodeID) Label() string {
	if label, _, ok := id.Unnumbered(); ok {
		return label
	}
	return string(id)
}

func (id NodeID) UnnumberedNodeType() (NodeType, bool) {
	_, t, ok := id.Unnumbered()
	return t, ok
}

func UnwrapNodeID(id NodeID, prefix, suffix string) (string, bool) {
	s := string(id)
	if len(s) <= len(prefix)+len(suffix) {
		return "", false
	}
	if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, suffix) {
		return "", false
	}
	return s[len(prefix) : len(s)-len(suffix)], true
}

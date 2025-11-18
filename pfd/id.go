package pfd

import (
	"fmt"
	"strconv"
	"strings"
)

type NodeID string

const ProcessIDPrefix = "P"
const DeliverableIDPrefix = "D"

const NodeIDContextDiagram NodeID = ProcessIDPrefix + "0"

func NewAtomicProcessID(num int) NodeID {
	return NodeID(fmt.Sprintf("%s%d", ProcessIDPrefix, num))
}

func NewCompositeProcessID(num int) NodeID {
	return NodeID(fmt.Sprintf("%s%d", ProcessIDPrefix, num))
}

func NewDeliverableID(num int) NodeID {
	return NodeID(fmt.Sprintf("%s%d", DeliverableIDPrefix, num))
}

func NewCompositeDeliverableID(num int) NodeID {
	return NodeID(fmt.Sprintf("%s%d", DeliverableIDPrefix, num))
}

func (a NodeID) Compare(b NodeID) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return strings.Compare(string(a), string(b))
}

func ParseNodeID(id NodeID) (int, error) {
	i, err := ParseProcessID(id)
	if err == nil {
		return i, nil
	}
	i, err = ParseDeliverableID(id)
	if err == nil {
		return i, nil
	}
	return 0, fmt.Errorf("pfd.ParseNodeID: invalid node ID: %q", id)
}

func ParseProcessID(id NodeID) (int, error) {
	if !strings.HasPrefix(string(id), ProcessIDPrefix) {
		return 0, fmt.Errorf("pfd.ParseProcessID: invalid node ID: %q", id)
	}
	return strconv.Atoi(string(id[1:]))
}

func ParseDeliverableID(id NodeID) (int, error) {
	if !strings.HasPrefix(string(id), DeliverableIDPrefix) {
		return 0, fmt.Errorf("pfd.ParseDeliverableID: invalid node ID: %q", id)
	}
	return strconv.Atoi(string(id[1:]))
}

package pfddrawio

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/pfd"
)

func FormatVertexValue(id pfd.NodeID, desc string) string {
	if desc == "" {
		return id.Label()
	}
	return fmt.Sprintf("%s: %s", id.Label(), desc)
}

func VertexElementID(id pfd.NodeID, desc string, t pfd.NodeType) pfd.NodeID {
	if id == "" || desc != "" {
		return id
	}
	if _, err := pfd.ParseNodeID(id); err == nil {
		return id
	}
	return pfd.NewUnnumberedNodeID(t, string(id))
}

func PageNameElementID(name string) pfd.NodeID {

	id, desc, err := ParseVertexValue(name)
	if err != nil {
		return ""
	}
	return VertexElementID(id, desc, pfd.NodeTypeCompositeProcess)
}

func VertexNodeTypeClass(style StyleMap) (pfd.NodeType, bool) {
	if style.IsRectangle() {
		return pfd.NodeTypeAtomicDeliverable, true
	}
	if style.IsEllipse() {
		return pfd.NodeTypeAtomicProcess, true
	}
	return "", false
}

package pfddot

import (
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/dot"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

func Write(w io.Writer, p *pfd.PFD) error {
	d, err := Dot(p)
	if err != nil {
		return fmt.Errorf("pfddot.Write: %w", err)
	}
	return d.Write(w)
}

func Dot(p *pfd.PFD) (*dot.Digraph, error) {
	res := &dot.Digraph{
		Name: p.Title,
		GraphAttributes: sets.New(
			dot.CompareAttribute,
			&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
				First:  dot.AttributeNameCharset,
				Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "UTF-8"},
			},
		),
		NodeAttributes: sets.New(
			dot.CompareAttribute,
			&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
				First:  dot.AttributeNameStyle,
				Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "filled"},
			},
		),
		EdgeAttributes: sets.New(dot.CompareAttribute),
		Nodes:          make([]*dot.Node, 0, p.Nodes.Len()),
		Edges:          make([]*dot.Edge, 0, p.Edges.Len()),
	}

	for _, node := range p.Nodes.Iter() {
		res.Nodes = append(res.Nodes, DotNode(node))
	}

	for _, edge := range p.Edges.Iter() {
		res.Edges = append(res.Edges, DotEdge(edge))
	}

	return res, nil
}

func DotNode(node *pfd.Node) *dot.Node {
	attrs := sets.New(
		dot.CompareAttribute,
		&pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameLabel,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: fmt.Sprintf("%s: %s", node.ID, node.Description)},
		},
	)

	switch node.Type {
	case pfd.NodeTypeCompositeProcess, pfd.NodeTypeCompositeDeliverable:
		break // NOTE: Composite processes and composite deliverables are not displayed.

	case pfd.NodeTypeAtomicProcess:
		attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameShape,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "ellipse"},
		})

	case pfd.NodeTypeAtomicDeliverable:
		attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameShape,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "box"},
		})
	}

	return &dot.Node{
		ID:         dot.NodeID(node.ID),
		Attributes: attrs,
	}
}

func DotEdge(edge *pfd.Edge) *dot.Edge {
	attrs := sets.New(dot.CompareAttribute)

	if edge.IsFeedback {
		attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameStyle,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "dashed"},
		})
	} else {
		attrs.Add(dot.CompareAttribute, &pairs.Pair[dot.AttributeName, dot.AttributeValue]{
			First:  dot.AttributeNameStyle,
			Second: dot.AttributeValue{Type: dot.AttributeValueTypeString, String: "solid"},
		})
	}

	return &dot.Edge{
		Source:     dot.NodeID(edge.Source),
		Target:     dot.NodeID(edge.Target),
		Attributes: attrs,
	}
}

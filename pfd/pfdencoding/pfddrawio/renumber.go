package pfddrawio

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

func Renumber(r io.Reader, logger *slog.Logger) ([]*xmldom.Node, error) {
	return RenumberWithBase(pfd.RenumberBase{})(r, logger)
}

func RenumberWithBase(base pfd.RenumberBase) func(io.Reader, *slog.Logger) ([]*xmldom.Node, error) {
	return func(r io.Reader, logger *slog.Logger) ([]*xmldom.Node, error) {
		nodes, err := xmldom.ParseXML(r)
		if err != nil {
			return nil, fmt.Errorf("pfddrawio.RenumberWithBase: %w", err)
		}

		plan, err := NewRenumberPlanWithBase(nodes, base, logger)
		if err != nil {
			return nil, fmt.Errorf("pfddrawio.RenumberWithBase: %w", err)
		}

		ApplyRenumberPlan(nodes, plan, logger)
		ApplyDuplicateMarksToNodes(nodes, AllPages, logger)

		return nodes, nil
	}
}

func RenumberByPlan(plan pfd.RenumberPlan) func(io.Reader, *slog.Logger) ([]*xmldom.Node, error) {
	return func(r io.Reader, logger *slog.Logger) ([]*xmldom.Node, error) {
		nodes, err := xmldom.ParseXML(r)
		if err != nil {
			return nil, fmt.Errorf("pfddrawio.RenumberByPlan: %w", err)
		}

		ApplyRenumberPlan(nodes, plan, logger)
		ApplyDuplicateMarksToNodes(nodes, AllPages, logger)

		return nodes, nil
	}
}

func NewRenumberPlanWithBase(nodes []*xmldom.Node, base pfd.RenumberBase, logger *slog.Logger) (pfd.RenumberPlan, error) {
	diagrams, err := ParseNodes(nodes, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.NewRenumberPlanWithBase: %w", err)
	}

	p, srcMapMap, err := NormalizeDiagrams("", diagrams, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.NewRenumberPlanWithBase: %w", err)
	}

	nodeMap := pfd.NewNodeMap(p.Nodes, logger)
	graphExceptFB := p.GraphExceptFeedback(nodeMap, logger)

	pfdErrs := make([]pfd.Error, 0)
	plan, ok := pfd.NewRenumberPlanWithBase(p, graphExceptFB, nodeMap, base, &pfdErrs)
	if !ok {
		return nil, NewCellErrorsByPFDErrors(pfdErrs, srcMapMap)
	}

	return plan, nil
}

func ApplyRenumberPlan(nodes []*xmldom.Node, plan pfd.RenumberPlan, logger *slog.Logger) {
	for _, node := range nodes {
		logger.Debug("Rewriting attributes...", "node", node.Start.Name.Local)
		node.RewriteAttr(func(node *xmldom.Node, attr xml.Attr) xml.Attr {
			id, _, ok := ParseVertexValueAttr(node, attr, logger)
			if !ok {

				id, ok = ParsePageNameAttr(node, attr)
			}
			if !ok {
				return attr
			}

			newNode, ok := plan[id]
			if !ok {

				return attr
			}

			return xml.Attr{
				Name:  attr.Name,
				Value: FormatVertexValue(newNode.ID, newNode.Description),
			}
		})
	}
}

func MaxNumberedIDs(nodes []*xmldom.Node, logger *slog.Logger) (int, int) {
	maxProcessNumber, maxDeliverableNumber := 0, 0
	for _, node := range nodes {
		node.Traverse(func(n *xmldom.Node) {
			for _, attr := range n.Start.Attr {
				id, _, ok := ParseVertexValueAttr(n, attr, logger)
				if !ok {
					continue
				}

				if num, err := pfd.ParseProcessID(id); err == nil {
					maxProcessNumber = max(maxProcessNumber, num)
					continue
				}
				if num, err := pfd.ParseDeliverableID(id); err == nil {
					maxDeliverableNumber = max(maxDeliverableNumber, num)
				}
			}
		}, nil)
	}
	return maxProcessNumber, maxDeliverableNumber
}

func ParsePageNameAttr(node *xmldom.Node, attr xml.Attr) (pfd.NodeID, bool) {
	if node.Start.Name.Local != "diagram" {
		return "", false
	}

	if attr.Name.Local != "name" || attr.Name.Space != "" {
		return "", false
	}

	if IsContextDiagramName(attr.Value) {
		return "", false
	}

	id := PageNameElementID(attr.Value)
	if id == "" {
		return "", false
	}

	return id, true
}

func ParseVertexValueAttr(node *xmldom.Node, attr xml.Attr, logger *slog.Logger) (pfd.NodeID, string, bool) {
	style, ok := node.GetAttr("style", "")
	if !ok {
		return "", "", false
	}

	styleMap, err := ParseStyle(style)
	if err != nil {
		return "", "", false
	}

	t, ok := VertexNodeTypeClass(styleMap)
	if !ok {
		return "", "", false
	}

	if attr.Name.Local != "value" || attr.Name.Space != "" {
		return "", "", false
	}

	sb := &strings.Builder{}
	if err := ParseValueHTML(ValueHTML(attr.Value), sb); err != nil {
		logger.Warn("pfddrawio.ParseVertexValueAttr: failed to parse value attribute", "attr.value", attr.Value)
		return "", "", false
	}

	id, description, err := ParseVertexValue(sb.String())
	if err != nil {
		logger.Debug("pfddrawio.ParseVertexValueAttr: skipping vertex, it looks like a comment", "value", attr.Value)
		return "", "", false
	}

	return VertexElementID(id, description, t), description, true
}

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
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Renumber: %w", err)
	}

	diagrams, err := ParseNodes(nodes, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Renumber: %w", err)
	}

	p, srcMapMap, err := NormalizeDiagrams("", diagrams, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Renumber: %w", err)
	}

	nodeMap := pfd.NewNodeMap(p.Nodes, logger)
	graphExceptFB := p.GraphExceptFeedback(nodeMap, logger)

	pfdErrs := make([]pfd.Error, 0)
	plan, ok := pfd.NewRenumberPlan(p, graphExceptFB, nodeMap, &pfdErrs)
	if !ok {
		return nil, NewCellErrorsByPFDErrors(pfdErrs, srcMapMap)
	}

	sb := &strings.Builder{}
	for _, node := range nodes {
		logger.Debug("Rewriting attributes...", "node", node.Start.Name.Local)
		node.RewriteAttr(func(node *xmldom.Node, attr xml.Attr) xml.Attr {
			style, ok := node.GetAttr("style", "")
			if !ok {
				return attr
			}

			styleMap, err := ParseStyle(style)
			if err != nil {
				return attr
			}

			if !styleMap.IsRectangle() && !styleMap.IsEllipse() {
				return attr
			}

			if attr.Name.Local != "value" || attr.Name.Space != "" {
				return attr
			}

			if err := ParseValueHTML(ValueHTML(attr.Value), sb); err != nil {
				logger.Warn("pfddrawio.Renumber: failed to parse value attribute", "attr.value", attr.Value)
				return attr
			}
			value := sb.String()
			sb.Reset()

			id, _, err := ParseVertexValue(value)
			if err != nil {
				logger.Debug("pfddrawio.Renumber: skipping vertex, it looks like a comment", "value", attr.Value)
				return attr
			}

			newNode, ok := plan[id]
			if !ok {
				// NOTE: Composite deliverable.
				return attr
			}

			return xml.Attr{
				Name:  attr.Name,
				Value: fmt.Sprintf("%s: %s", newNode.ID, newNode.Description),
			}
		})
	}

	return nodes, nil
}

package pfddrawio

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"golang.org/x/net/html"
)

func Parse(title string, r io.Reader, cdt *pfd.CompositeDeliverableTable, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	p, srcMap, err := ParseExceptCompositeDeliverables(title, r, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.Parse: %w", err)
	}

	p.DeliverableComposition = cdt.NodeIDMap(logger)

	newEdges := p.Edges.Clone()
	for _, edge := range p.Edges.Iter() {
		var ok bool
		var ns *sets.Set[pfd.NodeID]

		ns, ok = p.DeliverableComposition[edge.Source]
		if !ok {
			continue
		}
		for _, newSource := range ns.Iter() {
			newEdges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: newSource, Target: edge.Target, IsFeedback: edge.IsFeedback})
		}

		ns, ok = p.DeliverableComposition[edge.Target]
		if !ok {
			continue
		}
		for _, newTarget := range ns.Iter() {
			newEdges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: edge.Source, Target: newTarget, IsFeedback: edge.IsFeedback})
		}
	}

	p.Edges = newEdges
	return p, srcMap, nil
}

func ParseExceptCompositeDeliverables(title string, r io.Reader, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	ds, err := ParseDiagrams(r, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.ParseExceptCompositeDeliverables: %w", err)
	}
	p, srcMap, err := NormalizeDiagrams(title, ds, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.ParseExceptCompositeDeliverables: %w", err)
	}

	return p, srcMap, nil
}

func ParseDiagrams(r io.Reader, logger *slog.Logger) ([]Diagram, error) {
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.ParseDiagrams: %w", err)
	}

	diagrams, err := ParseNodes(nodes, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.ParseDiagrams: %w", err)
	}
	return diagrams, nil
}

func ParseNodes(nodes []*xmldom.Node, logger *slog.Logger) ([]Diagram, error) {
	sb := strings.Builder{}
	diagrams := make([]Diagram, 0)
	hasMxFile := false
	for _, node := range nodes {
		if node.Start.Name.Local == "mxfile" {
			hasMxFile = true
			diagrams = append(diagrams, ParseMxFile(node, &sb, logger)...)
		}
	}
	if !hasMxFile {
		return nil, errors.New("pfddrawio.ParseNodes: missing mxfile element")
	}
	return diagrams, nil
}

func ParseMxFile(root *xmldom.Node, sb *strings.Builder, logger *slog.Logger) []Diagram {
	if root.Start.Name.Local != "mxfile" {
		panic("pfddrawio.ParseMxFile: invalid root element")
	}

	diagrams := make([]Diagram, 0)

	for _, child := range root.Children {
		if child.Start.Name.Local == "diagram" {
			diagram, ok := ParseDiagram(child, sb, logger)
			if !ok {
				continue
			}
			diagrams = append(diagrams, diagram)
		}
	}
	return diagrams
}

func ParseDiagram(node *xmldom.Node, sb *strings.Builder, logger *slog.Logger) (Diagram, bool) {
	if node.Start.Name.Local != "diagram" {
		panic("pfddrawio.ParseDiagram: invalid diagram element")
	}

	idText, ok := node.GetAttr("id", "")
	if !ok {
		logger.Warn("pfddrawio.ParseDiagram: missing id attribute")
		return Diagram{}, false
	}
	id := DiagramID(idText)

	name, ok := node.GetAttr("name", "")
	if !ok {
		logger.Warn("pfddrawio.ParseDiagram: missing name attribute")
		return Diagram{}, false
	}

	for _, child := range node.Children {
		if child.Start.Name.Local == "mxGraphModel" {
			cells := ParseMxGraphModel(child, sb, logger)
			return Diagram{
				ID:    id,
				Name:  name,
				Cells: cells,
			}, true
		}
	}
	logger.Warn("pfddrawio.ParseDiagram: missing mxGraphModel element")
	return Diagram{}, false
}

func ParseMxGraphModel(node *xmldom.Node, sb *strings.Builder, logger *slog.Logger) []Cell {
	if node.Start.Name.Local != "mxGraphModel" {
		panic("pfddrawio.ParseMxGraphModel: invalid mxGraphModel element")
	}

	for _, child := range node.Children {
		if child.Start.Name.Local == "root" {
			return ParseRoot(child, sb, logger)
		}
	}
	logger.Warn("pfddrawio.ParseMxGraphModel: missing root element")
	return nil
}

func ParseRoot(node *xmldom.Node, sb *strings.Builder, logger *slog.Logger) []Cell {
	if node.Start.Name.Local != "root" {
		panic("pfddrawio.ParseRoot: invalid root element")
	}

	cells := make([]Cell, 0)
	for _, child := range node.Children {
		if child.Start.Name.Local == "mxCell" {
			cell, ok := ParseMxCell(child, sb, logger)
			if !ok {
				continue
			}
			cells = append(cells, cell)
		}
	}
	return cells
}

func ParseMxCell(node *xmldom.Node, sb *strings.Builder, logger *slog.Logger) (Cell, bool) {
	if node.Start.Name.Local != "mxCell" {
		panic("pfddrawio.ParseMxCell: invalid cell element")
	}

	idText, ok := node.GetAttr("id", "")
	if !ok {
		logger.Warn("pfddrawio.ParseMxCell: missing id attribute", "node", node)
		return Cell{}, false
	}
	id := CellID(idText)

	parentText, ok := node.GetAttr("parent", "")
	if !ok {
		return NewRoot(id), true
	}
	isLayer := parentText == "0"
	if isLayer {
		valueHTML, ok := node.GetAttr("value", "")
		if !ok {
			return NewLayer(id, ""), true
		}
		return NewLayer(id, valueHTML), true
	}
	parent := CellID(parentText)

	var styleMap StyleMap
	if style, ok := node.GetAttr("style", ""); ok {
		var err error
		styleMap, err = ParseStyle(style)
		if err != nil {
			logger.Warn("pfddrawio.ParseMxCell: failed to parse style", "error", err.Error(), "id", id)
			styleMap = make(StyleMap)
		}
	} else {
		styleMap = make(StyleMap)
	}

	vertex, ok := node.GetAttr("vertex", "")
	isVertex := ok && vertex != ""
	if isVertex {
		valueHTML, ok := node.GetAttr("value", "")
		if !ok {
			logger.Warn("pfddrawio.ParseMxCell: missing value attribute", "id", id)
			return Cell{}, false
		}
		sb.Reset()
		if err := ParseValueHTML(ValueHTML(valueHTML), sb); err != nil {
			logger.Warn("pfddrawio.ParseMxCell: failed to parse value", "error", err.Error(), "id", id)
			return Cell{}, false
		}
		value := sb.String()
		return NewVertex(id, parent, value, styleMap), true
	}

	edge, ok := node.GetAttr("edge", "")
	isEdge := ok && edge != ""
	if isEdge {
		sourceText, ok := node.GetAttr("source", "")
		if !ok {
			logger.Warn("pfddrawio.ParseMxCell: missing source attribute", "id", id)
			return Cell{}, false
		}
		targetText, ok := node.GetAttr("target", "")
		if !ok {
			logger.Warn("pfddrawio.ParseMxCell: missing target attribute", "id", id)
			return Cell{}, false
		}

		source := CellID(sourceText)
		target := CellID(targetText)

		return NewEdge(id, parent, source, target, styleMap), true
	}

	return Cell{}, false
}

func ParseStyle(style string) (StyleMap, error) {
	styleMap := make(StyleMap)
	for _, style := range strings.Split(style, ";") {
		if style == "" {
			continue
		}
		var name, value string
		parts := strings.Split(style, "=")
		switch len(parts) {
		case 1:
			name = parts[0]
		case 2:
			name = parts[0]
			value = parts[1]
		default:
			return nil, fmt.Errorf("pfddrawio.ParseStyle: invalid style: %s", style)
		}
		styleMap[name] = value
	}
	return styleMap, nil
}

func ParseValueHTML(value ValueHTML, sb *strings.Builder) error {
	node, err := html.Parse(strings.NewReader(string(value)))
	if err != nil {
		return err
	}
	TextContent(node, sb)
	return nil
}

func ParseVertexValue(s string) (pfd.NodeID, string, error) {
	parts := strings.SplitN(s, ":", 2)
	switch len(parts) {
	case 1:
		// NOTE: If there are unnumbered IDs, use the description as the ID.
		return pfd.NodeID(strings.TrimSpace(parts[0])), "", nil
	case 2:
		return pfd.NodeID(strings.TrimSpace(parts[0])), strings.TrimSpace(parts[1]), nil
	default:
		return "", "", fmt.Errorf("pfddrawio.ParseVertexValue: invalid value: %s", s)
	}
}

func TextContent(node *html.Node, sb *strings.Builder) {
	if node.Type == html.TextNode {
		sb.WriteString(node.Data)
	}

	if node.Data == "br" {
		sb.WriteString("\n")
	}

	if node.FirstChild != nil {
		TextContent(node.FirstChild, sb)
	}

	if node.NextSibling != nil {
		TextContent(node.NextSibling, sb)
	}
}

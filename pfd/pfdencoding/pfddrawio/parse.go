package pfddrawio

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"golang.org/x/net/html"
)

func Parse(title string, r io.Reader, cdt *pfd.CompositeDeliverableTable, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	return ParseWithOptions(title, r, cdt, NormalizeOptions{}, logger)
}

func ParseWithOptions(title string, r io.Reader, cdt *pfd.CompositeDeliverableTable, opts NormalizeOptions, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	p, srcMap, err := ParseExceptCompositeDeliverablesWithOptions(title, r, opts, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.ParseWithOptions: %w", err)
	}

	if err := p.SetExpandedDeliverableComposition(cdt.NodeIDMap(logger)); err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.ParseWithOptions: %w", err)
	}

	return p, srcMap, nil
}

func ParseExceptCompositeDeliverables(title string, r io.Reader, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	return ParseExceptCompositeDeliverablesWithOptions(title, r, NormalizeOptions{}, logger)
}

func ParseExceptCompositeDeliverablesWithOptions(title string, r io.Reader, opts NormalizeOptions, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	ds, err := ParseDiagrams(r, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.ParseExceptCompositeDeliverablesWithOptions: %w", err)
	}
	p, srcMap, err := NormalizeDiagramsWithOptions(title, ds, opts, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.ParseExceptCompositeDeliverablesWithOptions: %w", err)
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
			ds, err := ParseMxFile(node, &sb, logger)
			if err != nil {
				return nil, fmt.Errorf("pfddrawio.ParseNodes: %w", err)
			}
			diagrams = append(diagrams, ds...)
		}
	}
	if !hasMxFile {
		return nil, errors.New("pfddrawio.ParseNodes: missing mxfile element")
	}
	return diagrams, nil
}

func ParseMxFile(root *xmldom.Node, sb *strings.Builder, logger *slog.Logger) ([]Diagram, error) {
	if root.Start.Name.Local != "mxfile" {

		panic("pfddrawio.ParseMxFile: invalid root element")
	}

	diagrams := make([]Diagram, 0)

	for _, child := range root.Children {
		if child.Start.Name.Local == "diagram" {
			diagram, err := ParseDiagram(child, sb, logger)
			if err != nil {
				return nil, fmt.Errorf("pfddrawio.ParseMxFile: %w", err)
			}
			diagrams = append(diagrams, diagram)
		}
	}
	return diagrams, nil
}

func ParseDiagram(node *xmldom.Node, sb *strings.Builder, logger *slog.Logger) (Diagram, error) {
	if node.Start.Name.Local != "diagram" {

		panic("pfddrawio.ParseDiagram: invalid diagram element")
	}

	idText, ok := node.GetAttr("id", "")
	if !ok {
		return Diagram{}, errors.New("pfddrawio.ParseDiagram: ページに id 属性がありません。ページ id はページの同一性を示すので、すべてのページに必要です")
	}
	id := DiagramID(idText)

	name, ok := node.GetAttr("name", "")
	if !ok {
		return Diagram{}, fmt.Errorf("pfddrawio.ParseDiagram: ページ id %q に name 属性がありません。ページ名はコンテキスト図では %q、詳細図では内訳を示す複合プロセスの ID にしてください", id, pfd.NodeIDContextDiagram)
	}

	for _, child := range node.Children {
		if child.Start.Name.Local == "mxGraphModel" {
			cells, err := ParseMxGraphModel(child, sb, logger)
			if err != nil {
				return Diagram{}, fmt.Errorf("pfddrawio.ParseDiagram: ページ %q (id=%q): %w", name, id, err)
			}
			return Diagram{
				ID:    id,
				Name:  name,
				Cells: cells,
			}, nil
		}
	}

	return Diagram{}, fmt.Errorf("pfddrawio.ParseDiagram: ページ %q (id=%q) に mxGraphModel 要素がありません。draw.io の圧縮保存（ファイル > プロパティ > 圧縮）が有効な可能性があります。圧縮を無効にして保存してください", name, id)
}

func ParseMxGraphModel(node *xmldom.Node, sb *strings.Builder, logger *slog.Logger) ([]Cell, error) {
	if node.Start.Name.Local != "mxGraphModel" {

		panic("pfddrawio.ParseMxGraphModel: invalid mxGraphModel element")
	}

	for _, child := range node.Children {
		if child.Start.Name.Local == "root" {
			return ParseRoot(child, sb, logger), nil
		}
	}
	return nil, errors.New("pfddrawio.ParseMxGraphModel: mxGraphModel 要素に root 要素がありません")
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
	for _, token := range StyleTokens(style) {
		name, value, _ := strings.Cut(token, "=")
		if strings.Contains(value, "=") {
			return nil, fmt.Errorf("pfddrawio.ParseStyle: invalid style: %s", token)
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

func VertexNodeID(n *xmldom.Node) (pfd.NodeID, bool) {
	value, ok := n.GetAttr("value", "")
	if !ok {
		return "", false
	}
	var sb strings.Builder
	if err := ParseValueHTML(ValueHTML(value), &sb); err != nil {
		return "", false
	}
	id, desc, err := ParseVertexValue(sb.String())
	if err != nil {
		return "", false
	}

	style, ok := n.GetAttr("style", "")
	if !ok {
		return "", false
	}
	styleMap, err := ParseStyle(style)
	if err != nil {
		return "", false
	}
	t, ok := VertexNodeTypeClass(styleMap)
	if !ok {
		return "", false
	}
	return VertexElementID(id, desc, t), true
}

func ParseVertexValue(s string) (pfd.NodeID, string, error) {
	parts := strings.SplitN(s, ":", 2)
	switch len(parts) {
	case 1:

		id, _ := SplitDuplicateMark(pfd.NodeID(strings.TrimSpace(parts[0])))
		return id, "", nil
	case 2:
		id, _ := SplitDuplicateMark(pfd.NodeID(strings.TrimSpace(parts[0])))
		return id, strings.TrimSpace(parts[1]), nil
	default:
		return "", "", fmt.Errorf("pfddrawio.ParseVertexValue: invalid value: %s", s)
	}
}

func TextContent(node *html.Node, sb *strings.Builder) {
	if node.Type == html.TextNode {
		sb.WriteString(node.Data)
	}

	if node.FirstChild != nil {
		TextContent(node.FirstChild, sb)
	}

	if node.NextSibling != nil {
		TextContent(node.NextSibling, sb)
	}
}

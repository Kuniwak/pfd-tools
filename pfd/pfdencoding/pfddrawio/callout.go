package pfddrawio

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const EstimateBoxPlaceholder = "楽観: d 悲観: d"

const EstimateBoxStyle = "text;html=1;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;"

const EstimateOptimisticKeyword = "楽観"
const EstimatePessimisticKeyword = "悲観"

const (
	estimateBoxOffsetX = 5.0
	estimateBoxHeight  = 30.0
)

const (
	estimateBoxGapAbove = 10.0
	estimateBoxGapBelow = 40.0
)

func Callout(r io.Reader, logger *slog.Logger) ([]*xmldom.Node, error) {
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Callout: %w", err)
	}

	diagrams, err := ParseNodes(nodes, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Callout: %w", err)
	}
	p, srcMap, err := NormalizeDiagrams("", diagrams, logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.Callout: %w", err)
	}

	doms := CollectDiagramDOMs(nodes)
	nextID := MaxNumericCellID(nodes) + 1

	existing := make(map[DiagramID][]EstimateBox)
	for _, b := range EstimateBoxes(nodes) {
		existing[b.DiagramID] = append(existing[b.DiagramID], b)
	}

	for _, id := range p.AtomicProcesses().Slice() {
		locs, ok := srcMap.NodeIDMap[id]
		if !ok {
			logger.Warn("pfddrawio.Callout: atomic process has no source location; skipped", "process", id)
			continue
		}
		for _, loc := range locs.Slice() {
			dom, ok := doms[loc.DiagramID]
			if !ok {
				logger.Warn("pfddrawio.Callout: diagram not found in DOM; skipped", "process", id, "diagramID", loc.DiagramID)
				continue
			}
			procNode, ok := dom.CellsByID[loc.CellID]
			if !ok {
				logger.Warn("pfddrawio.Callout: process cell not found in DOM; skipped", "process", id, "diagramID", loc.DiagramID, "cellID", loc.CellID)
				continue
			}
			procRect, ok := GeometryRect(procNode)
			if !ok {
				logger.Warn("pfddrawio.Callout: process cell has no geometry; skipped", "process", id, "diagramID", loc.DiagramID, "cellID", loc.CellID)
				continue
			}

			if anyEstimateBoxBelow(existing[loc.DiagramID], procRect) {
				logger.Debug("pfddrawio.Callout: estimate box already exists; skipped", "process", id)
				continue
			}

			indent := dom.Root.ChildElementIndent()
			parent, _ := procNode.GetAttr("parent", "")
			box := newEstimateBoxNode(strconv.Itoa(nextID), parent, EstimateBoxRect(procRect), indent)
			dom.Root.AppendChildIndented(indent, box)
			logger.Debug("pfddrawio.Callout: added estimate box", "process", id, "cell", nextID)
			nextID++
		}
	}

	return nodes, nil
}

func GeometryRect(n *xmldom.Node) (geom.Rect, bool) {
	geo := n.FirstChildElement("mxGeometry")
	if geo == nil {
		return geom.Rect{}, false
	}
	w, okW := geo.FloatAttr("width")
	h, okH := geo.FloatAttr("height")
	if !okW || !okH {
		return geom.Rect{}, false
	}
	x, _ := geo.FloatAttr("x")
	y, _ := geo.FloatAttr("y")
	return geom.Rect{X: x, Y: y, Width: w, Height: h}, true
}

func EstimateBoxRect(process geom.Rect) geom.Rect {
	return geom.Rect{
		X:      process.X + estimateBoxOffsetX,
		Y:      process.Y + process.Height,
		Width:  process.Width - 2*estimateBoxOffsetX,
		Height: estimateBoxHeight,
	}
}

func IsBelow(process, box geom.Rect) bool {
	boxCenterX := box.X + box.Width/2
	if boxCenterX < process.X || boxCenterX > process.X+process.Width {
		return false
	}
	bottom := process.Y + process.Height
	return box.Y >= bottom-estimateBoxGapAbove && box.Y <= bottom+estimateBoxGapBelow
}

func IsEstimateBox(style StyleMap, value string) bool {
	return style.IsText() &&
		strings.Contains(value, EstimateOptimisticKeyword) &&
		strings.Contains(value, EstimatePessimisticKeyword)
}

type EstimateBox struct {
	DiagramID DiagramID
	CellID    CellID
	Value     string
	Rect      geom.Rect
}

func EstimateBoxes(nodes []*xmldom.Node) []EstimateBox {
	layers := NewLayerMapFromNodes(nodes)
	parents := NewParentMapFromNodes(nodes)

	var boxes []EstimateBox
	for diagramID, dom := range CollectDiagramDOMs(nodes) {
		for cellID, cell := range dom.CellsByID {
			if layers.IsCommentDescendant(parents, cellID) {
				continue
			}
			if vertex, _ := cell.GetAttr("vertex", ""); vertex != "1" {
				continue
			}
			value, _ := cell.GetAttr("value", "")
			styleStr, _ := cell.GetAttr("style", "")
			style, err := ParseStyle(styleStr)
			if err != nil || !IsEstimateBox(style, value) {
				continue
			}
			rect, ok := GeometryRect(cell)
			if !ok {
				continue
			}
			boxes = append(boxes, EstimateBox{DiagramID: diagramID, CellID: cellID, Value: value, Rect: rect})
		}
	}
	sort.Slice(boxes, func(i, j int) bool {
		if boxes[i].DiagramID != boxes[j].DiagramID {
			return boxes[i].DiagramID < boxes[j].DiagramID
		}
		if boxes[i].Rect.Y != boxes[j].Rect.Y {
			return boxes[i].Rect.Y < boxes[j].Rect.Y
		}
		return boxes[i].Rect.X < boxes[j].Rect.X
	})
	return boxes
}

func anyEstimateBoxBelow(boxes []EstimateBox, process geom.Rect) bool {
	for _, b := range boxes {
		if IsBelow(process, b.Rect) {
			return true
		}
	}
	return false
}

type DiagramDOM struct {
	Root      *xmldom.Node
	CellsByID map[CellID]*xmldom.Node
	Name      string
}

func PageNames(doms map[DiagramID]*DiagramDOM) []string {
	diagramIDs := make([]DiagramID, 0, len(doms))
	for id := range doms {
		diagramIDs = append(diagramIDs, id)
	}
	sort.Slice(diagramIDs, func(i, j int) bool {
		return CellID(diagramIDs[i]).Compare(CellID(diagramIDs[j])) < 0
	})

	names := make([]string, 0, len(diagramIDs))
	for _, id := range diagramIDs {
		names = append(names, doms[id].Name)
	}
	return names
}

func CollectDiagramDOMs(nodes []*xmldom.Node) map[DiagramID]*DiagramDOM {
	doms := make(map[DiagramID]*DiagramDOM)
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "diagram" {
				return
			}
			idText, _ := n.GetAttr("id", "")
			nameText, _ := n.GetAttr("name", "")
			model := n.FirstChildElement("mxGraphModel")
			if model == nil {
				return
			}
			rootEl := model.FirstChildElement("root")
			if rootEl == nil {
				return
			}
			cellsByID := make(map[CellID]*xmldom.Node)
			for _, ch := range rootEl.Children {
				if ch.Kind != xmldom.ElementNode || ch.Start.Name.Local != "mxCell" {
					continue
				}
				cellID, _ := ch.GetAttr("id", "")
				cellsByID[CellID(cellID)] = ch
			}
			doms[DiagramID(idText)] = &DiagramDOM{Root: rootEl, CellsByID: cellsByID, Name: nameText}
		}, nil)
	}
	return doms
}

func MaxNumericCellID(nodes []*xmldom.Node) int {
	maxID := 0
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			id, _ := n.GetAttr("id", "")
			if v, err := strconv.Atoi(id); err == nil && v > maxID {
				maxID = v
			}
		}, nil)
	}
	return maxID
}

func newEstimateBoxNode(id, parent string, rect geom.Rect, indent string) *xmldom.Node {
	geo := &xmldom.Node{
		Kind: xmldom.ElementNode,
		Start: xml.StartElement{
			Name: xml.Name{Local: "mxGeometry"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "x"}, Value: formatFloat(rect.X)},
				{Name: xml.Name{Local: "y"}, Value: formatFloat(rect.Y)},
				{Name: xml.Name{Local: "width"}, Value: formatFloat(rect.Width)},
				{Name: xml.Name{Local: "height"}, Value: formatFloat(rect.Height)},
				{Name: xml.Name{Local: "as"}, Value: "geometry"},
			},
		},
		End: xml.EndElement{Name: xml.Name{Local: "mxGeometry"}},
	}

	cell := &xmldom.Node{
		Kind: xmldom.ElementNode,
		Start: xml.StartElement{
			Name: xml.Name{Local: "mxCell"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "id"}, Value: id},
				{Name: xml.Name{Local: "value"}, Value: EstimateBoxPlaceholder},
				{Name: xml.Name{Local: "style"}, Value: EstimateBoxStyle},
				{Name: xml.Name{Local: "vertex"}, Value: "1"},
				{Name: xml.Name{Local: "parent"}, Value: parent},
			},
		},
		End: xml.EndElement{Name: xml.Name{Local: "mxCell"}},
		Children: []*xmldom.Node{
			{Kind: xmldom.TextNode, Data: []byte(indent + "    ")},
			geo,
			{Kind: xmldom.TextNode, Data: []byte(indent)},
		},
	}
	return cell
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

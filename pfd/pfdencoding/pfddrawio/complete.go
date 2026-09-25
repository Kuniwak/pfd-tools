package pfddrawio

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const (
	scaffoldNodeWidth  = "120"
	scaffoldNodeHeight = "80"
	scaffoldLeftX      = 40.0
	scaffoldRightX     = 360.0
	scaffoldStartY     = 40.0
	scaffoldRowStep    = 120.0
)

const (
	indentDiagram = "\n    "
	indentCell    = "\n                "
	indentGeom    = "\n                    "
)

type DeliverableAppearance struct {
	Value  string
	Style  string
	Width  string
	Height string
}

func CompleteCompositePages(r io.Reader, logger *slog.Logger) ([]*xmldom.Node, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.CompleteCompositePages: %w", err)
	}

	nodes, err := xmldom.ParseXML(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.CompleteCompositePages: %w", err)
	}

	p, _, err := ParseExceptCompositeDeliverables("", bytes.NewReader(input), logger)
	if err != nil {
		return nil, fmt.Errorf("pfddrawio.CompleteCompositePages: %w", err)
	}
	nodeMap := pfd.NewNodeMap(p.Nodes, logger)

	appearances := CollectDeliverableAppearances(nodes)

	mxfile := FindMxfile(nodes)
	if mxfile == nil {
		return nil, fmt.Errorf("pfddrawio.CompleteCompositePages: mxfile 要素が見つかりません")
	}

	diagramByPageID := diagramNodesByPageID(mxfile)
	usedDiagramIDs := diagramIDSet(mxfile)

	for _, cp := range ScaffoldTargetIDs(p) {
		inputs, outputs := BoundaryDeliverables(p, nodeMap, cp)

		description := ""
		if node, ok := nodeMap[cp]; ok {
			description = node.Description
		}
		pageName := FormatVertexValue(cp, description)

		page, hasPage := diagramByPageID[cp]
		if !hasPage {

			diagramID := uniqueID("cpcomp-"+cp.Label(), usedDiagramIDs)
			usedDiagramIDs[diagramID] = struct{}{}

			diagram, root, err := buildScaffoldPage(diagramID, pageName)
			if err != nil {
				return nil, fmt.Errorf("pfddrawio.CompleteCompositePages: %w", err)
			}

			usedCellIDs := map[string]struct{}{"0": {}, "1": {}}
			if err := placeBoundaries(root, cp, inputs, outputs, appearances, scaffoldLeftX, scaffoldRightX, columnYFor, "1", usedCellIDs); err != nil {
				return nil, err
			}

			mxfile.AppendChildIndented(indentDiagram, diagram)
			continue
		}

		present, bounds, layer, root := computePageState(page)
		if root == nil {
			logger.Warn("pfddrawio.CompleteCompositePages: ページに root がありません", "page", cp)
			continue
		}

		missingInputs := filterMissing(inputs, present)
		missingOutputs := filterMissing(outputs, present)
		if len(missingInputs) == 0 && len(missingOutputs) == 0 {
			continue
		}

		usedCellIDs := pageCellIDs(root)
		if !bounds.Found {

			if err := placeBoundaries(root, cp, missingInputs, missingOutputs, appearances, scaffoldLeftX, scaffoldRightX, columnYFor, layer, usedCellIDs); err != nil {
				return nil, err
			}
			continue
		}

		aboveY := func(i int) float64 { return bounds.MinY - float64(i+1)*scaffoldRowStep }
		if err := placeBoundaries(root, cp, missingInputs, missingOutputs, appearances, bounds.MinX, bounds.MaxX, aboveY, layer, usedCellIDs); err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

type pageBounds struct {
	MinX, MaxX, MinY float64
	Found            bool
}

func computePageState(diagram *xmldom.Node) (map[pfd.NodeID]bool, pageBounds, string, *xmldom.Node) {
	present := make(map[pfd.NodeID]bool)
	var bounds pageBounds

	layerMap := NewLayerMapFromNodes([]*xmldom.Node{diagram})
	parents := NewParentMapFromNodes([]*xmldom.Node{diagram})

	root := PageRoot(diagram)
	if root == nil {
		return present, bounds, "1", nil
	}
	layer := defaultLayerID(root, layerMap)

	for _, cell := range root.Children {
		if cell.Kind != xmldom.ElementNode || cell.Start.Name.Local != "mxCell" {
			continue
		}
		if v, _ := cell.GetAttr("vertex", ""); v == "" {
			continue
		}
		idAttr, _ := cell.GetAttr("id", "")
		if layerMap.IsCommentDescendant(parents, CellID(idAttr)) {
			continue
		}

		if id, ok := VertexNodeID(cell); ok {
			present[id] = true
		}

		geo := cell.FirstChildElement("mxGeometry")
		if geo == nil {
			continue
		}
		x, okX := geo.FloatAttr("x")
		y, okY := geo.FloatAttr("y")
		if !okX || !okY {
			continue
		}
		if !bounds.Found {
			bounds = pageBounds{MinX: x, MaxX: x, MinY: y, Found: true}
			continue
		}
		if x < bounds.MinX {
			bounds.MinX = x
		}
		if x > bounds.MaxX {
			bounds.MaxX = x
		}
		if y < bounds.MinY {
			bounds.MinY = y
		}
	}
	return present, bounds, layer, root
}

func PageRoot(diagram *xmldom.Node) *xmldom.Node {
	model := diagram.FirstChildElement("mxGraphModel")
	if model == nil {
		return nil
	}
	return model.FirstChildElement("root")
}

func defaultLayerID(root *xmldom.Node, layerMap LayerMap) string {
	for _, cell := range root.Children {
		if cell.Kind != xmldom.ElementNode || cell.Start.Name.Local != "mxCell" {
			continue
		}
		if parent, _ := cell.GetAttr("parent", ""); parent != "0" {
			continue
		}
		idAttr, _ := cell.GetAttr("id", "")
		if layerMap.IsCommentLayer(CellID(idAttr)) {
			continue
		}
		return idAttr
	}
	return "1"
}

func pageCellIDs(root *xmldom.Node) map[string]struct{} {
	used := make(map[string]struct{})
	for _, cell := range root.Children {
		if cell.Kind != xmldom.ElementNode || cell.Start.Name.Local != "mxCell" {
			continue
		}
		if id, ok := cell.GetAttr("id", ""); ok {
			used[id] = struct{}{}
		}
	}
	return used
}

func filterMissing(ids []pfd.NodeID, present map[pfd.NodeID]bool) []pfd.NodeID {
	missing := make([]pfd.NodeID, 0, len(ids))
	for _, id := range ids {
		if !present[id] {
			missing = append(missing, id)
		}
	}
	return missing
}

func BoundaryDeliverables(p *pfd.PFD, nodeMap map[pfd.NodeID]*pfd.Node, cp pfd.NodeID) ([]pfd.NodeID, []pfd.NodeID) {
	inputSet := p.InputsIncludingFeedback(cp)
	outputSet := p.OutputsIncludingFeedback(cp)

	isInput := make(map[pfd.NodeID]bool)
	inputs := make([]pfd.NodeID, 0)
	for _, id := range inputSet.Slice() {
		if isDeliverable(nodeMap, id) {
			inputs = append(inputs, id)
			isInput[id] = true
		}
	}

	outputs := make([]pfd.NodeID, 0)
	for _, id := range outputSet.Slice() {
		if isInput[id] {
			continue
		}
		if isDeliverable(nodeMap, id) {
			outputs = append(outputs, id)
		}
	}

	sortNodeIDs(inputs)
	sortNodeIDs(outputs)
	return inputs, outputs
}

func ScaffoldTargetIDs(p *pfd.PFD) []pfd.NodeID {
	seen := make(map[pfd.NodeID]struct{})
	ids := make([]pfd.NodeID, 0)
	add := func(id pfd.NodeID) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, n := range p.Nodes.Iter() {
		if n.Type == pfd.NodeTypeCompositeProcess {
			add(n.ID)
		}
	}

	for _, id := range p.ImplicitAtomicProcesses.Iter() {
		add(id)
	}

	sortNodeIDs(ids)
	return ids
}

func CollectDeliverableAppearances(nodes []*xmldom.Node) map[pfd.NodeID]DeliverableAppearance {
	result := make(map[pfd.NodeID]DeliverableAppearance)
	layerMap := NewLayerMapFromNodes(nodes)
	parents := NewParentMapFromNodes(nodes)

	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			if v, _ := n.GetAttr("vertex", ""); v == "" {
				return
			}
			idAttr, _ := n.GetAttr("id", "")
			if layerMap.IsCommentDescendant(parents, CellID(idAttr)) {
				return
			}
			styleStr, _ := n.GetAttr("style", "")
			styleMap, err := ParseStyle(styleStr)
			if err != nil || !styleMap.IsRectangle() {
				return
			}
			id, ok := VertexNodeID(n)
			if !ok {
				return
			}
			if _, exists := result[id]; exists {
				return
			}
			value, _ := n.GetAttr("value", "")

			width, height := scaffoldNodeWidth, scaffoldNodeHeight
			if geo := n.FirstChildElement("mxGeometry"); geo != nil {
				if w, ok := geo.GetAttr("width", ""); ok {
					width = w
				}
				if h, ok := geo.GetAttr("height", ""); ok {
					height = h
				}
			}
			result[id] = DeliverableAppearance{Value: value, Style: styleStr, Width: width, Height: height}
		}, nil)
	}
	return result
}

func columnYFor(i int) float64 {
	return scaffoldStartY + float64(i)*scaffoldRowStep
}

func placeBoundaries(root *xmldom.Node, cp pfd.NodeID, inputs, outputs []pfd.NodeID, appearances map[pfd.NodeID]DeliverableAppearance, inputX, outputX float64, yFor func(i int) float64, layer string, usedCellIDs map[string]struct{}) error {
	if err := placeStack(root, cp, inputs, appearances, inputX, yFor, layer, usedCellIDs); err != nil {
		return err
	}
	return placeStack(root, cp, outputs, appearances, outputX, yFor, layer, usedCellIDs)
}

func placeStack(root *xmldom.Node, cp pfd.NodeID, ids []pfd.NodeID, appearances map[pfd.NodeID]DeliverableAppearance, x float64, yFor func(i int) float64, layer string, usedCellIDs map[string]struct{}) error {
	for i, id := range ids {
		app, ok := appearances[id]
		if !ok {
			return fmt.Errorf("pfddrawio.CompleteCompositePages: 複合プロセス %s の境界成果物 %s に対応する頂点セルが見つかりません", cp, id)
		}

		cellID := uniqueID(fmt.Sprintf("cpcomp-%s-%s", cp, id), usedCellIDs)
		usedCellIDs[cellID] = struct{}{}
		cell := buildDeliverableCell(cellID, app, layer, x, yFor(i))
		root.AppendChildIndented(indentCell, cell)
	}
	return nil
}

func buildDeliverableCell(id string, app DeliverableAppearance, parent string, x, y float64) *xmldom.Node {
	cell := xmldom.NewElement("mxCell")
	cell.SetAttr("id", id)
	cell.SetAttr("value", app.Value)
	cell.SetAttr("style", app.Style)
	cell.SetAttr("parent", parent)
	cell.SetAttr("vertex", "1")

	geo := xmldom.NewElement("mxGeometry")
	geo.SetAttr("x", formatCoord(x))
	geo.SetAttr("y", formatCoord(y))
	geo.SetAttr("width", app.Width)
	geo.SetAttr("height", app.Height)
	geo.SetAttr("as", "geometry")

	cell.Children = []*xmldom.Node{
		xmldom.NewText(indentGeom),
		geo,
		xmldom.NewText(indentCell),
	}
	return cell
}

func buildScaffoldPage(diagramID, pageName string) (*xmldom.Node, *xmldom.Node, error) {
	const tmpl = `<diagram id="__ID__" name="__NAME__">
        <mxGraphModel dx="800" dy="600" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="827" pageHeight="1169" math="0" shadow="0">
            <root>
                <mxCell id="0" />
                <mxCell id="1" value="PFD" parent="0" />
            </root>
        </mxGraphModel>
    </diagram>`

	parsed, err := xmldom.ParseXML(strings.NewReader(tmpl))
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.buildScaffoldPage: %w", err)
	}

	var diagram *xmldom.Node
	for _, n := range parsed {
		if n.Kind == xmldom.ElementNode && n.Start.Name.Local == "diagram" {
			diagram = n
			break
		}
	}
	if diagram == nil {
		return nil, nil, fmt.Errorf("pfddrawio.buildScaffoldPage: テンプレートに diagram がありません")
	}
	diagram.SetAttr("id", diagramID)
	diagram.SetAttr("name", pageName)

	model := diagram.FirstChildElement("mxGraphModel")
	if model == nil {
		return nil, nil, fmt.Errorf("pfddrawio.buildScaffoldPage: mxGraphModel がありません")
	}
	root := model.FirstChildElement("root")
	if root == nil {
		return nil, nil, fmt.Errorf("pfddrawio.buildScaffoldPage: root がありません")
	}
	return diagram, root, nil
}

func FindMxfile(nodes []*xmldom.Node) *xmldom.Node {
	for _, n := range nodes {
		if n.Kind == xmldom.ElementNode && n.Start.Name.Local == "mxfile" {
			return n
		}
	}
	return nil
}

func diagramNodesByPageID(mxfile *xmldom.Node) map[pfd.NodeID]*xmldom.Node {
	result := make(map[pfd.NodeID]*xmldom.Node)
	for _, child := range mxfile.Children {
		if child.Kind != xmldom.ElementNode || child.Start.Name.Local != "diagram" {
			continue
		}
		name, _ := child.GetAttr("name", "")
		id := PageNameElementID(name)
		if _, exists := result[id]; !exists {
			result[id] = child
		}
	}
	return result
}

func diagramIDSet(mxfile *xmldom.Node) map[string]struct{} {
	result := make(map[string]struct{})
	for _, child := range mxfile.Children {
		if child.Kind != xmldom.ElementNode || child.Start.Name.Local != "diagram" {
			continue
		}
		if id, ok := child.GetAttr("id", ""); ok {
			result[id] = struct{}{}
		}
	}
	return result
}

func uniqueID(base string, used map[string]struct{}) string {
	if _, exists := used[base]; !exists {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func isDeliverable(nodeMap map[pfd.NodeID]*pfd.Node, id pfd.NodeID) bool {
	n, ok := nodeMap[id]
	return ok && n.Type.IsDeliverable()
}

func sortNodeIDs(ids []pfd.NodeID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i].Compare(ids[j]) < 0 })
}

func formatCoord(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

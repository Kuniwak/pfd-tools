package pfddrawio

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

type LayerMap map[CellID]string

func NewLayerMap(cells []Cell) LayerMap {
	layerMap := make(LayerMap)
	for _, cell := range cells {
		if !cell.IsLayer {
			continue
		}
		layerMap[cell.ID] = cell.Value
	}
	return layerMap
}

func NewLayerMapFromNodes(nodes []*xmldom.Node) LayerMap {
	layerMap := make(LayerMap)
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			if parent, _ := n.GetAttr("parent", ""); parent != "0" {
				return
			}
			id, _ := n.GetAttr("id", "")
			value, _ := n.GetAttr("value", "")
			layerMap[CellID(id)] = value
		}, nil)
	}
	return layerMap
}

type ParentMap map[CellID]CellID

func NewParentMap(cells []Cell) ParentMap {
	parents := make(ParentMap)
	for _, cell := range cells {
		parents[cell.ID] = cell.Parent
	}
	return parents
}

func NewParentMapFromNodes(nodes []*xmldom.Node) ParentMap {
	parents := make(ParentMap)
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			id, _ := n.GetAttr("id", "")
			parent, _ := n.GetAttr("parent", "")
			parents[CellID(id)] = CellID(parent)
		}, nil)
	}
	return parents
}

const CommentLayerNamePrefixEn = "comment"
const CommentLayerNamePrefixJa = "コメント"

func (l LayerMap) IsCommentLayer(id CellID) bool {
	value, ok := l[id]
	if !ok {
		return false
	}
	lower := strings.ToLower(value)
	return strings.HasPrefix(lower, CommentLayerNamePrefixEn) ||
		strings.HasPrefix(lower, CommentLayerNamePrefixJa)
}

func (l LayerMap) IsCommentDescendant(parents ParentMap, id CellID) bool {
	seen := make(map[CellID]struct{})
	for cur := id; cur != ""; {
		if _, dup := seen[cur]; dup {
			return false
		}
		seen[cur] = struct{}{}
		parent, ok := parents[cur]
		if !ok {
			return false
		}
		if l.IsCommentLayer(parent) {
			return true
		}
		cur = parent
	}
	return false
}

type NormalizeOptions struct {
	AllowDetachedDetailPage bool
}

func NormalizeDiagrams(title string, diagrams []Diagram, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	return NormalizeDiagramsWithOptions(title, diagrams, NormalizeOptions{}, logger)
}

func NormalizeDiagramsWithOptions(title string, diagrams []Diagram, opts NormalizeOptions, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
	p := &pfd.PFD{
		Title:                  title,
		Nodes:                  sets.New((*pfd.Node).Compare),
		Edges:                  sets.New((*pfd.Edge).Compare),
		ProcessComposition:     make(map[pfd.NodeID]*sets.Set[pfd.NodeID]),
		DeliverableComposition: make(map[pfd.NodeID]*sets.Set[pfd.NodeID]),
	}

	srcMap := &SourceMap{
		NodeIDMap: make(map[pfd.NodeID]*sets.Set[DrawIOLocation]),
		EdgeIDMap: make(map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation]),
	}
	diagramIDConv := make(map[DiagramID]map[CellID]pfd.NodeID)

	detailPageNames := DetailPageNames(diagrams)

	implicitAtomics := sets.New(pfd.NodeID.Compare)

	nodeTypes := make(NodeTypeMap)

	contextDiagramName := ""
	hasContextDiagram := false

	for _, diagram := range diagrams {
		idconv := make(map[CellID]pfd.NodeID)
		layerMap := NewLayerMap(diagram.Cells)
		parents := NewParentMap(diagram.Cells)

		compID := PageNameElementID(diagram.Name)
		isContextDiagram := IsContextDiagramName(diagram.Name)
		if isContextDiagram {

			if hasContextDiagram {
				return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: コンテキスト図のページが複数あります（%q と %q）。コンテキスト図は 1 つの PFD に 1 つだけです。詳細ページの名前は内訳を示す複合プロセスの ID（例: %q）にしてください", contextDiagramName, diagram.Name, pfd.NewCompositeProcessID(1))
			}
			contextDiagramName = diagram.Name
			hasContextDiagram = true
		} else {

			if _, ok := p.ProcessComposition[compID]; ok {
				return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: 詳細ページ名 %q が指す複合プロセス %q のページが複数あります。ページ名は内訳を示す複合プロセスの ID なので 1 つの PFD に 1 つだけです（ページ名に書いた説明が違っていても、ID 部が同じなら同じ複合プロセスを指します）。複合プロセスの内訳のバリエーションを結合した場合は、バリエーションごとに分けて扱ってください", diagram.Name, compID)
			}

			p.ProcessComposition[compID] = sets.New(pfd.NodeID.Compare)
		}

		for _, cell := range diagram.Cells {
			if cell.IsLayer {
				continue
			}

			if layerMap.IsCommentDescendant(parents, cell.ID) {
				continue
			}

			if cell.IsVertex {
				if cell.IsConnector() {

					continue
				}
				if cell.Style.IsRectangle() {
					isAtomicDeliverable := cell.Style.IsAtomicStrokeWidth()

					var t pfd.NodeType
					if isAtomicDeliverable {
						t = pfd.NodeTypeAtomicDeliverable
					} else {
						t = pfd.NodeTypeCompositeDeliverable
					}

					rawID, desc, err := ParseVertexValue(cell.Value)
					if err != nil {
						logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: cannot parse rectangle value", "diagramID", diagram.ID, "cellID", cell.ID)
						continue
					}
					id := VertexElementID(rawID, desc, t)

					if err := nodeTypes.Add(id, t, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID}); err != nil {
						return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: %w", err)
					}

					p.Nodes.Add((*pfd.Node).Compare, &pfd.Node{ID: id, Description: desc, Type: t})

					idconv[cell.ID] = id

					srcMap.AddNodeLocation(id, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
				} else if cell.Style.IsEllipse() {
					rawID, desc, err := ParseVertexValue(cell.Value)
					if err != nil {
						logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: cannot parse ellipse value", "diagramID", diagram.ID, "cellID", cell.ID)
						continue
					}

					id := VertexElementID(rawID, desc, pfd.NodeTypeAtomicProcess)

					var t pfd.NodeType
					if cell.Style.IsAtomicStrokeWidth() {
						t = pfd.NodeTypeAtomicProcess
					} else if detailPageNames.Contains(pfd.NodeID.Compare, id) {
						t = pfd.NodeTypeCompositeProcess
					} else {
						t = pfd.NodeTypeAtomicProcess
						implicitAtomics.Add(pfd.NodeID.Compare, id)
					}

					if err := nodeTypes.Add(id, t, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID}); err != nil {
						return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: %w", err)
					}

					p.Nodes.Add((*pfd.Node).Compare, &pfd.Node{ID: id, Description: desc, Type: t})
					if !isContextDiagram {
						p.ProcessComposition[compID].Add(pfd.NodeID.Compare, id)
					}

					idconv[cell.ID] = id

					srcMap.AddNodeLocation(id, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
				}
			}
		}

		if _, ok := diagramIDConv[diagram.ID]; ok {
			return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: ページ id %q が重複しています。複数のファイルを結合する場合は、ページ id を一意化する drawiocat を使ってください", diagram.ID)
		}
		diagramIDConv[diagram.ID] = idconv
	}

	nodeMap := pfd.NewNodeMap(p.Nodes, logger)
	for _, diagram := range diagrams {
		if IsContextDiagramName(diagram.Name) {
			continue
		}
		node, ok := nodeMap[PageNameElementID(diagram.Name)]
		if !ok && opts.AllowDetachedDetailPage {

			logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: ページ名が指す要素が PFD 中にありません（断片として読んでいるため続行します）", "page", diagram.Name)
			continue
		}
		if !ok || node.Type != pfd.NodeTypeCompositeProcess {
			return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: ページ名 %q が不正です。ページ名はコンテキスト図では %q、詳細図では図中に存在する複合プロセスの ID（例: %q）にしてください。ID の後ろには %q のように説明を書けます", diagram.Name, pfd.NodeIDContextDiagram, pfd.NewCompositeProcessID(1), FormatVertexValue(pfd.NewCompositeProcessID(1), "設計する"))
		}

		_, pageDesc, _ := ParseVertexValue(diagram.Name)
		if pageDesc != "" && node.Description != "" && pageDesc != node.Description {
			logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: ページ名の説明が複合プロセスのラベルと一致しません", "page", diagram.Name, "id", node.ID, "pageDesc", pageDesc, "nodeDesc", node.Description)
		}
	}

	flattened, err := pfd.FlattenProcessComposition(p.ProcessComposition, nodeMap)
	if err != nil {
		return nil, nil, fmt.Errorf("pfddrawio.NormalizeDiagramsWithOptions: %w", err)
	}
	p.ProcessComposition = flattened

	for _, diagram := range diagrams {
		idconv, ok := diagramIDConv[diagram.ID]
		if !ok {

			panic(fmt.Sprintf("pfddrawio.NormalizeDiagramsWithOptions: missing diagram ID: %q", diagram.ID))
		}

		layerMap := NewLayerMap(diagram.Cells)
		parents := NewParentMap(diagram.Cells)

		nonComment := make([]Cell, 0, len(diagram.Cells))
		for _, cell := range diagram.Cells {
			if layerMap.IsCommentDescendant(parents, cell.ID) {
				continue
			}
			nonComment = append(nonComment, cell)
		}

		for _, connID := range DanglingConnectors(nonComment) {
			logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: dangling connector (has only incoming or only outgoing edges); its edges are dropped", "diagramID", diagram.ID, "cellID", connID)
		}

		for _, eff := range SpliceConnectors(nonComment) {
			src, ok := idconv[eff.Source]
			if !ok {
				logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: missing source node", "diagramID", diagram.ID, "cellIDs", eff.Origins)
				continue
			}
			target, ok := idconv[eff.Target]
			if !ok {
				logger.Warn("pfddrawio.NormalizeDiagramsWithOptions: missing target node", "diagramID", diagram.ID, "cellIDs", eff.Origins)
				continue
			}
			p.Edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: src, Target: target, IsFeedback: eff.IsFeedback})

			for _, origin := range eff.Origins {
				srcMap.AddEdgeLocation(src, target, DrawIOLocation{DiagramID: diagram.ID, CellID: origin})
			}
		}
	}

	if implicitAtomics.Len() > 0 {
		p.ImplicitAtomicProcesses = implicitAtomics
	}

	return p, srcMap, nil
}

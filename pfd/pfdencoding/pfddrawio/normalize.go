package pfddrawio

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
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

func (l LayerMap) IsCommentLayer(id CellID) bool {
	value, ok := l[id]
	if !ok {
		return false
	}
	return strings.HasPrefix(strings.ToLower(value), "comment")
}

func NormalizeDiagrams(title string, diagrams []Diagram, logger *slog.Logger) (*pfd.PFD, *SourceMap, error) {
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

	for _, diagram := range diagrams {
		idconv := make(map[CellID]pfd.NodeID)
		layerMap := NewLayerMap(diagram.Cells)

		compID := pfd.NodeID(diagram.Name)
		isContextDiagram := diagram.Name == string(pfd.NodeIDContextDiagram) || diagram.Name == DefaultTopPageNameEn || diagram.Name == DefaultTopPageNameJa
		if !isContextDiagram {
			// NOTE: P0 is self-evident, so we don't keep it as data.
			p.ProcessComposition[compID] = sets.New(pfd.NodeID.Compare)
		}

		for _, cell := range diagram.Cells {
			if cell.IsLayer {
				continue
			}

			if layerMap.IsCommentLayer(cell.Parent) {
				continue
			}

			if cell.IsVertex {
				if cell.Style.IsRectangle() {
					isAtomicDeliverable := cell.Style.StrokeWidth() == 1

					var t pfd.NodeType
					if isAtomicDeliverable {
						t = pfd.NodeTypeAtomicDeliverable
					} else {
						t = pfd.NodeTypeCompositeDeliverable
					}

					id, desc, err := ParseVertexValue(cell.Value)
					if err != nil {
						logger.Warn("pfddrawio.NormalizeDiagrams: cannot parse rectangle value", "diagramID", diagram.ID, "cellID", cell.ID)
						continue
					}

					p.Nodes.Add((*pfd.Node).Compare, &pfd.Node{ID: id, Description: desc, Type: t})

					idconv[cell.ID] = id

					if srcMapEntry, ok := srcMap.NodeIDMap[id]; ok {
						srcMapEntry.Add(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
					} else {
						srcMap.NodeIDMap[id] = sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
					}
				} else if cell.Style.IsEllipse() {
					strokeWidth := cell.Style.StrokeWidth()
					isAtomicProcess := strokeWidth == 1

					var t pfd.NodeType
					if isAtomicProcess {
						t = pfd.NodeTypeAtomicProcess
					} else {
						t = pfd.NodeTypeCompositeProcess
					}

					id, desc, err := ParseVertexValue(cell.Value)
					if err != nil {
						logger.Warn("pfddrawio.NormalizeDiagrams: cannot parse ellipse value", "diagramID", diagram.ID, "cellID", cell.ID)
						continue
					}

					p.Nodes.Add((*pfd.Node).Compare, &pfd.Node{ID: id, Description: desc, Type: t})
					if compID != pfd.NodeIDContextDiagram {
						p.ProcessComposition[compID].Add(pfd.NodeID.Compare, id)
					}

					idconv[cell.ID] = id

					if srcMapEntry, ok := srcMap.NodeIDMap[id]; ok {
						srcMapEntry.Add(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
					} else {
						srcMap.NodeIDMap[id] = sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
					}
				}
			}
		}
		if _, ok := diagramIDConv[diagram.ID]; ok {
			panic(fmt.Sprintf("pfddrawio.NormalizeDiagrams: duplicate diagram ID: %q", diagram.ID))
		}
		diagramIDConv[diagram.ID] = idconv
	}

	for _, diagram := range diagrams {
		idconv, ok := diagramIDConv[diagram.ID]
		if !ok {
			panic(fmt.Sprintf("pfddrawio.NormalizeDiagrams: missing diagram ID: %q", diagram.ID))
		}

		layerMap := NewLayerMap(diagram.Cells)

		for _, cell := range diagram.Cells {
			if layerMap.IsCommentLayer(cell.Parent) {
				continue
			}

			if cell.IsEdge {
				isFeedback := cell.Style.IsDashed()

				src, ok := idconv[cell.Source]
				if !ok {
					logger.Warn("pfddrawio.NormalizeDiagrams: missing source node", "diagramID", diagram.ID, "cellID", cell.ID)
					continue
				}
				target, ok := idconv[cell.Target]
				if !ok {
					logger.Warn("pfddrawio.NormalizeDiagrams: missing target node", "diagramID", diagram.ID, "cellID", cell.ID)
					continue
				}
				p.Edges.Add((*pfd.Edge).Compare, &pfd.Edge{Source: src, Target: target, IsFeedback: isFeedback})

				if e1, ok := srcMap.EdgeIDMap[src]; ok {
					if e2, ok := e1[target]; ok {
						e2.Add(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
					} else {
						e1[target] = sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID})
					}
				} else {
					srcMap.EdgeIDMap[src] = map[pfd.NodeID]*sets.Set[DrawIOLocation]{
						target: sets.New(DrawIOLocation.Compare, DrawIOLocation{DiagramID: diagram.ID, CellID: cell.ID}),
					}
				}
			}
		}
	}

	return p, srcMap, nil
}

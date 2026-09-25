package pfddrawio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

type DiagramID string

type Diagram struct {
	ID    DiagramID `json:"id"`
	Name  string    `json:"name"`
	Cells []Cell    `json:"cells"`
}

type DrawIOLocation struct {
	DiagramID DiagramID `json:"diagramID"`
	CellID    CellID    `json:"cellID"`
}

func (a DrawIOLocation) Compare(b DrawIOLocation) int {
	cmp1 := strings.Compare(string(a.DiagramID), string(b.DiagramID))
	if cmp1 != 0 {
		return cmp1
	}
	return strings.Compare(string(a.CellID), string(b.CellID))
}

type NodeTypeMap map[pfd.NodeID]NodeTypeAt

type NodeTypeAt struct {
	Type     pfd.NodeType   `json:"type"`
	Location DrawIOLocation `json:"location"`
}

func (m NodeTypeMap) Add(id pfd.NodeID, t pfd.NodeType, loc DrawIOLocation) error {
	prev, ok := m[id]
	if !ok {
		m[id] = NodeTypeAt{Type: t, Location: loc}
		return nil
	}
	if prev.Type != t {

		return NewCellErrorByMessage(
			fmt.Sprintf(
				"pfddrawio.NodeTypeMap.Add: 要素 ID %q が「%s」と「%s」の両方で使われています。1 つの ID は 1 つの要素なので、どちらかの ID か見た目を直してください",
				id, NodeTypeShapeJa(prev.Type), NodeTypeShapeJa(t),
			),
			prev.Location, loc,
		)
	}
	return nil
}

func NodeTypeShapeJa(t pfd.NodeType) string {
	switch t {
	case pfd.NodeTypeAtomicDeliverable:
		return "細線の四角（原子成果物）"
	case pfd.NodeTypeCompositeDeliverable:
		return "太線の四角（複合成果物）"
	case pfd.NodeTypeAtomicProcess:
		return "細線の楕円（原子プロセス）"
	case pfd.NodeTypeCompositeProcess:
		return "太線の楕円（複合プロセス）"
	default:
		return string(t)
	}
}

type SourceMap struct {
	NodeIDMap map[pfd.NodeID]*sets.Set[DrawIOLocation]                `json:"nodeIDMap"`
	EdgeIDMap map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation] `json:"edgeIDMap"`
}

func (m *SourceMap) AddNodeLocation(id pfd.NodeID, loc DrawIOLocation) {
	if entry, ok := m.NodeIDMap[id]; ok {
		entry.Add(DrawIOLocation.Compare, loc)
	} else {
		m.NodeIDMap[id] = sets.New(DrawIOLocation.Compare, loc)
	}
}

func (m *SourceMap) AddEdgeLocation(src, target pfd.NodeID, loc DrawIOLocation) {
	targets, ok := m.EdgeIDMap[src]
	if !ok {
		targets = make(map[pfd.NodeID]*sets.Set[DrawIOLocation])
		m.EdgeIDMap[src] = targets
	}
	if entry, ok := targets[target]; ok {
		entry.Add(DrawIOLocation.Compare, loc)
	} else {
		targets[target] = sets.New(DrawIOLocation.Compare, loc)
	}
}

type CellID string

func (a CellID) Compare(b CellID) int {
	return graph.Node(a).Compare(graph.Node(b))
}

type Cell struct {
	ID       CellID   `json:"id"`
	Parent   CellID   `json:"parent,omitempty"`
	Value    string   `json:"value,omitempty"`
	Style    StyleMap `json:"style,omitempty"`
	Source   CellID   `json:"source,omitempty"`
	Target   CellID   `json:"target,omitempty"`
	IsEdge   bool     `json:"isEdge,omitempty"`
	IsVertex bool     `json:"isVertex,omitempty"`
	IsLayer  bool     `json:"isLayer,omitempty"`
	IsRoot   bool     `json:"isRoot,omitempty"`
}

func (c Cell) IsConnector() bool {
	return c.IsVertex && c.Value == "" && c.Style.IsEllipse() && c.Style.Get("aspect") == "fixed"
}

func NewRoot(id CellID) Cell {
	return Cell{ID: id, IsRoot: true}
}

func NewLayer(id CellID, value string) Cell {
	return Cell{
		ID:      id,
		Value:   value,
		IsLayer: true,
	}
}

func NewVertex(id CellID, parent CellID, value string, style StyleMap) Cell {
	return Cell{
		ID:       id,
		Parent:   parent,
		Style:    style,
		Value:    value,
		IsVertex: true,
	}
}

func NewEdge(id CellID, parent CellID, source CellID, target CellID, style StyleMap) Cell {
	return Cell{
		ID:     id,
		Parent: parent,
		Style:  style,
		Source: source,
		Target: target,
		IsEdge: true,
	}
}

type Style string
type StyleMap map[string]string

func (s StyleMap) Get(key string) string {
	return s[key]
}

func (s StyleMap) IsText() bool {
	text, ok := s["text"]
	return ok && text == ""
}

func (s StyleMap) IsRectangle() bool {
	rounded, ok := s["rounded"]
	return ok && rounded == "0" && !s.IsText()
}

func (s StyleMap) IsEllipse() bool {
	ellipse, ok := s["ellipse"]
	return ok && ellipse == ""
}

func (s StyleMap) StrokeWidth() int {
	width, ok := s["strokeWidth"]
	if !ok {
		return 1
	}
	widthInt, err := strconv.Atoi(width)
	if err != nil {
		return 1
	}
	return widthInt
}

func (s StyleMap) IsAtomicStrokeWidth() bool {
	return s.StrokeWidth() <= 1
}

func (s StyleMap) IsDashed() bool {
	dashed, ok := s["dashed"]
	return ok && dashed != ""
}

type ValueHTML string

const DefaultTopPageNameEn = "Page-1"
const DefaultTopPageNameJa = "ページ-1"

func IsContextDiagramName(name string) bool {

	id, _, _ := ParseVertexValue(name)
	return id == pfd.NodeIDContextDiagram ||
		id == DefaultTopPageNameEn ||
		id == DefaultTopPageNameJa
}

func DetailPageNames(diagrams []Diagram) *sets.Set[pfd.NodeID] {
	names := sets.New(pfd.NodeID.Compare)
	for _, diagram := range diagrams {
		if IsContextDiagramName(diagram.Name) {
			continue
		}
		names.Add(pfd.NodeID.Compare, PageNameElementID(diagram.Name))
	}
	return names
}

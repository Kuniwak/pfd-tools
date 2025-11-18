package pfddrawio

import (
	"strconv"
	"strings"

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

type SourceMap struct {
	NodeIDMap map[pfd.NodeID]*sets.Set[DrawIOLocation]                `json:"nodeIDMap"`
	EdgeIDMap map[pfd.NodeID]map[pfd.NodeID]*sets.Set[DrawIOLocation] `json:"edgeIDMap"`
}

type CellID string

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

func (s StyleMap) IsDashed() bool {
	dashed, ok := s["dashed"]
	return ok && dashed != ""
}

type ValueHTML string

const DefaultTopPageNameEn = "Page-1"
const DefaultTopPageNameJa = "ページ1"

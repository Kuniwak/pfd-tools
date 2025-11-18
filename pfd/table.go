package pfd

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/sets"
)

type TableType string

const (
	TableTypeAtomicProcess        TableType = "ATOMIC_PROCESS"
	TableTypeAtomicDeliverable    TableType = "ATOMIC_DELIVERABLE"
	TableTypeCompositeProcess     TableType = "COMPOSITE_PROCESS"
	TableTypeCompositeDeliverable TableType = "COMPOSITE_DELIVERABLE"
)

type AtomicProcessRow struct {
	ID          AtomicProcessID `json:"id"`
	Description string          `json:"description"`
	ExtraCells  []string        `json:"extra_cells"`
}

func (p *AtomicProcessRow) Clone() *AtomicProcessRow {
	return &AtomicProcessRow{ID: p.ID, Description: p.Description, ExtraCells: slices.Clone(p.ExtraCells)}
}

func (p *AtomicProcessRow) Compare(b *AtomicProcessRow) int {
	cmp1 := p.ID.Compare(b.ID)
	if cmp1 != 0 {
		return cmp1
	}
	return strings.Compare(p.Description, b.Description)
}

type AtomicProcessTable struct {
	ExtraHeaders []string            `json:"extra_headers"`
	Rows         []*AtomicProcessRow `json:"rows"`
}

func NewAtomicProcessTable(p *PFD, nodeMap map[NodeID]*Node) *AtomicProcessTable {
	rows := make([]*AtomicProcessRow, 0, p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type != NodeTypeAtomicProcess {
			continue
		}
		ap := AtomicProcessIDFromNodeID(node.ID, nodeMap)
		rows = append(rows, &AtomicProcessRow{ID: ap, Description: node.Description})
	}
	return &AtomicProcessTable{ExtraHeaders: []string{}, Rows: rows}
}

func (p *AtomicProcessTable) Clone() *AtomicProcessTable {
	rows := make([]*AtomicProcessRow, 0, len(p.Rows))
	for _, row := range p.Rows {
		rows = append(rows, row.Clone())
	}
	return &AtomicProcessTable{ExtraHeaders: slices.Clone(p.ExtraHeaders), Rows: rows}
}

func (p *AtomicProcessTable) AtomicProcesses() *sets.Set[AtomicProcessID] {
	s := sets.New(AtomicProcessID.Compare)
	for _, row := range p.Rows {
		s.Add(AtomicProcessID.Compare, row.ID)
	}
	return s
}

func (p *AtomicProcessTable) Header() []string {
	return append([]string{"ID", "Description"}, p.ExtraHeaders...)
}

func (p *AtomicProcessTable) Refresh(pfd *PFD, nodeMap map[NodeID]*Node) {
	rows := make([]*AtomicProcessRow, 0, pfd.Nodes.Len())

	actual := sets.NewWithCapacity[AtomicProcessID](len(p.Rows))
	for _, row := range p.Rows {
		actual.Add(AtomicProcessID.Compare, row.ID)
	}

	expected := sets.NewWithCapacity[AtomicProcessID](pfd.Nodes.Len())
	for _, node := range pfd.Nodes.Iter() {
		if node.Type != NodeTypeAtomicProcess {
			continue
		}
		ap := AtomicProcessIDFromNodeID(node.ID, nodeMap)
		expected.Add(AtomicProcessID.Compare, ap)
	}

	missings := expected.Clone()
	missings.Difference(AtomicProcessID.Compare, actual)

	extras := actual.Clone()
	extras.Difference(AtomicProcessID.Compare, expected)

	for _, row := range p.Rows {
		if extras.Contains(AtomicProcessID.Compare, row.ID) {
			continue
		}
		rows = append(rows, row)
	}

	for _, missing := range missings.Iter() {
		rows = append(rows, &AtomicProcessRow{ID: missing, Description: nodeMap[NodeID(missing)].Description, ExtraCells: make([]string, len(p.ExtraHeaders))})
	}

	slices.SortFunc(rows, (*AtomicProcessRow).Compare)

	p.Rows = rows
}

type AtomicDeliverableRow struct {
	ID          AtomicDeliverableID `json:"id"`
	Description string              `json:"description"`
	ExtraCells  []string            `json:"extra_cells"`
}

func (p *AtomicDeliverableRow) Clone() *AtomicDeliverableRow {
	return &AtomicDeliverableRow{ID: p.ID, Description: p.Description, ExtraCells: slices.Clone(p.ExtraCells)}
}

func (p *AtomicDeliverableRow) Compare(b *AtomicDeliverableRow) int {
	cmp1 := p.ID.Compare(b.ID)
	if cmp1 != 0 {
		return cmp1
	}
	return strings.Compare(p.Description, b.Description)
}

type AtomicDeliverableTable struct {
	ExtraHeaders []string                `json:"extra_headers"`
	Rows         []*AtomicDeliverableRow `json:"rows"`
}

func NewAtomicDeliverableTable(p *PFD, nodeMap map[NodeID]*Node) *AtomicDeliverableTable {
	rows := make([]*AtomicDeliverableRow, 0, p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type != NodeTypeAtomicDeliverable {
			continue
		}
		d := AtomicDeliverableIDFromNodeID(node.ID, nodeMap)
		rows = append(rows, &AtomicDeliverableRow{ID: d, Description: node.Description})
	}
	return &AtomicDeliverableTable{ExtraHeaders: []string{}, Rows: rows}
}

func (p *AtomicDeliverableTable) Clone() *AtomicDeliverableTable {
	rows := make([]*AtomicDeliverableRow, 0, len(p.Rows))
	for _, row := range p.Rows {
		rows = append(rows, row.Clone())
	}
	return &AtomicDeliverableTable{ExtraHeaders: slices.Clone(p.ExtraHeaders), Rows: rows}
}

func (p *AtomicDeliverableTable) Header() []string {
	return append([]string{"ID", "Description"}, p.ExtraHeaders...)
}

func (p *AtomicDeliverableTable) Refresh(pfd *PFD, nodeMap map[NodeID]*Node) {
	rows := make([]*AtomicDeliverableRow, 0, pfd.Nodes.Len())

	actual := sets.NewWithCapacity[AtomicDeliverableID](len(p.Rows))
	for _, row := range p.Rows {
		actual.Add(AtomicDeliverableID.Compare, row.ID)
	}

	expected := sets.NewWithCapacity[AtomicDeliverableID](pfd.Nodes.Len())
	for _, node := range pfd.Nodes.Iter() {
		if node.Type != NodeTypeAtomicDeliverable {
			continue
		}
		d := AtomicDeliverableIDFromNodeID(node.ID, nodeMap)
		expected.Add(AtomicDeliverableID.Compare, d)
	}

	extras := actual.Clone()
	extras.Difference(AtomicDeliverableID.Compare, expected)

	missings := expected.Clone()
	missings.Difference(AtomicDeliverableID.Compare, actual)

	for _, row := range p.Rows {
		if extras.Contains(AtomicDeliverableID.Compare, row.ID) {
			continue
		}
		rows = append(rows, row)
	}

	for _, missing := range missings.Iter() {
		rows = append(rows, &AtomicDeliverableRow{ID: missing, Description: nodeMap[NodeID(missing)].Description, ExtraCells: make([]string, len(p.ExtraHeaders))})
	}

	slices.SortFunc(rows, (*AtomicDeliverableRow).Compare)

	p.Rows = rows
}

type CompositeDeliverableTable struct {
	ExtraHeaders []string                   `json:"extra_headers"`
	Rows         []*CompositeDeliverableRow `json:"rows"`
}

func NewCompositeDeliverableTable(p *PFD, nodeMap map[NodeID]*Node) *CompositeDeliverableTable {
	rows := make([]*CompositeDeliverableRow, 0, p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type != NodeTypeCompositeDeliverable {
			continue
		}
		cd := CompositeDeliverableIDFromNodeID(node.ID, nodeMap)
		rows = append(rows, &CompositeDeliverableRow{ID: cd, Description: node.Description})
	}
	return &CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: rows}
}

func (p *CompositeDeliverableTable) Header() []string {
	return append([]string{"ID", "Description", "Deliverables"}, p.ExtraHeaders...)
}

func (p *CompositeDeliverableTable) NodeIDMap(logger *slog.Logger) map[NodeID]*sets.Set[NodeID] {
	m := make(map[NodeID]*sets.Set[NodeID])
	for _, row := range p.Rows {
		if _, ok := m[NodeID(row.ID)]; ok {
			logger.Warn("CompositeDeliverableTable.Map: duplicate composite deliverable ID", "id", row.ID)
			continue
		}
		s := sets.NewWithCapacity[NodeID](len(row.Deliverables))
		for _, deliverable := range row.Deliverables {
			s.Add(NodeID.Compare, NodeID(deliverable))
		}
		m[NodeID(row.ID)] = s
	}
	return m
}

func (p *CompositeDeliverableTable) Clone() *CompositeDeliverableTable {
	rows := make([]*CompositeDeliverableRow, 0, len(p.Rows))
	for _, row := range p.Rows {
		rows = append(rows, row.Clone())
	}
	return &CompositeDeliverableTable{ExtraHeaders: slices.Clone(p.ExtraHeaders), Rows: rows}
}

func (p *CompositeDeliverableTable) Refresh(pfd *PFD, nodeMap map[NodeID]*Node) {
	newRows := make([]*CompositeDeliverableRow, 0, pfd.Nodes.Len())

	actual := sets.NewWithCapacity[CompositeDeliverableID](len(p.Rows))
	for _, row := range p.Rows {
		actual.Add(CompositeDeliverableID.Compare, row.ID)
	}

	expected := sets.NewWithCapacity[CompositeDeliverableID](pfd.Nodes.Len())
	for _, node := range pfd.Nodes.Iter() {
		if node.Type != NodeTypeCompositeDeliverable {
			continue
		}
		cd := CompositeDeliverableIDFromNodeID(node.ID, nodeMap)
		expected.Add(CompositeDeliverableID.Compare, cd)
	}

	extras := actual.Clone()
	extras.Difference(CompositeDeliverableID.Compare, expected)

	missings := expected.Clone()
	missings.Difference(CompositeDeliverableID.Compare, actual)

	for _, row := range p.Rows {
		if extras.Contains(CompositeDeliverableID.Compare, row.ID) {
			continue
		}
		newRows = append(newRows, row)
	}

	for _, missing := range missings.Iter() {
		newRows = append(newRows, &CompositeDeliverableRow{ID: missing, Description: nodeMap[NodeID(missing)].Description, Deliverables: make([]AtomicDeliverableID, 0), ExtraCells: make([]string, len(p.ExtraHeaders))})
	}

	slices.SortFunc(newRows, (*CompositeDeliverableRow).Compare)
	p.Rows = newRows
}

type CompositeDeliverableRow struct {
	ID           CompositeDeliverableID `json:"composite_deliverable"`
	Description  string                 `json:"description"`
	Deliverables []AtomicDeliverableID  `json:"deliverables"`
	ExtraCells   []string               `json:"extra_cells"`
}

func (p *CompositeDeliverableRow) Clone() *CompositeDeliverableRow {
	return &CompositeDeliverableRow{ID: p.ID, Description: p.Description, Deliverables: slices.Clone(p.Deliverables), ExtraCells: slices.Clone(p.ExtraCells)}
}

func (p *CompositeDeliverableRow) Compare(b *CompositeDeliverableRow) int {
	c := p.ID.Compare(b.ID)
	if c != 0 {
		return c
	}
	c = slices.CompareFunc(p.Deliverables, b.Deliverables, AtomicDeliverableID.Compare)
	if c != 0 {
		return c
	}
	c = strings.Compare(p.Description, b.Description)
	if c != 0 {
		return c
	}
	return slices.CompareFunc(p.ExtraCells, b.ExtraCells, strings.Compare)
}

type CompositeProcessRow struct {
	ID          CompositeProcessID `json:"id"`
	Description string             `json:"description"`
	ExtraCells  []string           `json:"extra_cells"`
}

func (p *CompositeProcessRow) Clone() *CompositeProcessRow {
	return &CompositeProcessRow{ID: p.ID, Description: p.Description, ExtraCells: slices.Clone(p.ExtraCells)}
}

func (p *CompositeProcessRow) Compare(b *CompositeProcessRow) int {
	cmp1 := p.ID.Compare(b.ID)
	if cmp1 != 0 {
		return cmp1
	}
	return strings.Compare(p.Description, b.Description)
}

type CompositeProcessTable struct {
	ExtraHeaders []string               `json:"extra_headers"`
	Rows         []*CompositeProcessRow `json:"rows"`
}

func NewCompositeProcessTable(p *PFD, nodeMap map[NodeID]*Node) *CompositeProcessTable {
	rows := make([]*CompositeProcessRow, 0, p.Nodes.Len())
	for _, node := range p.Nodes.Iter() {
		if node.Type != NodeTypeCompositeProcess {
			continue
		}
		cp := CompositeProcessIDFromNodeID(node.ID, nodeMap)
		rows = append(rows, &CompositeProcessRow{ID: cp, Description: node.Description})
	}
	return &CompositeProcessTable{ExtraHeaders: []string{}, Rows: rows}
}

func (p *CompositeProcessTable) Clone() *CompositeProcessTable {
	rows := make([]*CompositeProcessRow, 0, len(p.Rows))
	for _, row := range p.Rows {
		rows = append(rows, row.Clone())
	}
	return &CompositeProcessTable{ExtraHeaders: slices.Clone(p.ExtraHeaders), Rows: rows}
}

func (p *CompositeProcessTable) Header() []string {
	return append([]string{"ID", "Description"}, p.ExtraHeaders...)
}

func (p *CompositeProcessTable) Refresh(pfd *PFD, nodeMap map[NodeID]*Node) {
	rows := make([]*CompositeProcessRow, 0, pfd.Nodes.Len())

	actual := sets.NewWithCapacity[CompositeProcessID](len(p.Rows))
	for _, row := range p.Rows {
		actual.Add(CompositeProcessID.Compare, row.ID)
	}

	expected := sets.NewWithCapacity[CompositeProcessID](pfd.Nodes.Len())
	for _, node := range pfd.Nodes.Iter() {
		if node.Type != NodeTypeCompositeProcess {
			continue
		}
		cp := CompositeProcessIDFromNodeID(node.ID, nodeMap)
		expected.Add(CompositeProcessID.Compare, cp)
	}

	extras := actual.Clone()
	extras.Difference(CompositeProcessID.Compare, expected)

	missings := expected.Clone()
	missings.Difference(CompositeProcessID.Compare, actual)

	for _, row := range p.Rows {
		if extras.Contains(CompositeProcessID.Compare, row.ID) {
			continue
		}
		rows = append(rows, row)
	}

	for _, missing := range missings.Iter() {
		rows = append(rows, &CompositeProcessRow{ID: missing, Description: nodeMap[NodeID(missing)].Description, ExtraCells: make([]string, len(p.ExtraHeaders))})
	}

	slices.SortFunc(rows, (*CompositeProcessRow).Compare)

	p.Rows = rows
}

// ColumnSelectFunc returns the corresponding index in ExtraCells from a header string slice. Returns -1 if there is no corresponding header string.
// If there are multiple identical header strings, which header string's index is returned is undefined.
type ColumnSelectFunc func(header []string) int

// ColumnMatchFunc returns a ColumnSelectFunc that matches any element in the given header string slice ss.
func ColumnMatchFunc(headerNameCandidates *sets.Set[string]) ColumnSelectFunc {
	return func(header []string) int {
		return slices.IndexFunc(header, func(s string) bool {
			return headerNameCandidates.Contains(strings.Compare, s)
		})
	}
}

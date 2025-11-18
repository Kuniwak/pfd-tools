package fsmtable

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

type ResourceTable struct {
	ExtraHeaders []string            `json:"extra_headers"`
	Rows         []*ResourceTableRow `json:"rows"`
}

func NewResourceTable() *ResourceTable {
	return &ResourceTable{ExtraHeaders: []string{}, Rows: []*ResourceTableRow{}}
}

func NewResourceTableByAtomicProcessTable(apTable *pfd.AtomicProcessTable, neededResourceSetsFunc fsm.NeededResourceSetsFunc) *ResourceTable {
	res := sets.New(fsm.ResourceID.Compare)
	rows := make([]*ResourceTableRow, 0, len(apTable.Rows))
	for _, row := range apTable.Rows {
		for _, entry := range neededResourceSetsFunc(row.ID).Iter() {
			for _, resource := range entry.Resources.Iter() {
				res.Add(fsm.ResourceID.Compare, resource)
			}
		}
	}

	for _, r := range res.Iter() {
		rows = append(rows, &ResourceTableRow{ID: r, Description: "", ExtraCells: make([]string, 0)})
	}

	return &ResourceTable{ExtraHeaders: []string{}, Rows: rows}
}

type ResourceTableRow struct {
	ID          fsm.ResourceID `json:"id"`
	Description string         `json:"description"`
	ExtraCells  []string       `json:"extra_cells"`
}

func (p *ResourceTable) Clone() *ResourceTable {
	rows := make([]*ResourceTableRow, 0, len(p.Rows))
	for _, row := range p.Rows {
		rows = append(rows, row.Clone())
	}
	return &ResourceTable{ExtraHeaders: slices.Clone(p.ExtraHeaders), Rows: rows}
}

func (p *ResourceTable) Header() []string {
	return append([]string{"ID", "Description"}, p.ExtraHeaders...)
}

func (p *ResourceTableRow) Clone() *ResourceTableRow {
	return &ResourceTableRow{ID: p.ID, Description: p.Description, ExtraCells: slices.Clone(p.ExtraCells)}
}

func (p *ResourceTable) Refresh(ns *sets.Set[pfd.NodeID], nodeMap map[pfd.NodeID]*pfd.Node, neededResourceSetsFunc fsm.NeededResourceSetsFunc) {
	actual := sets.NewWithCapacity[fsm.ResourceID](len(p.Rows))
	for _, row := range p.Rows {
		actual.Add(fsm.ResourceID.Compare, row.ID)
	}

	expected := sets.NewWithCapacity[fsm.ResourceID](0)
	for _, n := range ns.Iter() {
		if nodeMap[n].Type != pfd.NodeTypeAtomicProcess {
			continue
		}
		ap := pfd.AtomicProcessIDFromNodeID(n, nodeMap)

		for _, entry := range neededResourceSetsFunc(ap).Iter() {
			for _, resource := range entry.Resources.Iter() {
				expected.Add(fsm.ResourceID.Compare, resource)
			}
		}
	}

	extras := actual.Clone()
	extras.Difference(fsm.ResourceID.Compare, expected)

	missings := expected.Clone()
	missings.Difference(fsm.ResourceID.Compare, actual)

	rows := make([]*ResourceTableRow, 0, expected.Len())
	for _, row := range p.Rows {
		if extras.Contains(fsm.ResourceID.Compare, row.ID) {
			continue
		}
		rows = append(rows, row)
	}

	for _, missing := range missings.Iter() {
		rows = append(rows, &ResourceTableRow{ID: missing, Description: "", ExtraCells: make([]string, len(p.ExtraHeaders))})
	}

	slices.SortFunc(rows, func(a, b *ResourceTableRow) int {
		return a.ID.Compare(b.ID)
	})

	p.Rows = rows
}

func AvailableResources(t *ResourceTable) *sets.Set[fsm.ResourceID] {
	s := sets.New(fsm.ResourceID.Compare)
	for _, row := range t.Rows {
		s.Add(fsm.ResourceID.Compare, row.ID)
	}
	return s
}

const (
	NeededResourceSetsColumnHeaderJa = "必要資源"
	NeededResourceSetsColumnHeaderEn = "Needed Resources"
)

var DefaultNeededResourceSetsColumnSelectFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	NeededResourceSetsColumnHeaderJa,
	NeededResourceSetsColumnHeaderEn,
))

func ParseNeededResourceSetEntry(s string) (*sets.Set[fsm.AllocationElement], error) {
	res := sets.New(fsm.AllocationElement.Compare)
	for _, s1 := range strings.Split(s, ";") {
		s1 = strings.TrimSpace(s1)
		if s1 == "" {
			continue
		}

		ss1 := strings.SplitN(s1, ":", 2)
		if len(ss1) < 2 {
			return nil, fmt.Errorf("fsm.ParseNeededResourceSetEntry: must specify consumed volume: %v", ss1)
		}

		ss2 := strings.Split(ss1[0], ",")
		if len(ss2) < 1 {
			return nil, fmt.Errorf("fsm.ParseNeededResourceSetEntry: must specify at least one resource ID: %v", ss2)
		}
		rs := sets.New(fsm.ResourceID.Compare)
		for _, s2 := range ss2 {
			ridText := strings.TrimSpace(s2)
			if ridText == "" {
				continue
			}
			rs.Add(fsm.ResourceID.Compare, fsm.ResourceID(ridText))
		}
		consumedVolumeText := strings.TrimSpace(ss1[1])
		consumedVolume, err := strconv.ParseFloat(consumedVolumeText, 32)
		if err != nil {
			return nil, fmt.Errorf("fsm.ParseNeededResourceSetEntry: failed to parse consumed volume: %w: %q", err, s)
		}

		res.Add(fsm.AllocationElement.Compare, fsm.AllocationElement{
			Resources:      rs,
			ConsumedVolume: fsm.Volume(consumedVolume),
		})
	}
	return res, nil
}

func RawNeededResourceSetsMap(t *pfd.AtomicProcessTable, selectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]string, error) {
	m := make(map[pfd.AtomicProcessID]string, len(t.Rows))

	idx := selectFunc(t.ExtraHeaders)
	if idx < 0 {
		return nil, fmt.Errorf("fmt.NeededResourceSetsMap: missing needed resources column")
	}
	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[idx]
	}
	return m, nil
}

func ValidateNeededResourceSetsMap(m map[pfd.AtomicProcessID]string) (map[pfd.AtomicProcessID]*sets.Set[fsm.AllocationElement], error) {
	m2 := make(map[pfd.AtomicProcessID]*sets.Set[fsm.AllocationElement])

	for ap, resourceSetsText := range m {
		resourceSets, err := ParseNeededResourceSetEntry(resourceSetsText)
		if err != nil {
			return nil, fmt.Errorf("fmt.NeededResourceSetsMap: %w", err)
		}
		m2[ap] = resourceSets
	}

	return m2, nil
}

func NeededResourcesSetFuncByTable(t *pfd.AtomicProcessTable, selectFunc pfd.ColumnSelectFunc) (fsm.NeededResourceSetsFunc, error) {
	m, err := RawNeededResourceSetsMap(t, selectFunc)
	if err != nil {
		return nil, fmt.Errorf("fmt.NeededResourcesSetFuncByTable: %w", err)
	}
	m2, err := ValidateNeededResourceSetsMap(m)
	if err != nil {
		return nil, fmt.Errorf("fmt.NeededResourcesSetFuncByTable: %w", err)
	}
	return fsm.NeededResourceSetsFuncByMap(m2), nil
}

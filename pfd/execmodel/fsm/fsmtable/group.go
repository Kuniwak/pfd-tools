package fsmtable

import (
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/sets"
)

type GroupTable struct {
	ExtraHeaders []string         `json:"extra_headers"`
	Rows         []*GroupTableRow `json:"rows"`
}

func (t *GroupTable) Header() []string {
	return append([]string{"ID", "Description"}, t.ExtraHeaders...)
}

func (t *GroupTable) Groups() *sets.Set[fsmmasterschedule.Group] {
	groups := sets.New(fsmmasterschedule.Group.Compare)
	for _, row := range t.Rows {
		groups.Add(fsmmasterschedule.Group.Compare, row.ID)
	}
	return groups
}

func (t *GroupTable) DescriptionMap() map[fsmmasterschedule.Group]string {
	m := make(map[fsmmasterschedule.Group]string, len(t.Rows))
	for _, row := range t.Rows {
		m[row.ID] = row.Description
	}
	return m
}

type GroupTableRow struct {
	ID          fsmmasterschedule.Group `json:"id"`
	Description string                  `json:"description"`
	ExtraCells  []string                `json:"extra_cells"`
}

func (p *GroupTableRow) Compare(b *GroupTableRow) int {
	cmp1 := p.ID.Compare(b.ID)
	if cmp1 != 0 {
		return cmp1
	}
	return strings.Compare(p.Description, b.Description)
}

func (p *GroupTableRow) Row() []string {
	return append([]string{string(p.ID), p.Description}, p.ExtraCells...)
}

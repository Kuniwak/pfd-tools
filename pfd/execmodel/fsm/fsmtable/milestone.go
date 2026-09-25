package fsmtable

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/sets"
)

func NewMilestoneTableByAtomicProcessTable(ap *pfd.AtomicProcessTable) *MilestoneTable {
	empty := &MilestoneTable{ExtraHeaders: []string{}, Rows: []*MilestoneTableRow{}}
	milestoneIdx := DefaultBarColumnMatchFunc(ap.ExtraHeaders)
	if milestoneIdx < 0 {
		return empty
	}
	groupIdx := DefaultRowColumnMatchFunc(ap.ExtraHeaders)
	type milestoneInfo struct {
		groups *sets.Set[string]
	}
	milestoneMap := make(map[fsmmasterschedule.Milestone]*milestoneInfo)
	for _, row := range ap.Rows {
		milestoneText := strings.TrimSpace(row.ExtraCells[milestoneIdx])
		if milestoneText == "" {
			continue
		}
		mid := fsmmasterschedule.Milestone(milestoneText)
		info, ok := milestoneMap[mid]
		if !ok {
			info = &milestoneInfo{groups: sets.New(strings.Compare)}
			milestoneMap[mid] = info
		}
		if groupIdx >= 0 {
			groupText := strings.TrimSpace(row.ExtraCells[groupIdx])
			if groupText != "" {
				for _, g := range strings.Split(groupText, ",") {
					g = strings.TrimSpace(g)
					if g != "" {
						info.groups.Add(strings.Compare, g)
					}
				}
			}
		}
	}
	rows := make([]*MilestoneTableRow, 0, len(milestoneMap))
	for mid, info := range milestoneMap {
		groupParts := make([]string, 0, info.groups.Len())
		for _, g := range info.groups.Iter() {
			groupParts = append(groupParts, g)
		}
		rows = append(rows, &MilestoneTableRow{
			MilestoneID: mid,
			GroupIDs:    strings.Join(groupParts, ","),
			Description: "",
			Successors:  "",
			ExtraCells:  []string{},
		})
	}
	slices.SortFunc(rows, (*MilestoneTableRow).Compare)
	return &MilestoneTable{ExtraHeaders: []string{}, Rows: rows}
}

var DefaultBarColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	mastertsv.BarColumnHeader,
))

var DefaultRowColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	mastertsv.RowColumnHeader,
))

func RawGroupsMap(t *pfd.AtomicProcessTable, groupColumnSelectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]string, error) {
	m := make(map[pfd.AtomicProcessID]string, len(t.Rows))

	groupIdx := groupColumnSelectFunc(t.ExtraHeaders)
	if groupIdx < 0 {
		return nil, fmt.Errorf("fsmtable.RawGroupAndMilestonePairMap: missing group column")
	}

	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[groupIdx]
	}
	return m, nil
}

func ValidateGroupsMap(m map[pfd.AtomicProcessID]string) (map[pfd.AtomicProcessID]*sets.Set[fsmmasterschedule.Group], error) {
	m2 := make(map[pfd.AtomicProcessID]*sets.Set[fsmmasterschedule.Group], len(m))
	for ap, groupIDsText := range m {
		groupIDs, err := ParseGroups(groupIDsText)
		if err != nil {
			return nil, fmt.Errorf("fsmtable.ValidateGroupsMap: %w", err)
		}
		m2[ap] = groupIDs
	}
	return m2, nil
}

func RawMilestoneMap(t *pfd.AtomicProcessTable, milestoneColumnSelectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]string, error) {
	m := make(map[pfd.AtomicProcessID]string, len(t.Rows))

	milestoneIdx := milestoneColumnSelectFunc(t.ExtraHeaders)
	if milestoneIdx < 0 {
		return nil, fmt.Errorf("fsmtable.RawMilestoneMap: missing milestone column")
	}

	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[milestoneIdx]
	}
	return m, nil
}

func ValidateMilestoneMap(m map[pfd.AtomicProcessID]string) (map[pfd.AtomicProcessID]fsmmasterschedule.Milestone, error) {
	m2 := make(map[pfd.AtomicProcessID]fsmmasterschedule.Milestone, len(m))
	for ap, milestone := range m {
		m2[ap] = fsmmasterschedule.Milestone(milestone)
	}
	return m2, nil
}

func RawMilestoneEdgesMap(t *MilestoneTable) (map[fsmmasterschedule.Milestone]string, error) {
	m := make(map[fsmmasterschedule.Milestone]string, len(t.Rows))
	for _, row := range t.Rows {
		m[row.MilestoneID] = row.Successors
	}
	return m, nil
}

type MilestoneTable struct {
	ExtraHeaders []string             `json:"extra_headers"`
	Rows         []*MilestoneTableRow `json:"rows"`
}

func (t *MilestoneTable) Header() []string {
	return append([]string{"ID", "Description", "Groups", "Successors"}, t.ExtraHeaders...)
}

func (t *MilestoneTable) Milestones() *sets.Set[fsmmasterschedule.Milestone] {
	s := sets.New(fsmmasterschedule.Milestone.Compare)
	for _, row := range t.Rows {
		s.Add(fsmmasterschedule.Milestone.Compare, row.MilestoneID)
	}
	return s
}

func (t *MilestoneTable) DescriptionMap() map[fsmmasterschedule.Milestone]string {
	m := make(map[fsmmasterschedule.Milestone]string, len(t.Rows))
	for _, row := range t.Rows {
		m[row.MilestoneID] = row.Description
	}
	return m
}

type MilestoneTableRow struct {
	MilestoneID fsmmasterschedule.Milestone `json:"milestone_id"`
	GroupIDs    string                      `json:"group_ids"`
	Description string                      `json:"description"`
	Successors  string                      `json:"successors"`
	ExtraCells  []string                    `json:"extra_cells"`
}

func (p *MilestoneTableRow) Compare(b *MilestoneTableRow) int {
	c := p.MilestoneID.Compare(b.MilestoneID)
	if c != 0 {
		return c
	}
	c = strings.Compare(p.GroupIDs, b.GroupIDs)
	if c != 0 {
		return c
	}
	c = strings.Compare(p.Description, b.Description)
	if c != 0 {
		return c
	}
	c = strings.Compare(p.Successors, b.Successors)
	if c != 0 {
		return c
	}
	return slices.CompareFunc(p.ExtraCells, b.ExtraCells, strings.Compare)
}

func (p *MilestoneTableRow) Row() []string {
	return append([]string{string(p.MilestoneID), p.Description, p.GroupIDs, p.Successors}, p.ExtraCells...)
}

func ParseGroup(s string) (fsmmasterschedule.Group, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("fsmtable.ParseGroup: empty group")
	}
	return fsmmasterschedule.Group(strings.TrimSpace(s)), nil
}

func ParseGroups(s string) (*sets.Set[fsmmasterschedule.Group], error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sets.New(fsmmasterschedule.Group.Compare), nil
	}

	gs := strings.Split(s, ",")
	gs2 := sets.NewWithCapacity[fsmmasterschedule.Group](len(gs))
	for _, gText := range gs {
		g, err := ParseGroup(gText)
		if err != nil {
			return nil, fmt.Errorf("fsmtable.ParseGroups: %w", err)
		}
		gs2.Add(fsmmasterschedule.Group.Compare, g)
	}
	return gs2, nil
}

func ParseMilestone(s string) (fsmmasterschedule.Milestone, error) {
	if s == "" {
		return "", fmt.Errorf("fsmtable.ParseMilestone: empty milestone")
	}
	return fsmmasterschedule.Milestone(strings.TrimSpace(s)), nil
}

func ParseSuccessors(s string) (*sets.Set[fsmmasterschedule.Milestone], error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sets.New(fsmmasterschedule.Milestone.Compare), nil
	}

	ss := strings.Split(s, ",")
	ss2 := sets.NewWithCapacity[fsmmasterschedule.Milestone](len(ss))
	for _, sText := range ss {
		s, err := ParseMilestone(sText)
		if err != nil {
			return nil, fmt.Errorf("fsmtable.ParseSuccessors: %w", err)
		}
		ss2.Add(fsmmasterschedule.Milestone.Compare, s)
	}
	return ss2, nil
}

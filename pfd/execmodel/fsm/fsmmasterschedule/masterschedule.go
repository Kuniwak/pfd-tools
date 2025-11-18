package fsmmasterschedule

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/bizday"
	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/masterschedule"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

func NewMasterSchedule(timeline *Timeline, bizTimeFunc bizday.BusinessTimeFunc, startDay bizday.Day, gm map[Group]string, mm map[Milestone]string) *masterschedule.MasterSchedule {
	gs := slices.Collect(maps.Keys(*timeline))
	slices.SortFunc(gs, Group.Compare)

	ms := masterschedule.NewMasterSchedule()

	for _, g := range gs {
		m := (*timeline)[g]
		for milestone, ti := range *m {
			if ti.StartTime < 0 && ti.EndTime < 0 {
				continue
			}
			ms.Add(masterschedule.NewItem(
				bizTimeFunc(startDay, float64(ti.StartTime)),
				bizTimeFunc(startDay, float64(ti.EndTime)),
				string(g),
				gm[g],
				string(milestone),
				mm[milestone],
			))
		}
	}
	return ms
}

type Group string

func (g Group) Compare(h Group) int {
	return strings.Compare(string(g), string(h))
}

type Timeline map[Group]*MilestoneTimeline

func NewTimeline(gs *sets.Set[Group], msm map[Group]*sets.Set[Milestone]) *Timeline {
	t := make(Timeline, gs.Len())
	for g, milestones := range msm {
		t[g] = NewMilestoneTimeline(milestones)
	}
	return &t
}

func NewTimelineFromPlan(
	plan *fsm.Plan,
	aps *sets.Set[pfd.AtomicProcessID],
	gs *sets.Set[Group],
	gsm map[pfd.AtomicProcessID]*sets.Set[Group],
	mm map[pfd.AtomicProcessID]Milestone,
	mgm map[Group]*graph.Graph,
) (*Timeline, error) {
	msm := make(map[Group]*sets.Set[Milestone])
	for _, ap := range aps.Iter() {
		gs, ok := gsm[ap]
		if !ok {
			panic(fmt.Sprintf("fsmmasterschedule.NewTimelineFromPlan: missing group set for atomic process: %q", ap))
		}

		m, ok := mm[ap]
		if !ok {
			panic(fmt.Sprintf("fsmmasterschedule.NewTimelineFromPlan: missing milestone for atomic process: %q", ap))
		}

		for _, g := range gs.Iter() {
			if ms, ok := msm[g]; ok {
				ms.Add(Milestone.Compare, m)
			} else {
				msm[g] = sets.New(Milestone.Compare, m)
			}
		}
	}

	var lastNumOfReworksMap map[pfd.AtomicProcessID]int
	if len(plan.Transitions) == 0 {
		lastNumOfReworksMap = plan.InitialState.NumOfCompleteMap
	} else {
		lastNumOfReworksMap = plan.Transitions[len(plan.Transitions)-1].NextState.NumOfCompleteMap
	}

	timeline := NewTimeline(gs, msm)
	for g, ms := range msm {
		mg, ok := mgm[g]
		if !ok {
			panic(fmt.Sprintf("fsmmasterschedule.NewTimelineFromPlan: missing milestone graph for group: %q", g))
		}
		t, err := NewMilestoneTimelineFromPlan(
			plan,
			g,
			ms,
			NewMilestoneBeginFuncByPairs(mm, gsm),
			NewMilestoneEndFuncByPairs(aps, gs, mm, lastNumOfReworksMap),
			mg,
		)
		if err != nil {
			return nil, fmt.Errorf("fsmmasterschedule.NewTimelineFromPlan: %w", err)
		}
		timeline.SetMilestoneTimeline(g, t)
	}

	return timeline, nil
}

func NewTimelineWithBufferMultiplier(timeline *Timeline, bufferMultiplier float64) *Timeline {
	res := make(Timeline, len(*timeline))
	for g, ms := range *timeline {
		res[g] = NewMilestoneTimelineWithBufferMultiplier(ms, bufferMultiplier)
	}
	return &res
}

func (t *Timeline) SetMilestoneTimeline(g Group, ms *MilestoneTimeline) {
	(*t)[g] = ms
}

type MilestoneBeginFunc func(g Group, m Milestone, prevState fsm.State, alloc fsm.Allocation, nextState fsm.State) bool

func NewAtomicProcessToMilestoneMap(mm map[pfd.AtomicProcessID]Milestone, gsm map[pfd.AtomicProcessID]*sets.Set[Group]) map[Group]map[pfd.AtomicProcessID]Milestone {
	m1 := make(map[Group]*sets.Set[pfd.AtomicProcessID])
	for ap, gs := range gsm {
		for _, g := range gs.Iter() {
			if m, ok := m1[g]; ok {
				m.Add(pfd.AtomicProcessID.Compare, ap)
			} else {
				m1[g] = sets.New(pfd.AtomicProcessID.Compare, ap)
			}
		}
	}

	m2 := make(map[Group]map[pfd.AtomicProcessID]Milestone)
	for g, aps := range m1 {
		m3 := make(map[pfd.AtomicProcessID]Milestone, aps.Len())
		for _, ap := range aps.Iter() {
			m3[ap] = mm[ap]
		}
		m2[g] = m3
	}
	return m2
}

func NewMilestoneBeginFuncByPairs(mm map[pfd.AtomicProcessID]Milestone, gsm map[pfd.AtomicProcessID]*sets.Set[Group]) MilestoneBeginFunc {
	return NewMilestoneBeginFunc(NewAtomicProcessToMilestoneMap(mm, gsm))
}

func NewMilestoneBeginFunc(m1 map[Group]map[pfd.AtomicProcessID]Milestone) MilestoneBeginFunc {
	return func(g Group, milestone Milestone, prevState fsm.State, alloc fsm.Allocation, _ fsm.State) bool {
		m2, ok := m1[g]
		if !ok {
			panic(fmt.Sprintf("fsmmasterschedule.NewMilestoneBeginFunc: missing group: %q", g))
		}

		for ap := range alloc {
			m, ok := m2[ap]
			if !ok {
				// NOTE: Ignore atomic processes that do not belong to this group.
				continue
			}
			if m != milestone {
				// NOTE: Ignore if this atomic process's milestone is not this milestone.
				continue
			}
			return true
		}
		return false
	}
}

type MilestoneEndFunc func(g Group, m Milestone, prevState fsm.State, alloc fsm.Allocation, currentState fsm.State) bool

func NewMilestoneToAtomicProcessesMap(aps *sets.Set[pfd.AtomicProcessID], gs *sets.Set[Group], mm map[pfd.AtomicProcessID]Milestone) map[Group]map[Milestone]*sets.Set[pfd.AtomicProcessID] {
	res := make(map[Group]map[Milestone]*sets.Set[pfd.AtomicProcessID], gs.Len())
	for _, g := range gs.Iter() {
		m := make(map[Milestone]*sets.Set[pfd.AtomicProcessID], len(mm))
		for _, ap := range aps.Iter() {
			milestone, ok := mm[ap]
			if !ok {
				panic(fmt.Sprintf("fsmmasterschedule.NewMilestoneToAtomicProcessesMap: missing milestone for atomic process: %q", ap))
			}
			if s, ok := m[milestone]; ok {
				s.Add(pfd.AtomicProcessID.Compare, ap)
			} else {
				m[milestone] = sets.New(pfd.AtomicProcessID.Compare, ap)
			}
		}

		res[g] = m
	}

	return res
}

func NewMilestoneEndFuncByPairs(aps *sets.Set[pfd.AtomicProcessID], gs *sets.Set[Group], mm map[pfd.AtomicProcessID]Milestone, lastNumOfReworksMap map[pfd.AtomicProcessID]int) MilestoneEndFunc {
	return NewMilestoneEndFunc(NewMilestoneToAtomicProcessesMap(aps, gs, mm), lastNumOfReworksMap)
}

func NewMilestoneEndFunc(m1 map[Group]map[Milestone]*sets.Set[pfd.AtomicProcessID], lastNumOfReworksMap map[pfd.AtomicProcessID]int) MilestoneEndFunc {
	return func(g Group, milestone Milestone, _ fsm.State, alloc fsm.Allocation, currentState fsm.State) bool {
		m2, ok := m1[g]
		if !ok {
			panic(fmt.Sprintf("fsmmasterschedule.NewMilestoneEndFunc: missing group: %q", g))
		}

		aps, ok := m2[milestone]
		if !ok {
			panic(fmt.Sprintf("fsmmasterschedule.NewMilestoneEndFunc: missing milestone: %q", milestone))
		}

		for _, ap := range aps.Iter() {
			if currentState.NumOfCompleteMap[ap] < lastNumOfReworksMap[ap] {
				return false
			}
		}
		return true
	}
}

type Milestone string

func (m Milestone) Compare(h Milestone) int {
	return strings.Compare(string(m), string(h))
}

type MilestoneTimeline map[Milestone]*TimelineItem

func NewMilestoneTimeline(milestones *sets.Set[Milestone]) *MilestoneTimeline {
	t := make(MilestoneTimeline, milestones.Len())
	for _, m := range milestones.Iter() {
		t[m] = NewTimelineItem(-1, -1)
	}
	return &t
}

func NewMilestoneTimelineWithBufferMultiplier(milestoneTimeline *MilestoneTimeline, bufferMultiplier float64) *MilestoneTimeline {
	res := make(MilestoneTimeline, len(*milestoneTimeline))
	for m, ti := range *milestoneTimeline {
		res[m] = NewTimelineItem(ti.StartTime*execmodel.Time(bufferMultiplier), ti.EndTime*execmodel.Time(bufferMultiplier))
	}
	return &res
}

func (t *MilestoneTimeline) SetStartTime(m Milestone, startTime execmodel.Time) {
	(*t)[m].StartTime = startTime
}

func (t *MilestoneTimeline) SetEndTime(m Milestone, endTime execmodel.Time) {
	(*t)[m].EndTime = endTime
}

func (t *MilestoneTimeline) GetStartTime(m Milestone) (execmodel.Time, bool) {
	i, ok := (*t)[m]
	if !ok {
		// NOTE: Some groups may skip milestones, and skipped milestones have no start time.
		return -1, false
	}
	return i.StartTime, i.StartTime != -1
}

func (t *MilestoneTimeline) GetEndTime(m Milestone) (execmodel.Time, bool) {
	tt := (*t)[m].EndTime
	if tt == -1 {
		return -1, false
	}
	return tt, true
}

type TimelineItem struct {
	StartTime execmodel.Time
	EndTime   execmodel.Time
}

func (i *TimelineItem) Clone() *TimelineItem {
	return &TimelineItem{StartTime: i.StartTime, EndTime: i.EndTime}
}

func NewTimelineItem(startTime, endTime execmodel.Time) *TimelineItem {
	return &TimelineItem{StartTime: startTime, EndTime: endTime}
}

func NewMilestoneTimelineFromPlan(plan *fsm.Plan, g Group, ms *sets.Set[Milestone], begin MilestoneBeginFunc, end MilestoneEndFunc, mg *graph.Graph) (*MilestoneTimeline, error) {
	t1 := newMilestoneTimelineFromPlan(plan, g, ms, begin, end)
	t2, err := newMilestoneTimelineRemovingOverlapping(t1, mg)
	if err != nil {
		return nil, fmt.Errorf("fsmmasterschedule.NewMilestoneTimelineFromPlan: %w", err)
	}
	return t2, nil
}

func newMilestoneTimelineFromPlan(plan *fsm.Plan, g Group, ms *sets.Set[Milestone], begin MilestoneBeginFunc, end MilestoneEndFunc) *MilestoneTimeline {
	states := plan.States()

	timeline := NewMilestoneTimeline(ms)

	for idx, tr := range plan.Transitions {
		prevState := states[idx]

		for _, m := range ms.Iter() {
			if _, ok := timeline.GetStartTime(m); ok {
				continue
			}

			if begin(g, m, prevState, tr.Allocation, tr.NextState) {
				timeline.SetStartTime(m, prevState.Time)
			}
		}

		for _, m := range ms.Iter() {
			if _, ok := timeline.GetStartTime(m); !ok {
				continue
			}

			if _, ok := timeline.GetEndTime(m); ok {
				continue
			}

			if end(g, m, prevState, tr.Allocation, tr.NextState) {
				timeline.SetEndTime(m, tr.NextState.Time)
			}
		}
	}

	return timeline
}

func newMilestoneTimelineRemovingOverlapping(t *MilestoneTimeline, mg *graph.Graph) (*MilestoneTimeline, error) {
	if len(*t) < 2 {
		return t, nil
	}

	milestones := graphNodesToMilestones(mg.Nodes)
	maximals := graphNodesToMilestones(mg.Maximals())
	notMaximals := milestones.Clone()
	notMaximals.Difference(Milestone.Compare, maximals)
	successorsMap := make(map[Milestone]*sets.Set[Milestone])
	for _, e := range mg.Edges.Iter() {
		if s, ok := successorsMap[Milestone(e.First)]; ok {
			s.Add(Milestone.Compare, Milestone(e.Second))
		} else {
			successorsMap[Milestone(e.First)] = sets.New(Milestone.Compare, Milestone(e.Second))
		}
	}

	res := NewMilestoneTimeline(milestones)
	for _, m1 := range notMaximals.Iter() {
		startTime1, ok := t.GetStartTime(m1)
		if !ok {
			// NOTE: Some groups may skip milestones, and skipped milestones have no start time.
			continue
		}
		endTime1, ok := t.GetEndTime(m1)
		if !ok {
			// NOTE: Some groups may skip milestones, and skipped milestones have no start time.
			continue
		}

		res.SetStartTime(m1, startTime1)
		endTime2 := endTime1

		ms := sets.New(Milestone.Compare)
		collectSuccessorsTransitive(m1, successorsMap, ms)
		for _, m2 := range ms.Iter() {
			startTime2, ok := t.GetStartTime(m2)
			if !ok {
				// NOTE: Some groups may skip milestones, and skipped milestones have no start time.
				continue
			}
			if startTime1 > startTime2 {
				return nil, fmt.Errorf("fsmmasterschedule.NewMilestoneTimelineRemovingOverlapping: overlapping milestone: %q -> %q, %f > %f", m1, m2, startTime1, startTime2)
			}
			endTime2 = min(endTime2, startTime2)
		}

		res.SetEndTime(m1, endTime2)
	}

	for _, maximal := range maximals.Iter() {
		startTime, ok := t.GetStartTime(maximal)
		if !ok {
			continue
		}
		endTime, ok := t.GetEndTime(maximal)
		if !ok {
			continue
		}
		res.SetStartTime(maximal, startTime)
		res.SetEndTime(maximal, endTime)
	}
	return res, nil
}

func graphNodesToMilestones(s *sets.Set[graph.Node]) *sets.Set[Milestone] {
	res := sets.New(Milestone.Compare)
	for _, n := range s.Iter() {
		res.Add(Milestone.Compare, Milestone(n))
	}
	return res
}

func collectSuccessorsTransitive(m Milestone, succesorsMap map[Milestone]*sets.Set[Milestone], res *sets.Set[Milestone]) {
	successors, ok := succesorsMap[m]
	if !ok {
		return
	}
	for _, successor := range successors.Iter() {
		res.Add(Milestone.Compare, successor)
		collectSuccessorsTransitive(successor, succesorsMap, res)
	}
}

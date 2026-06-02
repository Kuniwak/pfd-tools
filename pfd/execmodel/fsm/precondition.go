package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

type PreconditionType string

const (
	PreconditionTypeFeedbackSourceCompleted                      PreconditionType = "FEEDBACK_SOURCE_COMPLETED"
	PreconditionTypeExecBetween                                  PreconditionType = "EXEC_BETWEEN"
	PreconditionTypeExecutable                                   PreconditionType = "EXECUTABLE"
	PreconditionTypeOr                                           PreconditionType = "OR"
	PreconditionTypeAnd                                          PreconditionType = "AND"
	PreconditionTypeTrue                                         PreconditionType = "TRUE"
	PreconditionTypeNot                                          PreconditionType = "NOT"
	PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted PreconditionType = "ALL_BACKWARD_REACHABLE_FEEDBACK_SOURCES_COMPLETED"
)

type RevisionBoundType string

const (
	RevisionBoundTypeInt         RevisionBoundType = "INT"
	RevisionBoundTypeInfinity    RevisionBoundType = "INFINITY"
	RevisionBoundTypeMaxRevision RevisionBoundType = "MAX_REVISION"
)

type RevisionBound struct {
	Type RevisionBoundType `json:"type"`

	IntValue int `json:"int_value,omitempty"`

	MaxRevisionDeliverable pfd.AtomicDeliverableID `json:"max_revision_deliverable,omitempty"`
}

func NewIntRevisionBound(v int) *RevisionBound {
	return &RevisionBound{
		Type:     RevisionBoundTypeInt,
		IntValue: v,
	}
}

func NewInfinityRevisionBound() *RevisionBound {
	return &RevisionBound{
		Type: RevisionBoundTypeInfinity,
	}
}

func NewMaxRevisionBound(d pfd.AtomicDeliverableID) *RevisionBound {
	return &RevisionBound{
		Type:                   RevisionBoundTypeMaxRevision,
		MaxRevisionDeliverable: d,
	}
}

type RevisionBoundEvalResult struct {
	Type RevisionBoundType `json:"type"`

	Value int `json:"value"`

	MaxRevisionDeliverable pfd.AtomicDeliverableID `json:"max_revision_deliverable,omitempty"`
}

func (b *RevisionBound) Eval(e *Env) *RevisionBoundEvalResult {
	switch b.Type {
	case RevisionBoundTypeInt:
		return &RevisionBoundEvalResult{
			Type:  b.Type,
			Value: b.IntValue,
		}
	case RevisionBoundTypeInfinity:
		return &RevisionBoundEvalResult{
			Type: b.Type,
		}
	case RevisionBoundTypeMaxRevision:
		maxRevision, ok := e.FeedbackSourceMaxRevision[b.MaxRevisionDeliverable]
		if !ok {
			panic(fmt.Sprintf("fsm.RevisionBound.Eval: missing max revision map: %q", b.MaxRevisionDeliverable))
		}
		return &RevisionBoundEvalResult{
			Type:                   b.Type,
			Value:                  maxRevision,
			MaxRevisionDeliverable: b.MaxRevisionDeliverable,
		}
	default:
		panic(fmt.Sprintf("fsm.RevisionBound.Eval: invalid type: %q", b.Type))
	}
}

func (b *RevisionBound) Write(w io.Writer) error {
	switch b.Type {
	case RevisionBoundTypeInt:
		_, _ = fmt.Fprint(w, b.IntValue)
	case RevisionBoundTypeInfinity:
		_, _ = io.WriteString(w, `\inf`)
	case RevisionBoundTypeMaxRevision:
		_, _ = io.WriteString(w, `\maxRev(`)
		_, _ = io.WriteString(w, string(b.MaxRevisionDeliverable))
		_, _ = io.WriteString(w, `)`)
	default:
		panic(fmt.Sprintf("fsm.RevisionBound.Write: invalid type: %q", b.Type))
	}
	return nil
}

type Precondition struct {
	Type PreconditionType `json:"type"`

	// FeedbackSource is true if the feedback loop has ended, false otherwise. Behavior is undefined when Type is other than PreconditionTypeFeedbackSourceCompleted.
	FeedbackSource pfd.AtomicDeliverableID `json:"feedback_source,omitempty"`

	ExecBetweenTarget pfd.AtomicDeliverableID `json:"exec_between_target,omitempty"`

	ExecBetweenBegin *RevisionBound `json:"exec_between_begin,omitempty"`

	ExecBetweenEnd *RevisionBound `json:"exec_between_end,omitempty"`

	// Executable is true if the specified atomic process is executable. Behavior is undefined when Type is other than PreconditionTypeExecutable.
	Executable pfd.AtomicProcessID `json:"executable,omitempty"`

	// Not is the NOT condition. Behavior is undefined when Type is other than PreconditionTypeNot.
	Not *Precondition `json:"not,omitempty"`

	// Or is the OR condition. Behavior is undefined when Type is other than PreconditionTypeOr.
	Or []*Precondition `json:"or,omitempty"`

	// And is the AND condition. Behavior is undefined when Type is other than PreconditionTypeAnd.
	And []*Precondition `json:"and,omitempty"`

	// AllBackwardReachableFeedbackSourcesCompletedTarget is true if all feedback loops from feedback deliverables reachable to the specified atomic process have ended, false otherwise. Behavior is undefined when Type is other than PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted.
	AllBackwardReachableFeedbackSourcesCompletedTarget pfd.AtomicProcessID `json:"all_backward_reachable_feedback_sources_completed_target,omitempty"`
}

func NewFeedbackSourceCompletedPrecondition(feedbackSource pfd.AtomicDeliverableID) *Precondition {
	return NewExecBetweenPrecondition(
		feedbackSource,
		NewMaxRevisionBound(feedbackSource),
		NewInfinityRevisionBound(),
	)
}

func NewExecBetweenPrecondition(target pfd.AtomicDeliverableID, begin, end *RevisionBound) *Precondition {
	return &Precondition{
		Type:              PreconditionTypeExecBetween,
		ExecBetweenTarget: target,
		ExecBetweenBegin:  begin,
		ExecBetweenEnd:    end,
	}
}

func NewExecutablePrecondition(executable pfd.AtomicProcessID) *Precondition {
	return &Precondition{
		Type:       PreconditionTypeExecutable,
		Executable: executable,
	}
}

func NewNotPrecondition(not *Precondition) *Precondition {
	return &Precondition{
		Type: PreconditionTypeNot,
		Not:  not,
	}
}

func NewOrPrecondition(or ...*Precondition) *Precondition {
	return &Precondition{
		Type: PreconditionTypeOr,
		Or:   or,
	}
}

func NewAndPrecondition(and ...*Precondition) *Precondition {
	return &Precondition{
		Type: PreconditionTypeAnd,
		And:  and,
	}
}

func NewTruePrecondition() *Precondition {
	return &Precondition{
		Type: PreconditionTypeTrue,
	}
}

func NewAllBackwardReachableFeedbackSourcesCompleted(target pfd.AtomicProcessID) *Precondition {
	return &Precondition{
		Type: PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted,
		AllBackwardReachableFeedbackSourcesCompletedTarget: target,
	}
}

type PreconditionEvalResult struct {
	Type   PreconditionType `json:"type"`
	Result bool             `json:"result"`

	// FeedbackSource is a feedback edge whose completion is specified in the precondition but is not yet completed. Behavior is undefined when Type is other than PreconditionTypeFeedbackSourceCompleted.
	FeedbackSource pfd.AtomicDeliverableID `json:"feedback_source,omitempty"`

	ExecBetweenTarget pfd.AtomicDeliverableID `json:"exec_between_target,omitempty"`

	ExecBetweenBegin *RevisionBoundEvalResult `json:"exec_between_begin,omitempty"`

	ExecBetweenEnd *RevisionBoundEvalResult `json:"exec_between_end,omitempty"`

	// Executable is true if the specified atomic process is executable, false otherwise. Behavior is undefined when Type is other than PreconditionTypeExecutable.
	Executable *AllocatabilityInfo `json:"executable,omitempty"`

	// Not is the result of the NOT condition. Behavior is undefined when Type is other than PreconditionTypeNot.
	Not *PreconditionEvalResult `json:"not,omitempty"`

	// Revision is the revision of the feedback source deliverable. Behavior is undefined when Type is other than PreconditionTypeFeedbackSourceCompleted.
	Revision int `json:"revision,omitempty"`

	// MaxRevision is the maximum revision of the feedback source deliverable. Behavior is undefined when Type is other than PreconditionTypeFeedbackSourceCompleted.
	MaxRevision int `json:"max_revision,omitempty"`

	// Or is the result of the OR condition. Behavior is undefined when Type is other than PreconditionTypeOr.
	Or []*PreconditionEvalResult `json:"or_result,omitempty"`

	// And is the result of the AND condition. Behavior is undefined when Type is other than PreconditionTypeAnd.
	And []*PreconditionEvalResult `json:"and_result,omitempty"`

	// AllBackwardReachableFeedbackSourcesCompleted is true if all feedback loops have ended, false otherwise. Behavior is undefined when Type is other than PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted.
	AllBackwardReachableFeedbackSourcesCompleted *PreconditionEvalResult `json:"all_backward_reachable_feedback_sources_completed,omitempty"`
}

func (r *PreconditionEvalResult) Write(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	return e.Encode(r)
}

func (p *Precondition) Eval(e *Env, remainedVolumeMap map[pfd.AtomicProcessID]Volume, revisionMap map[pfd.AtomicDeliverableID]int, allocationShouldContinue Allocation, updatedDeliverablesNotHandled map[pfd.AtomicProcessID]*sets.Set[pfd.AtomicDeliverableID]) *PreconditionEvalResult {
	switch p.Type {
	case PreconditionTypeFeedbackSourceCompleted:
		return NewFeedbackSourceCompletedPrecondition(p.FeedbackSource).Eval(e, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)

	case PreconditionTypeExecBetween:
		if !e.PFD.FeedbackSourceDeliverables().Contains(pfd.AtomicDeliverableID.Compare, p.ExecBetweenTarget) {
			panic(fmt.Sprintf("fsm.Precondition.Eval: missing feedback source deliverable: %q", p.ExecBetweenTarget))
		}

		revision, ok := revisionMap[p.ExecBetweenTarget]
		if !ok {
			panic(fmt.Sprintf("fsm.Precondition.Eval: missing revision map: %q", p.ExecBetweenTarget))
		}

		begin := p.ExecBetweenBegin.Eval(e)
		end := p.ExecBetweenEnd.Eval(e)
		result := begin.Value <= revision
		if end.Type != RevisionBoundTypeInfinity {
			result = result && revision < end.Value
		}

		return &PreconditionEvalResult{
			Type:              PreconditionTypeExecBetween,
			Result:            result,
			ExecBetweenTarget: p.ExecBetweenTarget,
			ExecBetweenBegin:  begin,
			ExecBetweenEnd:    end,
			Revision:          revision,
		}

	case PreconditionTypeNot:
		r := p.Not.Eval(e, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
		return &PreconditionEvalResult{
			Type:   PreconditionTypeNot,
			Result: !r.Result,
			Not:    r,
		}

	case PreconditionTypeExecutable:
		allocatability := e.AllocatabilityInfo(p.Executable, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
		return &PreconditionEvalResult{
			Type:       PreconditionTypeExecutable,
			Result:     allocatability.Allocatability.IsOK(),
			Executable: allocatability,
		}

	case PreconditionTypeOr:
		results := make([]*PreconditionEvalResult, len(p.Or))
		result := false
		for i, precondition := range p.Or {
			r := precondition.Eval(e, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
			results[i] = r
			if r.Result {
				result = true
			}
		}
		return &PreconditionEvalResult{
			Type:   PreconditionTypeOr,
			Result: result,
			Or:     results,
		}

	case PreconditionTypeAnd:
		results := make([]*PreconditionEvalResult, len(p.And))
		result := true
		for i, precondition := range p.And {
			r := precondition.Eval(e, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
			results[i] = r
			if !r.Result {
				result = false
			}
		}
		return &PreconditionEvalResult{
			Type:   PreconditionTypeAnd,
			Result: result,
			And:    results,
		}

	case PreconditionTypeTrue:
		return &PreconditionEvalResult{
			Type:   PreconditionTypeTrue,
			Result: true,
		}

	case PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted:
		p2 := p.Compile(e.PFD, e.Logger)
		r := p2.Eval(e, remainedVolumeMap, revisionMap, allocationShouldContinue, updatedDeliverablesNotHandled)
		return &PreconditionEvalResult{
			Type:   PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted,
			Result: r.Result,
			AllBackwardReachableFeedbackSourcesCompleted: r,
		}

	default:
		panic(fmt.Sprintf("fsm.Precondition.Eval: invalid type: %q", p.Type))
	}
}

func (p *Precondition) Compile(vp *pfd.ValidPFD, logger *slog.Logger) *Precondition {
	switch p.Type {
	case PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted:
		ads := sets.New(pfd.AtomicDeliverableID.Compare)
		vp.CollectBackwardReachableDeliverablesExceptFeedback(p.AllBackwardReachableFeedbackSourcesCompletedTarget, ads, logger)

		ps := make([]*Precondition, 0)

		for _, ad := range ads.Iter() {
			if vp.FeedbackDestinationAtomicProcesses(ad).Len() == 0 {
				continue
			}
			ps = append(ps, NewFeedbackSourceCompletedPrecondition(ad))
		}

		aps := sets.New(pfd.AtomicProcessID.Compare)
		vp.CollectBackwardReachableAtomicProcessesExceptFeedback(p.AllBackwardReachableFeedbackSourcesCompletedTarget, aps, logger)
		aps.Remove(pfd.AtomicProcessID.Compare, p.AllBackwardReachableFeedbackSourcesCompletedTarget)

		for _, ap := range aps.Iter() {
			ps = append(ps, NewNotPrecondition(NewExecutablePrecondition(ap)))
		}

		return NewAndPrecondition(ps...)
	default:
		return p
	}
}

func (p *Precondition) Write(w io.Writer) error {
	switch p.Type {
	case PreconditionTypeFeedbackSourceCompleted:
		return NewFeedbackSourceCompletedPrecondition(p.FeedbackSource).Write(w)
	case PreconditionTypeExecBetween:
		_, _ = io.WriteString(w, `\execBetween(`)
		_, _ = io.WriteString(w, string(p.ExecBetweenTarget))
		_, _ = io.WriteString(w, `, `)
		_ = p.ExecBetweenBegin.Write(w)
		_, _ = io.WriteString(w, `, `)
		_ = p.ExecBetweenEnd.Write(w)
		_, _ = io.WriteString(w, `)`)
	case PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted:
		_, _ = io.WriteString(w, `\complete(*)`)
	case PreconditionTypeExecutable:
		_, _ = io.WriteString(w, `\exec(`)
		_, _ = io.WriteString(w, string(p.Executable))
		_, _ = io.WriteString(w, `)`)
	case PreconditionTypeOr:
		for i, precondition := range p.Or {
			if i > 0 {
				_, _ = io.WriteString(w, ` || `)
			}
			_ = precondition.Write(w)
		}
	case PreconditionTypeAnd:
		for i, precondition := range p.And {
			if i > 0 {
				_, _ = io.WriteString(w, ` && `)
			}
			_ = precondition.Write(w)
		}
	case PreconditionTypeNot:
		_, _ = io.WriteString(w, `!`)
		_ = p.Not.Write(w)
	case PreconditionTypeTrue:
		_, _ = io.WriteString(w, `\true`)
	default:
		panic(fmt.Sprintf("fsm.Precondition.Write: invalid type: %q", p.Type))
	}
	return nil
}

func (p *Precondition) Traverse(f func(p *Precondition)) {
	switch p.Type {
	case PreconditionTypeFeedbackSourceCompleted:
		f(p)
	case PreconditionTypeExecBetween:
		f(p)
	case PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted:
		f(p)
	case PreconditionTypeExecutable:
		f(p)
	case PreconditionTypeOr:
		for _, precondition := range p.Or {
			precondition.Traverse(f)
		}
	case PreconditionTypeAnd:
		for _, precondition := range p.And {
			precondition.Traverse(f)
		}
	case PreconditionTypeNot:
		f(p)
		p.Not.Traverse(f)
	case PreconditionTypeTrue:
		f(p)
	default:
		panic(fmt.Sprintf("fsm.Precondition.Traverse: invalid type: %q", p.Type))
	}
}

func NewPreconditionMap(aps *sets.Set[pfd.AtomicProcessID], m map[pfd.AtomicProcessID]*Precondition) map[pfd.AtomicProcessID]*Precondition {
	for _, ap := range aps.Iter() {
		if _, ok := m[ap]; !ok {
			m[ap] = NewTruePrecondition()
		}
	}
	return m
}

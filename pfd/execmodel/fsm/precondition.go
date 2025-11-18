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
	PreconditionTypeExecutable                                   PreconditionType = "EXECUTABLE"
	PreconditionTypeOr                                           PreconditionType = "OR"
	PreconditionTypeAnd                                          PreconditionType = "AND"
	PreconditionTypeTrue                                         PreconditionType = "TRUE"
	PreconditionTypeNot                                          PreconditionType = "NOT"
	PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted PreconditionType = "ALL_BACKWARD_REACHABLE_FEEDBACK_SOURCES_COMPLETED"
)

type Precondition struct {
	Type PreconditionType `json:"type"`

	// FeedbackSource is true if the feedback loop has ended, false otherwise. Behavior is undefined when Type is other than PreconditionTypeFeedbackSourceCompleted.
	FeedbackSource pfd.AtomicDeliverableID `json:"feedback_source,omitempty"`

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
	return &Precondition{
		Type:           PreconditionTypeFeedbackSourceCompleted,
		FeedbackSource: feedbackSource,
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
		if !e.PFD.FeedbackSourceDeliverables().Contains(pfd.AtomicDeliverableID.Compare, p.FeedbackSource) {
			panic(fmt.Sprintf("fsm.Precondition.Eval: missing feedback source deliverable: %q", p.FeedbackSource))
		}

		revision, ok := revisionMap[p.FeedbackSource]
		if !ok {
			panic(fmt.Sprintf("fsm.Precondition.Eval: missing revision map: %q", p.FeedbackSource))
		}

		maxRevision, ok := e.FeedbackSourceMaxRevision[p.FeedbackSource]
		if !ok {
			panic(fmt.Sprintf("fsm.Precondition.Eval: missing max revision map: %q", p.FeedbackSource))
		}
		return &PreconditionEvalResult{
			Type:           PreconditionTypeFeedbackSourceCompleted,
			Result:         revision >= maxRevision,
			FeedbackSource: p.FeedbackSource,
			Revision:       revision,
			MaxRevision:    maxRevision,
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
		io.WriteString(w, `\complete(`)
		io.WriteString(w, string(p.FeedbackSource))
		io.WriteString(w, `)`)
	case PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted:
		io.WriteString(w, `\complete(*)`)
	case PreconditionTypeExecutable:
		io.WriteString(w, `\exec(`)
		io.WriteString(w, string(p.Executable))
		io.WriteString(w, `)`)
	case PreconditionTypeOr:
		for i, precondition := range p.Or {
			if i > 0 {
				io.WriteString(w, ` || `)
			}
			precondition.Write(w)
		}
	case PreconditionTypeAnd:
		for i, precondition := range p.And {
			if i > 0 {
				io.WriteString(w, ` && `)
			}
			precondition.Write(w)
		}
	case PreconditionTypeNot:
		io.WriteString(w, `!`)
		p.Not.Write(w)
	case PreconditionTypeTrue:
		io.WriteString(w, `\true`)
	default:
		panic(fmt.Sprintf("fsm.Precondition.Write: invalid type: %q", p.Type))
	}
	return nil
}

func (p *Precondition) Traverse(f func(p *Precondition)) {
	switch p.Type {
	case PreconditionTypeFeedbackSourceCompleted:
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

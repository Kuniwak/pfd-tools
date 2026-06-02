package fsmchecker

import (
	"fmt"
	"math"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/graph"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ValidPrecondition = checkers.AtomicChecker[*fsmcommon.Target]{
	ID: "valid-precondition",
	AvailableIfFunc: func(t *fsmcommon.Target) bool {
		return t.Memoized.HasPreconditionMap
	},
	CheckFunc: func(t *fsmcommon.Target, ch chan<- checkers.Problem) error {
		const malformedPreconditionProblemID = "malformed-precondition"
		const preconditionNotFeedbackSourceProblemID = "precondition-not-feedback-source"
		const preconditionNotAtomicProcessProblemID = "precondition-not-atomic-process"
		const preconditionCyclicExecutableReferenceProblemID = "precondition-cyclic-executable-reference"
		const preconditionReachableFeedbackSourceProblemID = "precondition-reachable-feedback-source"
		const preconditionReachableExecutableTargetProblemID = "precondition-reachable-executable-target"
		const preconditionInvalidExecBoundProblemID = "precondition-invalid-exec-bound"
		const preconditionInvalidExecRangeProblemID = "precondition-invalid-exec-range"

		m := make(map[pfd.AtomicProcessID]*fsm.Precondition)
		for ap, preconditionText := range t.Memoized.PreconditionMap {
			precondition, err := fsmtable.ParsePrecondition(preconditionText, ap)
			if err != nil {
				ch <- checkers.NewProblem(
					malformedPreconditionProblemID,
					checkers.SeverityError,
					fsmcommon.NewLocations(
						fsmcommon.NewLocation(
							fsmcommon.LocationTypeAtomicProcessTable,
							fsmcommon.NewAtomicProcessID(ap),
						),
					)...,
				)
				continue
			}
			m[ap] = precondition

			precondition.Traverse(func(p *fsm.Precondition) {
				switch p.Type {
				case fsm.PreconditionTypeFeedbackSourceCompleted:
					validateExecBetweenPrecondition(
						t,
						ap,
						p.FeedbackSource,
						fsm.NewMaxRevisionBound(p.FeedbackSource),
						fsm.NewInfinityRevisionBound(),
						ch,
						preconditionNotFeedbackSourceProblemID,
						preconditionReachableFeedbackSourceProblemID,
						preconditionInvalidExecBoundProblemID,
						preconditionInvalidExecRangeProblemID,
					)

				case fsm.PreconditionTypeExecBetween:
					validateExecBetweenPrecondition(
						t,
						ap,
						p.ExecBetweenTarget,
						p.ExecBetweenBegin,
						p.ExecBetweenEnd,
						ch,
						preconditionNotFeedbackSourceProblemID,
						preconditionReachableFeedbackSourceProblemID,
						preconditionInvalidExecBoundProblemID,
						preconditionInvalidExecRangeProblemID,
					)

				case fsm.PreconditionTypeExecutable:
					if !t.PFD.AtomicProcesses.Contains(pfd.AtomicProcessID.Compare, p.Executable) {
						ch <- checkers.NewProblem(
							preconditionNotAtomicProcessProblemID,
							checkers.SeverityError,
							fsmcommon.NewLocation(
								fsmcommon.LocationTypeAtomicProcessTable,
								fsmcommon.NewAtomicProcessID(ap),
								fsmcommon.NewAtomicProcessID(p.Executable),
							),
						)
						return
					}

					aps := sets.NewWithCapacity[pfd.AtomicProcessID](t.PFD.AtomicProcesses.Len())
					t.PFD.CollectReachableAtomicProcessesExceptFeedback(ap, aps, t.Logger)
					if aps.Contains(pfd.AtomicProcessID.Compare, p.Executable) {
						ch <- checkers.NewProblem(
							preconditionReachableExecutableTargetProblemID,
							checkers.SeverityError,
							fsmcommon.NewLocation(
								fsmcommon.LocationTypeAtomicProcessTable,
								fsmcommon.NewAtomicProcessID(ap),
								fsmcommon.NewAtomicProcessID(p.Executable),
							),
						)
					}

				case fsm.PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted,
					fsm.PreconditionTypeOr,
					fsm.PreconditionTypeAnd,
					fsm.PreconditionTypeTrue,
					fsm.PreconditionTypeNot:
					break

				default:
					panic(fmt.Sprintf("unexpected precondition type: %v", p.Type))
				}
			})
		}

		cycles := preconditionExecutableReferenceGraphCycles(m)
		for _, cycle := range cycles.Iter() {
			loc := fsmcommon.NewLocation(fsmcommon.LocationTypeAtomicProcessTable, cycle...)
			ch <- checkers.NewProblem(preconditionCyclicExecutableReferenceProblemID, checkers.SeverityError, loc)
		}
		return nil
	},
}

type resolvedRevisionBound struct {
	IsInfinity bool
	Value      int
	IDs        []fsmcommon.ID
}

func validateExecBetweenPrecondition(
	t *fsmcommon.Target,
	ap pfd.AtomicProcessID,
	target pfd.AtomicDeliverableID,
	begin *fsm.RevisionBound,
	end *fsm.RevisionBound,
	ch chan<- checkers.Problem,
	preconditionNotFeedbackSourceProblemID checkers.ProblemID,
	preconditionReachableFeedbackSourceProblemID checkers.ProblemID,
	preconditionInvalidExecBoundProblemID checkers.ProblemID,
	preconditionInvalidExecRangeProblemID checkers.ProblemID,
) {
	locIDs := []fsmcommon.ID{
		fsmcommon.NewAtomicProcessID(ap),
		fsmcommon.NewAtomicDeliverableID(target),
	}
	if !t.PFD.AtomicDeliverables.Contains(pfd.AtomicDeliverableID.Compare, target) || t.PFD.FeedbackDestinationAtomicProcesses(target).Len() == 0 {
		ch <- checkers.NewProblem(
			preconditionNotFeedbackSourceProblemID,
			checkers.SeverityError,
			fsmcommon.NewLocation(
				fsmcommon.LocationTypeAtomicProcessTable,
				locIDs...,
			),
		)
		return
	}

	ds := sets.NewWithCapacity[pfd.AtomicDeliverableID](t.PFD.AtomicDeliverables.Len())
	t.PFD.CollectReachableDeliverablesExceptFeedback(ap, ds, t.Logger)
	if ds.Contains(pfd.AtomicDeliverableID.Compare, target) {
		ch <- checkers.NewProblem(
			preconditionReachableFeedbackSourceProblemID,
			checkers.SeverityError,
			fsmcommon.NewLocation(
				fsmcommon.LocationTypeAtomicProcessTable,
				locIDs...,
			),
		)
	}

	beginResolved, beginErr := resolveRevisionBound(t, begin)
	if beginErr != nil {
		ch <- checkers.NewProblem(
			preconditionInvalidExecBoundProblemID,
			checkers.SeverityError,
			fsmcommon.NewLocation(
				fsmcommon.LocationTypeAtomicProcessTable,
				append(locIDs, beginErr...)...,
			),
		)
		return
	}

	endResolved, endErr := resolveRevisionBound(t, end)
	if endErr != nil {
		ch <- checkers.NewProblem(
			preconditionInvalidExecBoundProblemID,
			checkers.SeverityError,
			fsmcommon.NewLocation(
				fsmcommon.LocationTypeAtomicProcessTable,
				append(locIDs, endErr...)...,
			),
		)
		return
	}

	if compareResolvedRevisionBounds(beginResolved, endResolved) >= 0 {
		related := append([]fsmcommon.ID{}, locIDs...)
		related = append(related, beginResolved.IDs...)
		related = append(related, endResolved.IDs...)
		ch <- checkers.NewProblem(
			preconditionInvalidExecRangeProblemID,
			checkers.SeverityError,
			fsmcommon.NewLocation(
				fsmcommon.LocationTypeAtomicProcessTable,
				related...,
			),
		)
	}
}

func resolveRevisionBound(t *fsmcommon.Target, bound *fsm.RevisionBound) (resolvedRevisionBound, []fsmcommon.ID) {
	switch bound.Type {
	case fsm.RevisionBoundTypeInt:
		return resolvedRevisionBound{
			Value: bound.IntValue,
		}, nil
	case fsm.RevisionBoundTypeInfinity:
		return resolvedRevisionBound{
			IsInfinity: true,
			Value:      math.MaxInt,
		}, nil
	case fsm.RevisionBoundTypeMaxRevision:
		related := []fsmcommon.ID{fsmcommon.NewAtomicDeliverableID(bound.MaxRevisionDeliverable)}
		if !t.PFD.AtomicDeliverables.Contains(pfd.AtomicDeliverableID.Compare, bound.MaxRevisionDeliverable) || t.PFD.FeedbackDestinationAtomicProcesses(bound.MaxRevisionDeliverable).Len() == 0 {
			return resolvedRevisionBound{}, related
		}
		if !t.Memoized.HasMaxRevisionMap {
			return resolvedRevisionBound{}, related
		}
		maxRevisionText, ok := t.Memoized.MaxRevisionMap[bound.MaxRevisionDeliverable]
		if !ok {
			return resolvedRevisionBound{}, related
		}
		maxRevision, err := fsmtable.ValidateMaxRevision(maxRevisionText, true)
		if err != nil {
			return resolvedRevisionBound{}, related
		}
		return resolvedRevisionBound{
			Value: maxRevision,
			IDs:   related,
		}, nil
	default:
		panic(fmt.Sprintf("unexpected revision bound type: %v", bound.Type))
	}
}

func compareResolvedRevisionBounds(a, b resolvedRevisionBound) int {
	if a.IsInfinity {
		if b.IsInfinity {
			return 0
		}
		return 1
	}
	if b.IsInfinity {
		return -1
	}
	return a.Value - b.Value
}

func preconditionExecutableReferenceGraphCycles(preconditionMap map[pfd.AtomicProcessID]*fsm.Precondition) *sets.Set[[]fsmcommon.ID] {
	f1 := pairs.Compare(graph.Node.Compare, graph.Node.Compare)
	f2 := cmp2.CompareSlice[[]fsmcommon.ID](fsmcommon.ID.Compare)

	nodes := sets.NewWithCapacity[graph.Node](len(preconditionMap))
	for ap := range preconditionMap {
		nodes.Add(graph.Node.Compare, graph.Node(ap))
	}

	edges := sets.NewWithCapacity[*pairs.Pair[graph.Node, graph.Node]](0)
	for ap, precondition := range preconditionMap {
		precondition.Traverse(func(p *fsm.Precondition) {
			if p.Type != fsm.PreconditionTypeExecutable {
				return
			}
			edges.Add(f1, pairs.New(graph.Node(ap), graph.Node(p.Executable)))
		})
	}

	graph := graph.Graph{
		Nodes: nodes,
		Edges: edges,
	}

	res := sets.NewWithCapacity[[]fsmcommon.ID](0)
	for _, cycle := range graph.Cycles().Iter() {
		locs := make([]fsmcommon.ID, len(cycle))
		for i, node := range cycle {
			locs[i] = fsmcommon.NewAtomicProcessID(pfd.AtomicProcessID(node))
		}
		res.Add(f2, locs)
	}
	return res
}

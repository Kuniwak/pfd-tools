package fsmchecker

import (
	"fmt"

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
					if !t.PFD.AtomicDeliverables.Contains(pfd.AtomicDeliverableID.Compare, p.FeedbackSource) || t.PFD.FeedbackDestinationAtomicProcesses(p.FeedbackSource).Len() == 0 {
						ch <- checkers.NewProblem(
							preconditionNotFeedbackSourceProblemID,
							checkers.SeverityError,
							fsmcommon.NewLocations(
								fsmcommon.NewLocation(
									fsmcommon.LocationTypeAtomicProcessTable,
									fsmcommon.NewAtomicProcessID(ap),
									fsmcommon.NewAtomicDeliverableID(p.FeedbackSource),
								),
							)...,
						)
						return
					}

					ds := sets.NewWithCapacity[pfd.AtomicDeliverableID](t.PFD.AtomicDeliverables.Len())
					t.PFD.CollectReachableDeliverablesExceptFeedback(ap, ds, t.Logger)
					if ds.Contains(pfd.AtomicDeliverableID.Compare, p.FeedbackSource) {
						ch <- checkers.NewProblem(
							preconditionReachableFeedbackSourceProblemID,
							checkers.SeverityError,
							fsmcommon.NewLocation(
								fsmcommon.LocationTypeAtomicProcessTable,
								fsmcommon.NewAtomicProcessID(ap),
								fsmcommon.NewAtomicDeliverableID(p.FeedbackSource),
							),
						)
					}

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

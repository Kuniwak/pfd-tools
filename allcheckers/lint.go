package allcheckers

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
	"golang.org/x/sync/errgroup"

	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

type Target struct {
	PFD                       *pfd.PFD
	AtomicProcessTable        *pfd.AtomicProcessTable
	AtomicDeliverableTable    *pfd.AtomicDeliverableTable
	CompositeProcessTable     *pfd.CompositeProcessTable
	CompositeDeliverableTable *pfd.CompositeDeliverableTable
	ResourceTable             *fsmtable.ResourceTable
	MilestoneTable            *fsmtable.MilestoneTable
	GroupTable                *fsmtable.GroupTable
	Model                     execmodel.Model
}

type LintFunc func(t Target, ch chan<- checkers.Problem) error

func NewLintFunc(logger *slog.Logger) LintFunc {
	return func(t Target, ch chan<- checkers.Problem) error {

		defer close(ch)

		var eg errgroup.Group

		eg.Go(func() error {
			m := pfdcommon.NewMemoized(t.PFD, logger)
			if err := PFDCheckers.Check(pfdcommon.NewTarget(t.PFD, t.AtomicProcessTable, t.AtomicDeliverableTable, t.CompositeProcessTable, t.CompositeDeliverableTable, m), ch); err != nil {
				return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
			}

			return nil
		})
		eg.Go(func() error {
			if err := t.Model.Validate(); err != nil {
				return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
			}

			p, err := pfd.NewSafePFDByUnsafePFD(t.PFD)
			if err != nil {

				return nil
			}

			m, err := fsmcommon.NewMemoized(t.AtomicProcessTable, t.AtomicDeliverableTable, t.ResourceTable, t.MilestoneTable)
			if err != nil {
				return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
			}
			if err := FSMCheckers.Check(&fsmcommon.Target{PFD: p, AtomicProcessTable: t.AtomicProcessTable, AtomicDeliverableTable: t.AtomicDeliverableTable, ResourceTable: t.ResourceTable, MilestoneTable: t.MilestoneTable, GroupTable: t.GroupTable, Model: t.Model, Memoized: m, Logger: logger}, ch); err != nil {
				return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
			}

			return nil
		})

		if err := eg.Wait(); err != nil {
			return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
		}

		return nil
	}
}

func Lint(t Target, logger *slog.Logger) ([]checkers.Problem, error) {
	lintFunc := NewLintFunc(logger)
	ch := make(chan checkers.Problem)

	var eg errgroup.Group
	eg.Go(func() error {
		if err := lintFunc(t, ch); err != nil {
			return fmt.Errorf("allcheckers.Lint: %w", err)
		}
		return nil
	})

	ps := make([]checkers.Problem, 0)
	eg.Go(func() error {
		for problem := range ch {
			ps = append(ps, problem)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("allcheckers.Lint: %w", err)
	}

	return ps, nil
}

func CompositeDeliverableTableProblems(cdt *pfd.CompositeDeliverableTable, logger *slog.Logger) ([]checkers.Problem, error) {

	if _, err := pfd.FlattenDeliverableComposition(cdt.NodeIDMap(logger), sets.New((*pfd.Node).Compare)); err != nil {
		var cycleErr *pfd.CompositionCycleError
		if !errors.As(err, &cycleErr) {
			return nil, fmt.Errorf("allcheckers.CompositeDeliverableTableProblems: %w", err)
		}
		return []checkers.Problem{checkers.NewProblem("acyclic-cd-comp", checkers.SeverityError, pfdcommon.NewLocations(pfdcommon.NewLocation(pfdcommon.LocationTypeCompositeDeliverableTable, cycleErr.Cycle...))...)}, nil
	}
	return []checkers.Problem{}, nil
}

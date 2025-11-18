package allcheckers

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker/fsmcommon"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"golang.org/x/sync/errgroup"

	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers/pfdcommon"
)

type LintFunc func(
	up *pfd.PFD,
	apTable *pfd.AtomicProcessTable,
	adTable *pfd.AtomicDeliverableTable,
	cpTable *pfd.CompositeProcessTable,
	cdTable *pfd.CompositeDeliverableTable,
	rTable *fsmtable.ResourceTable,
	mt *fsmtable.MilestoneTable,
	gt *fsmtable.GroupTable,
	ch chan<- checkers.Problem,
) error

func NewLintFunc(logger *slog.Logger) LintFunc {
	return func(
		up *pfd.PFD,
		apTable *pfd.AtomicProcessTable,
		adTable *pfd.AtomicDeliverableTable,
		cpTable *pfd.CompositeProcessTable,
		cdTable *pfd.CompositeDeliverableTable,
		rTable *fsmtable.ResourceTable,
		mt *fsmtable.MilestoneTable,
		gt *fsmtable.GroupTable,
		ch chan<- checkers.Problem,
	) error {
		var eg errgroup.Group

		runningWorkers := 2
		var runningWorkersMu sync.Mutex

		eg.Go(func() error {
			defer func() {
				runningWorkersMu.Lock()
				runningWorkers -= 1
				runningWorkersMu.Unlock()
				if runningWorkers == 0 {
					close(ch)
				}
			}()

			m := pfdcommon.NewMemoized(up, logger)
			if err := PFDCheckers.Check(pfdcommon.NewTarget(up, apTable, adTable, cpTable, cdTable, m), ch); err != nil {
				return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
			}

			return nil
		})
		eg.Go(func() error {
			defer func() {
				runningWorkersMu.Lock()
				runningWorkers -= 1
				runningWorkersMu.Unlock()
				if runningWorkers == 0 {
					close(ch)
				}
			}()

			p, err := pfd.NewSafePFDByUnsafePFD(up)
			if err != nil {
				// NOTE: Should be reported by PFD Checker side, so skip.
				return nil
			}

			m, err := fsmcommon.NewMemoized(apTable, adTable, rTable, mt)
			if err != nil {
				return fmt.Errorf("allcheckers.NewLintFunc: %w", err)
			}
			if err := FSMCheckers.Check(fsmcommon.NewTarget(p, apTable, adTable, rTable, mt, gt, m, logger), ch); err != nil {
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

func Lint(
	p *pfd.PFD,
	apTable *pfd.AtomicProcessTable,
	adTable *pfd.AtomicDeliverableTable,
	cpTable *pfd.CompositeProcessTable,
	cdTable *pfd.CompositeDeliverableTable,
	rTable *fsmtable.ResourceTable,
	mt *fsmtable.MilestoneTable,
	gt *fsmtable.GroupTable,
	logger *slog.Logger,
) ([]checkers.Problem, error) {
	lintFunc := NewLintFunc(logger)
	ch := make(chan checkers.Problem)

	var eg errgroup.Group
	eg.Go(func() error {
		if err := lintFunc(p, apTable, adTable, cpTable, cdTable, rTable, mt, gt, ch); err != nil {
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

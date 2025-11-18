package cmd

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/allcheckers"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding/fsmtsv"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/version"
	"golang.org/x/sync/errgroup"
)

var ErrProblemsFound = errors.New("cmd.MainCommandByOptions: problems found")

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	opts, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(opts, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	return 0
}

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	parseOpts := &pfdfmt.ParseOptions{}
	var compositeDeliverableTable *pfd.CompositeDeliverableTable
	if opts.HasCompositeDeliverableTable {
		var err error
		compositeDeliverableTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		parseOpts.CompositeDeliverableTable = compositeDeliverableTable
	}

	p, err := pfdfmt.Parse("", opts.PFDReader, parseOpts, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	var atomicTable *pfd.AtomicProcessTable
	if opts.HasAtomicProcessTable {
		atomicTable, err = pfdtsv.ParseAtomicProcessTable(opts.AtomicProcessTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		atomicTable = nil
	}

	var atomicDeliverableTable *pfd.AtomicDeliverableTable
	if opts.HasAtomicDeliverableTable {
		atomicDeliverableTable, err = pfdtsv.ParseAtomicDeliverableTable(opts.AtomicDeliverableTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		atomicDeliverableTable = nil
	}

	var compositeProcessTable *pfd.CompositeProcessTable
	if opts.HasCompositeProcessTable {
		compositeProcessTable, err = pfdtsv.ParseCompositeProcessTable(opts.CompositeProcessTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		compositeProcessTable = nil
	}

	var resourceTable *fsmtable.ResourceTable
	if opts.HasResourceTable {
		resourceTable, err = fsmtsv.ParseResourceTable(opts.ResourceTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		resourceTable = nil
	}

	var milestoneTable *fsmtable.MilestoneTable
	if opts.HasMilestoneTable {
		milestoneTable, err = fsmtsv.ParseMilestoneTable(opts.MilestoneTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		milestoneTable = nil
	}
	var groupTable *fsmtable.GroupTable
	if opts.HasGroupTable {
		groupTable, err = fsmtsv.ParseGroupTable(opts.GroupTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		groupTable = nil
	}
	var eg errgroup.Group
	ch := make(chan checkers.Problem)
	lintFunc := allcheckers.NewLintFunc(logger)

	eg.Go(func() error {
		if err = lintFunc(p, atomicTable, atomicDeliverableTable, compositeProcessTable, compositeDeliverableTable, resourceTable, milestoneTable, groupTable, ch); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		return nil
	})

	var count int
	eg.Go(func() error {
		var err error
		count, err = opts.Reporter(ch)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if count > 0 {
		return ErrProblemsFound
	}

	return nil
}

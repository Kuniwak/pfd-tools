package cmd

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/masterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding/fsmtsv"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/version"
)

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

	plan, err := fsm.ParsePlan(opts.PlanReader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	atomicProcessTable, err := pfdtsv.ParseAtomicProcessTable(opts.AtomicProcessTableReader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	aps := atomicProcessTable.AtomicProcesses()

	milestoneTable, err := fsmtsv.ParseMilestoneTable(opts.MilestoneTableReader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	groupTable, err := fsmtsv.ParseGroupTable(opts.GroupTableReader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	gs := groupTable.Groups()

	mgm, err := fsmtable.MilestoneGraphByTable(milestoneTable, gs)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	rawGroupMap, err := fsmtable.RawGroupsMap(atomicProcessTable, fsmtable.DefaultGroupColumnMatchFunc)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	gsm, err := fsmtable.ValidateGroupsMap(rawGroupMap)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	rawMilestoneMap, err := fsmtable.RawMilestoneMap(atomicProcessTable, fsmtable.DefaultMilestoneColumnMatchFunc)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	mm, err := fsmtable.ValidateMilestoneMap(rawMilestoneMap)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	timeline1, err := fsmmasterschedule.NewTimelineFromPlan(plan, aps, gs, gsm, mm, mgm)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	timeline2 := fsmmasterschedule.NewTimelineWithBufferMultiplier(timeline1, opts.BufferMultiplier)

	master := fsmmasterschedule.NewMasterSchedule(
		timeline2,
		opts.BusinessTimeFuncOptions.BusinessTimeFunc,
		opts.BusinessTimeFuncOptions.StartDay,
		groupTable.DescriptionMap(),
		milestoneTable.DescriptionMap(),
	)

	if err := masterschedule.WriteGoogleSpreadsheetTSV(inout.Stdout, master); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	return nil
}

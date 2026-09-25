package cmd

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/version"
)

const ShortHelp = "実行計画と分類表から、行・バー別のマスタースケジュールを作成します。"

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

	if opts.CommonOptions.ShortHelp {
		fmt.Fprintln(inout.Stdout, ShortHelp)
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

	masterTable, err := mastertsv.Parse(opts.MasterTSVReader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	master, err := fsmmasterschedule.NewMasterScheduleFromPlan(
		plan,
		masterTable,
		opts.BufferMultiplier,
		opts.BusinessTimeFuncOptions.BusinessTimeFunc,
		opts.BusinessTimeFuncOptions.StartDay,
		opts.RowDescriptions,
		opts.BarDescriptions,
		opts.Logger,
	)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if err := opts.WriteMasterSchedule(inout.Stdout, master); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	return nil
}

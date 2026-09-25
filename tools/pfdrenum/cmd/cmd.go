package cmd

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/version"
)

const ShortHelp = "PFD の要素を採番します。既存の ID は維持されます。"

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

func NewRenumberBase(opts *Options) pfd.RenumberBase {
	return pfd.RenumberBase{
		Plan:                 opts.RenumberPlan,
		MaxProcessNumber:     opts.MaxProcessNumber,
		MaxDeliverableNumber: opts.MaxDeliverableNumber,
	}
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

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	input, err := io.ReadAll(opts.Reader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if opts.Mode == ModeEmitMaxID {
		nodes, err := pfdfmt.ReadDrawioNodes(input, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		maxProcessNumber, maxDeliverableNumber := pfddrawio.MaxNumberedIDs(nodes, logger)

		if _, err := fmt.Fprintln(inout.Stdout, pfd.FormatMaxIDs(maxProcessNumber, maxDeliverableNumber)); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		return nil
	}

	if opts.Mode == ModeEmitPlan || opts.Mode == ModeExtendPlan {

		if opts.Inplace {
			return fmt.Errorf("cmd.MainCommandByOptions: the renumber plan cannot be written in place")
		}

		nodes, err := pfdfmt.ReadDrawioNodes(input, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		base := NewRenumberBase(opts)

		plan, err := pfddrawio.NewRenumberPlanWithBase(nodes, base, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		out := pfd.MergeRenumberPlans(base.Plan, plan.NewlyNumbered())

		if err := pfdtsv.WriteRenumberPlan(inout.Stdout, out); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		return nil
	}

	transform := pfddrawio.RenumberWithBase(NewRenumberBase(opts))
	if opts.Mode == ModeApplyPlan {
		transform = pfddrawio.RenumberByPlan(opts.RenumberPlan)
	}

	out, err := pfdfmt.TransformDrawioBytes(input, transform, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	inplacePath := ""
	if opts.Inplace {
		inplacePath = opts.InputFilePath
	}
	if err := tools.WriteOutput(inout.Stdout, inplacePath, out); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	return nil
}

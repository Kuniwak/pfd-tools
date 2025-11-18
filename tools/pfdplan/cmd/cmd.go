package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(options, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	return 0
}

func MainCommandByOptions(options *Options, inout *cli.ProcInout) error {
	if options.CommonOptions.Help {
		return nil
	}
	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, options.CommonOptions.LogLevel))

	fsmEnvSeed, err := tools.ParseFSMEnvSeed(options.FSMOptions, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	if err := tools.ValidateFSMEnvSeed(fsmEnvSeed, logger, options.CommonOptions.Locale); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	env, err := tools.FSMPrepare(fsmEnvSeed, options.CommonOptions.Locale, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	plans, err := options.SearchFunc(env)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if options.OutDir == "" {
		firstPlan, ok := plans.At(0)
		if !ok {
			return fmt.Errorf("cmd.MainCommandByOptions: no plans")
		}

		if err := options.PlanReporter(inout.Stdout, firstPlan, env.PFD.AtomicProcessDescriptionMap); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	} else {
		if err := os.MkdirAll(options.OutDir, 0755); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		if plans.Len() == 0 {
			return fmt.Errorf("cmd.MainCommandByOptions: no plans")
		}

		var ext string
		switch options.OutputFormat {
		case tools.PlanOutputFormatGoogleSpreadsheetTSV:
			ext = ".tsv"
		case tools.PlanOutputFormatPlanJSON, tools.PlanOutputFormatTimelineJSON:
			ext = ".json"
		default:
			panic(fmt.Sprintf("cmd.MainCommandByOptions: invalid output format: %q", options.OutputFormat))
		}

		var dg sugar.DeferGroup
		for i, plan := range plans.Iter() {
			if err := dg.Run(func() error {
				outPath := filepath.Join(options.OutDir, fmt.Sprintf("plan_%03d%s", i, ext))
				f, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
				defer f.Close()
				if err := options.PlanReporter(f, plan, env.PFD.AtomicProcessDescriptionMap); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

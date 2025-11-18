package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/slograw"
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

	f, err := os.Open(options.PlanPath)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	defer f.Close()

	plan, err := fsm.ParsePlan(f)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if err := options.PlanReporter(inout.Stdout, plan, env.PFD.AtomicProcessDescriptionMap); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	return nil
}

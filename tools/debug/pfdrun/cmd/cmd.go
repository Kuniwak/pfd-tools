package cmd

import (
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmrun"
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

	var plan *fsm.Plan
	if options.HasPlan {
		plan, err = fsm.ParsePlan(options.PlanReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	}

	r := fsmrun.NewRunner(inout.Stdin, inout.Stdout, env, fsmrun.DefaultCommands)
	if err := r.Run(plan); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	return nil
}

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmviz"
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

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	fsmEnvSeed, err := tools.ParseFSMEnvSeed(opts.FSMOptions, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	if err := tools.ValidateFSMEnvSeed(fsmEnvSeed, logger, opts.CommonOptions.Locale); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	env, err := tools.FSMPrepare(fsmEnvSeed, opts.CommonOptions.Locale, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	plans, err := opts.SearchFunc(env)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	g, err := fsm.Graph(env, opts.MaxDepth, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	dot, err := fsmviz.Dot(env.IsCompleted, g, plans, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if err := dot.Write(inout.Stdout); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	return nil
}

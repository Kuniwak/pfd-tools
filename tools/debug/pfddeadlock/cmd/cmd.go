package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
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

	g, err := fsm.Graph(env, math.MaxInt, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	if err := os.MkdirAll(options.OutDir, 0755); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	ch := make(chan fsm.Deadlock)

	go func() {
		defer close(ch)
		g.FindDeadlocks(env.IsCompleted, ch)
	}()

	i := 0
	var dg sugar.DeferGroup
	for deadlock := range ch {
		i++
		if err := dg.Run(func() error {
			outPath := filepath.Join(options.OutDir, fmt.Sprintf("deadlock_S%03d_%03d.log", i, deadlock.StateID))
			logger.Info("cmd.MainCommandByOptions: writing deadlock", "outPath", outPath)

			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			defer f.Close()

			for _, pair := range deadlock.Path {
				s, ok := g.Nodes[pair.First]
				if !ok {
					panic("cmd.MainCommandByOptions: state not found")
				}
				io.WriteString(f, "State: ")
				s.Write(f)
				io.WriteString(f, "Allocation: ")
				pair.Second.Write(f)
				io.WriteString(f, "\n")
			}

			deadlockState, ok := g.Nodes[deadlock.StateID]
			if !ok {
				panic("cmd.MainCommandByOptions: state not found")
			}
			deadlockState.Write(f)
			io.WriteString(f, "\n")
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

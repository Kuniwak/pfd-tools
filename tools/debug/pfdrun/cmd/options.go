package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	FSMOptions    *tools.FSMOptions
	HasPlan       bool
	PlanReader    io.Reader
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdrun", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: pfdrun [options] [-p <pfd> [-a <atomic-process-table>] [-d <deliverable-table>] [-r <resource-table>]] [-f <config>] [-plan <plan>]")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
    $ pfdrun -f path/to/config.json

    $ pfdrun -p path/to/pfd.drawio -a path/to/atomic_proc.tsv -d path/to/deliv.tsv -r path/to/resource.tsv

    $ pfdrun -f path/to/config.json -plan path/to/plan.json
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var configShortPath, configLongPath string
	var fsmRawOptions tools.FSMRawOptions
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

	var planLongPath string
	flags.StringVar(&planLongPath, "plan", "", "path to the plan")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Options{CommonOptions: &tools.CommonOptions{Help: true}}, nil
		}
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	commonOptions, err := tools.ValidateCommonOptions(&commonRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	fsmOptions, err := tools.ValidateFSMOptionsOrConfig(&fsmRawOptions, &configShortPath, &configLongPath, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var planReader io.Reader
	var hasPlan bool
	if planLongPath != "" {
		hasPlan = true
		planReader, err = os.Open(planLongPath)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	} else {
		hasPlan = false
	}

	return &Options{
		CommonOptions: commonOptions,
		FSMOptions:    fsmOptions,
		HasPlan:       hasPlan,
		PlanReader:    planReader,
	}, nil
}

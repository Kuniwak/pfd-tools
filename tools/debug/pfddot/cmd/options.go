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
	CommonOptions                   *tools.CommonOptions
	PFDReader                       io.Reader
	CompositeDeliverableTableReader io.Reader
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfddot", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfddot [options] -p <pfd> -cd <composite-deliverable-table>", ShortHelp)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var pfdPathShortFlag, pfdPathLongFlag string
	var compositeDeliverableTablePathShortFlag, compositeDeliverableTablePathLongFlag string
	tools.DeclarePFDOptions(flags, &pfdPathShortFlag, &pfdPathLongFlag)
	tools.DeclareCompositeDeliverableTableOptions(flags, &compositeDeliverableTablePathShortFlag, &compositeDeliverableTablePathLongFlag)

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

	if commonOptions.ShortHelp {
		return &Options{CommonOptions: commonOptions}, nil
	}
	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	pfdReader, _, err := tools.ValidatePFDOptions(&pfdPathShortFlag, &pfdPathLongFlag, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}
	var compositeDeliverableTableReader io.Reader
	if compositeDeliverableTablePathShortFlag != "" || compositeDeliverableTablePathLongFlag != "" {
		compositeDeliverableTableReader, _, err = tools.ValidateCompositeDeliverableTableOptions(&compositeDeliverableTablePathShortFlag, &compositeDeliverableTablePathLongFlag, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	return &Options{CommonOptions: commonOptions, PFDReader: pfdReader, CompositeDeliverableTableReader: compositeDeliverableTableReader}, nil
}

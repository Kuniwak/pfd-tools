package cmd

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	HeadCount     int
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdres", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdres [-locale <ja|en>] <head-count>", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
  $ pfdres 3
  A:1;B:1;C:1

  $ pfdres 27
  A:1;B:1;...;Z:1;AA:1
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

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

	if flags.NArg() < 1 {
		flags.Usage()
		return nil, fmt.Errorf("cmd.ParseOptions: head count is required")
	}
	if flags.NArg() > 1 {
		return nil, fmt.Errorf("cmd.ParseOptions: too many arguments")
	}

	headCount, err := strconv.Atoi(flags.Arg(0))
	if err != nil || headCount < 1 {
		return nil, fmt.Errorf("cmd.ParseOptions: head count must be a positive integer: %q", flags.Arg(0))
	}

	return &Options{
		CommonOptions: commonOptions,
		HeadCount:     headCount,
	}, nil
}

package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/tools/pfdhelp/help"
)

type Options struct {
	CommonOptions *tools.CommonOptions

	Tools []help.ToolHelp

	Short bool
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdhelp", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdhelp [options] [tool ...]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
  $ pfdhelp
  # bizday
  ...
  # pfdlint
  ...

  $ pfdhelp pfdlint pfdplan
  # pfdlint
  ...
  # pfdplan
  ...
`)
	}
	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var short bool
	flags.BoolVar(&short, "short", false, "各ツールの一行説明を name<TAB>説明 の TSV でまとめて表示する")

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

	selectedTools, err := help.SelectTools(flags.Args())
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	return &Options{
		CommonOptions: commonOptions,
		Tools:         selectedTools,
		Short:         short,
	}, nil
}

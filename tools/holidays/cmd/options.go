package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("holidays", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: holidays -locale <locale>")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
  $ holidays -locale ja
  2025-01-01
  2025-01-02
  2025-01-03
  ...
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

	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	switch commonOptions.Locale {
	case locale.LocaleJa:
		break
	default:
		return nil, fmt.Errorf("cmd.ParseOptions: only Japanese holidays are supported")
	}

	return &Options{
		CommonOptions: commonOptions,
	}, nil
}

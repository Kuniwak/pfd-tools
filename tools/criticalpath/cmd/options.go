package cmd

import (
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	FSMOptions    *tools.FSMOptions
	SearchFunc    fsm.SearchFunc
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("criticalpath", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: criticalpath [options] <pfd>")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var configShortPath, configLongPath string
	var fsmRawOptions tools.FSMRawOptions
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

	var searchRawOptions tools.SearchRawOptions
	tools.DeclareSearchOptions(flags, &searchRawOptions, rand.Int64())

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

	searchFunc, err := tools.ValidateSearchOptions(&searchRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	return &Options{
		CommonOptions: commonOptions,
		FSMOptions:    fsmOptions,
		SearchFunc:    searchFunc,
	}, nil
}

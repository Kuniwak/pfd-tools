package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	FSMOptions    *tools.FSMOptions
	OutDir        string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfddeadlock", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfddeadlock [options] -p <pfd> [-ap <atomic-process-table>] [-ad <atomic-deliverable-table>] [-r <resource-table>]", ShortHelp)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var fsmRawOptions tools.FSMRawOptions
	var configShortPath, configLongPath string
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

	var outDir string
	flags.StringVar(&outDir, tools.OutDirFlag, "", "output directory")

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

	mergedOptions, basePath, err := tools.ReadFSMRawOptions(&configShortPath, &configLongPath, fsmRawOptions, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	fsmOptions, err := tools.ValidateAllFSMOptions(&mergedOptions, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if outDir == "" {
		return nil, fmt.Errorf("cmd.ParseOptions: -%s is required", tools.OutDirFlag)
	}

	s, err := os.Stat(outDir)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("cmd.ParseOptions: output directory is not a directory: %w", err)
	}

	return &Options{
		CommonOptions: commonOptions,
		FSMOptions:    fsmOptions,
		OutDir:        outDir,
	}, nil
}

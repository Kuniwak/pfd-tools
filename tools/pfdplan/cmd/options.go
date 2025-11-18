package cmd

import (
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmreporter"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	FSMOptions    *tools.FSMOptions
	PlanReporter  fsmreporter.PlanReporter
	SearchFunc    fsm.SearchFunc
	OutDir        string
	OutputFormat  tools.PlanOutputFormat
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdplan", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: pfdplan [-debug|-silent] -p <pfd> -a <atomic-process-table> -r <resource-table> -d <deliverable-table> [-start-time <start-time> -duration <duration> [-weekdays <weekdays>] [-not-biz-days <not-biz-days>]|-o plan-json|timeline-json|google-spreadsheet-tsv]")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
    $ pfdplan -p path/to/pfd.drawio -a path/to/atomic_proc.tsv -d path/to/deliv.tsv -r path/to/resource.tsv -start-time 10:00 -duration 9 -not-biz-days <(holidays -locale ja)
    AtomicProcess[NumOfComplete]     StartTime       EndTime
    P1[1]   2025-10-04T00:00:00+09:00       2025-10-11T04:30:00+09:00
    P1[2]   2025-10-04T04:30:00+09:00       2025-10-11T06:45:00+09:00
    P1[3]   2025-10-04T06:45:00+09:00       2025-10-11T06:45:00+09:00
	...
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var fsmRawOptions tools.FSMRawOptions
	var configShortPath, configLongPath string
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

	var planOutputFormatRawOptions tools.PlanOutputFormatRawOptions
	tools.DeclarePlanOutputFormatOptions(flags, &planOutputFormatRawOptions)

	outDirFlag := flags.String("out-dir", "", "output directory")

	var searchRawOptions tools.SearchRawOptions
	tools.DeclareSearchOptions(flags, &searchRawOptions, rand.Int64())

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Options{CommonOptions: &tools.CommonOptions{Help: true}, FSMOptions: nil}, nil
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

	planReporter, outputFormat, err := tools.ValidatePlanOutputFormat(&planOutputFormatRawOptions, commonOptions.Logger)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	searchFunc, err := tools.ValidateSearchOptions(&searchRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	outDir := *outDirFlag
	if outDir != "" {
		s, err := os.Stat(outDir)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		if !s.IsDir() {
			return nil, fmt.Errorf("cmd.ParseOptions: output directory is not a directory: %w", err)
		}
	}

	return &Options{
		CommonOptions: commonOptions,
		FSMOptions:    fsmOptions,
		PlanReporter:  planReporter,
		SearchFunc:    searchFunc,
		OutDir:        outDir,
		OutputFormat:  outputFormat,
	}, nil
}

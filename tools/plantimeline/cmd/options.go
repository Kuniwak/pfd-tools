package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmreporter"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	FSMOptions    *tools.FSMOptions
	PlanPath      string
	PlanReporter  fsmreporter.PlanReporter
	OutputFormat  tools.PlanOutputFormat
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("plantimeline", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: plantimeline [options] -f <project> <plan>")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
    $ plantimeline -f path/to/project.json path/to/plan.json
	AtomicProcess   NumOfComplete   AllocatedResources      Description     StartTime       EndTime Start   End
	P1      0       R1      Process 1        2025-11-18 10:00:00     2025-11-20 10:00:00     0       2
	...

	$ plantimeline -f path/to/project.json -out-format timeline-json path/to/plan.json
	AtomicProcess   NumOfComplete   AllocatedResources      Description     StartTime       EndTime Start   End
	P1      0       R1      Process 1        2025-11-18 10:00:00     2025-11-20 10:00:00     0       2
	...
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var planOutputFormatRawOptions tools.PlanOutputFormatRawOptions
	tools.DeclarePlanOutputFormatOptions(flags, &planOutputFormatRawOptions)

	var fsmRawOptions tools.FSMRawOptions
	var configShortPath, configLongPath string
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

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

	planPath := flags.Arg(0)
	if planPath == "" {
		return nil, fmt.Errorf("cmd.ParseOptions: plan path is required")
	}

	planReporter, outputFormat, err := tools.ValidatePlanOutputFormat(&planOutputFormatRawOptions, slog.New(slograw.NewHandler(inout.Stderr, commonOptions.LogLevel)))
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	fsmOptions, err := tools.ValidateFSMOptionsOrConfig(&fsmRawOptions, &configShortPath, &configLongPath, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	return &Options{
		CommonOptions: commonOptions,
		FSMOptions:    fsmOptions,
		PlanPath:      planPath,
		PlanReporter:  planReporter,
		OutputFormat:  outputFormat,
	}, nil
}

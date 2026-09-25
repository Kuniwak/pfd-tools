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
		tools.PrintUsageHeader(flags, "Usage: plantimeline [options] -f <project> [<plan>]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
    $ plantimeline -f path/to/project.json path/to/plan.json
	AtomicProcess   NumOfComplete   AllocatedResources      Description     StartTime       EndTime Start   End
	P1      0       R1      プロセス        2025-11-18 10:00:00     2025-11-20 10:00:00     0       2
	...

	$ plantimeline -f path/to/project.json -out-format timeline-json path/to/plan.json
	AtomicProcess   NumOfComplete   AllocatedResources      Description     StartTime       EndTime Start   End
	P1      0       R1      プロセス        2025-11-18 10:00:00     2025-11-20 10:00:00     0       2
	...

    # クリティカルパス上のバーを強調する（強調 ID 表は criticalpath の出力を qhs で絞って作る）
    $ criticalpath -poor -f path/to/project.json >cp.tsv
    $ qhs -H -O -t -T 'SELECT ID FROM cp.tsv WHERE "最大弾性値（全余裕）" < 0.0001' >em.tsv
    $ plantimeline -f path/to/project.json -out-format mermaid -em-tsv em.tsv path/to/plan.json
	gantt
	    dateFormat YYYY-MM-DD HH:mm
	    section P1 プロセス1
	    P1[0] R1 :crit, 2025-11-18 10:00, 2025-11-20 10:00
	    section P2 プロセス2
	    P2[0] R2 :2025-11-18 10:00, 2025-11-20 10:00
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

	if commonOptions.ShortHelp {
		return &Options{CommonOptions: commonOptions}, nil
	}
	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	if flags.NArg() > 1 {
		return nil, fmt.Errorf("cmd.ParseOptions: too many arguments: expected 1 plan, got %d", flags.NArg())
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	projectConfig, basePath, err := tools.ReadProjectConfig(&configShortPath, &configLongPath, fsmRawOptions, planOutputFormatRawOptions, cwd, flags)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	planPath := projectConfig.PlanPath
	if flags.Arg(0) != "" {
		planPath = tools.ResolvePath(cwd, flags.Arg(0))
	}
	if planPath == "" {
		return nil, fmt.Errorf("cmd.ParseOptions: plan path is required")
	}

	planReporter, outputFormat, err := tools.ValidatePlanOutputFormat(&projectConfig.PlanOutputFormatRawOptions, projectConfig, slog.New(slograw.NewHandler(inout.Stderr, commonOptions.LogLevel)))
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	fsmOptions, err := tools.ValidateAllFSMOptions(&projectConfig.FSMRawOptions, basePath)
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

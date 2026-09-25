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
	CommonOptions  *tools.CommonOptions
	FSMOptions     *tools.FSMOptions
	PlanReporter   fsmreporter.PlanReporter
	SearchFunc     fsm.SearchFunc
	OutDir         string
	OutputFormat   tools.PlanOutputFormat
	CPUProfilePath string
	MemProfilePath string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdplan", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdplan [options] (-f <project> | -p <pfd> -ap <atomic-process-table> -ad <atomic-deliverable-table> [-r <resource-table>]) [-out-format google-spreadsheet-tsv|plan-json|timeline-json|mermaid|plantuml]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
    $ pfdplan -p path/to/pfd.drawio -ap path/to/atomic_proc.tsv -ad path/to/deliv.tsv -r path/to/resource.tsv -start-time 10:00 -duration 9 -not-biz-days <(holidays -locale ja)
    AtomicProcess[NumOfComplete]     StartTime       EndTime
    P1[1]   2025-10-04T00:00:00+09:00       2025-10-11T04:30:00+09:00
    P1[2]   2025-10-04T04:30:00+09:00       2025-10-11T06:45:00+09:00
    P1[3]   2025-10-04T06:45:00+09:00       2025-10-11T06:45:00+09:00
	...

    # クリティカルパス上のバーを強調する（強調 ID 表は criticalpath の出力を qhs で絞って作る）
    $ criticalpath -poor -f path/to/project.json >cp.tsv
    $ qhs -H -O -t -T 'SELECT ID FROM cp.tsv WHERE "最大弾性値（全余裕）" < 0.0001' >em.tsv
    $ pfdplan -poor -f path/to/project.json -out-format mermaid -em-tsv em.tsv
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

	var fsmRawOptions tools.FSMRawOptions
	var configShortPath, configLongPath string
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

	var planOutputFormatRawOptions tools.PlanOutputFormatRawOptions
	tools.DeclarePlanOutputFormatOptions(flags, &planOutputFormatRawOptions)

	outDirFlag := flags.String(tools.OutDirFlag, "", "output directory")

	cpuProfileFlag := flags.String("cpuprofile", "", "write a CPU profile to the file. the profile is flushed even when interrupted by SIGINT/SIGTERM")
	memProfileFlag := flags.String("memprofile", "", "write a heap profile to the file. the profile is flushed even when interrupted by SIGINT/SIGTERM")

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

	projectConfig, basePath, err := tools.ReadProjectConfig(&configShortPath, &configLongPath, fsmRawOptions, planOutputFormatRawOptions, cwd, flags)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	fsmOptions, err := tools.ValidateAllFSMOptions(&projectConfig.FSMRawOptions, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	planReporter, outputFormat, err := tools.ValidatePlanOutputFormat(&projectConfig.PlanOutputFormatRawOptions, projectConfig, commonOptions.Logger)
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
		CommonOptions:  commonOptions,
		FSMOptions:     fsmOptions,
		PlanReporter:   planReporter,
		SearchFunc:     searchFunc,
		OutDir:         outDir,
		OutputFormat:   outputFormat,
		CPUProfilePath: *cpuProfileFlag,
		MemProfilePath: *memProfileFlag,
	}, nil
}

package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/allcheckers"
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Mode string

const (
	ModePFD Mode = "pfd"
	ModeFSM Mode = "fsm"
)

type Options struct {
	CommonOptions *tools.CommonOptions

	PFDReader io.Reader

	HasAtomicProcessTable    bool
	AtomicProcessTableReader io.Reader

	HasAtomicDeliverableTable    bool
	AtomicDeliverableTableReader io.Reader

	HasCompositeProcessTable    bool
	CompositeProcessTableReader io.Reader

	HasCompositeDeliverableTable    bool
	CompositeDeliverableTableReader io.Reader

	HasResourceTable    bool
	ResourceTableReader io.Reader

	HasMilestoneTable    bool
	MilestoneTableReader io.Reader

	HasGroupTable    bool
	GroupTableReader io.Reader

	Reporter allcheckers.Func
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("lint", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: pfdlint [options] [-f <config>] [-p <pfd>] [-cd <composite-deliverable-table>] [-a <atomic-process-table>] [-ad <atomic-deliverable-table>] [-cp <composite-process-table>] [-r <resource-table>] [-m <milestone-table>] [-g <group-table>]")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
  $ pfdlint -f ./path/to/project.json
  WARNING no-desc Please add a concise description.       [D2]
  ERROR   single-src      A deliverable should be output from only one process. This includes output through feedback edges.      [D3]

  $ pfdlint -p ./path/to/pfd.drawio -ap ./path/to/ap.tsv -ad ./path/to/ad.tsv -cd ./path/to/cd.tsv -cp ./path/to/cp.tsv -r ./path/to/r.tsv -m ./path/to/m.tsv -g ./path/to/g.tsv
  WARNING no-desc Please add a concise description.       [D2]
  ERROR   single-src      A deliverable should be output from only one process. This includes output through feedback edges.      [D3]

  $ pfdlint -locale en -f ./path/to/project.json
  WARNING no-desc Please add a concise description.       [D3]
  ERROR   single-src      A deliverable should be output from only one process. This includes output through feedback edges.      [D2]
`)

	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	formatFlag := flags.String("format", "tsv", "format of the fsmreporter")

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

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	mergedOptions, basePath, err := tools.ReadFSMRawOptions(&configShortPath, &configLongPath, fsmRawOptions, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	fsmOptions, err := tools.ValidatePossibleFSMOptions(&mergedOptions, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var rep allcheckers.Func
	switch *formatFlag {
	case "tsv":
		rep = allcheckers.NewTSV(inout.Stdout, commonOptions.Locale)
	case "json":
		rep = allcheckers.NewJSON(inout.Stdout, commonOptions.Locale)
	default:
		return nil, fmt.Errorf("cmd.ParseOptions: unknown format: %q", *formatFlag)
	}

	return &Options{
		PFDReader:                       fsmOptions.PFDReader,
		HasAtomicProcessTable:           fsmOptions.AtomicProcessTableReader != nil,
		AtomicProcessTableReader:        fsmOptions.AtomicProcessTableReader,
		HasAtomicDeliverableTable:       fsmOptions.AtomicDeliverableTableReader != nil,
		AtomicDeliverableTableReader:    fsmOptions.AtomicDeliverableTableReader,
		HasCompositeProcessTable:        fsmOptions.CompositeProcessTableReader != nil,
		CompositeProcessTableReader:     fsmOptions.CompositeProcessTableReader,
		HasCompositeDeliverableTable:    fsmOptions.CompositeDeliverableTableReader != nil,
		CompositeDeliverableTableReader: fsmOptions.CompositeDeliverableTableReader,
		HasResourceTable:                fsmOptions.ResourceTableReader != nil,
		ResourceTableReader:             fsmOptions.ResourceTableReader,
		HasMilestoneTable:               fsmOptions.MilestoneTableReader != nil,
		MilestoneTableReader:            fsmOptions.MilestoneTableReader,
		HasGroupTable:                   fsmOptions.GroupTableReader != nil,
		GroupTableReader:                fsmOptions.GroupTableReader,
		CommonOptions:                   commonOptions,
		Reporter:                        rep,
	}, nil
}

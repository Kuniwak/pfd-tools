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
  $ pfdlint -locale en -f ./path/to/project.json
  WARNING no-desc Please add a concise description.       [D3]
  ERROR   single-src      A deliverable should be output from only one process. This includes output through feedback edges.      [D2]

  $ pfdlint -p ./path/to/pfd.drawio -ap ./path/to/ap.tsv -ad ./path/to/ad.tsv -cd ./path/to/cd.tsv -cp ./path/to/cp.tsv -r ./path/to/r.tsv -m ./path/to/m.tsv -g ./path/to/g.tsv
  WARNING no-desc Please add a concise description.       [D3]
  ERROR   single-src      A deliverable should be output from only one process. This includes output through feedback edges.      [D2]

  $ pfdlint -locale ja -f ./path/to/project.json
  WARNING no-desc 端的な説明を追加してください。  [D2]
  ERROR   single-src      成果物が複数のプロセスから出力されています。成果物はただ1つのプロセスから出力されるべきです。   [D3]
`)

	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	formatFlag := flags.String("format", "tsv", "format of the fsmreporter")

	var pfdShortPath, pfdLongPath string
	tools.DeclarePFDOptions(flags, &pfdShortPath, &pfdLongPath)

	var atomicProcessTableShortPath, atomicProcessTableLongPath string
	tools.DeclareAtomicProcessTableOptions(flags, &atomicProcessTableShortPath, &atomicProcessTableLongPath)

	var atomicDeliverableTableShortPath, atomicDeliverableTableLongPath string
	tools.DeclareAtomicDeliverableTableOptions(flags, &atomicDeliverableTableShortPath, &atomicDeliverableTableLongPath)

	var compositeProcessTableShortPath, compositeProcessTableLongPath string
	tools.DeclareCompositeProcessTableOptions(flags, &compositeProcessTableShortPath, &compositeProcessTableLongPath)

	var compositeDeliverableTableShortPath, compositeDeliverableTableLongPath string
	tools.DeclareCompositeDeliverableTableOptions(flags, &compositeDeliverableTableShortPath, &compositeDeliverableTableLongPath)

	var resourceTableShortPath, resourceTableLongPath string
	tools.DeclareResourceTableOptions(flags, &resourceTableShortPath, &resourceTableLongPath)

	var milestoneTableShortPath, milestoneTableLongPath string
	tools.DeclareMilestoneTableOptions(flags, &milestoneTableShortPath, &milestoneTableLongPath)

	var groupTableShortPath, groupTableLongPath string
	tools.DeclareGroupTableOptions(flags, &groupTableShortPath, &groupTableLongPath)

	var configShortPath, configLongPath string
	tools.DeclareConfigOptions(flags, &configShortPath, &configLongPath)

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

	var pfdReader io.Reader
	var hasAtomicProcessTable bool
	var atomicProcessTableReader io.Reader
	var hasAtomicDeliverableTable bool
	var atomicDeliverableTableReader io.Reader
	var hasCompositeProcessTable bool
	var compositeProcessTableReader io.Reader
	var hasCompositeDeliverableTable bool
	var compositeDeliverableTableReader io.Reader
	var hasResourceTable bool
	var resourceTableReader io.Reader
	var hasMilestoneTable bool
	var milestoneTableReader io.Reader
	var hasGroupTable bool
	var groupTableReader io.Reader
	if configShortPath != "" || configLongPath != "" {
		fsmOptions, err := tools.ValidateFSMOptionsJSON(&configShortPath, &configLongPath, tools.FSMRawOptions{})
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		pfdReader = fsmOptions.PFDReader

		hasAtomicProcessTable = true
		atomicProcessTableReader = fsmOptions.AtomicProcessTableReader

		hasAtomicDeliverableTable = true
		atomicDeliverableTableReader = fsmOptions.AtomicDeliverableTableReader

		hasCompositeDeliverableTable = true
		compositeDeliverableTableReader = fsmOptions.CompositeDeliverableTableReader

		hasResourceTable = true
		resourceTableReader = fsmOptions.ResourceTableReader
	} else {
		pfdReader, _, err = tools.ValidatePFDOptions(&pfdShortPath, &pfdLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}

		hasAtomicProcessTable = atomicProcessTableShortPath != "" || atomicProcessTableLongPath != ""
		if hasAtomicProcessTable {
			atomicProcessTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		}

		hasAtomicDeliverableTable = atomicDeliverableTableShortPath != "" || atomicDeliverableTableLongPath != ""
		if hasAtomicDeliverableTable {
			atomicDeliverableTableReader, _, err = tools.ValidateAtomicDeliverableTableOptions(&atomicDeliverableTableShortPath, &atomicDeliverableTableLongPath, cwd)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		}

		hasCompositeDeliverableTable = compositeDeliverableTableShortPath != "" || compositeDeliverableTableLongPath != ""
		if hasCompositeDeliverableTable {
			compositeDeliverableTableReader, _, err = tools.ValidateCompositeDeliverableTableOptions(&compositeDeliverableTableShortPath, &compositeDeliverableTableLongPath, cwd)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		}

		hasResourceTable = resourceTableShortPath != "" || resourceTableLongPath != ""
		if hasResourceTable {
			resourceTableReader, _, err = tools.ValidateResourceTableOptions(&resourceTableShortPath, &resourceTableLongPath, cwd)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		}
	}

	hasCompositeProcessTable = compositeProcessTableShortPath != "" || compositeProcessTableLongPath != ""
	if hasCompositeProcessTable {
		compositeProcessTableReader, _, err = tools.ValidateCompositeProcessTableOptions(&compositeProcessTableShortPath, &compositeProcessTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	hasMilestoneTable = milestoneTableShortPath != "" || milestoneTableLongPath != ""
	if hasMilestoneTable {
		milestoneTableReader, _, err = tools.ValidateMilestoneTableOptions(&milestoneTableShortPath, &milestoneTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	hasGroupTable = groupTableShortPath != "" || groupTableLongPath != ""
	if hasGroupTable {
		groupTableReader, _, err = tools.ValidateGroupTableOptions(&groupTableShortPath, &groupTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
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
		PFDReader:                       pfdReader,
		HasAtomicProcessTable:           hasAtomicProcessTable,
		AtomicProcessTableReader:        atomicProcessTableReader,
		HasAtomicDeliverableTable:       hasAtomicDeliverableTable,
		AtomicDeliverableTableReader:    atomicDeliverableTableReader,
		HasCompositeProcessTable:        hasCompositeProcessTable,
		CompositeProcessTableReader:     compositeProcessTableReader,
		HasCompositeDeliverableTable:    hasCompositeDeliverableTable,
		CompositeDeliverableTableReader: compositeDeliverableTableReader,
		HasResourceTable:                hasResourceTable,
		ResourceTableReader:             resourceTableReader,
		HasMilestoneTable:               hasMilestoneTable,
		MilestoneTableReader:            milestoneTableReader,
		HasGroupTable:                   hasGroupTable,
		GroupTableReader:                groupTableReader,
		CommonOptions:                   commonOptions,
		Reporter:                        rep,
	}, nil
}

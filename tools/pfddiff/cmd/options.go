package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type OutputFormat string

const (
	OutputFormatDiff OutputFormat = "diff"
	OutputFormatJSON OutputFormat = "json"
)

type Options struct {
	CommonOptions                    *tools.CommonOptions
	ShowSame                         bool
	PFDReader1                       io.Reader
	PFDReader2                       io.Reader
	CompositeDeliverableTableReader1 io.Reader
	CompositeDeliverableTableReader2 io.Reader
	OutputFormat                     OutputFormat
	Prompt                           bool
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfddiff", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: pfddiff [options] -p1 <pfd-a> -cd1 <composite-deliverable-table-a> -p2 <pfd-b> -cd2 <composite-deliverable-table-b>")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
  $ pfddiff -p1 path/to/a.drawio -cd1 path/to/composite-deliverable-table-a.tsv -p2 path/to/b.drawio -cd2 path/to/composite-deliverable-table-b.tsv
  + P1 ----> D1
  - P2 ----> D2
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	outputFormatString := flags.String("format", "diff", "format of the output")
	showSameFlag := flags.Bool("show-same", false, "show same nodes and edges")

	var pfdPathShortFlag1, pfdPathLongFlag1 string
	var compositeDeliverableTablePathShortFlag1, compositeDeliverableTablePathLongFlag1 string
	tools.DeclarePFDOptionsWithFlagNames("p1", "pfd1", flags, &pfdPathShortFlag1, &pfdPathLongFlag1)
	tools.DeclareCompositeDeliverableTableOptionsWithFlagNames("cd1", "composite-deliverable1", flags, &compositeDeliverableTablePathShortFlag1, &compositeDeliverableTablePathLongFlag1)

	var pfdPathShortFlag2, pfdPathLongFlag2 string
	var compositeDeliverableTablePathShortFlag2, compositeDeliverableTablePathLongFlag2 string
	tools.DeclarePFDOptionsWithFlagNames("p2", "pfd2", flags, &pfdPathShortFlag2, &pfdPathLongFlag2)
	tools.DeclareCompositeDeliverableTableOptionsWithFlagNames("cd2", "composite-deliverable2", flags, &compositeDeliverableTablePathShortFlag2, &compositeDeliverableTablePathLongFlag2)

	var promptFlag bool
	flags.BoolVar(&promptFlag, "prompt", false, "output diff as a prompt for Agentic AIs")

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

	pfdReader1, _, err := tools.ValidatePFDOptions(&pfdPathShortFlag1, &pfdPathLongFlag1, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}
	compositeDeliverableTableReader1, _, err := tools.ValidateCompositeDeliverableTableOptions(&compositeDeliverableTablePathShortFlag1, &compositeDeliverableTablePathLongFlag1, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	pfdReader2, _, err := tools.ValidatePFDOptions(&pfdPathShortFlag2, &pfdPathLongFlag2, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}
	compositeDeliverableTableReader2, _, err := tools.ValidateCompositeDeliverableTableOptions(&compositeDeliverableTablePathShortFlag2, &compositeDeliverableTablePathLongFlag2, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var outputFormat OutputFormat
	switch *outputFormatString {
	case "diff", "":
		outputFormat = OutputFormatDiff
	case "json":
		outputFormat = OutputFormatJSON
	}

	return &Options{
		CommonOptions:                    commonOptions,
		PFDReader1:                       pfdReader1,
		CompositeDeliverableTableReader1: compositeDeliverableTableReader1,
		PFDReader2:                       pfdReader2,
		CompositeDeliverableTableReader2: compositeDeliverableTableReader2,
		OutputFormat:                     outputFormat,
		ShowSame:                         *showSameFlag,
		Prompt:                           promptFlag,
	}, nil
}

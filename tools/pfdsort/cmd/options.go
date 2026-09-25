package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions  *tools.CommonOptions
	Reader         io.Reader
	Inplace        bool
	InputFilePath  string
	DupRankSpan    int
	DupRowSpan     int
	HGap           float64
	VGap           float64
	OnlyPages      []string
	OnlyNodes      []pfd.NodeID
	PosRestriction pfddrawio.PosRestriction
	BreakCycles    bool
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdsort", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdsort [options] [path/to/pfd.drawio]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
  $ pfdsort path/to/pfd.drawio
  $ pfdsort -inplace path/to/pfd.drawio
  $ pfdsort -dup-rank-span 3 path/to/pfd.drawio
  $ pfdsort -dup-row-span 8 path/to/pfd.drawio
  $ pfdsort -only-page P123,P456 path/to/pfd.drawio
  $ pfdsort -only-node P1,D2 path/to/pfd.drawio
  $ pfdsort -only-node P1 -pos-restriction free-v path/to/pfd.drawio
  $ pfdsort -break-cycles path/to/pfd.drawio
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	defaults := pfddrawio.DefaultSortOptions()
	inplaceFlag := flags.Bool("inplace", false, "overwrite the file in place")
	dupRankSpanFlag := flags.Int("dup-rank-span", defaults.DupRankSpan, "duplicate a deliverable when an incident edge spans more than this many ranks (larger suppresses duplication)")
	dupRowSpanFlag := flags.Int("dup-row-span", defaults.DupRowSpan, "duplicate a deliverable when an incident edge spans more than this many rows (larger suppresses duplication)")
	hgapFlag := flags.Float64("hgap", defaults.HGap, "horizontal gap between rank columns")
	vgapFlag := flags.Float64("vgap", defaults.VGap, "vertical gap between nodes within a rank")
	onlyPageFlag := flags.String("only-page", "", `sort only the given pages (comma-separated page names, e.g. "P123,P456"; empty = all pages)`)
	onlyNodeFlag := flags.String("only-node", "", `sort only the given nodes locally (comma-separated PFD node IDs, e.g. "D1,P123"; empty = full sort). Non-selected nodes are constrained by -pos-restriction.`)
	posRestrictionFlag := flags.String("pos-restriction", "lock", `how non-selected nodes may move under -only-node: lock (fixed) | free-v (vertical only) | free-h (horizontal only) | free (minimal move)`)
	breakCyclesFlag := flags.Bool("break-cycles", defaults.BreakCycles, "lay out pages that still have cycles (e.g. mutually dependent composite processes) by cutting them for layout only and showing the cut edges as duplicated deliverables")

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

	if *hgapFlag < 0 {
		return nil, fmt.Errorf("cmd.ParseOptions: -hgap must not be negative: %v", *hgapFlag)
	}
	if *vgapFlag < 0 {
		return nil, fmt.Errorf("cmd.ParseOptions: -vgap must not be negative: %v", *vgapFlag)
	}

	posRestriction, err := pfddrawio.ParsePosRestriction(*posRestrictionFlag)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if commonOptions.ShortHelp {
		return &Options{CommonOptions: commonOptions}, nil
	}
	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	shouldReadFromStdin := false
	var inputFilePath string

	if flags.NArg() < 1 {
		shouldReadFromStdin = true
	} else if flags.NArg() > 1 {
		return nil, fmt.Errorf("cmd.ParseOptions: too many arguments")
	} else {
		inputFilePath = flags.Arg(0)
		if inputFilePath == "" {
			shouldReadFromStdin = true
		}
	}

	var r io.Reader
	if shouldReadFromStdin {
		if *inplaceFlag {
			return nil, fmt.Errorf("cmd.ParseOptions: -inplace requires a file path")
		}
		r = inout.Stdin
	} else {
		content, err := os.ReadFile(inputFilePath)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		r = bytes.NewReader(content)
	}

	var onlyNodes []pfd.NodeID
	for _, s := range tools.ParsePageList(*onlyNodeFlag) {
		onlyNodes = append(onlyNodes, pfd.NodeID(s))
	}

	return &Options{
		CommonOptions:  commonOptions,
		Reader:         r,
		Inplace:        *inplaceFlag,
		InputFilePath:  inputFilePath,
		DupRankSpan:    *dupRankSpanFlag,
		DupRowSpan:     *dupRowSpanFlag,
		HGap:           *hgapFlag,
		VGap:           *vgapFlag,
		OnlyPages:      tools.ParsePageList(*onlyPageFlag),
		OnlyNodes:      onlyNodes,
		PosRestriction: posRestriction,
		BreakCycles:    *breakCyclesFlag,
	}, nil
}

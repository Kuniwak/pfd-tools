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

type Options struct {
	CommonOptions            *tools.CommonOptions
	BusinessTimeFuncOptions  *tools.BusinessTimeFuncOptions
	BufferMultiplier         float64
	PlanReader               io.Reader
	AtomicProcessTableReader io.Reader
	MilestoneTableReader     io.Reader
	GroupTableReader         io.Reader
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("planmaster", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: planmaster [options] -p <plan> -ap <atomic-process-table> -m <milestone-table> -g <group-table> [-b <buffer-multiplier>]")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
    $ planmaster -p path/to/plan.json -ap path/to/atomic_proc.tsv -m path/to/milestone.tsv -g path/to/group.tsv
	Group   Description     Start   End
	G1      M1      2025-11-18 10:00:00     2025-11-21 14:30:00
	...

    $ planmaster -p path/to/plan.json -ap path/to/atomic_proc.tsv -m path/to/milestone.tsv -g path/to/group.tsv -b 1.5
	Group   Description     Start   End
	G1      M1      2025-11-18 10:00:00     2025-11-21 14:30:00
	...
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var atomicProcessTableShortPath, atomicProcessTableLongPath string
	tools.DeclareAtomicProcessTableOptions(flags, &atomicProcessTableShortPath, &atomicProcessTableLongPath)

	var businessTimeFuncRawOptions tools.BusinessTimeFuncRawOptions
	tools.DeclareBusinessTimeFuncOptions(flags, &businessTimeFuncRawOptions)

	var planShortPath, planLongPath string
	flags.StringVar(&planShortPath, "p", "", "path to the plan")
	flags.StringVar(&planLongPath, "plan", "", "path to the plan")

	var bufferMultiplierLong, bufferMultiplierShort float64
	flags.Float64Var(&bufferMultiplierShort, "b", -1.0, "plan multiplier (negative value means not specified. 1.0 if both -b and -buffer-multiplier are not specified)")
	flags.Float64Var(&bufferMultiplierLong, "buffer-multiplier", -1.0, "plan multiplier (negative value means not specified. 1.0 if both -b and -buffer-multiplier are not specified)")

	var milestoneTableShortPath, milestoneTableLongPath string
	flags.StringVar(&milestoneTableShortPath, "m", "", "path to the milestone table")
	flags.StringVar(&milestoneTableLongPath, "milestone", "", "path to the milestone table")

	var groupTableShortPath, groupTableLongPath string
	flags.StringVar(&groupTableShortPath, "g", "", "path to the group table")
	flags.StringVar(&groupTableLongPath, "group", "", "path to the group table")

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

	businessTimeFuncOptions, err := tools.ValidateBusinessTimeFuncOptions(&businessTimeFuncRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var milestoneTablePath string
	if milestoneTableLongPath != "" {
		milestoneTablePath = milestoneTableLongPath
	} else if milestoneTableShortPath != "" {
		milestoneTablePath = milestoneTableShortPath
	} else {
		return nil, fmt.Errorf("cmd.ParseOptions: milestone table path is required")
	}
	milestoneTableReader, err := os.OpenFile(milestoneTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var groupTablePath string
	if groupTableLongPath != "" {
		groupTablePath = groupTableLongPath
	} else if groupTableShortPath != "" {
		groupTablePath = groupTableShortPath
	} else {
		return nil, fmt.Errorf("cmd.ParseOptions: group table path is required")
	}
	groupTableReader, err := os.OpenFile(groupTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var planPath string
	if planLongPath != "" {
		planPath = planLongPath
	} else if planShortPath != "" {
		planPath = planShortPath
	} else {
		return nil, fmt.Errorf("cmd.ParseOptions: plan path is required")
	}

	var bufferMultiplier float64
	if bufferMultiplierLong >= 0 {
		bufferMultiplier = bufferMultiplierLong
	} else if bufferMultiplierShort >= 0 {
		bufferMultiplier = bufferMultiplierShort
	} else {
		bufferMultiplier = 1.0
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	atomicProcessTableReader, _, err := tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	planReader, err := os.Open(planPath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	return &Options{
		CommonOptions:            commonOptions,
		BusinessTimeFuncOptions:  businessTimeFuncOptions,
		BufferMultiplier:         bufferMultiplier,
		PlanReader:               planReader,
		AtomicProcessTableReader: atomicProcessTableReader,
		MilestoneTableReader:     milestoneTableReader,
		GroupTableReader:         groupTableReader,
	}, nil
}

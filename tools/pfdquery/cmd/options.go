package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding/fsmtsv"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	Queries                              []string                       `json:"queries"`
	BackwardReachable                    bool                           `json:"backward_reachable"`
	BackwardReachableFeedbackDestination bool                           `json:"backward_reachable_feedback_destination"`
	Reachable                            bool                           `json:"reachable"`
	CommonOptions                        *tools.CommonOptions           `json:"common_options"`
	PFD                                  *pfd.PFD                       `json:"pfd"`
	FSMOptions                           *tools.FSMOptions              `json:"fsm_options"`
	AtomicProcessTable                   *pfd.AtomicProcessTable        `json:"atomic_process_table"`
	AtomicDeliverableTable               *pfd.AtomicDeliverableTable    `json:"atomic_deliverable_table"`
	CompositeProcessTable                *pfd.CompositeProcessTable     `json:"composite_process_table"`
	CompositeDeliverableTable            *pfd.CompositeDeliverableTable `json:"composite_deliverable_table"`
	MilestoneTable                       *fsmtable.MilestoneTable       `json:"milestone_table"`
	GroupTable                           *fsmtable.GroupTable           `json:"group_table"`
	ResourceTable                        *fsmtable.ResourceTable        `json:"resource_table"`
}

func (o *Options) String() string {
	sb := &strings.Builder{}
	if err := o.Write(sb); err != nil {
		return ""
	}
	return sb.String()
}

func (o *Options) Write(w io.Writer) error {
	io.WriteString(w, "queries: [")
	for i, query := range o.Queries {
		if i > 0 {
			io.WriteString(w, ", ")
		}
		io.WriteString(w, query)
	}
	io.WriteString(w, "], has_pfd: ")
	if o.PFD != nil {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}
	io.WriteString(w, ", has_atomic_process_table: ")
	if o.AtomicProcessTable != nil {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}
	io.WriteString(w, ", has_atomic_deliverable_table: ")
	if o.AtomicDeliverableTable != nil {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}
	io.WriteString(w, ", has_composite_process_table: ")
	if o.CompositeProcessTable != nil {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}
	io.WriteString(w, ", has_composite_deliverable_table: ")
	if o.CompositeDeliverableTable != nil {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}
	io.WriteString(w, ", has_resource_table: ")
	if o.ResourceTable != nil {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}

	return nil
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdquery", flag.ContinueOnError)

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var configShortPath, configLongPath string
	var fsmRawOptions tools.FSMRawOptions
	tools.DeclareFSMOptions(flags, &fsmRawOptions, &configShortPath, &configLongPath)

	var backwardReachable bool
	flags.BoolVar(&backwardReachable, "backward-reachable", false, "backward reachable")

	var backwardReachableFeedbackDestination bool
	flags.BoolVar(&backwardReachableFeedbackDestination, "backward-reachable-fb", false, "backward reachable feedback destination")

	var reachable bool
	flags.BoolVar(&reachable, "reachable", false, "reachable")

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

	var atomicProcessTable *pfd.AtomicProcessTable
	if fsmOptions.AtomicProcessTableReader != nil {
		atomicProcessTable, err = pfdtsv.ParseAtomicProcessTable(fsmOptions.AtomicProcessTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var atomicDeliverableTable *pfd.AtomicDeliverableTable
	if fsmOptions.AtomicDeliverableTableReader != nil {
		atomicDeliverableTable, err = pfdtsv.ParseAtomicDeliverableTable(fsmOptions.AtomicDeliverableTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var compositeProcessTable *pfd.CompositeProcessTable
	if fsmOptions.CompositeProcessTableReader != nil {
		compositeProcessTable, err = pfdtsv.ParseCompositeProcessTable(fsmOptions.CompositeProcessTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var compositeDeliverableTable *pfd.CompositeDeliverableTable
	if fsmOptions.CompositeDeliverableTableReader != nil {
		compositeDeliverableTable, err = pfdtsv.ParseCompositeDeliverableTable(fsmOptions.CompositeDeliverableTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var p *pfd.PFD
	if fsmOptions.PFDReader != nil && compositeDeliverableTable != nil {
		p, err = pfdfmt.Parse("", fsmOptions.PFDReader, &pfdfmt.ParseOptions{CompositeDeliverableTable: compositeDeliverableTable}, commonOptions.Logger)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var resourceTable *fsmtable.ResourceTable
	if fsmOptions.ResourceTableReader != nil {
		resourceTable, err = fsmtsv.ParseResourceTable(fsmOptions.ResourceTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var milestoneTable *fsmtable.MilestoneTable
	if fsmOptions.MilestoneTableReader != nil {
		milestoneTable, err = fsmtsv.ParseMilestoneTable(fsmOptions.MilestoneTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var groupTable *fsmtable.GroupTable
	if fsmOptions.GroupTableReader != nil {
		groupTable, err = fsmtsv.ParseGroupTable(fsmOptions.GroupTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	if flags.NArg() == 0 {
		return nil, fmt.Errorf("cmd.ParseOptions: query is required")
	}

	queries := make([]string, 0, flags.NArg())
	for i := 0; i < flags.NArg(); i++ {
		queries = append(queries, flags.Arg(i))
	}

	return &Options{
		CommonOptions:                        commonOptions,
		Queries:                              queries,
		PFD:                                  p,
		AtomicProcessTable:                   atomicProcessTable,
		AtomicDeliverableTable:               atomicDeliverableTable,
		CompositeProcessTable:                compositeProcessTable,
		CompositeDeliverableTable:            compositeDeliverableTable,
		MilestoneTable:                       milestoneTable,
		GroupTable:                           groupTable,
		ResourceTable:                        resourceTable,
		BackwardReachable:                    backwardReachable,
		BackwardReachableFeedbackDestination: backwardReachableFeedbackDestination,
		Reachable:                            reachable,
	}, nil
}

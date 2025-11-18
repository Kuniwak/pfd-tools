package tools

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/allcheckers"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding/fsmtsv"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
)

type FSMEnvSeed struct {
	PFD                                  *pfd.PFD
	AtomicProcessTable                   *pfd.AtomicProcessTable
	AtomicDeliverableTable               *pfd.AtomicDeliverableTable
	CompositeDeliverableTable            *pfd.CompositeDeliverableTable
	ResourceTable                        *fsmtable.ResourceTable
	MilestoneTable                       *fsmtable.MilestoneTable
	GroupTable                           *fsmtable.GroupTable
	MaximalAvailableAllocationsThreshold int
}

func ParseFSMEnvSeed(fsOpts *FSMOptions, logger *slog.Logger) (*FSMEnvSeed, error) {
	var compositeDeliverableTable *pfd.CompositeDeliverableTable
	if fsOpts.CompositeDeliverableTableReader != nil {
		var err error
		compositeDeliverableTable, err = pfdtsv.ParseCompositeDeliverableTable(fsOpts.CompositeDeliverableTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	}
	parseOpts := &pfdfmt.ParseOptions{
		CompositeDeliverableTable: compositeDeliverableTable,
	}

	up, err := pfdfmt.Parse("", fsOpts.PFDReader, parseOpts, logger)
	if err != nil {
		return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	atomicProcessTable, err := pfdtsv.ParseAtomicProcessTable(fsOpts.AtomicProcessTableReader)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseFSMTable: %w", err)
	}
	atomicDeliverableTable, err := pfdtsv.ParseAtomicDeliverableTable(fsOpts.AtomicDeliverableTableReader)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseFSMTable: %w", err)
	}
	resourceTable, err := fsmtsv.ParseResourceTable(fsOpts.ResourceTableReader)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseFSMTable: %w", err)
	}
	var milestoneTable *fsmtable.MilestoneTable
	if fsOpts.MilestoneTableReader != nil {
		milestoneTable, err = fsmtsv.ParseMilestoneTable(fsOpts.MilestoneTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseFSMTable: %w", err)
		}
	}
	var groupTable *fsmtable.GroupTable
	if fsOpts.GroupTableReader != nil {
		groupTable, err = fsmtsv.ParseGroupTable(fsOpts.GroupTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseFSMTable: %w", err)
		}
	}
	return &FSMEnvSeed{
		PFD:                                  up,
		AtomicProcessTable:                   atomicProcessTable,
		AtomicDeliverableTable:               atomicDeliverableTable,
		CompositeDeliverableTable:            compositeDeliverableTable,
		ResourceTable:                        resourceTable,
		MilestoneTable:                       milestoneTable,
		GroupTable:                           groupTable,
		MaximalAvailableAllocationsThreshold: fsOpts.MaximalAvailableAllocationsThreshold,
	}, nil
}

func ValidateFSMEnvSeed(fsmEnvSeed *FSMEnvSeed, logger *slog.Logger, locale locale.Locale) error {
	ps, err := allcheckers.Lint(
		fsmEnvSeed.PFD,
		fsmEnvSeed.AtomicProcessTable,
		fsmEnvSeed.AtomicDeliverableTable,
		nil,
		fsmEnvSeed.CompositeDeliverableTable,
		fsmEnvSeed.ResourceTable,
		fsmEnvSeed.MilestoneTable,
		fsmEnvSeed.GroupTable,
		logger,
	)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	hasError := false
	for _, p := range ps {
		if p.Severity == checkers.SeverityError {
			hasError = true
		}
	}
	if hasError {
		sb := &strings.Builder{}
		allcheckers.WriteTSV(sb, ps, locale)
		return fmt.Errorf("cmd.MainCommandByOptions: linting pfd: error\n%s", sb.String())
	}

	return nil
}

func FSMPrepare(fsmEnvSeed *FSMEnvSeed, locale locale.Locale, logger *slog.Logger) (*fsm.Env, error) {
	if err := ValidateFSMEnvSeed(fsmEnvSeed, logger, locale); err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: validate fsm env seed: %w", err)
	}

	p, err := pfd.NewSafePFDByUnsafePFD(fsmEnvSeed.PFD)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: new safe pfd: %w", err)
	}

	availableResources := fsmtable.AvailableResources(fsmEnvSeed.ResourceTable)

	initialVolumeFunc, err := fsmtable.InitialVolumeByTableFunc(fsmEnvSeed.AtomicProcessTable, fsmtable.DefaultInitialVolumeColumnMatchFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: initial volume func: %w", err)
	}

	reworkVolumeFunc, err := fsmtable.ReworkVolumeFuncByTableFunc(fsmEnvSeed.AtomicProcessTable, fsmtable.DefaultReworkVolumeRatioColumnMatchFunc, initialVolumeFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: rework volume func: %w", err)
	}

	maxRevisionMap, err := fsmtable.MaxRevisionMapByTableFunc(fsmEnvSeed.AtomicDeliverableTable, fsmtable.DefaultMaxRevisionColumnMatchFunc, p.FeedbackSourceDeliverables())
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: max revision map: %w", err)
	}

	neededResourceSetsFunc, err := fsmtable.NeededResourcesSetFuncByTable(fsmEnvSeed.AtomicProcessTable, fsmtable.DefaultNeededResourceSetsColumnSelectFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: needed resource sets func: %w", err)
	}

	atomicDeliverableAvailableTimeFunc, err := fsmtable.AvailableTimeFuncByTable(fsmEnvSeed.AtomicDeliverableTable, fsmtable.DefaultAvailableTimeColumnMatchFunc, p.InitialDeliverables())
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: atomic deliverable available time func: %w", err)
	}

	preconditionFunc, err := fsmtable.PreconditionFuncByTableFunc(fsmEnvSeed.AtomicProcessTable, fsmtable.DefaultPreconditionColumnMatchFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: precondition func: %w", err)
	}

	availableAllocationsFunc := fsm.NewThresholdAvailableAllocationsFunc(fsmEnvSeed.MaximalAvailableAllocationsThreshold, neededResourceSetsFunc, logger)

	env := fsm.NewEnv(
		p,
		availableResources,
		availableAllocationsFunc,
		initialVolumeFunc,
		reworkVolumeFunc,
		maxRevisionMap,
		preconditionFunc,
		neededResourceSetsFunc,
		atomicDeliverableAvailableTimeFunc,
		logger,
	)

	return env, nil
}

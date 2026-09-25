package tools

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Kuniwak/pfd-tools/allcheckers"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
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
	Model                                execmodel.Model
}

func ParseFSMEnvSeed(fsOpts *FSMOptions, logger *slog.Logger) (*FSMEnvSeed, error) {
	compositeDeliverableTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(fsOpts.CompositeDeliverableTableReader)
	if err != nil {
		return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
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
	var resourceTable *fsmtable.ResourceTable
	if fsOpts.ResourceTableReader != nil {
		resourceTable, err = fsmtsv.ParseResourceTable(fsOpts.ResourceTableReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseFSMTable: %w", err)
		}
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
		Model:                                fsOpts.Model,
	}, nil
}

func ValidateFSMEnvSeed(fsmEnvSeed *FSMEnvSeed, logger *slog.Logger, locale locale.Locale) error {
	ps, err := allcheckers.Lint(allcheckers.Target{
		PFD:                       fsmEnvSeed.PFD,
		AtomicProcessTable:        fsmEnvSeed.AtomicProcessTable,
		AtomicDeliverableTable:    fsmEnvSeed.AtomicDeliverableTable,
		CompositeDeliverableTable: fsmEnvSeed.CompositeDeliverableTable,
		ResourceTable:             fsmEnvSeed.ResourceTable,
		MilestoneTable:            fsmEnvSeed.MilestoneTable,
		GroupTable:                fsmEnvSeed.GroupTable,
		Model:                     fsmEnvSeed.Model,
	}, logger)
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

	initialVolumeFunc, err := fsmtable.InitialVolumeByTableFunc(fsmEnvSeed.AtomicProcessTable, fsmtable.DefaultInitialVolumeColumnMatchFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: initial volume func: %w", err)
	}

	resourceAspect, err := fsmtable.NewResourceAspect(fsmEnvSeed.Model.Resource, fsmEnvSeed.AtomicProcessTable, fsmEnvSeed.ResourceTable)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: resource aspect: %w", err)
	}

	feedbackAspect, err := fsmtable.NewFeedbackAspect(fsmEnvSeed.Model.Feedback, p, fsmEnvSeed.AtomicProcessTable, fsmEnvSeed.AtomicDeliverableTable, initialVolumeFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: feedback aspect: %w", err)
	}

	availableAllocationsFunc := fsm.NewThresholdAvailableAllocationsFunc(fsmEnvSeed.MaximalAvailableAllocationsThreshold, resourceAspect.NeededResourceSetsFunc, logger)

	atomicDeliverableAvailableTimeFunc, err := fsmtable.AvailableTimeFuncByTable(fsmEnvSeed.AtomicDeliverableTable, fsmtable.DefaultAvailableTimeColumnMatchFunc, p.InitialDeliverables())
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: atomic deliverable available time func: %w", err)
	}

	preconditionFunc, err := fsmtable.PreconditionFuncByTableFunc(fsmEnvSeed.AtomicProcessTable, fsmtable.DefaultPreconditionColumnMatchFunc)
	if err != nil {
		return nil, fmt.Errorf("tools.FSMPrepare: precondition func: %w", err)
	}

	env := fsm.NewEnv(
		p,
		resourceAspect.AvailableResources,
		availableAllocationsFunc,
		initialVolumeFunc,
		feedbackAspect.ReworkVolumeFunc,
		feedbackAspect.FeedbackSourceMaxRevision,
		preconditionFunc,
		resourceAspect.NeededResourceSetsFunc,
		atomicDeliverableAvailableTimeFunc,
		logger,
	)

	return env, nil
}

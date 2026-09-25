package fsmtable_test

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
)

func TestNewResourceAspect(t *testing.T) {
	apTable := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{fsmtable.NeededResourceSetsColumnHeaderEn},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "Process 1", ExtraCells: []string{"R1:2"}},
		},
	}
	apTableWithoutResources := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "Process 1", ExtraCells: []string{}},
		},
	}
	rTable := &fsmtable.ResourceTable{
		ExtraHeaders: []string{},
		Rows:         []*fsmtable.ResourceTableRow{{ID: "R1", Description: "Resource 1"}},
	}

	testCases := map[string]struct {
		Mode                   execmodel.ResourceMode
		AtomicProcessTable     *pfd.AtomicProcessTable
		ResourceTable          *fsmtable.ResourceTable
		HasError               bool
		ExpectedResources      []fsm.ResourceID
		ExpectedConsumedVolume fsm.Volume
	}{
		"finite reads the resource table and the needed resources column": {
			Mode:                   execmodel.ResourceModeFinite,
			AtomicProcessTable:     apTable,
			ResourceTable:          rTable,
			ExpectedResources:      []fsm.ResourceID{"R1"},
			ExpectedConsumedVolume: 2,
		},
		"finite without a resource table is an error": {
			Mode:               execmodel.ResourceModeFinite,
			AtomicProcessTable: apTable,
			ResourceTable:      nil,
			HasError:           true,
		},
		"infinite ignores both": {
			Mode:                   execmodel.ResourceModeInfinite,
			AtomicProcessTable:     apTable,
			ResourceTable:          rTable,
			ExpectedResources:      []fsm.ResourceID{},
			ExpectedConsumedVolume: 1,
		},
		"infinite works without a resource table nor the needed resources column": {
			Mode:                   execmodel.ResourceModeInfinite,
			AtomicProcessTable:     apTableWithoutResources,
			ResourceTable:          nil,
			ExpectedResources:      []fsm.ResourceID{},
			ExpectedConsumedVolume: 1,
		},
		"unknown mode is an error": {
			Mode:               execmodel.ResourceMode("none"),
			AtomicProcessTable: apTable,
			ResourceTable:      rTable,
			HasError:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := fsmtable.NewResourceAspect(tc.Mode, tc.AtomicProcessTable, tc.ResourceTable)
			if tc.HasError {
				if err == nil {
					t.Fatal("expected an error, but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectedResources := sets.New(fsm.ResourceID.Compare, tc.ExpectedResources...)
			if sets.Compare(fsm.ResourceID.Compare)(actual.AvailableResources, expectedResources) != 0 {
				t.Errorf("expected available resources %v, but got %v", expectedResources.Slice(), actual.AvailableResources.Slice())
			}
			elems := actual.NeededResourceSetsFunc("P1")
			if elems.Len() != 1 {
				t.Fatalf("expected exactly 1 allocation element, but got %d", elems.Len())
			}
			if got := elems.Slice()[0].ConsumedVolume; got != tc.ExpectedConsumedVolume {
				t.Errorf("expected consumed volume %v, but got %v", tc.ExpectedConsumedVolume, got)
			}
		})
	}
}

func TestNewFeedbackAspect(t *testing.T) {

	withoutFeedback := pfd.MustNewSafePFDByUnsafePFD(pfd.PresetSmallest)

	withFeedback := pfd.MustNewSafePFDByUnsafePFD(pfd.PresetSmallestLoop)

	apTable := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{fsmtable.ReworkVolumeRatioColumnHeaderEn},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "Process 1", ExtraCells: []string{"0.5"}},
		},
	}
	apTableWithoutRework := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "Process 1", ExtraCells: []string{}},
		},
	}
	adTable := &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{fsmtable.MaxRevisionHeaderEn},
		Rows: []*pfd.AtomicDeliverableRow{
			{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{""}},
			{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{"3"}},
		},
	}
	adTableWithoutMaxRevision := &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{},
		Rows: []*pfd.AtomicDeliverableRow{
			{ID: "D1", Description: "Deliverable 1", ExtraCells: []string{}},
			{ID: "D2", Description: "Deliverable 2", ExtraCells: []string{}},
		},
	}
	initialVolumeFunc := fsm.ConstInitialVolumeFunc(4)

	testCases := map[string]struct {
		Mode                   execmodel.FeedbackMode
		PFD                    *pfd.ValidPFD
		AtomicProcessTable     *pfd.AtomicProcessTable
		AtomicDeliverableTable *pfd.AtomicDeliverableTable
		HasError               bool
		ExpectedReworkVolume   fsm.Volume
		ExpectedMaxRevision    map[pfd.AtomicDeliverableID]int
	}{
		"enabled reads the rework ratio and the max revision": {
			Mode:                   execmodel.FeedbackModeEnabled,
			PFD:                    withFeedback,
			AtomicProcessTable:     apTable,
			AtomicDeliverableTable: adTable,
			ExpectedReworkVolume:   2,
			ExpectedMaxRevision:    map[pfd.AtomicDeliverableID]int{"D2": 3},
		},
		"disabled reads neither": {
			Mode:                   execmodel.FeedbackModeDisabled,
			PFD:                    withoutFeedback,
			AtomicProcessTable:     apTableWithoutRework,
			AtomicDeliverableTable: adTableWithoutMaxRevision,
			ExpectedReworkVolume:   4,
			ExpectedMaxRevision:    map[pfd.AtomicDeliverableID]int{},
		},
		"disabled rejects a pfd with feedback edges": {
			Mode:                   execmodel.FeedbackModeDisabled,
			PFD:                    withFeedback,
			AtomicProcessTable:     apTableWithoutRework,
			AtomicDeliverableTable: adTableWithoutMaxRevision,
			HasError:               true,
		},
		"enabled without the rework ratio column is an error": {
			Mode:                   execmodel.FeedbackModeEnabled,
			PFD:                    withFeedback,
			AtomicProcessTable:     apTableWithoutRework,
			AtomicDeliverableTable: adTable,
			HasError:               true,
		},
		"unknown mode is an error": {
			Mode:                   execmodel.FeedbackMode("on"),
			PFD:                    withoutFeedback,
			AtomicProcessTable:     apTable,
			AtomicDeliverableTable: adTable,
			HasError:               true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := fsmtable.NewFeedbackAspect(tc.Mode, tc.PFD, tc.AtomicProcessTable, tc.AtomicDeliverableTable, initialVolumeFunc)
			if tc.HasError {
				if err == nil {
					t.Fatal("expected an error, but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := actual.ReworkVolumeFunc("P1", 1); got != tc.ExpectedReworkVolume {
				t.Errorf("expected rework volume %v, but got %v", tc.ExpectedReworkVolume, got)
			}
			if len(actual.FeedbackSourceMaxRevision) != len(tc.ExpectedMaxRevision) {
				t.Errorf("expected max revision %v, but got %v", tc.ExpectedMaxRevision, actual.FeedbackSourceMaxRevision)
			}
			for d, expected := range tc.ExpectedMaxRevision {
				if actual.FeedbackSourceMaxRevision[d] != expected {
					t.Errorf("expected max revision of %q to be %d, but got %d", d, expected, actual.FeedbackSourceMaxRevision[d])
				}
			}
		})
	}
}

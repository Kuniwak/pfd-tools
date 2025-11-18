package fsmcommon

import (
	"fmt"
	"io"
	"strings"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
)

type LocationType string

const (
	LocationTypePFD                       LocationType = "PFD"
	LocationTypeAtomicProcessTable        LocationType = "ATOMIC_PROCESS_TABLE"
	LocationTypeAtomicDeliverableTable    LocationType = "ATOMIC_DELIVERABLE_TABLE"
	LocationTypeCompositeProcessTable     LocationType = "COMPOSITE_PROCESS_TABLE"
	LocationTypeCompositeDeliverableTable LocationType = "COMPOSITE_DELIVERABLE_TABLE"
	LocationTypeResourceTable             LocationType = "RESOURCE_TABLE"
	LocationTypeMilestoneTable            LocationType = "MILESTONE_TABLE"
	LocationTypeGroupTable                LocationType = "GROUP_TABLE"
)

type IDType string

const (
	IDTypeAtomicProcess        IDType = "ATOMIC_PROCESS"
	IDTypeAtomicDeliverable    IDType = "ATOMIC_DELIVERABLE"
	IDTypeCompositeProcess     IDType = "COMPOSITE_PROCESS"
	IDTypeCompositeDeliverable IDType = "COMPOSITE_DELIVERABLE"
	IDTypeResource             IDType = "RESOURCE"
	IDTypeMilestone            IDType = "MILESTONE"
	IDTypeGroup                IDType = "GROUP"
)

type ID struct {
	Type                   IDType
	AtomicProcessID        pfd.AtomicProcessID
	AtomicDeliverableID    pfd.AtomicDeliverableID
	CompositeProcessID     pfd.CompositeProcessID
	CompositeDeliverableID pfd.CompositeDeliverableID
	ResourceID             fsm.ResourceID
	MilestoneID            fsmmasterschedule.Milestone
	GroupID                fsmmasterschedule.Group
}

func (i ID) Write(w io.Writer) error {
	switch i.Type {
	case IDTypeAtomicProcess:
		if _, err := io.WriteString(w, string(i.AtomicProcessID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
		return nil
	case IDTypeAtomicDeliverable:
		if _, err := io.WriteString(w, string(i.AtomicDeliverableID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
		return nil
	case IDTypeCompositeProcess:
		if _, err := io.WriteString(w, string(i.CompositeProcessID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
		return nil
	case IDTypeCompositeDeliverable:
		if _, err := io.WriteString(w, string(i.CompositeDeliverableID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
	case IDTypeResource:
		if _, err := io.WriteString(w, string(i.ResourceID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
	case IDTypeMilestone:
		if _, err := io.WriteString(w, string(i.MilestoneID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
		return nil
	case IDTypeGroup:
		if _, err := io.WriteString(w, string(i.GroupID)); err != nil {
			return fmt.Errorf("fsmcommon.ID.Write: %w", err)
		}
	default:
		panic(fmt.Sprintf("fsmcommon.ID.Write: invalid ID type: %q", i.Type))
	}
	return nil
}

func NewAtomicProcessID(atomicProcessID pfd.AtomicProcessID) ID {
	return ID{Type: IDTypeAtomicProcess, AtomicProcessID: atomicProcessID}
}

func NewAtomicDeliverableID(atomicDeliverableID pfd.AtomicDeliverableID) ID {
	return ID{Type: IDTypeAtomicDeliverable, AtomicDeliverableID: atomicDeliverableID}
}

func NewCompositeProcessID(compositeProcessID pfd.CompositeProcessID) ID {
	return ID{Type: IDTypeCompositeProcess, CompositeProcessID: compositeProcessID}
}

func NewCompositeDeliverableID(compositeDeliverableID pfd.CompositeDeliverableID) ID {
	return ID{Type: IDTypeCompositeDeliverable, CompositeDeliverableID: compositeDeliverableID}
}

func NewResourceID(resourceID fsm.ResourceID) ID {
	return ID{Type: IDTypeResource, ResourceID: resourceID}
}

func NewMilestoneID(milestoneID fsmmasterschedule.Milestone) ID {
	return ID{Type: IDTypeMilestone, MilestoneID: milestoneID}
}

func NewGroupID(groupID fsmmasterschedule.Group) ID {
	return ID{Type: IDTypeGroup, GroupID: groupID}
}

func (a ID) Compare(b ID) int {
	if a.Type != b.Type {
		return strings.Compare(string(a.Type), string(b.Type))
	}
	switch a.Type {
	case IDTypeAtomicProcess:
		return a.AtomicProcessID.Compare(b.AtomicProcessID)
	case IDTypeAtomicDeliverable:
		return a.AtomicDeliverableID.Compare(b.AtomicDeliverableID)
	case IDTypeCompositeProcess:
		return a.CompositeProcessID.Compare(b.CompositeProcessID)
	case IDTypeCompositeDeliverable:
		return a.CompositeDeliverableID.Compare(b.CompositeDeliverableID)
	case IDTypeResource:
		return a.ResourceID.Compare(b.ResourceID)
	case IDTypeMilestone:
		return a.MilestoneID.Compare(b.MilestoneID)
	case IDTypeGroup:
		return a.GroupID.Compare(b.GroupID)
	default:
		panic(fmt.Sprintf("fsmcommon.ID.Compare: invalid ID type: %q", a.Type))
	}
}

func NewLocations(ls ...Location) []checkers.Location {
	ls2 := make([]checkers.Location, len(ls))
	for i, l := range ls {
		ls2[i] = l
	}
	return ls2
}

type Location struct {
	RelatedIDs []ID
	Type       LocationType
}

var _ checkers.Location = Location{}

func NewLocation(t LocationType, relatedIDs ...ID) Location {
	return Location{
		Type:       t,
		RelatedIDs: relatedIDs,
	}
}

var (
	openSqBracket  = []byte{'['}
	closeSqBracket = []byte{']'}
	sep            = []byte(", ")
)

func (l Location) Write(w io.Writer) error {
	if _, err := io.WriteString(w, string(l.Type)); err != nil {
		return fmt.Errorf("fsmcommon.Location.Write: %w", err)
	}
	if _, err := w.Write(openSqBracket); err != nil {
		return fmt.Errorf("fsmcommon.Location.Write: %w", err)
	}
	for i, id := range l.RelatedIDs {
		if i > 0 {
			if _, err := w.Write(sep); err != nil {
				return fmt.Errorf("fsmcommon.Location.Write: %w", err)
			}
		}
		if err := id.Write(w); err != nil {
			return fmt.Errorf("fsmcommon.Location.Write: %w", err)
		}
	}
	if _, err := w.Write(closeSqBracket); err != nil {
		return fmt.Errorf("fsmcommon.Location.Write: %w", err)
	}
	return nil
}

package pfdcommon

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
)

func NewLocations(ls ...Location) []checkers.Location {
	ls2 := make([]checkers.Location, len(ls))
	for i, l := range ls {
		ls2[i] = l
	}
	return ls2
}

type LocationType string

const (
	LocationTypePFD                   LocationType = "PFD"
	LocationTypeAtomicProcessTable    LocationType = "ATOMIC_PROCESS_TABLE"
	LocationTypeDeliverableTable      LocationType = "DELIVERABLE_TABLE"
	LocationTypeCompositeProcessTable LocationType = "COMPOSITE_PROCESS_TABLE"
)

type Location struct {
	RelatedIDs []pfd.NodeID
	Type       LocationType
}

func CompareLocation(l1, l2 Location) int {
	if l1.Type != l2.Type {
		return strings.Compare(string(l1.Type), string(l2.Type))
	}
	return slices.CompareFunc(l1.RelatedIDs, l2.RelatedIDs, pfd.NodeID.Compare)
}

var (
	openSqBracket  = []byte{'['}
	closeSqBracket = []byte{']'}
	sep            = []byte(", ")
)

var _ checkers.Location = Location{}

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
		if _, err := io.WriteString(w, string(id)); err != nil {
			return fmt.Errorf("fsmcommon.Location.Write: %w", err)
		}
	}
	if _, err := w.Write(closeSqBracket); err != nil {
		return fmt.Errorf("fsmcommon.Location.Write: %w", err)
	}
	return nil
}

func NewLocation(t LocationType, relatedIDs ...pfd.NodeID) Location {
	return Location{RelatedIDs: relatedIDs, Type: t}
}

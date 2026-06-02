// Package pfdtable provides the extra-column profile (Mode) for pfd tables and related utilities.
//
// Mode defines extra columns only for AtomicProcessTable / AtomicDeliverableTable.
// CompositeProcessTable / CompositeDeliverableTable have no Mode-derived extra columns,
// so specifying ModePlan / ModePlanMaster has no effect on them (equivalent to ModeMinimal).
package pfdtable

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
)

type Mode string

const (
	ModeMinimal    Mode = "minimal"
	ModePlan       Mode = "plan"
	ModePlanMaster Mode = "plan-master"
)

// ColumnAliases represents the set of alternative notations for a column header that share the same meaning.
// Canonical returns the canonical notation for a given locale, and HasMatch is used to test whether an arbitrary notation matches.
type ColumnAliases struct {
	Ja string
	En string
}

func (c ColumnAliases) Canonical(l locale.Locale) string {
	switch l {
	case locale.LocaleJa:
		return c.Ja
	case locale.LocaleEn:
		return c.En
	default:
		panic(fmt.Sprintf("pfdtable.ColumnAliases.Canonical: unsupported locale: %q", l))
	}
}

func (c ColumnAliases) HasMatch(name string) bool {
	return name == c.Ja || name == c.En
}

var (
	APInitialVolume      = ColumnAliases{Ja: fsmtable.InitialVolumeColumnHeaderJa, En: fsmtable.InitialVolumeColumnHeaderEn}
	APReworkVolumeRatio  = ColumnAliases{Ja: fsmtable.ReworkVolumeRatioColumnHeaderJa, En: fsmtable.ReworkVolumeRatioColumnHeaderEn}
	APNeededResourceSets = ColumnAliases{Ja: fsmtable.NeededResourceSetsColumnHeaderJa, En: fsmtable.NeededResourceSetsColumnHeaderEn}
	APPrecondition       = ColumnAliases{Ja: fsmtable.PreconditionColumnHeaderJa, En: fsmtable.PreconditionColumnHeaderEn}
	APMilestone          = ColumnAliases{Ja: fsmtable.MilestoneColumnHeaderJa, En: fsmtable.MilestoneColumnHeaderEn}
	APGroup              = ColumnAliases{Ja: fsmtable.GroupColumnHeaderJa, En: fsmtable.GroupColumnHeaderEn}
	ADAvailableTime      = ColumnAliases{Ja: fsmtable.AvailableTimeHeaderJa, En: fsmtable.AvailableTimeHeaderEn}
	ADMaxRevision        = ColumnAliases{Ja: fsmtable.MaxRevisionHeaderJa, En: fsmtable.MaxRevisionHeaderEn}
)

// APProfileColumns returns the extra-column definitions of AtomicProcessTable corresponding to the mode.
// It panics on an unknown Mode (to suppress out-of-spec fallbacks).
func APProfileColumns(mode Mode) []ColumnAliases {
	switch mode {
	case ModeMinimal:
		return nil
	case ModePlan:
		return []ColumnAliases{APInitialVolume, APReworkVolumeRatio, APNeededResourceSets, APPrecondition}
	case ModePlanMaster:
		return []ColumnAliases{APInitialVolume, APReworkVolumeRatio, APNeededResourceSets, APPrecondition, APMilestone, APGroup}
	default:
		panic(fmt.Sprintf("pfdtable.APProfileColumns: unknown mode: %q", mode))
	}
}

// ADProfileColumns returns the extra-column definitions of AtomicDeliverableTable corresponding to the mode.
// It panics on an unknown Mode (to suppress out-of-spec fallbacks).
func ADProfileColumns(mode Mode) []ColumnAliases {
	switch mode {
	case ModeMinimal:
		return nil
	case ModePlan, ModePlanMaster:
		return []ColumnAliases{ADAvailableTime, ADMaxRevision}
	default:
		panic(fmt.Sprintf("pfdtable.ADProfileColumns: unknown mode: %q", mode))
	}
}

// APExtraHeaders returns the extra header columns of AtomicProcessTable corresponding to the mode.
func APExtraHeaders(mode Mode, l locale.Locale) []string {
	return CanonicalNames(APProfileColumns(mode), l)
}

// ADExtraHeaders returns the extra header columns of AtomicDeliverableTable corresponding to the mode.
func ADExtraHeaders(mode Mode, l locale.Locale) []string {
	return CanonicalNames(ADProfileColumns(mode), l)
}

// CanonicalNames projects a set of ColumnAliases onto a slice of canonical names for the given locale.
// It can be used generically for help text, I18n templates, validation message formatting, and so on.
func CanonicalNames(cols []ColumnAliases, l locale.Locale) []string {
	res := make([]string, 0, len(cols))
	for _, c := range cols {
		res = append(res, c.Canonical(l))
	}
	return res
}

// ApplyAPMode writes the ExtraHeaders corresponding to the mode and the ExtraCells of every row
// (filled with empty strings) to a freshly bootstrapped AtomicProcessTable.
//
// Precondition: t.ExtraHeaders is empty and each t.Rows[i].ExtraCells is empty.
// It panics if these are not satisfied (to suppress out-of-spec fallbacks).
// To add mode columns to an existing table, use EnsureAPExtraHeaders instead.
func ApplyAPMode(t *pfd.AtomicProcessTable, mode Mode, l locale.Locale) {
	if len(t.ExtraHeaders) != 0 {
		panic(fmt.Sprintf("pfdtable.ApplyAPMode: expected empty ExtraHeaders, got %v", t.ExtraHeaders))
	}
	for i, row := range t.Rows {
		if len(row.ExtraCells) != 0 {
			panic(fmt.Sprintf("pfdtable.ApplyAPMode: expected empty ExtraCells on row %d (ID=%q), got %v", i, row.ID, row.ExtraCells))
		}
	}
	headers := APExtraHeaders(mode, l)
	t.ExtraHeaders = headers
	for _, row := range t.Rows {
		row.ExtraCells = make([]string, len(headers))
	}
}

// ApplyADMode is the AtomicDeliverableTable version of ApplyAPMode. The preconditions are the same.
func ApplyADMode(t *pfd.AtomicDeliverableTable, mode Mode, l locale.Locale) {
	if len(t.ExtraHeaders) != 0 {
		panic(fmt.Sprintf("pfdtable.ApplyADMode: expected empty ExtraHeaders, got %v", t.ExtraHeaders))
	}
	for i, row := range t.Rows {
		if len(row.ExtraCells) != 0 {
			panic(fmt.Sprintf("pfdtable.ApplyADMode: expected empty ExtraCells on row %d (ID=%q), got %v", i, row.ID, row.ExtraCells))
		}
	}
	headers := ADExtraHeaders(mode, l)
	t.ExtraHeaders = headers
	for _, row := range t.Rows {
		row.ExtraCells = make([]string, len(headers))
	}
}

// EnsureAPExtraHeaders appends to the end of an existing table whichever mode-required columns are missing.
// It treats the Ja and En notations as equivalent and does not add a column if the corresponding one already exists. It does not change the order of existing headers.
//
// Precondition: at call time, len(row.ExtraCells) == len(t.ExtraHeaders)
// (a state where the headers and each row are consistent, e.g. via pfd.AtomicProcessTable.Refresh).
// If this precondition is not satisfied, the number of cells in the added columns will not match the headers, causing column misalignment.
func EnsureAPExtraHeaders(t *pfd.AtomicProcessTable, mode Mode, l locale.Locale) {
	for _, col := range APProfileColumns(mode) {
		if HasColumn(t.ExtraHeaders, col) {
			continue
		}
		t.ExtraHeaders = append(t.ExtraHeaders, col.Canonical(l))
		for _, row := range t.Rows {
			row.ExtraCells = append(row.ExtraCells, "")
		}
	}
}

// EnsureADExtraHeaders is the AtomicDeliverableTable version of EnsureAPExtraHeaders. The preconditions are the same.
func EnsureADExtraHeaders(t *pfd.AtomicDeliverableTable, mode Mode, l locale.Locale) {
	for _, col := range ADProfileColumns(mode) {
		if HasColumn(t.ExtraHeaders, col) {
			continue
		}
		t.ExtraHeaders = append(t.ExtraHeaders, col.Canonical(l))
		for _, row := range t.Rows {
			row.ExtraCells = append(row.ExtraCells, "")
		}
	}
}

// HasColumn returns whether the existing headers contain a column matching any notation of the ColumnAliases.
func HasColumn(headers []string, col ColumnAliases) bool {
	for _, h := range headers {
		if col.HasMatch(h) {
			return true
		}
	}
	return false
}

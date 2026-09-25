package pfdtable

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/table/tabletsv"
)

type Mode string

const (
	ModeMinimal    Mode = "minimal"
	ModePlan       Mode = "plan"
	ModePlanMaster Mode = "plan-master"
)

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

var IDColumn = ColumnAliases{Ja: tabletsv.IDColumnHeader, En: tabletsv.IDColumnHeader}

var (
	APInitialVolume      = ColumnAliases{Ja: fsmtable.InitialVolumeColumnHeaderJa, En: fsmtable.InitialVolumeColumnHeaderEn}
	APReworkVolumeRatio  = ColumnAliases{Ja: fsmtable.ReworkVolumeRatioColumnHeaderJa, En: fsmtable.ReworkVolumeRatioColumnHeaderEn}
	APNeededResourceSets = ColumnAliases{Ja: fsmtable.NeededResourceSetsColumnHeaderJa, En: fsmtable.NeededResourceSetsColumnHeaderEn}
	APPrecondition       = ColumnAliases{Ja: fsmtable.PreconditionColumnHeaderJa, En: fsmtable.PreconditionColumnHeaderEn}

	APBar           = ColumnAliases{Ja: mastertsv.BarColumnHeader, En: mastertsv.BarColumnHeader}
	APRow           = ColumnAliases{Ja: mastertsv.RowColumnHeader, En: mastertsv.RowColumnHeader}
	ADAvailableTime = ColumnAliases{Ja: fsmtable.AvailableTimeHeaderJa, En: fsmtable.AvailableTimeHeaderEn}
	ADMaxRevision   = ColumnAliases{Ja: fsmtable.MaxRevisionHeaderJa, En: fsmtable.MaxRevisionHeaderEn}
)

func APProfileColumns(mode Mode) []ColumnAliases {
	switch mode {
	case ModeMinimal:
		return nil
	case ModePlan:
		return []ColumnAliases{APInitialVolume, APReworkVolumeRatio, APNeededResourceSets, APPrecondition}
	case ModePlanMaster:
		return []ColumnAliases{APInitialVolume, APReworkVolumeRatio, APNeededResourceSets, APPrecondition, APRow, APBar}
	default:
		panic(fmt.Sprintf("pfdtable.APProfileColumns: unknown mode: %q", mode))
	}
}

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

func FilterColumnsByModel(cols []ColumnAliases, model execmodel.Model) []ColumnAliases {
	if err := model.Validate(); err != nil {
		panic(fmt.Sprintf("pfdtable.FilterColumnsByModel: %s", err))
	}
	res := make([]ColumnAliases, 0, len(cols))
	for _, c := range cols {
		if model.Resource == execmodel.ResourceModeInfinite && c == APNeededResourceSets {
			continue
		}
		if model.Feedback == execmodel.FeedbackModeDisabled && (c == APReworkVolumeRatio || c == ADMaxRevision) {
			continue
		}
		res = append(res, c)
	}
	return res
}

func APExtraHeaders(mode Mode, model execmodel.Model, l locale.Locale) []string {
	return CanonicalNames(FilterColumnsByModel(APProfileColumns(mode), model), l)
}

func ADExtraHeaders(mode Mode, model execmodel.Model, l locale.Locale) []string {
	return CanonicalNames(FilterColumnsByModel(ADProfileColumns(mode), model), l)
}

func CanonicalNames(cols []ColumnAliases, l locale.Locale) []string {
	res := make([]string, 0, len(cols))
	for _, c := range cols {
		res = append(res, c.Canonical(l))
	}
	return res
}

func ApplyAPMode(t *pfd.AtomicProcessTable, mode Mode, model execmodel.Model, l locale.Locale) {
	if len(t.ExtraHeaders) != 0 {
		panic(fmt.Sprintf("pfdtable.ApplyAPMode: expected empty ExtraHeaders, got %v", t.ExtraHeaders))
	}
	for i, row := range t.Rows {
		if len(row.ExtraCells) != 0 {
			panic(fmt.Sprintf("pfdtable.ApplyAPMode: expected empty ExtraCells on row %d (ID=%q), got %v", i, row.ID, row.ExtraCells))
		}
	}
	headers := APExtraHeaders(mode, model, l)
	t.ExtraHeaders = headers
	for _, row := range t.Rows {
		row.ExtraCells = make([]string, len(headers))
	}
}

func ApplyADMode(t *pfd.AtomicDeliverableTable, mode Mode, model execmodel.Model, l locale.Locale) {
	if len(t.ExtraHeaders) != 0 {
		panic(fmt.Sprintf("pfdtable.ApplyADMode: expected empty ExtraHeaders, got %v", t.ExtraHeaders))
	}
	for i, row := range t.Rows {
		if len(row.ExtraCells) != 0 {
			panic(fmt.Sprintf("pfdtable.ApplyADMode: expected empty ExtraCells on row %d (ID=%q), got %v", i, row.ID, row.ExtraCells))
		}
	}
	headers := ADExtraHeaders(mode, model, l)
	t.ExtraHeaders = headers
	for _, row := range t.Rows {
		row.ExtraCells = make([]string, len(headers))
	}
}

func EnsureAPExtraHeaders(t *pfd.AtomicProcessTable, mode Mode, model execmodel.Model, l locale.Locale) {
	for _, col := range FilterColumnsByModel(APProfileColumns(mode), model) {
		if HasColumn(t.ExtraHeaders, col) {
			continue
		}
		t.ExtraHeaders = append(t.ExtraHeaders, col.Canonical(l))
		for _, row := range t.Rows {
			row.ExtraCells = append(row.ExtraCells, "")
		}
	}
}

func EnsureADExtraHeaders(t *pfd.AtomicDeliverableTable, mode Mode, model execmodel.Model, l locale.Locale) {
	for _, col := range FilterColumnsByModel(ADProfileColumns(mode), model) {
		if HasColumn(t.ExtraHeaders, col) {
			continue
		}
		t.ExtraHeaders = append(t.ExtraHeaders, col.Canonical(l))
		for _, row := range t.Rows {
			row.ExtraCells = append(row.ExtraCells, "")
		}
	}
}

func HasColumn(headers []string, col ColumnAliases) bool {
	for _, h := range headers {
		if col.HasMatch(h) {
			return true
		}
	}
	return false
}

package pfdtable_test

import (
	"slices"
	"testing"

	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable"
)

func TestAPExtraHeaders(t *testing.T) {
	cases := []struct {
		name   string
		mode   pfdtable.Mode
		locale locale.Locale
		want   []string
	}{
		{"minimal/ja", pfdtable.ModeMinimal, locale.LocaleJa, []string{}},
		{"plan/ja", pfdtable.ModePlan, locale.LocaleJa, []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件"}},
		{"plan-master/ja", pfdtable.ModePlanMaster, locale.LocaleJa, []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件", "マイルストーン", "グループ"}},
		{"plan/en", pfdtable.ModePlan, locale.LocaleEn, []string{"Est. Work Volume", "Est. Rework Volume Ratio", "Needed Resources", "Start Condition"}},
		{"plan-master/en", pfdtable.ModePlanMaster, locale.LocaleEn, []string{"Est. Work Volume", "Est. Rework Volume Ratio", "Needed Resources", "Start Condition", "Milestone", "Group"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := pfdtable.APExtraHeaders(c.mode, c.locale)
			if !slices.Equal(got, c.want) {
				t.Errorf("APExtraHeaders(%v, %v) = %v, want %v", c.mode, c.locale, got, c.want)
			}
		})
	}
}

func TestADExtraHeaders(t *testing.T) {
	cases := []struct {
		name   string
		mode   pfdtable.Mode
		locale locale.Locale
		want   []string
	}{
		{"minimal/ja", pfdtable.ModeMinimal, locale.LocaleJa, []string{}},
		{"plan/ja", pfdtable.ModePlan, locale.LocaleJa, []string{"利用可能時刻", "最大版"}},
		{"plan-master/ja", pfdtable.ModePlanMaster, locale.LocaleJa, []string{"利用可能時刻", "最大版"}},
		{"plan/en", pfdtable.ModePlan, locale.LocaleEn, []string{"Available Time", "Max Revision"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := pfdtable.ADExtraHeaders(c.mode, c.locale)
			if !slices.Equal(got, c.want) {
				t.Errorf("ADExtraHeaders(%v, %v) = %v, want %v", c.mode, c.locale, got, c.want)
			}
		})
	}
}

func TestApplyAPMode(t *testing.T) {
	tbl := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "implement"},
			{ID: "P2", Description: "review"},
		},
	}
	pfdtable.ApplyAPMode(tbl, pfdtable.ModePlan, locale.LocaleJa)

	wantHeaders := []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件"}
	if !slices.Equal(tbl.ExtraHeaders, wantHeaders) {
		t.Errorf("ExtraHeaders = %v, want %v", tbl.ExtraHeaders, wantHeaders)
	}
	for _, row := range tbl.Rows {
		if len(row.ExtraCells) != 4 {
			t.Errorf("row %s: ExtraCells len = %d, want 4", row.ID, len(row.ExtraCells))
		}
		for _, cell := range row.ExtraCells {
			if cell != "" {
				t.Errorf("row %s: expected empty cell, got %q", row.ID, cell)
			}
		}
	}
}

func TestApplyAPMode_PanicsOnNonEmptyExtraHeaders(t *testing.T) {
	tbl := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{"existing"},
		Rows:         []*pfd.AtomicProcessRow{{ID: "P1", Description: "implement"}},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ApplyAPMode should panic on non-empty ExtraHeaders, but did not")
		}
	}()
	pfdtable.ApplyAPMode(tbl, pfdtable.ModePlan, locale.LocaleJa)
}

func TestApplyAPMode_PanicsOnNonEmptyExtraCells(t *testing.T) {
	tbl := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "implement", ExtraCells: []string{"x"}},
		},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ApplyAPMode should panic on non-empty ExtraCells, but did not")
		}
	}()
	pfdtable.ApplyAPMode(tbl, pfdtable.ModePlan, locale.LocaleJa)
}

func TestApplyADMode_PanicsOnNonEmptyExtraHeaders(t *testing.T) {
	tbl := &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{"existing"},
		Rows:         []*pfd.AtomicDeliverableRow{{ID: "D1", Description: "doc"}},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ApplyADMode should panic on non-empty ExtraHeaders, but did not")
		}
	}()
	pfdtable.ApplyADMode(tbl, pfdtable.ModePlan, locale.LocaleJa)
}

func TestApplyADMode_PanicsOnNonEmptyExtraCells(t *testing.T) {
	tbl := &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{},
		Rows: []*pfd.AtomicDeliverableRow{
			{ID: "D1", Description: "doc", ExtraCells: []string{"x"}},
		},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ApplyADMode should panic on non-empty ExtraCells, but did not")
		}
	}()
	pfdtable.ApplyADMode(tbl, pfdtable.ModePlan, locale.LocaleJa)
}

func TestEnsureAPExtraHeaders_AppendsMissing(t *testing.T) {
	tbl := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{"Note"},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "implement", ExtraCells: []string{"keep me"}},
		},
	}
	pfdtable.EnsureAPExtraHeaders(tbl, pfdtable.ModePlan, locale.LocaleJa)

	wantHeaders := []string{"Note", "予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件"}
	if !slices.Equal(tbl.ExtraHeaders, wantHeaders) {
		t.Errorf("ExtraHeaders = %v, want %v", tbl.ExtraHeaders, wantHeaders)
	}
	wantCells := []string{"keep me", "", "", "", ""}
	if !slices.Equal(tbl.Rows[0].ExtraCells, wantCells) {
		t.Errorf("row P1 cells = %v, want %v", tbl.Rows[0].ExtraCells, wantCells)
	}
}

func TestEnsureAPExtraHeaders_NoDuplicateAcrossLocale(t *testing.T) {
	// When columns already exist in their Ja notation, calling Ensure with locale=en does not add the English columns
	tbl := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件"},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "implement", ExtraCells: []string{"1", "0.1", "R1:1", ""}},
		},
	}
	pfdtable.EnsureAPExtraHeaders(tbl, pfdtable.ModePlan, locale.LocaleEn)

	want := []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件"}
	if !slices.Equal(tbl.ExtraHeaders, want) {
		t.Errorf("ExtraHeaders = %v, want %v", tbl.ExtraHeaders, want)
	}
}

func TestEnsureAPExtraHeaders_PlanMasterOverPlan(t *testing.T) {
	tbl := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件"},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "implement", ExtraCells: []string{"1", "0.1", "R1:1", ""}},
		},
	}
	pfdtable.EnsureAPExtraHeaders(tbl, pfdtable.ModePlanMaster, locale.LocaleJa)

	want := []string{"予想作業量", "予想手戻り作業量割合", "必要資源", "開始条件", "マイルストーン", "グループ"}
	if !slices.Equal(tbl.ExtraHeaders, want) {
		t.Errorf("ExtraHeaders = %v, want %v", tbl.ExtraHeaders, want)
	}
	if len(tbl.Rows[0].ExtraCells) != 6 {
		t.Errorf("row P1: ExtraCells len = %d, want 6", len(tbl.Rows[0].ExtraCells))
	}
	if tbl.Rows[0].ExtraCells[4] != "" || tbl.Rows[0].ExtraCells[5] != "" {
		t.Errorf("row P1: appended cells should be empty, got %v", tbl.Rows[0].ExtraCells)
	}
}

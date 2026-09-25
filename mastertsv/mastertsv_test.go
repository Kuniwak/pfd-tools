package mastertsv

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestParseDescriptions(t *testing.T) {
	testCases := map[string]struct {
		TSV      string
		Expected map[string]string
		WantErr  string
	}{
		"英語ヘッダ（milestone.tsv 形式）": {
			TSV:      "ID\tDescription\tGroups\tSuccessors\nM1\tマイルストーン1\tG1\tM2\n",
			Expected: map[string]string{"M1": "マイルストーン1"},
		},
		"日本語ヘッダ（group.tsv 形式）": {
			TSV:      "ID\t説明\nG1\tグループ1\n",
			Expected: map[string]string{"G1": "グループ1"},
		},
		"ID が空の行は無視する": {
			TSV:      "ID\t説明\n\tグループ1\nG1\tグループ1\n",
			Expected: map[string]string{"G1": "グループ1"},
		},
		"ID 列がなければエラー": {
			TSV:     "Key\t説明\nG1\tグループ1\n",
			WantErr: "ID",
		},
		"説明列がなければエラー": {
			TSV:     "ID\tGroups\nM1\tG1\n",
			WantErr: "Description",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseDescriptions(strings.NewReader(tc.TSV))

			if tc.WantErr != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil (result: %v)", tc.WantErr, got)
				}
				if !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("want error containing %q, got: %s", tc.WantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDescriptions: %s", err)
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got))
			}
		})
	}
}

func TestRequireKnownIDs(t *testing.T) {
	known := sets.New(pfd.AtomicProcessID.Compare, "P1", "P2")

	testCases := map[string]struct {
		Table   *Table
		WantErr string
	}{
		"すべて既知なら成功": {
			Table: &Table{
				Classifications: []*Classification{
					{ID: "P1", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M1"},
					{ID: "P2", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M2"},
				},
			},
		},
		"一部だけ分類していても成功": {
			Table: &Table{
				Classifications: []*Classification{
					{ID: "P2", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M2"},
				},
			},
		},
		"未知の ID はエラー": {
			Table: &Table{
				Classifications: []*Classification{
					{ID: "P1", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M1"},
					{ID: "P9", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M1"},
				},
			},
			WantErr: "P9",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := RequireKnownIDs(tc.Table, known)

			if tc.WantErr == "" {
				if err != nil {
					t.Fatalf("RequireKnownIDs: %s", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tc.WantErr)
			}
			if !strings.Contains(err.Error(), tc.WantErr) {
				t.Fatalf("want error containing %q, got: %s", tc.WantErr, err)
			}
		})
	}
}

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		TSV      string
		Expected *Table
		WantErr  string
	}{
		"正常系": {
			TSV: "ID\tMasterRow\tMasterBar\nP1\tG1\tM1\nP2\tG2\tM2\n",
			Expected: &Table{
				Classifications: []*Classification{
					{ID: "P1", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M1"},
					{ID: "P2", Rows: sets.New(Row.Compare, Row("G2")), Bar: "M2"},
				},
			},
		},
		"多値 Row は空白を除去して分解する": {
			TSV: "ID\tMasterRow\tMasterBar\nP1\tG1, G2\tM1\n",
			Expected: &Table{
				Classifications: []*Classification{
					{ID: "P1", Rows: sets.New(Row.Compare, Row("G1"), Row("G2")), Bar: "M1"},
				},
			},
		},
		"未知の列は無視し列順は自由": {
			TSV: "MasterBar\tメモ\tID\tMasterRow\nM1\tなんでも\tP1\tG1\n",
			Expected: &Table{
				Classifications: []*Classification{
					{ID: "P1", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M1"},
				},
			},
		},
		"ID が空の行は無視する": {
			TSV: "ID\tMasterRow\tMasterBar\n\tG1\tM1\nP1\tG1\tM1\n",
			Expected: &Table{
				Classifications: []*Classification{
					{ID: "P1", Rows: sets.New(Row.Compare, Row("G1")), Bar: "M1"},
				},
			},
		},
		"Row が空の行は分類しない": {
			TSV: "ID\tMasterRow\tMasterBar\nP1\t\tM1\n",
			Expected: &Table{
				Classifications: []*Classification{},
			},
		},
		"Bar が空の行は分類しない": {
			TSV: "ID\tMasterRow\tMasterBar\nP1\tG1\t\n",
			Expected: &Table{
				Classifications: []*Classification{},
			},
		},
		"ID 列がなければエラー": {
			TSV:     "MasterRow\tMasterBar\nG1\tM1\n",
			WantErr: "ID",
		},
		"Row 列がなければエラー": {
			TSV:     "ID\tMasterBar\nP1\tM1\n",
			WantErr: "MasterRow",
		},
		"Bar 列がなければエラー": {
			TSV:     "ID\tMasterRow\nP1\tG1\n",
			WantErr: "MasterBar",
		},
		"ID が重複していればエラー": {
			TSV:     "ID\tMasterRow\tMasterBar\nP1\tG1\tM1\nP1\tG2\tM2\n",
			WantErr: "P1",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tc.TSV))

			if tc.WantErr != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil (result: %v)", tc.WantErr, got)
				}
				if !strings.Contains(err.Error(), tc.WantErr) {
					t.Fatalf("want error containing %q, got: %s", tc.WantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse: %s", err)
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Error(cmp.Diff(tc.Expected, got, cmp.Comparer(func(a, b *sets.Set[Row]) bool {
					return sets.IsEqual(Row.Compare, a, b)
				})))
			}
		})
	}
}

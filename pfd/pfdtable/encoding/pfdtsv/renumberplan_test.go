package pfdtsv

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/google/go-cmp/cmp"
)

func TestWriteRenumberPlan(t *testing.T) {
	testCases := map[string]struct {
		Plan pfd.RenumberPlan
		Want string
	}{
		"newly numbered process": {
			Plan: pfd.RenumberPlan{
				"(テストを書く)": &pfd.Node{ID: "P12", Description: "テストを書く", Type: pfd.NodeTypeAtomicProcess},
			},
			Want: "Key\tID\n(テストを書く)\tP12\n",
		},
		"an empty plan yields a header only": {
			Plan: pfd.RenumberPlan{},
			Want: "Key\tID\n",
		},
		"rows are sorted by key": {
			Plan: pfd.RenumberPlan{
				"(テストを書く)": &pfd.Node{ID: "P12", Description: "テストを書く", Type: pfd.NodeTypeAtomicProcess},
				"[テストコード]": &pfd.Node{ID: "D9", Description: "テストコード", Type: pfd.NodeTypeAtomicDeliverable},
				"(あとしまつ)":  &pfd.Node{ID: "P13", Description: "あとしまつ", Type: pfd.NodeTypeAtomicProcess},
			},
			Want: "Key\tID\n(あとしまつ)\tP13\n(テストを書く)\tP12\n[テストコード]\tD9\n",
		},
		"rows are sorted lexicographically, not by length": {
			Plan: pfd.RenumberPlan{
				"(ん)":  &pfd.Node{ID: "P1", Description: "ん", Type: pfd.NodeTypeAtomicProcess},
				"(あい)": &pfd.Node{ID: "P2", Description: "あい", Type: pfd.NodeTypeAtomicProcess},
			},
			Want: "Key\tID\n(あい)\tP2\n(ん)\tP1\n",
		},
		"a key containing a tab is quoted": {
			Plan: pfd.RenumberPlan{
				"(テスト\tを書く)": &pfd.Node{ID: "P12", Description: "テスト\tを書く", Type: pfd.NodeTypeAtomicProcess},
			},
			Want: "Key\tID\n\"(テスト\tを書く)\"\tP12\n",
		},
		"a key containing a newline is quoted": {
			Plan: pfd.RenumberPlan{
				"(テスト\nを書く)": &pfd.Node{ID: "P12", Description: "テスト\nを書く", Type: pfd.NodeTypeAtomicProcess},
			},
			Want: "Key\tID\n\"(テスト\nを書く)\"\tP12\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			if err := WriteRenumberPlan(w, tc.Plan); err != nil {
				t.Fatalf("WriteRenumberPlan: %v", err)
			}
			if got := w.String(); got != tc.Want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tc.Want)
			}
		})
	}
}

func TestParseRenumberPlan(t *testing.T) {
	testCases := map[string]struct {
		TSV  string
		Want pfd.RenumberPlan
	}{

		"a newly numbered process": {
			TSV: "Key\tID\n(テストを書く)\tP12\n",
			Want: pfd.RenumberPlan{
				"(テストを書く)": &pfd.Node{ID: "P12", Description: "テストを書く"},
			},
		},
		"a newly numbered deliverable": {
			TSV: "Key\tID\n[テストコード]\tD9\n",
			Want: pfd.RenumberPlan{
				"[テストコード]": &pfd.Node{ID: "D9", Description: "テストコード"},
			},
		},

		"the same label as a process and a deliverable are different keys": {
			TSV: "Key\tID\n(実装する)\tP1\n[実装する]\tD1\n",
			Want: pfd.RenumberPlan{
				"(実装する)": &pfd.Node{ID: "P1", Description: "実装する"},
				"[実装する]": &pfd.Node{ID: "D1", Description: "実装する"},
			},
		},
		"header only": {
			TSV:  "Key\tID\n",
			Want: pfd.RenumberPlan{},
		},

		"an empty key is kept": {
			TSV: "Key\tID\n\"\"\tD1\n",
			Want: pfd.RenumberPlan{
				"": &pfd.Node{ID: "D1", Description: ""},
			},
		},
		"a quoted key containing a tab": {
			TSV: "Key\tID\n\"(テスト\tを書く)\"\tP12\n",
			Want: pfd.RenumberPlan{
				"(テスト\tを書く)": &pfd.Node{ID: "P12", Description: "テスト\tを書く"},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseRenumberPlan(strings.NewReader(tc.TSV))
			if err != nil {
				t.Fatalf("ParseRenumberPlan: %v", err)
			}
			if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRenumberPlanRoundTrip(t *testing.T) {
	plan := pfd.RenumberPlan{
		"(テストを書く)":   &pfd.Node{ID: "P12", Description: "テストを書く"},
		"[テスト\tコード]": &pfd.Node{ID: "D9", Description: "テスト\tコード"},
		"(あと\nしまつ)":  &pfd.Node{ID: "P13", Description: "あと\nしまつ"},
	}

	w := bytes.NewBuffer(nil)
	if err := WriteRenumberPlan(w, plan); err != nil {
		t.Fatalf("WriteRenumberPlan: %v", err)
	}

	got, err := ParseRenumberPlan(w)
	if err != nil {
		t.Fatalf("ParseRenumberPlan: %v", err)
	}

	if diff := cmp.Diff(plan, got); diff != "" {
		t.Error(diff)
	}
}

func TestParseRenumberPlan_Invalid(t *testing.T) {
	testCases := map[string]struct {
		TSV string
	}{

		"missing header": {
			TSV: "(設計する)\tP2\n(実装する)\tP3\n",
		},
		"unexpected header": {
			TSV: "Description\tID\n(設計する)\tP2\n",
		},
		"empty input": {
			TSV: "",
		},
		"too few columns": {
			TSV: "Key\tID\n(設計する)\n",
		},
		"invalid id": {
			TSV: "Key\tID\n(設計する)\tnot-an-id\n",
		},

		"already numbered key": {
			TSV: "Key\tID\nD1\tD1\n",
		},

		"a legacy key without a kind": {
			TSV: "Key\tID\n設計する\tP2\n",
		},

		"a process key with a deliverable ID": {
			TSV: "Key\tID\n(設計する)\tD9\n",
		},
		"a deliverable key with a process ID": {
			TSV: "Key\tID\n[設計書]\tP9\n",
		},

		"duplicate key": {
			TSV: "Key\tID\n(設計する)\tP2\n(設計する)\tP9\n",
		},

		"duplicate ID": {
			TSV: "Key\tID\n(設計する)\tP2\n(調査する)\tP2\n",
		},
		"the same ID for a process key and a deliverable key": {
			TSV: "Key\tID\n(設計する)\tP2\n[設計書]\tP2\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseRenumberPlan(strings.NewReader(tc.TSV)); err == nil {
				t.Error("ParseRenumberPlan: err = nil, want an error")
			}
		})
	}
}

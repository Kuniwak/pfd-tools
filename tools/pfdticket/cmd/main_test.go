package cmd

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestHelpWritesToInoutStderr(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-h"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
	if got := spy.Stderr.String(); !strings.Contains(got, "Usage") {
		t.Errorf("stderr should contain the help text, got %q", got)
	}
}

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		args     []string
		contains []string
	}{
		"json via individual flags": {
			args: []string{"-p", "testdata/basic/pfd.drawio.png", "-ad", "testdata/basic/ad.tsv"},
			contains: []string{
				`"id": "P1"`,
				`"summary": "P1: Process"`,
				"# 入力成果物の一覧",
				"# 出力成果物の一覧と品質基準とレビューア",
				"品質基準: レビュー指摘が全てクローズ",
				"レビューア: @alice",
				"成果物リンク: （完了時に URL を記入）",
			},
		},
		"without ad table still lists deliverables": {
			args:     []string{"-p", "testdata/basic/pfd.drawio.png"},
			contains: []string{"- D2: Input", "成果物リンク: （完了時に URL を記入）"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.args, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}
			out := spy.Stdout.String()
			for _, want := range tc.contains {
				if !strings.Contains(out, want) {
					t.Errorf("stdout does not contain %q\n--- stdout ---\n%s", want, out)
				}
			}
		})
	}
}

func TestMainCommandByArgsTSV(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-p", "testdata/basic/pfd.drawio.png", "-ad", "testdata/basic/ad.tsv", "-out-format", "tsv"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}

	reader := csv.NewReader(strings.NewReader(spy.Stdout.String()))
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("csv.Reader failed to parse tsv output: %v\n--- stdout ---\n%s", err, spy.Stdout.String())
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2 (header + P1)", len(records))
	}
	if got, want := records[0], []string{"ID", "Summary", "Description"}; !reflect.DeepEqual(got, want) {
		t.Errorf("header = %v, want %v", got, want)
	}
	row := records[1]
	if row[0] != "P1" {
		t.Errorf("ID = %q, want %q", row[0], "P1")
	}
	if row[1] != "P1: Process" {
		t.Errorf("Summary = %q, want %q", row[1], "P1: Process")
	}
	description := row[2]
	for _, want := range []string{
		"# 入力成果物の一覧",
		"# 出力成果物の一覧と品質基準とレビューア",
		"品質基準: レビュー指摘が全てクローズ",
		"成果物リンク: （完了時に URL を記入）",
	} {
		if !strings.Contains(description, want) {
			t.Errorf("Description does not contain %q\n--- Description ---\n%s", want, description)
		}
	}
	if !strings.Contains(description, "\n") {
		t.Errorf("Description should keep real newlines after a csv round-trip (not \\n-escaped)\n%s", description)
	}
}

func TestMainCommandByArgsErrors(t *testing.T) {
	testCases := map[string][]string{
		"invalid out-format": {"-p", "testdata/basic/pfd.drawio.png", "-out-format", "yaml"},
		"missing pfd":        {"-ad", "testdata/basic/ad.tsv"},
	}
	for name, args := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("exitStatus = 0, want non-zero")
			}
		})
	}
}

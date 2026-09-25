package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	writeTSV := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "master.tsv")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write master tsv: %v", err)
		}
		return path
	}

	testCases := map[string]struct {
		Args           func(t *testing.T) []string
		WantExitStatus int
		WantStdout     []string
		WantStderr     []string
	}{
		"既定の出力形式は google-spreadsheet-tsv": {
			Args: func(t *testing.T) []string {
				return []string{"-tsv", "testdata/loop/atomic_proc.tsv", "-b", "1.5", "testdata/loop/plan.json"}
			},
			WantExitStatus: 0,
			WantStdout:     []string{"MasterRow\tMasterRowDescription\tMasterBar\tMasterBarDescription\tStart\tEnd\n"},
		},
		"mermaid 形式": {
			Args: func(t *testing.T) []string {
				return []string{"-tsv", "testdata/loop/atomic_proc.tsv", "-out-format", "mermaid", "testdata/loop/plan.json"}
			},
			WantExitStatus: 0,
			WantStdout:     []string{"gantt\n", "section G1"},
		},
		"plantuml 形式": {
			Args: func(t *testing.T) []string {
				return []string{"-tsv", "testdata/loop/atomic_proc.tsv", "-out-format", "plantuml", "testdata/loop/plan.json"}
			},
			WantExitStatus: 0,
			WantStdout:     []string{"@startgantt\n", "@endgantt", "-- G1"},
		},
		"未対応の出力形式はエラー": {
			Args: func(t *testing.T) []string {
				return []string{"-tsv", "testdata/loop/atomic_proc.tsv", "-out-format", "xml", "testdata/loop/plan.json"}
			},
			WantExitStatus: 1,
		},
		"-tsv は必須": {
			Args: func(t *testing.T) []string {
				return []string{"testdata/loop/plan.json"}
			},
			WantExitStatus: 1,
			WantStderr:     []string{"-tsv is required"},
		},
		"実行計画にない ID を分類するとエラー": {
			Args: func(t *testing.T) []string {
				return []string{"-tsv", writeTSV(t, "ID\tMasterRow\tMasterBar\nP9\tG1\tM1\n"), "testdata/loop/plan.json"}
			},
			WantExitStatus: 1,
			WantStderr:     []string{"P9"},
		},
		"分類していない原子プロセスは描画しない": {
			Args: func(t *testing.T) []string {
				return []string{"-tsv", writeTSV(t, "ID\tMasterRow\tMasterBar\nP1\tG1\tM1\n"), "-start", "2026-06-01", "testdata/loop/plan.json"}
			},
			WantExitStatus: 0,
			WantStdout: []string{
				"MasterRow\tMasterRowDescription\tMasterBar\tMasterBarDescription\tStart\tEnd\n" +
					"G1\t\tM1\t\t2026-06-01 10:00:00\t2026-06-03 10:00:00\n",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.Args(t), spy.NewProcInout())
			if exitStatus != tc.WantExitStatus {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Fatalf("exitStatus = %d, want %d", exitStatus, tc.WantExitStatus)
			}
			for _, want := range tc.WantStdout {
				if !strings.Contains(spy.Stdout.String(), want) {
					t.Errorf("expected stdout to contain %q, but got:\n%s", want, spy.Stdout.String())
				}
			}
			for _, want := range tc.WantStderr {
				if !strings.Contains(spy.Stderr.String(), want) {
					t.Errorf("expected stderr to mention %q, but got:\n%s", want, spy.Stderr.String())
				}
			}
		})
	}
}

func TestMainCommandByArgs_FirstExecutionOnly(t *testing.T) {
	args := []string{
		"-tsv", "testdata/loop/atomic_proc.tsv",
		"-row-meta", "testdata/loop/group.tsv",
		"-bar-meta", "testdata/loop/milestone.tsv",
		"-start", "2026-06-01",
		"testdata/loop/plan.json",
	}
	want := "MasterRow\tMasterRowDescription\tMasterBar\tMasterBarDescription\tStart\tEnd\n" +
		"G1\tグループ\tM1\tマイルストーン1\t2026-06-01 10:00:00\t2026-06-03 10:00:00\n" +
		"G1\tグループ\tM2\tマイルストーン2\t2026-06-04 14:30:00\t2026-06-08 14:30:00\n"

	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs(args, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}
	if spy.Stdout.String() != want {
		t.Errorf("stdout = %q, want %q", spy.Stdout.String(), want)
	}
}

func TestMainCommandByArgs_ProjectConfig(t *testing.T) {
	absPath := func(path string) string {
		t.Helper()
		res, err := filepath.Abs(path)
		if err != nil {
			t.Fatalf("abs %q: %v", path, err)
		}
		return res
	}

	writeConfig := func(t *testing.T, extra map[string]any) string {
		t.Helper()
		merged, err := json.Marshal(extra)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}
		path := filepath.Join(t.TempDir(), "project.json")
		if err := os.WriteFile(path, merged, 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		return path
	}

	baseArgs := []string{"-tsv", "testdata/loop/atomic_proc.tsv"}

	testCases := map[string]struct {
		Extra            map[string]any
		Args             []string
		WantExitStatus   int
		ExpectedStdout   []string
		ExpectedMentions []string
	}{
		"実行計画のパスは構成ファイルから読める": {
			Extra:          map[string]any{"plan": absPath("testdata/loop/plan.json")},
			WantExitStatus: 0,
			ExpectedStdout: []string{"MasterRow\t"},
		},
		"positional の実行計画パスは構成ファイルより優先される": {
			Extra:          map[string]any{"plan": "no_such_plan.json"},
			Args:           []string{"testdata/loop/plan.json"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"MasterRow\t"},
		},
		"出力形式は構成ファイルから読める": {
			Extra:          map[string]any{"output_format": "mermaid"},
			Args:           []string{"testdata/loop/plan.json"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"gantt", "section G1"},
		},
		"業務時間の設定は構成ファイルから読める": {
			Extra:          map[string]any{"output_format": "mermaid", "start_day": "2026-06-01"},
			Args:           []string{"testdata/loop/plan.json"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"2026-06-01"},
		},
		"コマンドラインは構成ファイルより優先される": {
			Extra:          map[string]any{"output_format": "mermaid", "start_day": "2026-06-01"},
			Args:           []string{"-out-format", "plantuml", "-start", "2026-07-01", "testdata/loop/plan.json"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"@startgantt", "2026-07-01"},
		},
		"planmaster が対応しない出力形式は理由つきで弾く": {
			Extra:            map[string]any{"output_format": "plan-json"},
			Args:             []string{"testdata/loop/plan.json"},
			WantExitStatus:   1,
			ExpectedMentions: []string{"plan-json"},
		},
		"positional argument は 1 つまで": {
			Extra:            map[string]any{},
			Args:             []string{"testdata/loop/plan.json", "testdata/loop/plan.json"},
			WantExitStatus:   1,
			ExpectedMentions: []string{"too many arguments"},
		},
		"実行計画のパスは必須": {
			Extra:            map[string]any{},
			WantExitStatus:   1,
			ExpectedMentions: []string{"plan path is required"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := append([]string{"-f", writeConfig(t, tc.Extra)}, baseArgs...)
			args = append(args, tc.Args...)

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != tc.WantExitStatus {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Fatalf("exitStatus = %d, want %d", exitStatus, tc.WantExitStatus)
			}
			for _, want := range tc.ExpectedStdout {
				if !strings.Contains(spy.Stdout.String(), want) {
					t.Errorf("expected stdout to contain %q, but got:\n%s", want, spy.Stdout.String())
				}
			}
			for _, mention := range tc.ExpectedMentions {
				if !strings.Contains(spy.Stderr.String(), mention) {
					t.Errorf("expected stderr to mention %q, but got:\n%s", mention, spy.Stderr.String())
				}
			}
		})
	}
}

func TestMainCommandByArgs_EmphasisTSV(t *testing.T) {
	writeEmphasisTSV := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "em.tsv")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write emphasis tsv: %v", err)
		}
		return path
	}

	testCases := map[string]struct {
		OutFormat      string
		EmphasisTSV    string
		WantExitStatus int
		WantStdout     []string
		WantStderr     []string
	}{
		"mermaid では crit タグが付く": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "ID\nM1\n",
			WantExitStatus: 0,
			WantStdout:     []string{"M1 マイルストーン1 :crit, "},
		},
		"plantuml では色が付く": {
			OutFormat:      "plantuml",
			EmphasisTSV:    "ID\nM1\n",
			WantExitStatus: 0,
			WantStdout:     []string{"] is colored in Salmon"},
		},
		"google-spreadsheet-tsv では Emphasis 列が付く": {
			OutFormat:      "google-spreadsheet-tsv",
			EmphasisTSV:    "ID\nM1\n",
			WantExitStatus: 0,
			WantStdout:     []string{"\tEmphasis\n", "\tTRUE\n"},
		},
		"ID 列がない表はエラーになる": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "Key\nM1\n",
			WantExitStatus: 1,
		},
		"マスタースケジュールにない ID は警告して無視する": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "ID\nM1\nM99\n",
			WantExitStatus: 0,
			WantStdout:     []string{"M1 マイルストーン1 :crit, "},
			WantStderr:     []string{"M99"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := []string{
				"-tsv", "testdata/loop/atomic_proc.tsv",
				"-row-meta", "testdata/loop/group.tsv",
				"-bar-meta", "testdata/loop/milestone.tsv",
				"-out-format", tc.OutFormat,
				"-em-tsv", writeEmphasisTSV(t, tc.EmphasisTSV),
				"testdata/loop/plan.json",
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != tc.WantExitStatus {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want %d", exitStatus, tc.WantExitStatus)
			}
			for _, want := range tc.WantStdout {
				if !strings.Contains(spy.Stdout.String(), want) {
					t.Errorf("stdout = %q, want to contain %q", spy.Stdout.String(), want)
				}
			}
			for _, want := range tc.WantStderr {
				if !strings.Contains(spy.Stderr.String(), want) {
					t.Errorf("stderr = %q, want to contain %q", spy.Stderr.String(), want)
				}
			}
		})
	}
}

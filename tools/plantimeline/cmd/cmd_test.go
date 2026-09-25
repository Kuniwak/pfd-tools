package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func TestMainCommandByArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-f", "testdata/loop/config.json", "testdata/loop/plan.json"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}

func TestMainCommandByArgs_PlanPathFromProjectConfig(t *testing.T) {
	writeConfig := func(t *testing.T, planPath string) string {
		t.Helper()
		raw, err := os.ReadFile("testdata/loop/config.json")
		if err != nil {
			t.Fatalf("read config: %v", err)
		}
		var config map[string]any
		if err := json.Unmarshal(raw, &config); err != nil {
			t.Fatalf("unmarshal config: %v", err)
		}
		config["plan"] = planPath
		merged, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}

		path := filepath.Join("testdata/loop", "project_with_plan.json")
		if err := os.WriteFile(path, merged, 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		t.Cleanup(func() { os.Remove(path) })
		return path
	}

	testCases := map[string]struct {
		ConfigPlanPath string
		Args           []string
		WantExitStatus int
	}{
		"the plan path comes from the project config": {
			ConfigPlanPath: "plan.json",
			WantExitStatus: 0,
		},
		"the positional plan path overrides the project config": {
			ConfigPlanPath: "no_such_plan.json",
			Args:           []string{"testdata/loop/plan.json"},
			WantExitStatus: 0,
		},
		"a missing plan path in the project config is an error": {
			ConfigPlanPath: "no_such_plan.json",
			WantExitStatus: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := append([]string{"-f", writeConfig(t, tc.ConfigPlanPath)}, tc.Args...)
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != tc.WantExitStatus {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want %d", exitStatus, tc.WantExitStatus)
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
		OutFormat       string
		EmphasisTSV     string
		WantExitStatus  int
		WantStdout      []string
		WantNotInStdout []string
		WantStderr      []string
	}{
		"mermaid では crit タグが付く": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "ID\nP1\n",
			WantExitStatus: 0,
			WantStdout:     []string{"P1[0] R1 :crit, "},
		},
		"plantuml では色が付く": {
			OutFormat:      "plantuml",
			EmphasisTSV:    "ID\nP1\n",
			WantExitStatus: 0,
			WantStdout:     []string{"] is colored in Salmon"},
		},
		"google-spreadsheet-tsv では Emphasis 列が付く": {
			OutFormat:      "google-spreadsheet-tsv",
			EmphasisTSV:    "ID\nP1\n",
			WantExitStatus: 0,
			WantStdout:     []string{"\tEmphasis\n", "\tTRUE\n"},
		},
		"plan-json では強調できないのでエラーになる": {
			OutFormat:      "plan-json",
			EmphasisTSV:    "ID\nP1\n",
			WantExitStatus: 1,
		},
		"timeline-json では強調できないのでエラーになる": {
			OutFormat:      "timeline-json",
			EmphasisTSV:    "ID\nP1\n",
			WantExitStatus: 1,
		},
		"ID 列がない表はエラーになる": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "Key\nP1\n",
			WantExitStatus: 1,
		},
		"実行計画にない ID は警告して無視する": {
			OutFormat:       "mermaid",
			EmphasisTSV:     "ID\nP1\nP99\n",
			WantExitStatus:  0,
			WantStdout:      []string{"P1[0] R1 :crit, "},
			WantNotInStdout: []string{"P99"},
			WantStderr:      []string{"P99"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := []string{
				"-f", "testdata/loop/config.json",
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
			for _, notWant := range tc.WantNotInStdout {
				if strings.Contains(spy.Stdout.String(), notWant) {
					t.Errorf("stdout = %q, want not to contain %q", spy.Stdout.String(), notWant)
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

func TestMainCommandByArgs_EmphasisTSVFromProjectConfig(t *testing.T) {
	writeConfigWithEmphasis := func(t *testing.T) string {
		t.Helper()
		raw, err := os.ReadFile("testdata/loop/config.json")
		if err != nil {
			t.Fatalf("read config: %v", err)
		}
		var config map[string]any
		if err := json.Unmarshal(raw, &config); err != nil {
			t.Fatalf("unmarshal config: %v", err)
		}
		config["plan"] = "plan.json"
		config["emphasis_tsv_path"] = "em.tsv"
		merged, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}

		emPath := filepath.Join("testdata/loop", "em.tsv")
		if err := os.WriteFile(emPath, []byte("ID\nP1\n"), 0644); err != nil {
			t.Fatalf("write emphasis tsv: %v", err)
		}
		t.Cleanup(func() { os.Remove(emPath) })
		path := filepath.Join("testdata/loop", "project_with_emphasis.json")
		if err := os.WriteFile(path, merged, 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		t.Cleanup(func() { os.Remove(path) })
		return path
	}

	testCases := map[string]struct {
		OutFormat       string
		WantExitStatus  int
		WantStdout      []string
		WantNotInStdout []string
	}{
		"mermaid では構成ファイルの強調 ID 表が使われる": {
			OutFormat:      "mermaid",
			WantExitStatus: 0,
			WantStdout:     []string{"P1[0] R1 :crit, "},
		},
		"plan-json では強調できないが、構成ファイル由来なら単に使われない": {
			OutFormat:       "plan-json",
			WantExitStatus:  0,
			WantNotInStdout: []string{"crit"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := []string{"-f", writeConfigWithEmphasis(t), "-out-format", tc.OutFormat}
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
			for _, notWant := range tc.WantNotInStdout {
				if strings.Contains(spy.Stdout.String(), notWant) {
					t.Errorf("stdout = %q, want not to contain %q", spy.Stdout.String(), notWant)
				}
			}
		})
	}
}

func TestMainCommandByArgs_PNGInput(t *testing.T) {
	pngPath := pngtest.WriteTempPNG(t, "testdata/loop/pfd.drawio")
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs([]string{"-f", "testdata/loop/config.json", "-p", pngPath, "testdata/loop/plan.json"}, spy.NewProcInout())
	if exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Log(spy.Stdout.String())
		t.Errorf("exitStatus = %d, want 0", exitStatus)
	}
}

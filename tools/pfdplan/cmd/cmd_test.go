package cmd

import (
	"encoding/json"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("-poor", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-poor"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-poor -random-seed is reproducible", func(t *testing.T) {
		args := []string{"-f", "testdata/simple/config.json", "-poor", "-random-seed", "42"}

		spy1 := cli.SpyProcInout()
		if exitStatus := MainCommandByArgs(args, spy1.NewProcInout()); exitStatus != 0 {
			t.Log(spy1.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		spy2 := cli.SpyProcInout()
		if exitStatus := MainCommandByArgs(args, spy2.NewProcInout()); exitStatus != 0 {
			t.Log(spy2.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		if spy1.Stdout.String() != spy2.Stdout.String() {
			t.Errorf("same -random-seed produced different output:\nfirst:\n%s\nsecond:\n%s", spy1.Stdout.String(), spy2.Stdout.String())
		}
	})
	t.Run("-poorest", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-poorest"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-poorest is deterministic", func(t *testing.T) {
		args := []string{"-f", "testdata/simple/config.json", "-poorest"}

		spy1 := cli.SpyProcInout()
		if exitStatus := MainCommandByArgs(args, spy1.NewProcInout()); exitStatus != 0 {
			t.Log(spy1.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		spy2 := cli.SpyProcInout()
		if exitStatus := MainCommandByArgs(args, spy2.NewProcInout()); exitStatus != 0 {
			t.Log(spy2.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}

		if spy1.Stdout.String() != spy2.Stdout.String() {
			t.Errorf("-poorest produced different output across runs:\nfirst:\n%s\nsecond:\n%s", spy1.Stdout.String(), spy2.Stdout.String())
		}
	})
	t.Run("-better", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-better", "-quality", "small"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-best", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-best"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-out-format mermaid with -start", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "mermaid", "-start", "2025-01-06", "-poor"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
		stdout := spy.Stdout.String()
		if !strings.HasPrefix(stdout, "gantt\n") {
			t.Errorf("stdout does not start with gantt header: %q", stdout[:min(len(stdout), 100)])
		}
		if !strings.Contains(stdout, "dateFormat YYYY-MM-DD HH:mm") {
			t.Errorf("stdout does not contain dateFormat: %q", stdout[:min(len(stdout), 200)])
		}
		if !strings.Contains(stdout, "section") {
			t.Errorf("stdout does not contain section: %q", stdout[:min(len(stdout), 200)])
		}
	})
	t.Run("-out-format plantuml with -start", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plantuml", "-start", "2025-01-06", "-poor"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
		stdout := spy.Stdout.String()
		if !strings.HasPrefix(stdout, "@startgantt\n") {
			t.Errorf("stdout does not start with @startgantt: %q", stdout[:min(len(stdout), 100)])
		}
		if !strings.Contains(stdout, "@endgantt") {
			t.Errorf("stdout does not contain @endgantt: %q", stdout[:min(len(stdout), 200)])
		}
		if !strings.Contains(stdout, "-- P1") {
			t.Errorf("stdout does not contain section separator: %q", stdout[:min(len(stdout), 200)])
		}
	})
	t.Run("-out-format plan-json without business-time flags should succeed", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plan-json", "-poor"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
	t.Run("-out-format plan-json with -start should fail", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plan-json", "-start", "2025-01-01", "-poor"}, spy.NewProcInout())
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
	})
	t.Run("-out-format plan-json with -start-time should fail", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plan-json", "-start-time", "09:00", "-poor"}, spy.NewProcInout())
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
	})
	t.Run("-out-format plan-json with -duration should fail", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plan-json", "-duration", "8", "-poor"}, spy.NewProcInout())
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
	})
	t.Run("-out-format plan-json with -weekdays should fail", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plan-json", "-weekdays", "mon,tue,wed", "-poor"}, spy.NewProcInout())
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
	})
	t.Run("-out-format plan-json with -not-biz-days should fail", func(t *testing.T) {
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-out-format", "plan-json", "-not-biz-days", "dummy.txt", "-poor"}, spy.NewProcInout())
		if exitStatus != 1 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 1", exitStatus)
		}
	})
	t.Run("-cpuprofile and -memprofile write non-empty profiles", func(t *testing.T) {
		dir := t.TempDir()
		cpuPath := filepath.Join(dir, "cpu.pprof")
		memPath := filepath.Join(dir, "mem.pprof")
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-poor", "-cpuprofile", cpuPath, "-memprofile", memPath}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Fatalf("exitStatus = %d, want 0", exitStatus)
		}
		for _, p := range []string{cpuPath, memPath} {
			info, err := os.Stat(p)
			if err != nil {
				t.Fatalf("stat %q: %v", p, err)
			}
			if info.Size() == 0 {
				t.Fatalf("profile %q is empty", p)
			}
		}
	})
	t.Run("png input", func(t *testing.T) {
		pngPath := pngtest.WriteTempPNG(t, "testdata/simple/pfd.drawio")
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-f", "testdata/simple/config.json", "-p", pngPath, "-poor"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})
}

func TestMainCommandByArgs_CompositeDeliverable(t *testing.T) {
	testCases := map[string]struct {
		Config string
	}{

		"atomic process outputs a composite deliverable": {Config: "testdata/compdeliv/config.json"},

		"nested composite deliverable": {Config: "testdata/compdeliv_nested/config.json"},

		"composite deliverable used only as a member": {Config: "testdata/compdeliv_member_only/config.json"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-f", tc.Config, "-best"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	}
}

func TestMainCommandByArgs_ExecModel(t *testing.T) {
	branchTables := []string{
		"-p", "testdata/branch/pfd.drawio",
		"-ap", "testdata/branch/atomic_proc.tsv",
		"-ad", "testdata/branch/deliv.tsv",
		"-cd", "testdata/branch/comp_deliv.tsv",
	}

	testCases := map[string]struct {
		Args             []string
		WantExitStatus   int
		ExpectedMentions []string
	}{
		"default model needs neither the resource table nor the rework columns": {
			Args:           append(append([]string{}, branchTables...), "-poor"),
			WantExitStatus: 0,
		},
		"infinite resources ignore an explicit resource table": {
			Args:           append(append([]string{}, branchTables...), "-r", "testdata/branch/resource.tsv", "-res", "infinite", "-poor"),
			WantExitStatus: 0,
		},
		"finite resources require the resource table": {
			Args:             append(append([]string{}, branchTables...), "-res", "finite", "-poor"),
			WantExitStatus:   1,
			ExpectedMentions: []string{"resource"},
		},
		"finite resources with the resource table": {
			Args:           append(append([]string{}, branchTables...), "-r", "testdata/branch/resource.tsv", "-res", "finite", "-fb", "enabled", "-poor"),
			WantExitStatus: 0,
		},
		"feedback edges are rejected when feedback is disabled": {
			Args:           []string{"-f", "testdata/simple/config.json", "-fb", "disabled", "-poor"},
			WantExitStatus: 1,

			ExpectedMentions: []string{"feedback-edge-not-available", "PFD[D2, P1]"},
		},
		"feedback edges are allowed when feedback is enabled": {
			Args:           []string{"-f", "testdata/simple/config.json", "-fb", "enabled", "-poor"},
			WantExitStatus: 0,
		},
		"unknown resource mode is an error": {
			Args:             append(append([]string{}, branchTables...), "-res", "unlimited", "-poor"),
			WantExitStatus:   1,
			ExpectedMentions: []string{"unknown resource mode"},
		},
		"unknown feedback mode is an error": {
			Args:             append(append([]string{}, branchTables...), "-fb", "on", "-poor"),
			WantExitStatus:   1,
			ExpectedMentions: []string{"unknown feedback mode"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.Args, spy.NewProcInout())
			if exitStatus != tc.WantExitStatus {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Fatalf("exitStatus = %d, want %d", exitStatus, tc.WantExitStatus)
			}
			for _, mention := range tc.ExpectedMentions {
				if !strings.Contains(spy.Stderr.String(), mention) {
					t.Errorf("expected stderr to mention %q, but got:\n%s", mention, spy.Stderr.String())
				}
			}
		})
	}
}

func TestMainCommandByArgs_ExecModelLarge(t *testing.T) {
	testCases := map[string]struct {
		Config               string
		AtomicProcessTable   string
		ExpectedProcessCount int
	}{
		"finite resources with feedback edges": {
			Config:               "testdata/large/project.json",
			AtomicProcessTable:   "testdata/large/ap.tsv",
			ExpectedProcessCount: 22,
		},
		"infinite resources with feedback edges": {
			Config:               "testdata/large/project_infinite.json",
			AtomicProcessTable:   "testdata/large/ap_infinite.tsv",
			ExpectedProcessCount: 22,
		},
		"finite resources without feedback edges": {
			Config:               "testdata/large_nofb/project_finite.json",
			AtomicProcessTable:   "testdata/large_nofb/ap_finite.tsv",
			ExpectedProcessCount: 22,
		},
		"infinite resources without feedback edges": {
			Config:               "testdata/large_nofb/project_infinite.json",
			AtomicProcessTable:   "testdata/large_nofb/ap_infinite.tsv",
			ExpectedProcessCount: 22,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			plan := planJSONByConfig(t, tc.Config)

			if leadtime := plan.leadtime(); leadtime <= 0 {
				t.Errorf("expected a positive leadtime, but got %v", leadtime)
			}

			expected := sortedAtomicProcessIDs(t, tc.AtomicProcessTable)

			if len(expected) != tc.ExpectedProcessCount {
				t.Fatalf("expected %d atomic processes in %q, but got %d", tc.ExpectedProcessCount, tc.AtomicProcessTable, len(expected))
			}
			if actual := plan.allocatedAtomicProcesses(); !slices.Equal(actual, expected) {
				t.Errorf("expected every atomic process to be planned\nplanned:  %v\nexpected: %v", actual, expected)
			}
		})
	}
}

func TestMainCommandByArgs_ExecModelLargeLeadtimeOrder(t *testing.T) {
	testCases := map[string]struct {
		ShorterConfig string
		LongerConfig  string
	}{
		"infinite resources are shorter than finite resources (with feedback edges)": {
			ShorterConfig: "testdata/large/project_infinite.json",
			LongerConfig:  "testdata/large/project.json",
		},
		"infinite resources are shorter than finite resources (without feedback edges)": {
			ShorterConfig: "testdata/large_nofb/project_infinite.json",
			LongerConfig:  "testdata/large_nofb/project_finite.json",
		},
		"no feedback edges is shorter than feedback edges (finite resources)": {
			ShorterConfig: "testdata/large_nofb/project_finite.json",
			LongerConfig:  "testdata/large/project.json",
		},
		"no feedback edges is shorter than feedback edges (infinite resources)": {
			ShorterConfig: "testdata/large_nofb/project_infinite.json",
			LongerConfig:  "testdata/large/project_infinite.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			shorter := planJSONByConfig(t, tc.ShorterConfig).leadtime()
			longer := planJSONByConfig(t, tc.LongerConfig).leadtime()
			if !(shorter < longer) {
				t.Errorf("expected the leadtime of %q (%v) to be shorter than %q (%v)", tc.ShorterConfig, shorter, tc.LongerConfig, longer)
			}
		})
	}
}

func planJSONByConfig(t *testing.T, config string) planJSON {
	t.Helper()
	spy := cli.SpyProcInout()
	args := []string{"-f", config, "-poorest", "-out-format", "plan-json"}
	if exitStatus := MainCommandByArgs(args, spy.NewProcInout()); exitStatus != 0 {
		t.Log(spy.Stderr.String())
		t.Fatalf("exitStatus = %d, want 0", exitStatus)
	}

	var plan planJSON
	if err := json.Unmarshal([]byte(spy.Stdout.String()), &plan); err != nil {
		t.Fatalf("unmarshal plan json: %v", err)
	}
	if len(plan.Transitions) == 0 {
		t.Fatal("expected at least 1 transition, but got none")
	}
	return plan
}

type planJSON struct {
	Transitions []struct {
		Allocation map[string]struct{} `json:"allocation"`
		NextState  struct {
			Time float64 `json:"time"`
		} `json:"next_state"`
	} `json:"transitions"`
}

func (p planJSON) leadtime() float64 {
	if len(p.Transitions) == 0 {
		return 0
	}
	return p.Transitions[len(p.Transitions)-1].NextState.Time
}

func (p planJSON) allocatedAtomicProcesses() []string {
	seen := make(map[string]struct{})
	for _, t := range p.Transitions {
		for ap := range t.Allocation {
			seen[ap] = struct{}{}
		}
	}
	res := slices.Collect(maps.Keys(seen))
	slices.Sort(res)
	return res
}

func sortedAtomicProcessIDs(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %q: %v", path, err)
	}
	defer f.Close()

	table, err := pfdtsv.ParseAtomicProcessTable(f)
	if err != nil {
		t.Fatalf("parse %q: %v", path, err)
	}

	res := make([]string, 0, len(table.Rows))
	for _, id := range table.IDs() {
		res = append(res, string(id))
	}
	slices.Sort(res)
	return res
}

func TestMainCommandByArgs_OptionResolution(t *testing.T) {
	testCases := map[string]struct {
		Args             []string
		WantExitStatus   int
		ExpectedMentions []string
	}{
		"atomic process table is required": {
			Args: []string{
				"-p", "testdata/branch/pfd.drawio",
				"-ad", "testdata/branch/deliv.tsv",
				"-cd", "testdata/branch/comp_deliv.tsv",
				"-poor",
			},
			WantExitStatus:   1,
			ExpectedMentions: []string{"-ap is required"},
		},
		"atomic deliverable table is required": {
			Args: []string{
				"-p", "testdata/branch/pfd.drawio",
				"-ap", "testdata/branch/atomic_proc.tsv",
				"-cd", "testdata/branch/comp_deliv.tsv",
				"-poor",
			},
			WantExitStatus:   1,
			ExpectedMentions: []string{"-ad is required"},
		},

		"a command line path is resolved from the current directory": {
			Args: []string{
				"-f", "testdata/branch/config.json",
				"-ap", "testdata/branch/atomic_proc.tsv",
				"-poor",
			},
			WantExitStatus: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.Args, spy.NewProcInout())
			if exitStatus != tc.WantExitStatus {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Fatalf("exitStatus = %d, want %d", exitStatus, tc.WantExitStatus)
			}
			for _, mention := range tc.ExpectedMentions {
				if !strings.Contains(spy.Stderr.String(), mention) {
					t.Errorf("expected stderr to mention %q, but got:\n%s", mention, spy.Stderr.String())
				}
			}
		})
	}
}

func TestMainCommandByArgs_PlanOutputFromProjectConfig(t *testing.T) {
	writeConfig := func(t *testing.T, extra map[string]any) string {
		t.Helper()
		raw, err := os.ReadFile("testdata/branch/config.json")
		if err != nil {
			t.Fatalf("read config: %v", err)
		}
		var config map[string]any
		if err := json.Unmarshal(raw, &config); err != nil {
			t.Fatalf("unmarshal config: %v", err)
		}

		for _, key := range []string{"pfd", "atomic_process_table", "atomic_deliverable_table", "composite_deliverable_table", "resource_table"} {
			if v, ok := config[key].(string); ok {
				abs, err := filepath.Abs(filepath.Join("testdata/branch", v))
				if err != nil {
					t.Fatalf("abs %q: %v", v, err)
				}
				config[key] = abs
			}
		}
		for k, v := range extra {
			config[k] = v
		}
		merged, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}
		path := filepath.Join(t.TempDir(), "project.json")
		if err := os.WriteFile(path, merged, 0644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		return path
	}

	testCases := map[string]struct {
		Extra            map[string]any
		ExtraArgs        []string
		WantExitStatus   int
		ExpectedStdout   []string
		ExpectedMentions []string
	}{
		"output format comes from the project config": {
			Extra:          map[string]any{"output_format": "mermaid", "start_day": "2026-06-01"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"gantt", "2026-06-01 10:00"},
		},
		"business time settings come from the project config": {
			Extra:          map[string]any{"output_format": "mermaid", "start_day": "2026-06-01", "start_time": "09:30"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"2026-06-01 09:30"},
		},
		"the command line overrides the project config": {
			Extra:          map[string]any{"output_format": "mermaid", "start_day": "2026-06-01"},
			ExtraArgs:      []string{"-out-format", "plantuml", "-start", "2026-07-01"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"@startgantt", "2026-07-01"},
		},

		"business time settings in the project config are ignored by plan-json": {
			Extra:          map[string]any{"output_format": "plan-json", "start_day": "2026-06-01"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"\"transitions\""},
		},
		"business time settings in the project config are ignored by timeline-json": {
			Extra:          map[string]any{"output_format": "timeline-json", "start_day": "2026-06-01"},
			WantExitStatus: 0,
		},
		"business time flags on the command line cannot be used with plan-json": {
			Extra:            map[string]any{"output_format": "mermaid"},
			ExtraArgs:        []string{"-out-format", "plan-json", "-start", "2026-06-01"},
			WantExitStatus:   1,
			ExpectedMentions: []string{"plan-json", "-start"},
		},
		"plan-json works without business time settings": {
			Extra:          map[string]any{"output_format": "plan-json"},
			WantExitStatus: 0,
			ExpectedStdout: []string{"\"transitions\""},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := append([]string{"-f", writeConfig(t, tc.Extra), "-poor"}, tc.ExtraArgs...)
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

func TestFixture_LargeNoFeedbackIsLargeWithoutFeedbackEdges(t *testing.T) {
	const removedCompositeDeliverable = "D3"

	large := parsePFDFixture(t, "testdata/large/pfd.drawio", "testdata/large/cd.tsv")
	noFeedback := parsePFDFixture(t, "testdata/large_nofb/pfd.drawio", "testdata/large_nofb/cd.tsv")

	expectedNodes := make([]string, 0, large.Nodes.Len())
	for _, n := range large.Nodes.Iter() {
		if string(n.ID) == removedCompositeDeliverable {
			continue
		}
		expectedNodes = append(expectedNodes, string(n.ID))
	}
	slices.Sort(expectedNodes)

	actualNodes := make([]string, 0, noFeedback.Nodes.Len())
	for _, n := range noFeedback.Nodes.Iter() {
		actualNodes = append(actualNodes, string(n.ID))
	}
	slices.Sort(actualNodes)

	if !slices.Equal(actualNodes, expectedNodes) {
		t.Errorf("node set mismatch\ntestdata/large_nofb: %v\nexpected:            %v", actualNodes, expectedNodes)
	}

	if !slices.Equal(solidEdges(t, noFeedback), solidEdges(t, large)) {
		t.Errorf("solid edge set mismatch\ntestdata/large_nofb: %v\ntestdata/large:      %v", solidEdges(t, noFeedback), solidEdges(t, large))
	}

	for _, e := range noFeedback.Edges.Iter() {
		if e.IsFeedback {
			t.Errorf("expected no feedback edge in testdata/large_nofb, but got %s -> %s", e.Source, e.Target)
		}
	}
	feedbackCount := 0
	for _, e := range large.Edges.Iter() {
		if e.IsFeedback {
			feedbackCount++
		}
	}
	if feedbackCount == 0 {
		t.Error("expected testdata/large to have feedback edges, but got none")
	}
}

func solidEdges(t *testing.T, p *pfd.PFD) []string {
	t.Helper()
	res := make([]string, 0, p.Edges.Len())
	for _, e := range p.Edges.Iter() {
		if e.IsFeedback {
			continue
		}
		res = append(res, string(e.Source)+" -> "+string(e.Target))
	}
	slices.Sort(res)
	return res
}

func parsePFDFixture(t *testing.T, pfdPath string, compositeDeliverableTablePath string) *pfd.PFD {
	t.Helper()

	cdFile, err := os.Open(compositeDeliverableTablePath)
	if err != nil {
		t.Fatalf("open %q: %v", compositeDeliverableTablePath, err)
	}
	defer cdFile.Close()

	cdTable, err := pfdtsv.ParseCompositeDeliverableTable(cdFile)
	if err != nil {
		t.Fatalf("parse %q: %v", compositeDeliverableTablePath, err)
	}

	pfdFile, err := os.Open(pfdPath)
	if err != nil {
		t.Fatalf("open %q: %v", pfdPath, err)
	}
	defer pfdFile.Close()

	p, err := pfdfmt.Parse("", pfdFile, &pfdfmt.ParseOptions{CompositeDeliverableTable: cdTable}, slog.New(slogtest.NewTestHandler(t)))
	if err != nil {
		t.Fatalf("parse %q: %v", pfdPath, err)
	}
	return p
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
	}{
		"mermaid では crit タグが付く": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "ID\nP1\n",
			WantExitStatus: 0,
			WantStdout:     []string{":crit, "},
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
		"ID 列がない表はエラーになる": {
			OutFormat:      "mermaid",
			EmphasisTSV:    "Key\nP1\n",
			WantExitStatus: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := []string{
				"-f", "testdata/simple/config.json",
				"-poor",
				"-out-format", tc.OutFormat,
				"-em-tsv", writeEmphasisTSV(t, tc.EmphasisTSV),
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
		})
	}
}

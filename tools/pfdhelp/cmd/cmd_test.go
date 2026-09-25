package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/tools/pfdhelp/help"
	"github.com/Kuniwak/pfd-tools/version"
)

func TestMainCommandByOptionsWarnsOnToolFailure(t *testing.T) {
	spy := cli.SpyProcInout()
	opts := &Options{
		CommonOptions: &tools.CommonOptions{},
		Tools: []help.ToolHelp{
			{Name: "broken", Command: func(_ []string, _ *cli.ProcInout) int { return 1 }},
		},
	}
	if err := MainCommandByOptions(opts, spy.NewProcInout()); err != nil {
		t.Fatalf("err = %v, want nil (best-effort, exit 0)", err)
	}
	if !strings.Contains(spy.Stderr.String(), "broken") {
		t.Errorf("stderr should warn about the failed tool, got %q", spy.Stderr.String())
	}
}

func TestMainCommandByArgsShowsAllTools(t *testing.T) {
	spy := cli.SpyProcInout()
	exitStatus := MainCommandByArgs(nil, spy.NewProcInout())
	if exitStatus != 0 {
		t.Fatalf("exitStatus = %d, want 0\n%s", exitStatus, spy.Stderr.String())
	}

	out := spy.Stdout.String()

	starts := make([]int, len(help.Tools))
	pos := 0
	for i, tool := range help.Tools {
		header := "# " + tool.Name + "\n"
		idx := strings.Index(out[pos:], header)
		if idx < 0 {
			t.Fatalf("header %q not found in expected order", header)
		}
		starts[i] = pos + idx
		pos = starts[i] + len(header)
	}

	for i, tool := range help.Tools {
		end := len(out)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		section := out[starts[i]:end]
		t.Run(tool.Name, func(t *testing.T) {
			if !strings.Contains(section, "Usage") {
				t.Errorf("help section for %q does not contain Usage\n--- section ---\n%s", tool.Name, section)
			}
		})
	}
}

func TestMainCommandByArgs(t *testing.T) {
	testCases := map[string]struct {
		args              []string
		wantExit          int
		stdoutContains    []string
		stdoutNotContains []string
		stderrContains    []string
		stdoutEquals      string
	}{
		"filter shows only specified tools": {
			args:              []string{"pfdlint", "pfdplan"},
			wantExit:          0,
			stdoutContains:    []string{"# pfdlint\n", "# pfdplan\n"},
			stdoutNotContains: []string{"# bizday\n"},
		},
		"unknown tool errors": {
			args:           []string{"nope"},
			wantExit:       1,
			stderrContains: []string{"nope"},
		},
		"help shows own usage": {
			args:           []string{"-h"},
			wantExit:       0,
			stderrContains: []string{"Usage: pfdhelp"},
		},
		"version shows version": {
			args:         []string{"-version"},
			wantExit:     0,
			stdoutEquals: version.Version,
		},
		"short lists all tools as tsv": {
			args:              []string{"-short"},
			wantExit:          0,
			stdoutContains:    []string{"bizday\t", "pfdlint\t", "pfdplan\t"},
			stdoutNotContains: []string{"Usage:", "# pfdlint"},
		},
		"short with filter shows only specified tools": {
			args:              []string{"-short", "pfdlint", "pfdplan"},
			wantExit:          0,
			stdoutContains:    []string{"pfdlint\t", "pfdplan\t"},
			stdoutNotContains: []string{"bizday\t", "# "},
		},
		"short-help shows own description": {
			args:         []string{"--short-help"},
			wantExit:     0,
			stdoutEquals: ShortHelp,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs(tc.args, spy.NewProcInout())
			if exitStatus != tc.wantExit {
				t.Fatalf("exitStatus = %d, want %d\n%s", exitStatus, tc.wantExit, spy.Stderr.String())
			}

			stdout := spy.Stdout.String()
			for _, want := range tc.stdoutContains {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout does not contain %q", want)
				}
			}
			for _, notWant := range tc.stdoutNotContains {
				if strings.Contains(stdout, notWant) {
					t.Errorf("stdout should not contain %q", notWant)
				}
			}
			for _, want := range tc.stderrContains {
				if !strings.Contains(spy.Stderr.String(), want) {
					t.Errorf("stderr does not contain %q, got %q", want, spy.Stderr.String())
				}
			}
			if tc.stdoutEquals != "" {
				if got := strings.TrimSpace(stdout); got != tc.stdoutEquals {
					t.Errorf("stdout = %q, want %q", got, tc.stdoutEquals)
				}
			}
		})
	}
}

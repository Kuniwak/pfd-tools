package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestMainCommandByArgs(t *testing.T) {
	baseArgs := []string{"-p", "testdata/loop/plan.json", "-ap", "testdata/loop/atomic_proc.tsv", "-m", "testdata/loop/milestone.tsv", "-g", "testdata/loop/group.tsv", "-b", "1.5"}

	cases := []struct {
		name           string
		extraArgs      []string
		wantExitStatus int
		checkStdout    func(t *testing.T, stdout string)
	}{
		{
			name:           "DefaultFormat",
			extraArgs:      nil,
			wantExitStatus: 0,
			checkStdout: func(t *testing.T, stdout string) {
				if !strings.HasPrefix(stdout, "Group\t") {
					t.Errorf("expected TSV header, got: %q", stdout[:min(len(stdout), 50)])
				}
			},
		},
		{
			name:           "ExplicitGoogleSpreadsheetTSV",
			extraArgs:      []string{"-out-format", "google-spreadsheet-tsv"},
			wantExitStatus: 0,
			checkStdout: func(t *testing.T, stdout string) {
				if !strings.HasPrefix(stdout, "Group\t") {
					t.Errorf("expected TSV header, got: %q", stdout[:min(len(stdout), 50)])
				}
			},
		},
		{
			name:           "MermaidFormat",
			extraArgs:      []string{"-out-format", "mermaid"},
			wantExitStatus: 0,
			checkStdout: func(t *testing.T, stdout string) {
				if !strings.HasPrefix(stdout, "gantt\n") {
					t.Errorf("expected mermaid gantt header, got: %q", stdout[:min(len(stdout), 50)])
				}
				if !strings.Contains(stdout, "section G1") {
					t.Errorf("expected section G1 in output, got: %q", stdout)
				}
			},
		},
		{
			name:           "PlantUMLFormat",
			extraArgs:      []string{"-out-format", "plantuml"},
			wantExitStatus: 0,
			checkStdout: func(t *testing.T, stdout string) {
				if !strings.HasPrefix(stdout, "@startgantt\n") {
					t.Errorf("expected @startgantt header, got: %q", stdout[:min(len(stdout), 50)])
				}
				if !strings.Contains(stdout, "@endgantt") {
					t.Errorf("expected @endgantt in output, got: %q", stdout)
				}
				if !strings.Contains(stdout, "-- G1") {
					t.Errorf("expected section separator -- G1 in output, got: %q", stdout)
				}
			},
		},
		{
			name:           "InvalidFormat",
			extraArgs:      []string{"-out-format", "xml"},
			wantExitStatus: 1,
			checkStdout:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			args := append(baseArgs, tc.extraArgs...)
			exitStatus := MainCommandByArgs(args, spy.NewProcInout())
			if exitStatus != tc.wantExitStatus {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want %d", exitStatus, tc.wantExitStatus)
			}
			if tc.checkStdout != nil {
				tc.checkStdout(t, spy.Stdout.String())
			}
		})
	}
}

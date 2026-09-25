package help

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
)

func TestWriteHelpReportsFailedTools(t *testing.T) {
	tools := []ToolHelp{
		{Name: "ok", Command: func(_ []string, inout *cli.ProcInout) int {
			fmt.Fprintln(inout.Stdout, "Usage: ok")
			return 0
		}},
		{Name: "broken", Command: func(_ []string, _ *cli.ProcInout) int {
			return 1
		}},
	}
	var buf bytes.Buffer
	failed := WriteHelp(&buf, tools)
	if len(failed) != 1 || failed[0] != "broken" {
		t.Errorf("failed = %v, want [broken]", failed)
	}
	for _, header := range []string{"# ok\n", "# broken\n"} {
		if !strings.Contains(buf.String(), header) {
			t.Errorf("output does not contain header %q\n%s", header, buf.String())
		}
	}
}

func TestWriteShortHelp(t *testing.T) {
	testCases := map[string]struct {
		tools []ToolHelp
		want  string
	}{
		"plain descriptions": {
			tools: []ToolHelp{{Name: "bizday", Short: "説明1"}, {Name: "pfdlint", Short: "説明2"}},
			want:  "bizday\t説明1\npfdlint\t説明2\n",
		},
		"tab in description is quoted so columns do not shift": {
			tools: []ToolHelp{{Name: "x", Short: "a\tb"}},
			want:  "x\t\"a\tb\"\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := WriteShortHelp(&buf, tc.tools); err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if got := buf.String(); got != tc.want {
				t.Errorf("output = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAllToolsSupportShortHelp(t *testing.T) {
	for _, tool := range Tools {
		t.Run(tool.Name, func(t *testing.T) {
			spy := cli.SpyProcInout()
			code := tool.Command([]string{"--short-help"}, spy.NewProcInout())
			if code != 0 {
				t.Fatalf("exit = %d, want 0 (stderr=%q)", code, spy.Stderr.String())
			}
			out := spy.Stdout.String()
			if strings.TrimSpace(out) == "" {
				t.Errorf("stdout is empty, want a one-line description")
			}
			if strings.Contains(out, "Usage:") {
				t.Errorf("stdout should be the short description, not full help; got %q", out)
			}
			if n := strings.Count(strings.TrimRight(out, "\n"), "\n"); n != 0 {
				t.Errorf("stdout should be a single line, got %d extra newline(s): %q", n, out)
			}
		})
	}
}

func TestToolsRegistryMatchesFilesystem(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	toolsRoot := filepath.Join(wd, "..", "..")

	want := make(map[string]struct{})
	err = filepath.WalkDir(toolsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "main.go" {
			return nil
		}
		name := filepath.Base(filepath.Dir(path))
		if name == "pfdhelp" {
			return nil
		}
		want[name] = struct{}{}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	got := make(map[string]struct{}, len(Tools))
	for _, tool := range Tools {
		got[tool.Name] = struct{}{}
	}

	for name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("tool %q has tools/**/main.go but is missing from Tools registry", name)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("tool %q is in Tools registry but has no tools/**/main.go", name)
		}
	}
}

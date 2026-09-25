package tools

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type validateTablePathFunc func(shortPath *string, path *string, basePath string) (io.Reader, string, error)

func TestValidatePathOptionsResolvesPath(t *testing.T) {
	validators := map[string]validateTablePathFunc{
		"ValidatePFDOptions":                       ValidatePFDOptions,
		"ValidateAtomicProcessTableOptions":        ValidateAtomicProcessTableOptions,
		"ValidateAtomicDeliverableTableOptions":    ValidateAtomicDeliverableTableOptions,
		"ValidateResourceTableOptions":             ValidateResourceTableOptions,
		"ValidateCompositeProcessTableOptions":     ValidateCompositeProcessTableOptions,
		"ValidateCompositeDeliverableTableOptions": ValidateCompositeDeliverableTableOptions,
		"ValidateMilestoneTableOptions":            ValidateMilestoneTableOptions,
		"ValidateGroupTableOptions":                ValidateGroupTableOptions,
	}

	for name, validate := range validators {
		t.Run(name+"/absolute path is used as-is", func(t *testing.T) {

			fileDir := t.TempDir()
			absPath := filepath.Join(fileDir, "table.tsv")
			if err := os.WriteFile(absPath, []byte("x"), 0644); err != nil {
				t.Fatal(err)
			}
			basePath := t.TempDir()

			shortPath := ""
			longPath := absPath
			r, resolved, err := validate(&shortPath, &longPath, basePath)
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			closeReader(r)
			if resolved != absPath {
				t.Errorf("resolved = %q, want %q (basePath must not be prepended to an absolute path)", resolved, absPath)
			}
		})

		t.Run(name+"/relative path is joined with basePath", func(t *testing.T) {
			basePath := t.TempDir()
			relPath := "table.tsv"
			if err := os.WriteFile(filepath.Join(basePath, relPath), []byte("x"), 0644); err != nil {
				t.Fatal(err)
			}

			shortPath := ""
			longPath := relPath
			r, resolved, err := validate(&shortPath, &longPath, basePath)
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			closeReader(r)
			if want := filepath.Join(basePath, relPath); resolved != want {
				t.Errorf("resolved = %q, want %q", resolved, want)
			}
		})
	}
}

func closeReader(r io.Reader) {
	if c, ok := r.(io.Closer); ok {
		c.Close()
	}
}

func TestPrintUsageHeaderContainsParts(t *testing.T) {
	testCases := map[string]struct {
		usageLine string
		shortHelp string
	}{
		"usage with options and arg": {"Usage: example [options] <arg>", "これは例のツールです。"},
		"usage without options":      {"Usage: example", "短い説明。"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			flags := flag.NewFlagSet("example", flag.ContinueOnError)
			flags.SetOutput(&buf)
			flags.Bool("foo", false, "foo flag")

			PrintUsageHeader(flags, tc.usageLine, tc.shortHelp)

			got := buf.String()
			for _, want := range []string{tc.usageLine, tc.shortHelp, "Options", "-foo", "foo flag"} {
				if !strings.Contains(got, want) {
					t.Errorf("output does not contain %q\n%s", want, got)
				}
			}
		})
	}
}

func TestReadProjectConfigResolvesPlanPath(t *testing.T) {
	dir := t.TempDir()

	testCases := map[string]struct {
		PlanPath string
		Want     string
	}{
		"a relative path is joined with the config directory": {
			PlanPath: "plan.json",
			Want:     filepath.Join(dir, "plan.json"),
		},
		"an absolute path is used as-is": {
			PlanPath: filepath.Join(dir, "elsewhere", "plan.json"),
			Want:     filepath.Join(dir, "elsewhere", "plan.json"),
		},
		"an empty path stays empty": {
			PlanPath: "",
			Want:     "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			configPath := filepath.Join(dir, "project.json")
			if err := os.WriteFile(configPath, []byte(`{"plan": `+strconv.Quote(tc.PlanPath)+`}`), 0644); err != nil {
				t.Fatalf("write config: %v", err)
			}

			var configShortPath, configLongPath = configPath, ""
			config, _, err := ReadProjectConfig(&configShortPath, &configLongPath, FSMRawOptions{}, PlanOutputFormatRawOptions{}, dir, NoFlags())
			if err != nil {
				t.Fatalf("ReadProjectConfig: %v", err)
			}
			if config.PlanPath != tc.Want {
				t.Errorf("PlanPath = %q, want %q", config.PlanPath, tc.Want)
			}
		})
	}
}

func TestParsePageList(t *testing.T) {
	testCases := map[string]struct {
		In   string
		Want []string
	}{
		"empty is all pages (nil)":     {In: "", Want: nil},
		"single":                       {In: "P123", Want: []string{"P123"}},
		"comma separated":              {In: "P123,P456", Want: []string{"P123", "P456"}},
		"trims spaces and drops empty": {In: " P1 , ,P2 ", Want: []string{"P1", "P2"}},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := ParsePageList(tc.In)
			if len(got) != len(tc.Want) {
				t.Fatalf("ParsePageList(%q) = %v, want %v", tc.In, got, tc.Want)
			}
			for i := range got {
				if got[i] != tc.Want[i] {
					t.Errorf("ParsePageList(%q)[%d] = %q, want %q", tc.In, i, got[i], tc.Want[i])
				}
			}
		})
	}
}

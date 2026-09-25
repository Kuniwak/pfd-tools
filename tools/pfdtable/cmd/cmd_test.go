package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

func absTestdataPath(t *testing.T, path string) string {
	t.Helper()
	res, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %q: %v", path, err)
	}
	return res
}

func TestMainCommandByArgs(t *testing.T) {
	t.Run("atomic process", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-f supplies the existing table from the project config", func(t *testing.T) {
			spy1 := cli.SpyProcInout()
			MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy1.NewProcInout())

			spy2 := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", "testdata/loop/project.json"}, spy2.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy2.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if spy1.Stdout.String() != spy2.Stdout.String() {
				t.Errorf("-f output differs from -existing output\n-existing: %q\n-f: %q", spy1.Stdout.String(), spy2.Stdout.String())
			}
		})
		t.Run("-existing overrides the project config", func(t *testing.T) {

			const distinctDescription = "既存表から引き継いだ説明"
			raw, err := os.ReadFile("testdata/loop/atomic_proc.tsv")
			if err != nil {
				t.Fatal(err)
			}
			existingPath := filepath.Join(t.TempDir(), "atomic_proc.tsv")
			if err := os.WriteFile(existingPath, []byte(strings.Replace(string(raw), "プロセス", distinctDescription, 1)), 0644); err != nil {
				t.Fatal(err)
			}

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", "testdata/loop/project.json", "-existing", existingPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if !strings.Contains(spy.Stdout.String(), distinctDescription) {
				t.Errorf("-existing did not override the project config, got: %q", spy.Stdout.String())
			}
		})
		t.Run("-existing works even when the project config has no entry for the table", func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "project.json")

			if err := os.WriteFile(configPath, []byte(`{"pfd": "`+absTestdataPath(t, "testdata/loop/pfd.drawio")+`"}`), 0644); err != nil {
				t.Fatal(err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", configPath, "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if !strings.Contains(spy.Stdout.String(), "MasterBar") {
				t.Errorf("expected the existing atomic process table to be used, got: %q", spy.Stdout.String())
			}
		})
		t.Run("-f with a table instead of a project config is an error", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-f", "testdata/loop/atomic_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 1 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 1", exitStatus)
			}
		})
	})
	t.Run("atomic deliverable", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ad", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ad", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
	t.Run("composite process", func(t *testing.T) {
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cp", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cp", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/comp_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
	t.Run("milestone", func(t *testing.T) {
		t.Run("header only without ap", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "m"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if !strings.Contains(spy.Stdout.String(), "ID\tDescription\tGroups\tSuccessors") {
				t.Errorf("expected header row, got: %q", spy.Stdout.String())
			}
		})
		t.Run("from ap with milestone column", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "m", "-ap", "testdata/loop/atomic_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			out := spy.Stdout.String()
			if !strings.Contains(out, "M1") {
				t.Errorf("expected M1 in output, got: %q", out)
			}
			if !strings.Contains(out, "M2") {
				t.Errorf("expected M2 in output, got: %q", out)
			}
		})
	})
	t.Run("group", func(t *testing.T) {
		t.Run("header only without ap", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "g"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if !strings.Contains(spy.Stdout.String(), "ID\tDescription") {
				t.Errorf("expected header row, got: %q", spy.Stdout.String())
			}
		})
		t.Run("from ap with group column", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "g", "-ap", "testdata/loop/atomic_proc.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if !strings.Contains(spy.Stdout.String(), "G1") {
				t.Errorf("expected G1 in output, got: %q", spy.Stdout.String())
			}
		})
	})
	t.Run("all", func(t *testing.T) {
		t.Run("with pfd and ap", func(t *testing.T) {
			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{
				"-t", "all",
				"-p", "testdata/loop/pfd.drawio",
				"-cd", "testdata/loop/comp_deliv.tsv",
				"-ap", "testdata/loop/atomic_proc.tsv",
				"-out-dir", outDir,
				"-res", "finite",
				"-fb", "enabled",
			}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			for _, name := range []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv", "project.json"} {
				path := filepath.Join(outDir, name)
				info, err := os.Stat(path)
				if err != nil {
					t.Errorf("expected file %s to exist: %v", name, err)
					continue
				}
				if info.Size() == 0 {
					t.Errorf("expected file %s to be non-empty", name)
				}
			}
			projectJSON, err := os.ReadFile(filepath.Join(outDir, "project.json"))
			if err != nil {
				t.Fatalf("failed to read project.json: %v", err)
			}
			content := string(projectJSON)
			if !strings.Contains(content, "pfd.drawio") {
				t.Errorf("project.json should contain pfd path, got: %s", content)
			}
			if !strings.Contains(content, "ap.tsv") {
				t.Errorf("project.json should contain ap.tsv, got: %s", content)
			}
		})
		t.Run("with pfd no cd", func(t *testing.T) {
			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{
				"-t", "all",
				"-p", "testdata/loop/pfd.drawio",
				"-out-dir", outDir,
				"-res", "finite",
				"-fb", "enabled",
			}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			for _, name := range []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv", "project.json"} {
				path := filepath.Join(outDir, name)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("expected file %s to exist: %v", name, err)
				}
			}
			cdContent, err := os.ReadFile(filepath.Join(outDir, "cd.tsv"))
			if err != nil {
				t.Fatalf("failed to read cd.tsv: %v", err)
			}
			if !strings.Contains(string(cdContent), "ID\tDescription\tDeliverables") {
				t.Errorf("expected cd.tsv to contain header row, got: %q", string(cdContent))
			}
		})
		t.Run("-f project.json -out-dir", func(t *testing.T) {

			srcDir := t.TempDir()
			spy0 := cli.SpyProcInout()
			MainCommandByArgs([]string{
				"-t", "all",
				"-p", "testdata/loop/pfd.drawio",
				"-cd", "testdata/loop/comp_deliv.tsv",
				"-ap", "testdata/loop/atomic_proc.tsv",
				"-out-dir", srcDir,
				"-res", "finite",
				"-fb", "enabled",
			}, spy0.NewProcInout())

			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{
				"-t", "all",
				"-f", filepath.Join(srcDir, "project.json"),
				"-out-dir", outDir,
			}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			for _, name := range []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv", "project.json"} {
				if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
					t.Errorf("expected %s to exist in out-dir: %v", name, err)
				}
			}

			srcProjRaw, _ := os.ReadFile(filepath.Join(srcDir, "project.json"))
			outProjRaw, _ := os.ReadFile(filepath.Join(outDir, "project.json"))

			if len(outProjRaw) == 0 {
				t.Errorf("outDir/project.json is empty")
			}
			_ = srcProjRaw
		})
		t.Run("-f project.json -inplace", func(t *testing.T) {

			srcDir := t.TempDir()
			spy0 := cli.SpyProcInout()
			MainCommandByArgs([]string{
				"-t", "all",
				"-p", "testdata/loop/pfd.drawio",
				"-cd", "testdata/loop/comp_deliv.tsv",
				"-ap", "testdata/loop/atomic_proc.tsv",
				"-out-dir", srcDir,
				"-res", "finite",
				"-fb", "enabled",
			}, spy0.NewProcInout())

			projJSONPath := filepath.Join(srcDir, "project.json")
			origProjRaw, _ := os.ReadFile(projJSONPath)

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{
				"-t", "all",
				"-f", projJSONPath,
				"-inplace",
			}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}

			afterProjRaw, _ := os.ReadFile(projJSONPath)
			if string(origProjRaw) != string(afterProjRaw) {
				t.Errorf("project.json was modified by -t all -inplace (should not be)")
			}

			for _, name := range []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv"} {
				if _, err := os.Stat(filepath.Join(srcDir, name)); err != nil {
					t.Errorf("expected %s to still exist after inplace: %v", name, err)
				}
			}
		})
		t.Run("-f project.json without -out-dir or -inplace is error", func(t *testing.T) {
			projJSON := "testdata/loop/project.json"
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "all", "-f", projJSON}, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("expected non-zero exit for -t all -f without -out-dir or -inplace")
			}
		})
		t.Run("-f project.json -out-dir and -inplace together is error", func(t *testing.T) {
			projJSON := "testdata/loop/project.json"
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "all", "-f", projJSON, "-out-dir", t.TempDir(), "-inplace"}, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("expected non-zero exit for -t all -f -out-dir -inplace (mutually exclusive)")
			}
		})
	})
	t.Run("composite deliverable", func(t *testing.T) {
		t.Run("header only without pfd", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if !strings.Contains(spy.Stdout.String(), "ID\tDescription\tDeliverables") {
				t.Errorf("expected header row, got: %q", spy.Stdout.String())
			}
		})
		t.Run("no -existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
		t.Run("-existing", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Log(spy.Stdout.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})
	})
	t.Run("png input", func(t *testing.T) {
		pngPath := pngtest.WriteTempPNG(t, "testdata/loop/pfd.drawio")
		spy := cli.SpyProcInout()
		exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", pngPath, "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
		if exitStatus != 0 {
			t.Log(spy.Stderr.String())
			t.Log(spy.Stdout.String())
			t.Errorf("exitStatus = %d, want 0", exitStatus)
		}
	})

	t.Run("mode bootstrap", func(t *testing.T) {
		cases := []struct {
			name    string
			tType   string
			model   []string
			wantSub string
		}{

			{"ap-plan (default model)", "ap-plan", nil, "ID\tDescription\t予想作業量\t開始条件"},
			{"ad-plan (default model)", "ad-plan", nil, "ID\tDescription\t利用可能時刻"},
			{"ap-plan", "ap-plan", []string{"-res", "finite", "-fb", "enabled"}, "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件"},
			{"ap-plan-master", "ap-plan-master", []string{"-res", "finite", "-fb", "enabled"}, "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tMasterRow\tMasterBar"},
			{"ad-plan", "ad-plan", []string{"-res", "finite", "-fb", "enabled"}, "ID\tDescription\t利用可能時刻\t最大版"},
			{"ad-plan-master", "ad-plan-master", []string{"-res", "finite", "-fb", "enabled"}, "ID\tDescription\t利用可能時刻\t最大版"},
			{"ap-plan with infinite resources only", "ap-plan", []string{"-res", "infinite", "-fb", "enabled"}, "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t開始条件"},
			{"ap-plan with feedback disabled only", "ap-plan", []string{"-res", "finite", "-fb", "disabled"}, "ID\tDescription\t予想作業量\t必要資源\t開始条件"},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				spy := cli.SpyProcInout()
				args := append([]string{"-t", c.tType, "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, c.model...)
				exitStatus := MainCommandByArgs(args, spy.NewProcInout())
				if exitStatus != 0 {
					t.Log(spy.Stderr.String())
					t.Log(spy.Stdout.String())
					t.Errorf("exitStatus = %d, want 0", exitStatus)
				}
				if !strings.Contains(spy.Stdout.String(), c.wantSub) {
					t.Errorf("expected header row %q, got: %q", c.wantSub, spy.Stdout.String())
				}
			})
		}
		t.Run("ap-plan/en", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-locale", "en", "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			want := "ID\tDescription\tEst. Work Volume\tEst. Rework Volume Ratio\tNeeded Resources\tStart Condition"
			if !strings.Contains(spy.Stdout.String(), want) {
				t.Errorf("expected header row %q, got: %q", want, spy.Stdout.String())
			}
		})
	})

	t.Run("mode all", func(t *testing.T) {
		cases := []struct {
			name       string
			tType      string
			wantAPHead string
			wantADHead string
		}{
			{
				"all-plan",
				"all-plan",
				"ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件",
				"ID\tDescription\t利用可能時刻\t最大版",
			},
			{
				"all-plan-master",
				"all-plan-master",
				"ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tMasterRow\tMasterBar",
				"ID\tDescription\t利用可能時刻\t最大版",
			},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				outDir := t.TempDir()
				spy := cli.SpyProcInout()
				exitStatus := MainCommandByArgs([]string{
					"-t", c.tType,
					"-p", "testdata/loop/pfd.drawio",
					"-cd", "testdata/loop/comp_deliv.tsv",
					"-out-dir", outDir,
					"-res", "finite",
					"-fb", "enabled",
				}, spy.NewProcInout())
				if exitStatus != 0 {
					t.Log(spy.Stderr.String())
					t.Errorf("exitStatus = %d, want 0", exitStatus)
				}
				apBytes, err := os.ReadFile(filepath.Join(outDir, "ap.tsv"))
				if err != nil {
					t.Fatalf("read ap.tsv: %v", err)
				}
				if !strings.Contains(string(apBytes), c.wantAPHead) {
					t.Errorf("ap.tsv missing %q, got: %q", c.wantAPHead, string(apBytes))
				}
				adBytes, err := os.ReadFile(filepath.Join(outDir, "ad.tsv"))
				if err != nil {
					t.Fatalf("read ad.tsv: %v", err)
				}
				if !strings.Contains(string(adBytes), c.wantADHead) {
					t.Errorf("ad.tsv missing %q, got: %q", c.wantADHead, string(adBytes))
				}
			})
		}
		t.Run("a alias", func(t *testing.T) {
			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "a", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-out-dir", outDir, "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			for _, name := range []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv", "project.json"} {
				if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
					t.Errorf("expected file %s: %v", name, err)
				}
			}
		})
		t.Run("a-plan-master alias", func(t *testing.T) {
			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "a-plan-master", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-out-dir", outDir, "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			apBytes, _ := os.ReadFile(filepath.Join(outDir, "ap.tsv"))
			want := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tMasterRow\tMasterBar"
			if !strings.Contains(string(apBytes), want) {
				t.Errorf("ap.tsv missing %q, got: %q", want, string(apBytes))
			}
		})
	})

	t.Run("mode existing", func(t *testing.T) {
		t.Run("ap-plan appends to minimal", func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "ap.tsv")
			if err := os.WriteFile(tmp, []byte("ID\tDescription\nP1\timplement\nP2\treview\n"), 0644); err != nil {
				t.Fatalf("write tmp ap.tsv: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", tmp, "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			want := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件"
			if !strings.Contains(spy.Stdout.String(), want) {
				t.Errorf("expected header row %q, got: %q", want, spy.Stdout.String())
			}
		})
		t.Run("ap-plan no duplicate when already present", func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "ap.tsv")
			content := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\nP1\timplement\t1\t0.1\tR1:1\t\nP2\treview\t1\t0.1\tR1:1\t\n"
			if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
				t.Fatalf("write tmp ap.tsv: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", tmp, "-res", "finite", "-fb", "enabled", "-locale", "en", "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			out := spy.Stdout.String()
			if strings.Contains(out, "Est. Work Volume") {
				t.Errorf("English header should not be added when Japanese counterpart exists: %q", out)
			}
		})
		t.Run("ap-plan-master extends plan", func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "ap.tsv")
			content := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\nP1\timplement\t1\t0.1\tR1:1\t\nP2\treview\t1\t0.1\tR1:1\t\n"
			if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
				t.Fatalf("write tmp ap.tsv: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan-master", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", tmp, "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			want := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tMasterRow\tMasterBar"
			if !strings.Contains(spy.Stdout.String(), want) {
				t.Errorf("expected %q, got: %q", want, spy.Stdout.String())
			}
		})
	})

	t.Run("project.json", func(t *testing.T) {
		projJSON := "testdata/loop/project.json"

		t.Run("-f project.json resolves ap", func(t *testing.T) {
			spy1 := cli.SpyProcInout()
			MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy1.NewProcInout())

			spy2 := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", projJSON}, spy2.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy2.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if spy1.Stdout.String() != spy2.Stdout.String() {
				t.Errorf("project.json output differs from explicit flags output\nexplicit: %q\nproject.json: %q", spy1.Stdout.String(), spy2.Stdout.String())
			}
		})

		t.Run("PFD fallback from project.json", func(t *testing.T) {
			spy := cli.SpyProcInout()

			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (PFD should come from project.json)", exitStatus)
			}
		})

		t.Run("CD fallback from project.json", func(t *testing.T) {
			spy := cli.SpyProcInout()

			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (CD should come from project.json)", exitStatus)
			}
		})

		t.Run("CLI -p overrides project.json pfd", func(t *testing.T) {
			spy1 := cli.SpyProcInout()
			MainCommandByArgs([]string{"-t", "ap", "-f", projJSON}, spy1.NewProcInout())

			spy2 := cli.SpyProcInout()

			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-f", projJSON}, spy2.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy2.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if spy1.Stdout.String() != spy2.Stdout.String() {
				t.Errorf("explicit -p output differs from fallback output")
			}
		})

		t.Run("-f project.json resolves ad", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ad", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-f project.json resolves cd", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-f project.json resolves r", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "r", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-f project.json resolves m", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "m", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-f project.json resolves g", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "g", "-f", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("missing type entry in project.json", func(t *testing.T) {

			dir := t.TempDir()
			partialJSON := filepath.Join(dir, "project.json")
			absPFD, _ := filepath.Abs("testdata/loop/pfd.drawio")
			absAP, _ := filepath.Abs("testdata/loop/atomic_proc.tsv")
			content := `{"pfd":"` + absPFD + `","atomic_process_table":"` + absAP + `"}`
			if err := os.WriteFile(partialJSON, []byte(content), 0644); err != nil {
				t.Fatalf("write: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "r", "-existing", partialJSON}, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("expected non-zero exit for missing resource_table in project.json")
			}
		})

		t.Run("unknown fields in project.json are ignored", func(t *testing.T) {
			dir := t.TempDir()
			projJSONPath := filepath.Join(dir, "project.json")
			absPFD, _ := filepath.Abs("testdata/loop/pfd.drawio")
			absAP, _ := filepath.Abs("testdata/loop/atomic_proc.tsv")
			absCD, _ := filepath.Abs("testdata/loop/comp_deliv.tsv")
			content := `{"pfd":"` + absPFD + `","atomic_process_table":"` + absAP + `","composite_deliverable_table":"` + absCD + `","unknown_future_field":"whatever"}`
			if err := os.WriteFile(projJSONPath, []byte(content), 0644); err != nil {
				t.Fatalf("write: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", projJSONPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (unknown fields should be ignored)", exitStatus)
			}
		})

		t.Run("absolute paths in project.json", func(t *testing.T) {
			dir := t.TempDir()
			projJSONPath := filepath.Join(dir, "project.json")
			absPFD, _ := filepath.Abs("testdata/loop/pfd.drawio")
			absAP, _ := filepath.Abs("testdata/loop/atomic_proc.tsv")
			absCD, _ := filepath.Abs("testdata/loop/comp_deliv.tsv")
			content := `{"pfd":"` + absPFD + `","atomic_process_table":"` + absAP + `","composite_deliverable_table":"` + absCD + `"}`
			if err := os.WriteFile(projJSONPath, []byte(content), 0644); err != nil {
				t.Fatalf("write: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-f", projJSONPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (absolute paths should resolve correctly)", exitStatus)
			}
		})

		t.Run("-t ap -inplace via project.json", func(t *testing.T) {
			dir := t.TempDir()

			apSrc, _ := os.ReadFile("testdata/loop/atomic_proc.tsv")
			apDst := filepath.Join(dir, "atomic_proc.tsv")
			if err := os.WriteFile(apDst, apSrc, 0644); err != nil {
				t.Fatalf("write: %v", err)
			}
			pfdAbs, _ := filepath.Abs("testdata/loop/pfd.drawio")
			cdAbs, _ := filepath.Abs("testdata/loop/comp_deliv.tsv")
			projJSONPath := filepath.Join(dir, "project.json")
			content := `{"pfd":"` + pfdAbs + `","atomic_process_table":"atomic_proc.tsv","composite_deliverable_table":"` + cdAbs + `"}`
			if err := os.WriteFile(projJSONPath, []byte(content), 0644); err != nil {
				t.Fatalf("write: %v", err)
			}
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-inplace", "-f", projJSONPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}

			raw, _ := os.ReadFile(projJSONPath)
			if string(raw) != content {
				t.Errorf("project.json was modified unexpectedly")
			}

			if _, err := os.ReadFile(apDst); err != nil {
				t.Errorf("AP file missing after inplace: %v", err)
			}
		})
	})
}

const brokenDrawio = `<mxfile host="x">
    <diagram id="d1" name="P1">
        <mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel>
    </diagram>
</mxfile>
`

func TestMainCommandByArgs_InplaceUpdatesFile(t *testing.T) {
	testCases := map[string]struct {
		TableType    string
		ExistingFile string
	}{
		"atomic process":     {TableType: "ap", ExistingFile: "testdata/loop/atomic_proc.tsv"},
		"atomic deliverable": {TableType: "ad", ExistingFile: "testdata/loop/deliv.tsv"},
		"composite process":  {TableType: "cp", ExistingFile: "testdata/loop/comp_proc.tsv"},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(testCase.ExistingFile)
			if err != nil {
				t.Fatal(err)
			}
			existingPath := filepath.Join(t.TempDir(), "existing.tsv")
			if err := os.WriteFile(existingPath, src, 0644); err != nil {
				t.Fatal(err)
			}

			args := []string{"-t", testCase.TableType, "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", existingPath}

			stdoutSpy := cli.SpyProcInout()
			if exit := MainCommandByArgs(args, stdoutSpy.NewProcInout()); exit != 0 {
				t.Fatalf("exit = %d, want 0\nstderr: %s", exit, stdoutSpy.Stderr.String())
			}

			spy := cli.SpyProcInout()
			if exit := MainCommandByArgs(append(args, "-inplace"), spy.NewProcInout()); exit != 0 {
				t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
			}
			if spy.Stdout.Len() != 0 {
				t.Errorf("stdout should be empty in -inplace mode, got %q", spy.Stdout.String())
			}

			got, err := os.ReadFile(existingPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != stdoutSpy.Stdout.String() {
				t.Errorf("in-place content = %q, want %q (same as stdout)", string(got), stdoutSpy.Stdout.String())
			}

			if string(got) == string(src) {
				t.Errorf("in-place content is unchanged from the existing table: %q", string(got))
			}
		})
	}
}

func TestMainCommandByArgs_InplaceKeepsFileOnError(t *testing.T) {
	t.Run("single table", func(t *testing.T) {
		dir := t.TempDir()
		existingPath := filepath.Join(dir, "atomic_proc.tsv")
		src, err := os.ReadFile("testdata/loop/atomic_proc.tsv")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(existingPath, src, 0644); err != nil {
			t.Fatal(err)
		}
		brokenPFDPath := filepath.Join(dir, "broken.drawio")
		if err := os.WriteFile(brokenPFDPath, []byte(brokenDrawio), 0644); err != nil {
			t.Fatal(err)
		}

		spy := cli.SpyProcInout()
		exit := MainCommandByArgs([]string{"-t", "ap", "-p", brokenPFDPath, "-existing", existingPath, "-inplace"}, spy.NewProcInout())
		if exit != 1 {
			t.Fatalf("exit = %d, want 1\nstderr: %s", exit, spy.Stderr.String())
		}

		got, err := os.ReadFile(existingPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(src) {
			t.Errorf("the existing table must not change on error\ngot:  %q\nwant: %q", string(got), string(src))
		}
	})

	t.Run("-t all with project.json", func(t *testing.T) {
		dir := t.TempDir()
		bootstrap := cli.SpyProcInout()
		if exit := MainCommandByArgs([]string{
			"-t", "all",
			"-p", "testdata/loop/pfd.drawio",
			"-cd", "testdata/loop/comp_deliv.tsv",
			"-ap", "testdata/loop/atomic_proc.tsv",
			"-out-dir", dir,
			"-res", "finite",
			"-fb", "enabled",
		}, bootstrap.NewProcInout()); exit != 0 {
			t.Fatalf("bootstrap exit = %d, want 0\nstderr: %s", exit, bootstrap.Stderr.String())
		}

		apPath := filepath.Join(dir, "ap.tsv")
		if err := os.WriteFile(apPath, []byte("ID\tDescription\n"), 0644); err != nil {
			t.Fatal(err)
		}

		adPath := filepath.Join(dir, "ad.tsv")
		brokenAD := "ID\tDescription\nD1\n"
		if err := os.WriteFile(adPath, []byte(brokenAD), 0644); err != nil {
			t.Fatal(err)
		}

		before := make(map[string][]byte)
		names := []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv"}
		for _, name := range names {
			bs, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			before[name] = bs
		}

		spy := cli.SpyProcInout()
		exit := MainCommandByArgs([]string{"-t", "all", "-f", filepath.Join(dir, "project.json"), "-inplace", "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
		if exit != 1 {
			t.Fatalf("exit = %d, want 1\nstderr: %s", exit, spy.Stderr.String())
		}

		for _, name := range names {
			got, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(before[name]) {
				t.Errorf("%s must not change on error\ngot:  %q\nwant: %q", name, string(got), string(before[name]))
			}
		}
	})
}

func TestMainCommandByArgs_AllInplaceRejectsEmptyTable(t *testing.T) {
	dir := t.TempDir()
	bootstrap := cli.SpyProcInout()
	if exit := MainCommandByArgs([]string{
		"-t", "all",
		"-p", "testdata/loop/pfd.drawio",
		"-cd", "testdata/loop/comp_deliv.tsv",
		"-ap", "testdata/loop/atomic_proc.tsv",
		"-out-dir", dir,
		"-res", "finite",
		"-fb", "enabled",
	}, bootstrap.NewProcInout()); exit != 0 {
		t.Fatalf("bootstrap exit = %d, want 0\nstderr: %s", exit, bootstrap.Stderr.String())
	}

	pfdAbs, err := filepath.Abs("testdata/loop/pfd.drawio")
	if err != nil {
		t.Fatal(err)
	}
	projPath := filepath.Join(dir, "project.json")
	proj := `{"pfd":"` + pfdAbs + `","atomic_deliverable_table":"ad.tsv","composite_deliverable_table":"cd.tsv","resource_table":"r.tsv","milestone_table":"m.tsv","group_table":"g.tsv"}`
	if err := os.WriteFile(projPath, []byte(proj), 0644); err != nil {
		t.Fatal(err)
	}

	names := []string{"ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv"}
	before := make(map[string][]byte)
	for _, name := range names {
		bs, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		before[name] = bs
	}

	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-t", "all", "-f", projPath, "-inplace", "-res", "finite", "-fb", "enabled"}, spy.NewProcInout())
	if exit != 1 {
		t.Fatalf("exit = %d, want 1\nstderr: %s", exit, spy.Stderr.String())
	}

	for _, name := range names {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(before[name]) {
			t.Errorf("%s must not change\ngot:  %q\nwant: %q", name, string(got), string(before[name]))
		}
	}
}

func TestMainCommandByArgs_AllKeepsFilesOnError(t *testing.T) {
	dir := t.TempDir()

	badAPPath := filepath.Join(dir, "bad_ap.tsv")
	if err := os.WriteFile(badAPPath, []byte("ID\t説明\t必要資源\nP1\tプロセス\tR1:x\n"), 0644); err != nil {
		t.Fatal(err)
	}

	existingAPPath := filepath.Join(dir, "ap.tsv")
	if err := os.WriteFile(existingAPPath, []byte("previous\n"), 0644); err != nil {
		t.Fatal(err)
	}

	spy := cli.SpyProcInout()

	exit := MainCommandByArgs([]string{
		"-t", "all",
		"-p", "testdata/loop/pfd.drawio",
		"-cd", "testdata/loop/comp_deliv.tsv",
		"-ap", badAPPath,
		"-out-dir", dir,
		"-res", "finite",
		"-fb", "enabled",
	}, spy.NewProcInout())
	if exit != 1 {
		t.Fatalf("exit = %d, want 1\nstderr: %s", exit, spy.Stderr.String())
	}

	got, err := os.ReadFile(existingAPPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "previous\n" {
		t.Errorf("ap.tsv must not change on error\ngot:  %q\nwant: %q", string(got), "previous\n")
	}
	for _, name := range []string{"ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv", "project.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("%s must not be written on error", name)
		}
	}
}

func TestMainCommandByArgs_InplaceRejectsEmptyTable(t *testing.T) {
	testCases := map[string]struct {
		TableType    string
		ExistingFile string
	}{

		"milestone without -ap": {TableType: "m", ExistingFile: "testdata/loop/milestone.tsv"},
		"group without -ap":     {TableType: "g", ExistingFile: "testdata/loop/group.tsv"},

		"composite deliverable without -p": {TableType: "cd", ExistingFile: "testdata/loop/comp_deliv.tsv"},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(testCase.ExistingFile)
			if err != nil {
				t.Fatal(err)
			}
			existingPath := filepath.Join(t.TempDir(), "existing.tsv")
			if err := os.WriteFile(existingPath, src, 0644); err != nil {
				t.Fatal(err)
			}

			spy := cli.SpyProcInout()
			exit := MainCommandByArgs([]string{"-t", testCase.TableType, "-existing", existingPath, "-inplace"}, spy.NewProcInout())
			if exit != 1 {
				t.Fatalf("exit = %d, want 1\nstderr: %s", exit, spy.Stderr.String())
			}

			got, err := os.ReadFile(existingPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(src) {
				t.Errorf("the existing table must not be emptied\ngot:  %q\nwant: %q", string(got), string(src))
			}
		})
	}
}

func TestMainCommandByArgs_AllExecModel(t *testing.T) {
	testCases := map[string]struct {
		Args             []string
		WantFiles        []string
		WantMissingFiles []string
		WantJSONContains []string
		WantJSONExcludes []string
	}{
		"default model generates no resource table": {
			Args:             []string{"-t", "all"},
			WantFiles:        []string{"ap.tsv", "ad.tsv", "cd.tsv", "m.tsv", "g.tsv", "project.json"},
			WantMissingFiles: []string{"r.tsv"},
			WantJSONContains: []string{`"resource_mode": "infinite"`, `"feedback_mode": "disabled"`},
			WantJSONExcludes: []string{"resource_table"},
		},
		"finite resources generate the resource table": {
			Args:             []string{"-t", "all", "-res", "finite", "-fb", "enabled"},
			WantFiles:        []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv", "project.json"},
			WantJSONContains: []string{`"resource_table": "r.tsv"`, `"resource_mode": "finite"`, `"feedback_mode": "enabled"`},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			args := append(append([]string{}, tc.Args...),
				"-p", "testdata/loop/pfd.drawio",
				"-cd", "testdata/loop/comp_deliv.tsv",
				"-ap", "testdata/loop/atomic_proc.tsv",
				"-out-dir", outDir,
			)
			if exitStatus := MainCommandByArgs(args, spy.NewProcInout()); exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Fatalf("exitStatus = %d, want 0", exitStatus)
			}
			for _, name := range tc.WantFiles {
				if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
					t.Errorf("expected %s to exist: %v", name, err)
				}
			}
			for _, name := range tc.WantMissingFiles {
				if _, err := os.Stat(filepath.Join(outDir, name)); err == nil {
					t.Errorf("expected %s not to exist", name)
				}
			}
			projectJSON, err := os.ReadFile(filepath.Join(outDir, "project.json"))
			if err != nil {
				t.Fatalf("read project.json: %v", err)
			}
			for _, want := range tc.WantJSONContains {
				if !strings.Contains(string(projectJSON), want) {
					t.Errorf("expected project.json to contain %q, got:\n%s", want, string(projectJSON))
				}
			}
			for _, unwanted := range tc.WantJSONExcludes {
				if strings.Contains(string(projectJSON), unwanted) {
					t.Errorf("expected project.json not to contain %q, got:\n%s", unwanted, string(projectJSON))
				}
			}
		})
	}
}

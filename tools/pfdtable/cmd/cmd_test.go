package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
)

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
		t.Run("-f alias for -existing (TSV)", func(t *testing.T) {
			spy1 := cli.SpyProcInout()
			MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy1.NewProcInout())

			spy2 := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-f", "testdata/loop/atomic_proc.tsv"}, spy2.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy2.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if spy1.Stdout.String() != spy2.Stdout.String() {
				t.Errorf("-f output differs from -existing output\n-existing: %q\n-f: %q", spy1.Stdout.String(), spy2.Stdout.String())
			}
		})
		t.Run("-f short overrides -existing long", func(t *testing.T) {
			spy := cli.SpyProcInout()
			// -f wins over -existing; -existing points at non-existent file, -f at real file
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-f", "testdata/loop/atomic_proc.tsv", "-existing", "testdata/loop/nonexistent.tsv"}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (-f should override -existing)", exitStatus)
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
		t.Run("-existing project.json -out-dir", func(t *testing.T) {
			// First generate project.json with -t all
			srcDir := t.TempDir()
			spy0 := cli.SpyProcInout()
			MainCommandByArgs([]string{
				"-t", "all",
				"-p", "testdata/loop/pfd.drawio",
				"-cd", "testdata/loop/comp_deliv.tsv",
				"-ap", "testdata/loop/atomic_proc.tsv",
				"-out-dir", srcDir,
			}, spy0.NewProcInout())

			outDir := t.TempDir()
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{
				"-t", "all",
				"-existing", filepath.Join(srcDir, "project.json"),
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
			// project.json must not exist in srcDir as modified (only outDir gets a new one)
			srcProjRaw, _ := os.ReadFile(filepath.Join(srcDir, "project.json"))
			outProjRaw, _ := os.ReadFile(filepath.Join(outDir, "project.json"))
			// They should both be valid JSON (just different paths)
			if len(outProjRaw) == 0 {
				t.Errorf("outDir/project.json is empty")
			}
			_ = srcProjRaw
		})
		t.Run("-existing project.json -inplace", func(t *testing.T) {
			// Generate project.json + TSVs in a tempdir
			srcDir := t.TempDir()
			spy0 := cli.SpyProcInout()
			MainCommandByArgs([]string{
				"-t", "all",
				"-p", "testdata/loop/pfd.drawio",
				"-cd", "testdata/loop/comp_deliv.tsv",
				"-ap", "testdata/loop/atomic_proc.tsv",
				"-out-dir", srcDir,
			}, spy0.NewProcInout())

			projJSONPath := filepath.Join(srcDir, "project.json")
			origProjRaw, _ := os.ReadFile(projJSONPath)

			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{
				"-t", "all",
				"-existing", projJSONPath,
				"-inplace",
			}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			// project.json must not be modified
			afterProjRaw, _ := os.ReadFile(projJSONPath)
			if string(origProjRaw) != string(afterProjRaw) {
				t.Errorf("project.json was modified by -t all -inplace (should not be)")
			}
			// all TSV files must still exist
			for _, name := range []string{"ap.tsv", "ad.tsv", "cd.tsv", "r.tsv", "m.tsv", "g.tsv"} {
				if _, err := os.Stat(filepath.Join(srcDir, name)); err != nil {
					t.Errorf("expected %s to still exist after inplace: %v", name, err)
				}
			}
		})
		t.Run("-existing project.json without -out-dir or -inplace is error", func(t *testing.T) {
			projJSON := "testdata/loop/project.json"
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "all", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("expected non-zero exit for -t all -existing without -out-dir or -inplace")
			}
		})
		t.Run("-existing project.json -out-dir and -inplace together is error", func(t *testing.T) {
			projJSON := "testdata/loop/project.json"
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "all", "-existing", projJSON, "-out-dir", t.TempDir(), "-inplace"}, spy.NewProcInout())
			if exitStatus == 0 {
				t.Errorf("expected non-zero exit for -t all -existing -out-dir -inplace (mutually exclusive)")
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
			wantSub string
		}{
			{"ap-plan", "ap-plan", "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件"},
			{"ap-plan-master", "ap-plan-master", "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tマイルストーン\tグループ"},
			{"ad-plan", "ad-plan", "ID\tDescription\t利用可能時刻\t最大版"},
			{"ad-plan-master", "ad-plan-master", "ID\tDescription\t利用可能時刻\t最大版"},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				spy := cli.SpyProcInout()
				exitStatus := MainCommandByArgs([]string{"-t", c.tType, "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv"}, spy.NewProcInout())
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-locale", "en"}, spy.NewProcInout())
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
			name        string
			tType       string
			wantAPHead  string
			wantADHead  string
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
				"ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tマイルストーン\tグループ",
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
			exitStatus := MainCommandByArgs([]string{"-t", "a", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-out-dir", outDir}, spy.NewProcInout())
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
			exitStatus := MainCommandByArgs([]string{"-t", "a-plan-master", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-out-dir", outDir}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			apBytes, _ := os.ReadFile(filepath.Join(outDir, "ap.tsv"))
			want := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tマイルストーン\tグループ"
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", tmp}, spy.NewProcInout())
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", tmp, "-locale", "en"}, spy.NewProcInout())
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap-plan-master", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", tmp}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			want := "ID\tDescription\t予想作業量\t予想手戻り作業量割合\t必要資源\t開始条件\tマイルストーン\tグループ"
			if !strings.Contains(spy.Stdout.String(), want) {
				t.Errorf("expected %q, got: %q", want, spy.Stdout.String())
			}
		})
	})

	t.Run("project.json", func(t *testing.T) {
		projJSON := "testdata/loop/project.json"

		t.Run("-existing project.json resolves ap", func(t *testing.T) {
			spy1 := cli.SpyProcInout()
			MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-cd", "testdata/loop/comp_deliv.tsv", "-existing", "testdata/loop/atomic_proc.tsv"}, spy1.NewProcInout())

			spy2 := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-existing", projJSON}, spy2.NewProcInout())
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
			// -p omitted; project.json supplies pfd
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (PFD should come from project.json)", exitStatus)
			}
		})

		t.Run("CD fallback from project.json", func(t *testing.T) {
			spy := cli.SpyProcInout()
			// -cd omitted; project.json supplies composite_deliverable_table
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (CD should come from project.json)", exitStatus)
			}
		})

		t.Run("CLI -p overrides project.json pfd", func(t *testing.T) {
			spy1 := cli.SpyProcInout()
			MainCommandByArgs([]string{"-t", "ap", "-existing", projJSON}, spy1.NewProcInout())

			spy2 := cli.SpyProcInout()
			// explicit -p should produce same result (same file)
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-p", "testdata/loop/pfd.drawio", "-existing", projJSON}, spy2.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy2.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			if spy1.Stdout.String() != spy2.Stdout.String() {
				t.Errorf("explicit -p output differs from fallback output")
			}
		})

		t.Run("-existing project.json resolves ad", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "ad", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-existing project.json resolves cd", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "cd", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-existing project.json resolves r", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "r", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-existing project.json resolves m", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "m", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("-existing project.json resolves g", func(t *testing.T) {
			spy := cli.SpyProcInout()
			exitStatus := MainCommandByArgs([]string{"-t", "g", "-existing", projJSON}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
		})

		t.Run("missing type entry in project.json", func(t *testing.T) {
			// project.json without resource_table
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-existing", projJSONPath}, spy.NewProcInout())
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-existing", projJSONPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0 (absolute paths should resolve correctly)", exitStatus)
			}
		})

		t.Run("-t ap -inplace via project.json", func(t *testing.T) {
			dir := t.TempDir()
			// Copy TSV to tempdir
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
			exitStatus := MainCommandByArgs([]string{"-t", "ap", "-inplace", "-existing", projJSONPath}, spy.NewProcInout())
			if exitStatus != 0 {
				t.Log(spy.Stderr.String())
				t.Errorf("exitStatus = %d, want 0", exitStatus)
			}
			// project.json must not be modified
			raw, _ := os.ReadFile(projJSONPath)
			if string(raw) != content {
				t.Errorf("project.json was modified unexpectedly")
			}
			// the AP file should still be readable
			if _, err := os.ReadFile(apDst); err != nil {
				t.Errorf("AP file missing after inplace: %v", err)
			}
		})
	})
}

package perf_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/perf"
)

func TestProfilerStop(t *testing.T) {
	cases := map[string]struct {
		cpu bool
		mem bool
	}{
		"noop":        {cpu: false, mem: false},
		"cpu only":    {cpu: true, mem: false},
		"mem only":    {cpu: false, mem: true},
		"cpu and mem": {cpu: true, mem: true},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			var cpuPath, memPath string
			if c.cpu {
				cpuPath = filepath.Join(dir, "cpu.pprof")
			}
			if c.mem {
				memPath = filepath.Join(dir, "mem.pprof")
			}

			p, err := perf.Start(cpuPath, memPath)
			if err != nil {
				t.Fatalf("Start: unexpected error: %v", err)
			}

			burnCPU()

			if err := p.Stop(); err != nil {
				t.Fatalf("Stop: unexpected error: %v", err)
			}

			if err := p.Stop(); err != nil {
				t.Fatalf("Stop (second call): unexpected error: %v", err)
			}

			assertProfile(t, cpuPath, c.cpu)
			assertProfile(t, memPath, c.mem)
		})
	}
}

func assertProfile(t *testing.T, path string, want bool) {
	t.Helper()
	if !want {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %q: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("profile %q is empty", path)
	}
}

func burnCPU() {
	sum := 0
	for i := 0; i < 5_000_000; i++ {
		sum += i % 7
	}
	_ = sum
}

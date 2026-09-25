//go:build unix

package perf_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/Kuniwak/pfd-tools/perf"
)

func TestProfilerStopOnSignal(t *testing.T) {
	dir := t.TempDir()
	cpuPath := filepath.Join(dir, "cpu.pprof")
	memPath := filepath.Join(dir, "mem.pprof")

	p, err := perf.Start(cpuPath, memPath)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	done := p.StopOnSignal(syscall.SIGUSR1)

	burnCPU()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StopOnSignal did not complete after signal")
	}

	for _, path := range []string{cpuPath, memPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %q: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("profile %q is empty after signal", path)
		}
	}
}

func TestStartWithSignalHandler(t *testing.T) {
	dir := t.TempDir()
	cpuPath := filepath.Join(dir, "cpu.pprof")
	memPath := filepath.Join(dir, "mem.pprof")

	signaled := make(chan struct{})
	p, err := perf.StartWithSignalHandler(cpuPath, memPath, func() { close(signaled) }, syscall.SIGUSR1)
	if err != nil {
		t.Fatalf("StartWithSignalHandler: %v", err)
	}
	defer func() { _ = p.Stop() }()

	burnCPU()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	select {
	case <-signaled:
	case <-time.After(5 * time.Second):
		t.Fatal("onSignal was not called after signal")
	}

	for _, path := range []string{cpuPath, memPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %q: %v", path, err)
		}
		if info.Size() == 0 {
			t.Fatalf("profile %q is empty after signal", path)
		}
	}
}

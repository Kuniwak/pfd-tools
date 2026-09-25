package perf

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
)

type Profiler struct {
	cpuFile *os.File
	memPath string
	once    sync.Once
	stopErr error
}

func Start(cpuPath, memPath string) (*Profiler, error) {
	p := &Profiler{memPath: memPath}
	if cpuPath != "" {
		f, err := os.Create(cpuPath)
		if err != nil {
			return nil, fmt.Errorf("perf.Start: create cpu profile: %w", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("perf.Start: start cpu profile: %w", err)
		}
		p.cpuFile = f
	}
	return p, nil
}

func (p *Profiler) Stop() error {
	p.once.Do(func() {
		var errs []error
		if p.cpuFile != nil {
			pprof.StopCPUProfile()
			if err := p.cpuFile.Close(); err != nil {
				errs = append(errs, fmt.Errorf("perf.Profiler.Stop: close cpu profile: %w", err))
			}
		}
		if p.memPath != "" {
			f, err := os.Create(p.memPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("perf.Profiler.Stop: create mem profile: %w", err))
			} else {
				runtime.GC()
				if err := pprof.WriteHeapProfile(f); err != nil {
					errs = append(errs, fmt.Errorf("perf.Profiler.Stop: write heap profile: %w", err))
				}
				if err := f.Close(); err != nil {
					errs = append(errs, fmt.Errorf("perf.Profiler.Stop: close mem profile: %w", err))
				}
			}
		}
		p.stopErr = errors.Join(errs...)
	})
	return p.stopErr
}

func (p *Profiler) StopOnSignal(sigs ...os.Signal) <-chan struct{} {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	done := make(chan struct{})
	go func() {
		<-ch
		_ = p.Stop()
		close(done)
	}()
	return done
}

func StartWithSignalHandler(cpuPath, memPath string, onSignal func(), sigs ...os.Signal) (*Profiler, error) {
	p, err := Start(cpuPath, memPath)
	if err != nil {
		return nil, err
	}
	if cpuPath != "" || memPath != "" {
		go func() {
			<-p.StopOnSignal(sigs...)
			onSignal()
		}()
	}
	return p, nil
}

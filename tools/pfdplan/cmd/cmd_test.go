package cmd

import (
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng/pngtest"
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

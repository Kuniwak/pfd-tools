package allcheckers

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddot"
	"github.com/Kuniwak/pfd-tools/pfd/pfdfuzz"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"pgregory.net/rapid"
)

func TestAll(t *testing.T) {
	for _, p := range pfd.PresetsAll {
		t.Run(p.Title, func(t *testing.T) {
			ch := make(chan checkers.Problem)
			go func() {
				logger := slog.New(slogtest.NewTestHandler(t))
				lint := NewLintFunc(logger)

				cdt := pfd.NewCompositeDeliverableTable(p, pfd.NewNodeMap(p.Nodes, logger))
				if err := lint(Target{PFD: p, CompositeDeliverableTable: cdt, Model: execmodel.Model{Resource: execmodel.ResourceModeFinite, Feedback: execmodel.FeedbackModeEnabled}}, ch); err != nil {
					t.Errorf("NewLintFunc: %v", err)
				}
			}()
			got := chans.Slice(ch)
			if len(got) > 0 {
				t.Errorf("got %v, expected nil", got)
			}
		})
	}
}

func FuzzLint(t *testing.F) {
	sb := &strings.Builder{}
	t.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
		p := pfdfuzz.AnyValidPFD(t, 100)
		ch := make(chan checkers.Problem)
		go func() {
			lint := NewLintFunc(slog.New(slogtest.NewRapidHandler(t)))
			if err := lint(Target{PFD: p, Model: execmodel.Model{Resource: execmodel.ResourceModeFinite, Feedback: execmodel.FeedbackModeEnabled}}, ch); err != nil {
				t.Errorf("NewLintFunc: %v", err)
			}
		}()
		got := chans.Slice(ch)
		if len(got) > 0 {
			sb.Reset()
			pfddot.Write(sb, p)
			t.Log(sb.String())
			t.Errorf("got %v, expected nil", got)
		}
	}))
}

func TestLintRejectsUnknownExecModel(t *testing.T) {
	testCases := map[string]execmodel.Model{
		"zero value":            {},
		"unknown resource mode": {Resource: execmodel.ResourceMode("none"), Feedback: execmodel.FeedbackModeDisabled},
		"unknown feedback mode": {Resource: execmodel.ResourceModeFinite, Feedback: execmodel.FeedbackMode("on")},
	}
	for name, model := range testCases {
		t.Run(name, func(t *testing.T) {
			if _, err := Lint(Target{PFD: pfd.PresetSmallest, Model: model}, slog.New(slogtest.NewTestHandler(t))); err == nil {
				t.Errorf("Lint should fail for %+v, but did not", model)
			}
		})
	}
}

func TestLintFuncClosesChannelOnce(t *testing.T) {
	testCases := map[string]struct {
		model     execmodel.Model
		expectErr bool
	}{
		"both checks run": {
			model:     execmodel.Model{Resource: execmodel.ResourceModeFinite, Feedback: execmodel.FeedbackModeEnabled},
			expectErr: false,
		},

		"model validation fails immediately": {
			model:     execmodel.Model{},
			expectErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lint := NewLintFunc(slog.New(slogtest.NewTestHandler(t)))
			for range 100 {
				ch := make(chan checkers.Problem)
				errCh := make(chan error, 1)

				go func() {
					errCh <- lint(Target{PFD: pfd.PresetSmallest, Model: tc.model}, ch)
				}()
				for range ch {
				}
				err := <-errCh

				if (err != nil) != tc.expectErr {
					t.Fatalf("lint() error = %v, expectErr %v", err, tc.expectErr)
				}
			}
		})
	}
}

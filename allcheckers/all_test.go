package allcheckers

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/chans"
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddot"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"pgregory.net/rapid"
)

func TestAll(t *testing.T) {
	for _, p := range pfd.PresetsAll {
		t.Run(p.Title, func(t *testing.T) {
			ch := make(chan checkers.Problem)
			go func() {
				lint := NewLintFunc(slog.New(slogtest.NewTestHandler(t)))
				if err := lint(p, nil, nil, nil, nil, nil, nil, nil, ch); err != nil {
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
		p := pfd.AnyValidPFD(t, 100)
		ch := make(chan checkers.Problem)
		go func() {
			lint := NewLintFunc(slog.New(slogtest.NewRapidHandler(t)))
			if err := lint(p, nil, nil, nil, nil, nil, nil, nil, ch); err != nil {
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

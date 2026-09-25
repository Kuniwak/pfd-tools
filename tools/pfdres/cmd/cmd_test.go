package cmd

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/spreadsheet"
	"github.com/Kuniwak/pfd-tools/version"
)

func TestMainCommandByArgs(t *testing.T) {
	t.Run("prints the needed resources for the head count", func(t *testing.T) {
		cases := map[string]struct {
			args []string
			want string
		}{
			"single":            {args: []string{"3"}, want: "A:1;B:1;C:1\n"},
			"one":               {args: []string{"1"}, want: "A:1\n"},
			"locale en is same": {args: []string{"-locale", "en", "3"}, want: "A:1;B:1;C:1\n"},
			"locale ja is same": {args: []string{"-locale", "ja", "3"}, want: "A:1;B:1;C:1\n"},
			"rolls over past Z": {args: []string{"27"}, want: NeededResourceSets(27) + "\n"},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				spy := cli.SpyProcInout()
				exitStatus := MainCommandByArgs(c.args, spy.NewProcInout())
				if exitStatus != 0 {
					t.Log(spy.Stderr.String())
					t.Fatalf("exitStatus = %d, want 0", exitStatus)
				}
				if spy.Stdout.String() != c.want {
					t.Errorf("stdout = %q, want %q", spy.Stdout.String(), c.want)
				}
			})
		}
	})

	t.Run("fails on invalid head count", func(t *testing.T) {
		cases := map[string]struct {
			args []string
		}{
			"no argument":    {args: []string{}},
			"zero":           {args: []string{"0"}},
			"negative":       {args: []string{"-1"}},
			"not an integer": {args: []string{"abc"}},
			"too many args":  {args: []string{"3", "5"}},
			"unknown locale": {args: []string{"-locale", "fr", "3"}},
		}

		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				spy := cli.SpyProcInout()
				exitStatus := MainCommandByArgs(c.args, spy.NewProcInout())
				if exitStatus == 0 {
					t.Errorf("exitStatus = 0, want non-zero (stdout = %q)", spy.Stdout.String())
				}
				if spy.Stderr.String() == "" {
					t.Errorf("stderr is empty, want an error message")
				}
			})
		}
	})

	t.Run("prints the version", func(t *testing.T) {
		for _, arg := range []string{"-v", "-version"} {
			t.Run(arg, func(t *testing.T) {
				spy := cli.SpyProcInout()
				exitStatus := MainCommandByArgs([]string{arg}, spy.NewProcInout())
				if exitStatus != 0 {
					t.Log(spy.Stderr.String())
					t.Fatalf("exitStatus = %d, want 0", exitStatus)
				}
				want := version.Version + "\n"
				if spy.Stdout.String() != want {
					t.Errorf("stdout = %q, want %q", spy.Stdout.String(), want)
				}
			})
		}
	})
}

func TestNeededResourceSets_SatisfiesParserContract(t *testing.T) {
	for _, headCount := range []int{1, 2, 26, 27, 100} {
		s := NeededResourceSets(headCount)
		set, err := fsmtable.ParseNeededResourceSetEntry(s)
		if err != nil {
			t.Errorf("ParseNeededResourceSetEntry(%q) failed: %v", s, err)
			continue
		}
		if set.Len() != headCount {
			t.Errorf("ParseNeededResourceSetEntry(%q).Len() = %d, want %d", s, set.Len(), headCount)
			continue
		}

		want := make(map[fsm.ResourceID]bool, headCount)
		for i := 1; i <= headCount; i++ {
			want[fsm.ResourceID(spreadsheet.ColumnLabel(i))] = true
		}

		got := make(map[fsm.ResourceID]bool, headCount)
		for _, elem := range set.Slice() {
			if elem.Resources.Len() != 1 {
				t.Errorf("headCount=%d: an entry has %d resources, want 1 (no AND groups)", headCount, elem.Resources.Len())
				continue
			}
			if !elem.ConsumedVolume.ApproximateEqual(fsm.Volume(1)) {
				t.Errorf("headCount=%d: consumed volume = %v, want 1", headCount, elem.ConsumedVolume)
			}
			for _, id := range elem.Resources.Slice() {
				got[id] = true
			}
		}

		if len(got) != len(want) {
			t.Errorf("headCount=%d: resource IDs = %v, want %v", headCount, got, want)
			continue
		}
		for id := range want {
			if !got[id] {
				t.Errorf("headCount=%d: missing resource ID %q (got %v)", headCount, id, got)
			}
		}
	}
}

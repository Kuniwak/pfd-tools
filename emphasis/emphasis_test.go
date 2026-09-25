package emphasis_test

import (
	"bytes"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/emphasis"
)

func TestParse(t *testing.T) {
	cases := map[string]struct {
		tsv     string
		wantIDs []string
		wantErr bool
	}{
		"ID 列だけの表": {
			tsv:     "ID\nP1\nP2\n",
			wantIDs: []string{"P1", "P2"},
		},
		"criticalpath の出力をそのまま渡した表": {
			tsv:     "ID\t最大弾性値（全余裕）\t最小弾性値\nP2\t0.00\t1.00\nP1\t0.00\t-\n",
			wantIDs: []string{"P1", "P2"},
		},
		"ID 列が先頭でない表": {
			tsv:     "最大弾性値（全余裕）\tID\n0.00\tP1\n",
			wantIDs: []string{"P1"},
		},
		"重複した ID": {
			tsv:     "ID\nP1\nP1\n",
			wantIDs: []string{"P1"},
		},
		"空の ID のセル": {
			tsv:     "ID\t最大弾性値（全余裕）\nP1\t0.00\n\t0.00\n",
			wantIDs: []string{"P1"},
		},
		"ヘッダだけの表": {
			tsv:     "ID\n",
			wantIDs: []string{},
		},
		"ID 列がない表": {
			tsv:     "Key\tValue\nP1\t1\n",
			wantErr: true,
		},
		"空の入力": {
			tsv:     "",
			wantErr: true,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := emphasis.Parse(strings.NewReader(c.tsv))
			if c.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) = %v, want error", c.tsv, got.IDs())
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) = %v, want no error", c.tsv, err)
			}
			if !slices.Equal(got.IDs(), c.wantIDs) {
				t.Errorf("Parse(%q).IDs() = %v, want %v", c.tsv, got.IDs(), c.wantIDs)
			}
		})
	}
}

func TestSet_Enabled(t *testing.T) {
	cases := map[string]struct {
		set  *emphasis.Set
		want bool
	}{
		"強調 ID 表が指定されている":    {set: emphasis.MustParse("ID\nP1\n"), want: true},
		"空の強調 ID 表も指定されている":  {set: emphasis.MustParse("ID\n"), want: true},
		"強調なし（nil）は指定されていない": {set: nil, want: false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := c.set.Enabled(); got != c.want {
				t.Errorf("Enabled() = %v, want %v", got, c.want)
			}
			if got := c.set.IDs(); c.set == nil && got != nil {
				t.Errorf("IDs() = %v, want nil", got)
			}
		})
	}
}

func TestMustParse_PanicsOnBrokenTable(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("MustParse did not panic, want panic")
		}
	}()
	emphasis.MustParse("Key\nP1\n")
}

func TestSet_Contains(t *testing.T) {
	parsed, err := emphasis.Parse(strings.NewReader("ID\nP1\n"))
	if err != nil {
		t.Fatalf("Parse = %v, want no error", err)
	}

	cases := map[string]struct {
		set  *emphasis.Set
		id   string
		want bool
	}{
		"強調対象の ID":              {set: parsed, id: "P1", want: true},
		"強調対象でない ID":            {set: parsed, id: "P2", want: false},
		"強調なし（nil）はどの ID も含まない": {set: nil, id: "P1", want: false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := c.set.Contains(c.id); got != c.want {
				t.Errorf("Contains(%q) = %v, want %v", c.id, got, c.want)
			}
		})
	}
}

func TestSet_WarnUnknownIDs(t *testing.T) {
	parsed, err := emphasis.Parse(strings.NewReader("ID\nP1\nP9\nP8\n"))
	if err != nil {
		t.Fatalf("Parse = %v, want no error", err)
	}

	cases := map[string]struct {
		set      *emphasis.Set
		known    []string
		wantLogs []string
		wantNot  []string
	}{
		"未知の ID があれば警告する": {
			set:      parsed,
			known:    []string{"P1", "P2"},
			wantLogs: []string{"WARN", "P8", "P9"},
			wantNot:  []string{"P1"},
		},
		"すべて既知なら警告しない": {
			set:     parsed,
			known:   []string{"P1", "P8", "P9"},
			wantNot: []string{"WARN"},
		},
		"強調なし（nil）なら警告しない": {
			set:     nil,
			known:   []string{"P1"},
			wantNot: []string{"WARN"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
			c.set.WarnUnknownIDs(logger, c.known)
			got := buf.String()
			for _, want := range c.wantLogs {
				if !strings.Contains(got, want) {
					t.Errorf("log = %q, want to contain %q", got, want)
				}
			}
			for _, notWant := range c.wantNot {
				if strings.Contains(got, notWant) {
					t.Errorf("log = %q, want not to contain %q", got, notWant)
				}
			}
		})
	}
}

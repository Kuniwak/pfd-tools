package pfddrawio

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestRenumber(t *testing.T) {
	tests := []struct {
		FilePath     string
		ExpectedPath string
	}{
		{
			FilePath:     "testdata/sequential_without_id.drawio",
			ExpectedPath: "testdata/sequential_with_id.drawio",
		},
	}
	for _, test := range tests {
		t.Run(test.FilePath, func(t *testing.T) {
			f1, err := os.OpenFile(test.FilePath, os.O_RDONLY, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer f1.Close()

			bs, err := os.ReadFile(test.ExpectedPath)
			if err != nil {
				t.Fatal(err)
			}

			nodes, err := Renumber(f1, slog.New(slogtest.NewTestHandler(t)))
			if err != nil {
				t.Fatal(err)
			}

			w := bytes.NewBuffer(nil)
			for _, node := range nodes {
				if err := node.Write(w); err != nil {
					t.Fatal(err)
				}
			}

			expected := string(bs)
			if expected != w.String() {
				t.Logf("got: %s", w.String())
				t.Error(cmp.Diff(expected, w.String()))
			}
		})
	}
}

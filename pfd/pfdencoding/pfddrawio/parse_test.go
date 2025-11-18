package pfddrawio

import (
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		FilePath string
		Expected []Diagram
	}{
		"without XML declaration": {
			FilePath: "testdata/example.drawio",
			Expected: exampleFile,
		},
		"with XML declaration": {
			FilePath: "testdata/example_with_xml_decl.drawio",
			Expected: exampleFile,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			exampleFile, err := os.Open(testCase.FilePath)
			if err != nil {
				t.Fatal(err)
			}
			defer sugar.IgnoreError(exampleFile.Close)
			logger := slog.New(slogtest.NewTestHandler(t))
			actual, err := ParseDiagrams(exampleFile, logger)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestParseVertexValue(t *testing.T) {
	testCases := map[string]struct {
		Input        string
		ExpectedID   pfd.NodeID
		ExpectedDesc string
	}{
		"no descriptions": {
			Input:        "D1",
			ExpectedID:   "D1",
			ExpectedDesc: "",
		},
		"description": {
			Input:        "D1: Implementation",
			ExpectedID:   "D1",
			ExpectedDesc: "Implementation",
		},
		"empty description": {
			Input:        "D1: ",
			ExpectedID:   "D1",
			ExpectedDesc: "",
		},
		"spaces before colon": {
			Input:        "D1 : Implementation",
			ExpectedID:   "D1",
			ExpectedDesc: "Implementation",
		},
		"description with colons": {
			Input:        "D1: Implementation: Implementation",
			ExpectedID:   "D1",
			ExpectedDesc: "Implementation: Implementation",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actualID, actualDesc, err := ParseVertexValue(testCase.Input)
			if err != nil {
				t.Fatal(err)
			}
			if actualID != testCase.ExpectedID {
				t.Errorf("expected ID: %q, actual: %q", testCase.ExpectedID, actualID)
			}
			if actualDesc != testCase.ExpectedDesc {
				t.Errorf("expected description: %q, actual: %q", testCase.ExpectedDesc, actualDesc)
			}
		})
	}
}

func TestParseValueHTML(t *testing.T) {
	testCases := map[string]struct {
		Input    ValueHTML
		Expected string
	}{
		"no tags": {
			Input:    "D1: Implementation",
			Expected: "D1: Implementation",
		},
		"wrapping": {
			Input:    "D1: <b>Implementation</b>",
			Expected: "D1: Implementation",
		},
		"embedded tags": {
			Input:    "D1: Implement<br>ation",
			Expected: "D1: Implement\nation",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			sb := strings.Builder{}
			err := ParseValueHTML(testCase.Input, &sb)
			if err != nil {
				t.Fatal(err)
			}
			if sb.String() != testCase.Expected {
				t.Errorf("expected: %q, actual: %q", testCase.Expected, sb.String())
			}
		})
	}
}

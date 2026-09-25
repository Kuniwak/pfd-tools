package pfddrawio

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
)

func TestPageNameElementID(t *testing.T) {
	testCases := map[string]struct {
		Name     string
		Expected pfd.NodeID
	}{
		"numbered": {
			Name:     "P1",
			Expected: "P1",
		},
		"numbered with description": {
			Name:     "P1: 設計する",
			Expected: "P1",
		},
		"numbered with spaces around the separator": {
			Name:     "P1 : 設計する",
			Expected: "P1",
		},
		"numbered with a colon in the description": {
			Name:     "P1: 設計する: 詳細",
			Expected: "P1",
		},
		"unnumbered": {
			Name:     "設計する",
			Expected: "(設計する)",
		},
		"full width colon is not a separator": {
			Name:     "P1： 設計する",
			Expected: "(P1： 設計する)",
		},
		"context diagram": {
			Name:     "P0: コンテキストダイアグラム",
			Expected: "P0",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := PageNameElementID(testCase.Name)
			if actual != testCase.Expected {
				t.Errorf("expected %q, but got %q", testCase.Expected, actual)
			}
		})
	}
}

func TestFormatVertexValue(t *testing.T) {
	testCases := map[string]struct {
		ID       pfd.NodeID
		Desc     string
		Expected string
	}{
		"numbered with description": {
			ID:       "P1",
			Desc:     "設計する",
			Expected: "P1: 設計する",
		},
		"numbered without description": {
			ID:       "P1",
			Desc:     "",
			Expected: "P1",
		},
		"unnumbered process without description": {
			ID:       "(設計する)",
			Desc:     "",
			Expected: "設計する",
		},
		"unnumbered deliverable without description": {
			ID:       "[設計書]",
			Desc:     "",
			Expected: "設計書",
		},
		"unnumbered with description": {
			ID:       "(設計する)",
			Desc:     "設計する",
			Expected: "設計する: 設計する",
		},
		"empty": {
			ID:       "",
			Desc:     "",
			Expected: "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := FormatVertexValue(testCase.ID, testCase.Desc)
			if actual != testCase.Expected {
				t.Errorf("expected %q, but got %q", testCase.Expected, actual)
			}
		})
	}
}

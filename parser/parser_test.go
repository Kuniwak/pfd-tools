package parser

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
)

func TestSkipRune(t *testing.T) {
	testCases := map[string]struct {
		Input         string
		Index         int
		SkippingRunes *sets.Set[rune]
		Expected      int
	}{
		"empty input": {
			Input:         "",
			Index:         0,
			SkippingRunes: sets.New(cmp.Compare, 'H'),
			Expected:      0,
		},
		"last": {
			Input:         "Hello, world!",
			Index:         12,
			SkippingRunes: sets.New(cmp.Compare, 'H'),
			Expected:      12,
		},
		"contains single skipping runes": {
			Input:         "Hello, world!",
			Index:         6,
			SkippingRunes: sets.New(cmp.Compare, ' '),
			Expected:      7,
		},
		"contains multiple skipping runes": {
			Input:         "Hello,    world!",
			Index:         6,
			SkippingRunes: sets.New(cmp.Compare, ' ', 'o'),
			Expected:      10,
		},
		"not contains skipping runes": {
			Input:         "Hello, world!",
			Index:         6,
			SkippingRunes: sets.New(cmp.Compare, 'X'),
			Expected:      6,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			newIndex := SkipRune(testCase.SkippingRunes, []rune(testCase.Input), testCase.Index)
			if newIndex != testCase.Expected {
				t.Errorf("SkipRune: expected %d, got %d", testCase.Expected, newIndex)
			}
		})
	}
}

func TestExpectKeyword(t *testing.T) {
	testCases := map[string]struct {
		Input            string
		Keyword          string
		Index            int
		ExpectedOk       bool
		ExpectedNewIndex int
	}{
		"empty input": {
			Input:            "",
			Keyword:          "Hello",
			Index:            0,
			ExpectedOk:       false,
			ExpectedNewIndex: 0,
		},
		"last": {
			Input:            "Hello, world!",
			Keyword:          "world",
			Index:            12,
			ExpectedOk:       false,
			ExpectedNewIndex: 12,
		},
		"not found": {
			Input:            "Hello, world!",
			Keyword:          "WORLD",
			Index:            7,
			ExpectedOk:       false,
			ExpectedNewIndex: 7,
		},
		"found at first": {
			Input:            "Hello, world",
			Keyword:          "Hello",
			Index:            0,
			ExpectedOk:       true,
			ExpectedNewIndex: 5,
		},
		"found at last": {
			Input:            "Hello, world",
			Keyword:          "world",
			Index:            7,
			ExpectedOk:       true,
			ExpectedNewIndex: 12,
		},
		"found at middle": {
			Input:            "Hello, world!",
			Keyword:          "world",
			Index:            7,
			ExpectedOk:       true,
			ExpectedNewIndex: 12,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ok, newIndex := ExpectKeyword([]rune(testCase.Keyword), []rune(testCase.Input), testCase.Index)
			if ok != testCase.ExpectedOk {
				t.Fatalf("ExpectKeyword: expected %t, got %t", testCase.ExpectedOk, ok)
			}
			if newIndex != testCase.ExpectedNewIndex {
				t.Fatalf("ExpectKeyword: expected %d, got %d", testCase.ExpectedNewIndex, newIndex)
			}
		})
	}
}

func TestIsDigit(t *testing.T) {
	testCases := map[string]struct {
		Input    rune
		Expected bool
	}{
		"not digit": {
			Input:    'a',
			Expected: false,
		},
	}
	for i := 0; i < 10; i++ {
		testCases[fmt.Sprintf("digit %d", i)] = struct {
			Input    rune
			Expected bool
		}{
			Input:    rune('0' + i),
			Expected: true,
		}
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := IsDigit(testCase.Input)
			if actual != testCase.Expected {
				t.Errorf("IsDigit: expected %t, got %t", testCase.Expected, actual)
			}
		})
	}
}

func TestAdvanceUntil(t *testing.T) {
	testCases := map[string]struct {
		Input            string
		Index            int
		Predicate        func(rune) bool
		Expected         []rune
		ExpectedNewIndex int
	}{
		"empty input": {
			Input:            "",
			Index:            0,
			Predicate:        IsDigit,
			Expected:         nil,
			ExpectedNewIndex: 0,
		},
		"last": {
			Input:            "0123456789",
			Index:            10,
			Predicate:        IsDigit,
			Expected:         nil,
			ExpectedNewIndex: 10,
		},
		"not found": {
			Input:            "abcdefghijklmnopqrstuvwxyz",
			Index:            10,
			Predicate:        IsDigit,
			Expected:         nil,
			ExpectedNewIndex: 10,
		},
		"found": {
			Input:            "0123456789abcd",
			Index:            4,
			Predicate:        IsDigit,
			Expected:         []rune("456789"),
			ExpectedNewIndex: 10,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, newIndex := AdvanceUntil(testCase.Predicate, []rune(testCase.Input), testCase.Index)
			if !slices.Equal(actual, testCase.Expected) {
				t.Fatalf("AdvanceUntil: expected %v, got %v", testCase.Expected, actual)
			}
			if newIndex != testCase.ExpectedNewIndex {
				t.Fatalf("AdvanceUntil: expected %d, got %d", testCase.ExpectedNewIndex, newIndex)
			}
		})
	}
}

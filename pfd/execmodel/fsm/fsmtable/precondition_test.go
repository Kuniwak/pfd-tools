package fsmtable

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/google/go-cmp/cmp"
)

func TestParsePrecondition(t *testing.T) {
	testCases := map[string]struct {
		Input   string
		Context pfd.AtomicProcessID
		Want    *fsm.Precondition
	}{
		"true without spaces": {
			Input: `\true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeTrue,
			},
		},
		"true with spaces": {
			Input: `\true `,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeTrue,
			},
		},
		"executable without spaces": {
			Input: `\exec(P123)`,
			Want: &fsm.Precondition{
				Type:       fsm.PreconditionTypeExecutable,
				Executable: pfd.AtomicProcessID("P123"),
			},
		},
		"executable with spaces": {
			Input: `\exec( P123 ) `,
			Want: &fsm.Precondition{
				Type:       fsm.PreconditionTypeExecutable,
				Executable: pfd.AtomicProcessID("P123"),
			},
		},
		"complete without spaces": {
			Input: `\complete(D123)`,
			Want: &fsm.Precondition{
				Type:           fsm.PreconditionTypeFeedbackSourceCompleted,
				FeedbackSource: pfd.AtomicDeliverableID("D123"),
			},
		},
		"complete with spaces": {
			Input: `\complete( D123 ) `,
			Want: &fsm.Precondition{
				Type:           fsm.PreconditionTypeFeedbackSourceCompleted,
				FeedbackSource: pfd.AtomicDeliverableID("D123"),
			},
		},
		"complete asterisk without spaces": {
			Input:   `\complete(*)`,
			Context: pfd.AtomicProcessID("P123"),
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted,
				AllBackwardReachableFeedbackSourcesCompletedTarget: pfd.AtomicProcessID("P123"),
			},
		},
		"complete asterisk with spaces": {
			Input:   `\complete( * ) `,
			Context: pfd.AtomicProcessID("P123"),
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeAllBackwardReachableFeedbackSourcesCompleted,
				AllBackwardReachableFeedbackSourcesCompletedTarget: pfd.AtomicProcessID("P123"),
			},
		},
		"or without spaces": {
			Input: `\true||\true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeOr, Or: []*fsm.Precondition{
					{Type: fsm.PreconditionTypeTrue},
					{Type: fsm.PreconditionTypeTrue},
				},
			},
		},
		"or with spaces": {
			Input: `\true || \true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeOr, Or: []*fsm.Precondition{
					{Type: fsm.PreconditionTypeTrue},
					{Type: fsm.PreconditionTypeTrue},
				},
			},
		},
		"and with spaces": {
			Input: `\true && \true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeAnd, And: []*fsm.Precondition{
					{Type: fsm.PreconditionTypeTrue},
					{Type: fsm.PreconditionTypeTrue},
				},
			},
		},
		"and without spaces": {
			Input: `\true&&\true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeAnd, And: []*fsm.Precondition{
					{Type: fsm.PreconditionTypeTrue},
					{Type: fsm.PreconditionTypeTrue},
				},
			},
		},
		"not without spaces": {
			Input: `!\true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeNot,
				Not: &fsm.Precondition{
					Type: fsm.PreconditionTypeTrue,
				},
			},
		},
		"not with spaces": {
			Input: `! \true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeNot,
				Not: &fsm.Precondition{
					Type: fsm.PreconditionTypeTrue,
				},
			},
		},
		"precedence 1": {
			Input: `\true && \true || \true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeOr,
				Or: []*fsm.Precondition{
					{
						Type: fsm.PreconditionTypeAnd,
						And: []*fsm.Precondition{
							{Type: fsm.PreconditionTypeTrue},
							{Type: fsm.PreconditionTypeTrue},
						},
					},
					{Type: fsm.PreconditionTypeTrue},
				},
			},
		},
		"precedence 2": {
			Input: `\true || \true && \true`,
			Want: &fsm.Precondition{
				Type: fsm.PreconditionTypeOr,
				Or: []*fsm.Precondition{
					{Type: fsm.PreconditionTypeTrue},
					{
						Type: fsm.PreconditionTypeAnd,
						And: []*fsm.Precondition{
							{Type: fsm.PreconditionTypeTrue},
							{Type: fsm.PreconditionTypeTrue},
						},
					},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := ParsePrecondition(testCase.Input, testCase.Context)
			if err != nil {
				t.Fatalf("ParsePrecondition: %v", err)
			}
			if !reflect.DeepEqual(got, testCase.Want) {
				t.Fatal(cmp.Diff(testCase.Want, got))
			}
		})
	}
}

func TestParseNodeID(t *testing.T) {
	testCases := map[string]struct {
		Input     string
		Want      pfd.NodeID
		WantIndex int
	}{
		"simple": {
			Input:     "P123",
			Want:      pfd.NodeID("P123"),
			WantIndex: 4,
		},
		"with spaces": {
			Input:     "P123 ",
			Want:      pfd.NodeID("P123"),
			WantIndex: 5,
		},
		"trailing hyphen": {
			Input:     "P12-",
			Want:      pfd.NodeID("P12-"),
			WantIndex: 4,
		},
		"trailing hyphen with not trailing space": {
			Input:     "P12-&",
			Want:      pfd.NodeID("P12-"),
			WantIndex: 4,
		},
		"trailing hyphen with space": {
			Input:     "P12- ",
			Want:      pfd.NodeID("P12-"),
			WantIndex: 5,
		},
		"trailing arrow": {
			Input:     "P12->",
			Want:      pfd.NodeID("P12"),
			WantIndex: 3,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			rs := []rune(testCase.Input)
			ok, got, gotIndex := parseNodeID(rs, 0)
			if !ok {
				t.Fatalf("parseNodeID: failed to parse node ID: %q", testCase.Input)
			}
			if got != testCase.Want {
				t.Errorf("want: %q, got: %q", testCase.Want, got)
			}
			if gotIndex != testCase.WantIndex {
				t.Errorf("want index: %d, got: %d", testCase.WantIndex, gotIndex)
			}
		})
	}
}

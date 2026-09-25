package pfd

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/google/go-cmp/cmp"
)

func TestRenumberPlan(t *testing.T) {
	testCases := map[string]struct {
		PFD           *PFD
		ExpectedOneOf []RenumberPlan
	}{
		"empty": {
			PFD: &PFD{
				Nodes: sets.New((*Node).Compare),
				Edges: sets.New((*Edge).Compare),
			},
			ExpectedOneOf: []RenumberPlan{{}},
		},
		"simplest": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "入力成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "プロセス", Type: NodeTypeAtomicProcess},
					&Node{ID: "出力成果物", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "入力成果物", Target: "プロセス"},
					&Edge{Source: "プロセス", Target: "出力成果物"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"出力成果物": {ID: "D1", Description: "出力成果物", Type: NodeTypeAtomicDeliverable},
					"入力成果物": {ID: "D2", Description: "入力成果物", Type: NodeTypeAtomicDeliverable},
					"プロセス":  {ID: "P1", Description: "プロセス", Type: NodeTypeAtomicProcess},
				},
			},
		},

		"the same label as a process and a deliverable": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: NewUnnumberedNodeID(NodeTypeAtomicDeliverable, "要件"), Type: NodeTypeAtomicDeliverable},
					&Node{ID: NewUnnumberedNodeID(NodeTypeAtomicProcess, "実装する"), Type: NodeTypeAtomicProcess},
					&Node{ID: NewUnnumberedNodeID(NodeTypeAtomicDeliverable, "実装する"), Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "[要件]", Target: "(実装する)"},
					&Edge{Source: "(実装する)", Target: "[実装する]"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"[実装する]": {ID: "D1", Description: "実装する", Type: NodeTypeAtomicDeliverable},
					"[要件]":   {ID: "D2", Description: "要件", Type: NodeTypeAtomicDeliverable},
					"(実装する)": {ID: "P1", Description: "実装する", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"like Y shape": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "入力成果物1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "入力成果物2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "プロセス", Type: NodeTypeAtomicProcess},
					&Node{ID: "出力成果物", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "入力成果物1", Target: "プロセス"},
					&Edge{Source: "入力成果物2", Target: "プロセス"},
					&Edge{Source: "プロセス", Target: "出力成果物"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"出力成果物":  {ID: "D1", Description: "出力成果物", Type: NodeTypeAtomicDeliverable},
					"入力成果物1": {ID: "D2", Description: "入力成果物1", Type: NodeTypeAtomicDeliverable},
					"入力成果物2": {ID: "D3", Description: "入力成果物2", Type: NodeTypeAtomicDeliverable},
					"プロセス":   {ID: "P1", Description: "プロセス", Type: NodeTypeAtomicProcess},
				},
				{
					"出力成果物":  {ID: "D1", Description: "出力成果物", Type: NodeTypeAtomicDeliverable},
					"入力成果物2": {ID: "D2", Description: "入力成果物2", Type: NodeTypeAtomicDeliverable},
					"入力成果物1": {ID: "D3", Description: "入力成果物1", Type: NodeTypeAtomicDeliverable},
					"プロセス":   {ID: "P1", Description: "プロセス", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"like reversed Y shape": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "入力成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "プロセス", Type: NodeTypeAtomicProcess},
					&Node{ID: "出力成果物1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "出力成果物2", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "入力成果物", Target: "プロセス"},
					&Edge{Source: "プロセス", Target: "出力成果物1"},
					&Edge{Source: "プロセス", Target: "出力成果物2"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"出力成果物1": {ID: "D1", Description: "出力成果物1", Type: NodeTypeAtomicDeliverable},
					"出力成果物2": {ID: "D2", Description: "出力成果物2", Type: NodeTypeAtomicDeliverable},
					"入力成果物":  {ID: "D3", Description: "入力成果物", Type: NodeTypeAtomicDeliverable},
					"プロセス":   {ID: "P1", Description: "プロセス", Type: NodeTypeAtomicProcess},
				},
				{
					"出力成果物2": {ID: "D1", Description: "出力成果物2", Type: NodeTypeAtomicDeliverable},
					"出力成果物1": {ID: "D2", Description: "出力成果物1", Type: NodeTypeAtomicDeliverable},
					"入力成果物":  {ID: "D3", Description: "入力成果物", Type: NodeTypeAtomicDeliverable},
					"プロセス":   {ID: "P1", Description: "プロセス", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"sequential": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "入力成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "プロセス1", Type: NodeTypeAtomicProcess},
					&Node{ID: "中間成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "プロセス2", Type: NodeTypeAtomicProcess},
					&Node{ID: "出力成果物", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "入力成果物", Target: "プロセス1"},
					&Edge{Source: "プロセス1", Target: "中間成果物"},
					&Edge{Source: "中間成果物", Target: "プロセス2"},
					&Edge{Source: "プロセス2", Target: "出力成果物"},
				),
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"出力成果物": {ID: "D1", Description: "出力成果物", Type: NodeTypeAtomicDeliverable},
					"入力成果物": {ID: "D2", Description: "入力成果物", Type: NodeTypeAtomicDeliverable},
					"中間成果物": {ID: "D3", Description: "中間成果物", Type: NodeTypeAtomicDeliverable},
					"プロセス1": {ID: "P1", Description: "プロセス1", Type: NodeTypeAtomicProcess},
					"プロセス2": {ID: "P2", Description: "プロセス2", Type: NodeTypeAtomicProcess},
				},
			},
		},
		"composite process": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "入力成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "中間成果物1", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "中間成果物2", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "中間成果物3", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "出力成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "原子プロセス1", Type: NodeTypeAtomicProcess},
					&Node{ID: "原子プロセス2", Type: NodeTypeAtomicProcess},
					&Node{ID: "原子プロセス3", Type: NodeTypeAtomicProcess},
					&Node{ID: "原子プロセス4", Type: NodeTypeAtomicProcess},
					&Node{ID: "複合プロセス1", Type: NodeTypeCompositeProcess},
					&Node{ID: "複合プロセス2", Type: NodeTypeCompositeProcess},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "入力成果物", Target: "原子プロセス1"},
					&Edge{Source: "原子プロセス1", Target: "中間成果物1"},
					&Edge{Source: "中間成果物1", Target: "原子プロセス2"},
					&Edge{Source: "原子プロセス2", Target: "中間成果物2"},
					&Edge{Source: "中間成果物2", Target: "原子プロセス3"},
					&Edge{Source: "原子プロセス3", Target: "中間成果物3"},
					&Edge{Source: "中間成果物3", Target: "原子プロセス4"},
					&Edge{Source: "原子プロセス4", Target: "出力成果物"},

					&Edge{Source: "入力成果物", Target: "複合プロセス1"},
					&Edge{Source: "複合プロセス1", Target: "中間成果物2"},
					&Edge{Source: "中間成果物2", Target: "複合プロセス2"},
					&Edge{Source: "複合プロセス2", Target: "出力成果物"},
				),
				ProcessComposition: map[NodeID]*sets.Set[NodeID]{
					"複合プロセス1": sets.New(NodeID.Compare, "原子プロセス1", "原子プロセス2"),
					"複合プロセス2": sets.New(NodeID.Compare, "原子プロセス3", "原子プロセス4"),
				},
			},
			ExpectedOneOf: []RenumberPlan{
				{
					"出力成果物":   {ID: "D1", Description: "出力成果物", Type: NodeTypeAtomicDeliverable},
					"入力成果物":   {ID: "D2", Description: "入力成果物", Type: NodeTypeAtomicDeliverable},
					"中間成果物1":  {ID: "D3", Description: "中間成果物1", Type: NodeTypeAtomicDeliverable},
					"中間成果物2":  {ID: "D4", Description: "中間成果物2", Type: NodeTypeAtomicDeliverable},
					"中間成果物3":  {ID: "D5", Description: "中間成果物3", Type: NodeTypeAtomicDeliverable},
					"原子プロセス1": {ID: "P1", Description: "原子プロセス1", Type: NodeTypeAtomicProcess},
					"原子プロセス2": {ID: "P2", Description: "原子プロセス2", Type: NodeTypeAtomicProcess},
					"原子プロセス3": {ID: "P3", Description: "原子プロセス3", Type: NodeTypeAtomicProcess},
					"原子プロセス4": {ID: "P4", Description: "原子プロセス4", Type: NodeTypeAtomicProcess},
					"複合プロセス1": {ID: "P5", Description: "複合プロセス1", Type: NodeTypeCompositeProcess},
					"複合プロセス2": {ID: "P6", Description: "複合プロセス2", Type: NodeTypeCompositeProcess},
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodeMap := NewNodeMap(testCase.PFD.Nodes, logger)
			graphExceptFB := testCase.PFD.GraphExceptFeedback(nodeMap, logger)
			errs := make([]Error, 0)

			plan, ok := NewRenumberPlan(testCase.PFD, graphExceptFB, nodeMap, &errs)
			if !ok {
				t.Fatalf("NewRenumberPlan: expected true, got false: %s", Errors(errs).Error())
			}

			matched := false
			for _, expected := range testCase.ExpectedOneOf {
				if reflect.DeepEqual(plan, expected) {
					matched = true
					break
				}
			}
			if !matched {
				for _, expected := range testCase.ExpectedOneOf {
					t.Log(cmp.Diff(expected, plan))
				}
				t.Error("expected one of the plans")
			}
		})
	}
}

func TestRenumberPlanError(t *testing.T) {
	testCases := map[string]struct {
		PFD *PFD
	}{
		"loop": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "入力成果物", Type: NodeTypeAtomicDeliverable},
					&Node{ID: "プロセス", Type: NodeTypeAtomicProcess},
					&Node{ID: "出力成果物", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New(
					(*Edge).Compare,
					&Edge{Source: "入力成果物", Target: "プロセス"},
					&Edge{Source: "プロセス", Target: "出力成果物"},
					&Edge{Source: "出力成果物", Target: "プロセス"},
				),
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodeMap := NewNodeMap(testCase.PFD.Nodes, logger)
			graphExceptFB := testCase.PFD.GraphExceptFeedback(nodeMap, logger)

			errs := make([]Error, 0)

			_, ok := NewRenumberPlan(testCase.PFD, graphExceptFB, nodeMap, &errs)
			if ok {
				t.Errorf("NewRenumberPlan: expected false, got true: %s", Errors(errs).Error())
			}
		})
	}
}

func TestNewRenumberPlanWithBaseError(t *testing.T) {
	testCases := map[string]struct {
		PFD  *PFD
		Base RenumberBase
	}{

		"the base assigns a deliverable ID to a process": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "(設計する)", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New((*Edge).Compare),
			},
			Base: RenumberBase{Plan: RenumberPlan{
				"(設計する)": {ID: "D9", Description: "設計する"},
			}},
		},

		"the base assigns a process ID to a deliverable": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "[設計書]", Type: NodeTypeAtomicDeliverable},
				),
				Edges: sets.New((*Edge).Compare),
			},
			Base: RenumberBase{Plan: RenumberPlan{
				"[設計書]": {ID: "P9", Description: "設計書"},
			}},
		},

		"the base ID collides with an assigned ID in the input": {
			PFD: &PFD{
				Nodes: sets.New(
					(*Node).Compare,
					&Node{ID: "P1", Description: "実装する", Type: NodeTypeAtomicProcess},
					&Node{ID: "(設計する)", Type: NodeTypeAtomicProcess},
				),
				Edges: sets.New((*Edge).Compare),
			},
			Base: RenumberBase{Plan: RenumberPlan{
				"(設計する)": {ID: "P1", Description: "設計する"},
			}},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodeMap := NewNodeMap(testCase.PFD.Nodes, logger)
			graphExceptFB := testCase.PFD.GraphExceptFeedback(nodeMap, logger)

			errs := make([]Error, 0)
			if _, ok := NewRenumberPlanWithBase(testCase.PFD, graphExceptFB, nodeMap, testCase.Base, &errs); ok {
				t.Error("NewRenumberPlanWithBase: ok = true, want false")
			}
		})
	}
}

func TestRenumberPlanCompositeTieBreak(t *testing.T) {
	p := &PFD{
		Nodes: sets.New(
			(*Node).Compare,
			&Node{ID: "入力成果物", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "中間成果物1", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "中間成果物2", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "出力成果物", Type: NodeTypeAtomicDeliverable},
			&Node{ID: "原子プロセス1", Type: NodeTypeAtomicProcess},
			&Node{ID: "原子プロセス2", Type: NodeTypeAtomicProcess},
			&Node{ID: "原子プロセス3", Type: NodeTypeAtomicProcess},
			&Node{ID: "複合プロセスA", Type: NodeTypeCompositeProcess},
			&Node{ID: "複合プロセスB", Type: NodeTypeCompositeProcess},
		),
		Edges: sets.New(
			(*Edge).Compare,
			&Edge{Source: "入力成果物", Target: "原子プロセス1"},
			&Edge{Source: "原子プロセス1", Target: "中間成果物1"},
			&Edge{Source: "中間成果物1", Target: "原子プロセス2"},
			&Edge{Source: "原子プロセス2", Target: "中間成果物2"},
			&Edge{Source: "中間成果物2", Target: "原子プロセス3"},
			&Edge{Source: "原子プロセス3", Target: "出力成果物"},
		),

		ProcessComposition: map[NodeID]*sets.Set[NodeID]{
			"複合プロセスA": sets.New(NodeID.Compare, "原子プロセス1", "原子プロセス2", "原子プロセス3"),
			"複合プロセスB": sets.New(NodeID.Compare, "原子プロセス1", "原子プロセス2"),
		},
	}

	expected := RenumberPlan{
		"出力成果物":   {ID: "D1", Description: "出力成果物", Type: NodeTypeAtomicDeliverable},
		"入力成果物":   {ID: "D2", Description: "入力成果物", Type: NodeTypeAtomicDeliverable},
		"中間成果物1":  {ID: "D3", Description: "中間成果物1", Type: NodeTypeAtomicDeliverable},
		"中間成果物2":  {ID: "D4", Description: "中間成果物2", Type: NodeTypeAtomicDeliverable},
		"原子プロセス1": {ID: "P1", Description: "原子プロセス1", Type: NodeTypeAtomicProcess},
		"原子プロセス2": {ID: "P2", Description: "原子プロセス2", Type: NodeTypeAtomicProcess},
		"原子プロセス3": {ID: "P3", Description: "原子プロセス3", Type: NodeTypeAtomicProcess},
		"複合プロセスA": {ID: "P4", Description: "複合プロセスA", Type: NodeTypeCompositeProcess},
		"複合プロセスB": {ID: "P5", Description: "複合プロセスB", Type: NodeTypeCompositeProcess},
	}

	logger := slog.New(slogtest.NewTestHandler(t))
	for i := 0; i < 100; i++ {
		nodeMap := NewNodeMap(p.Nodes, logger)
		graphExceptFB := p.GraphExceptFeedback(nodeMap, logger)
		errs := make([]Error, 0)

		plan, ok := NewRenumberPlan(p, graphExceptFB, nodeMap, &errs)
		if !ok {
			t.Fatalf("NewRenumberPlan: expected true, got false: %s", Errors(errs).Error())
		}
		if !reflect.DeepEqual(plan, expected) {
			t.Fatalf("run %d: unexpected plan:\n%s", i, cmp.Diff(expected, plan))
		}
	}
}

func TestRenumberPlanNewlyNumbered(t *testing.T) {
	testCases := map[string]struct {
		Plan RenumberPlan
		Want RenumberPlan
	}{
		"a newly numbered entry is kept": {
			Plan: RenumberPlan{
				"テストを書く": &Node{ID: "P12", Description: "テストを書く", Type: NodeTypeAtomicProcess},
			},
			Want: RenumberPlan{
				"テストを書く": &Node{ID: "P12", Description: "テストを書く", Type: NodeTypeAtomicProcess},
			},
		},
		"an already numbered entry is dropped": {
			Plan: RenumberPlan{
				"P1": &Node{ID: "P1", Description: "要件を洗い出す", Type: NodeTypeAtomicProcess},
			},
			Want: RenumberPlan{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.Want, tc.Plan.NewlyNumbered()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestParseMaxIDs(t *testing.T) {
	testCases := map[string]struct {
		Text                  string
		WantProcessNumber     int
		WantDeliverableNumber int
		WantErr               bool
	}{
		"empty text yields no floor": {
			Text: "",
		},
		"the output of FormatMaxIDs": {
			Text:                  FormatMaxIDs(12, 34),
			WantProcessNumber:     12,
			WantDeliverableNumber: 34,
		},
		"a comma separated list": {
			Text:                  "P12, D34",
			WantProcessNumber:     12,
			WantDeliverableNumber: 34,
		},
		"only one kind": {
			Text:              "P12",
			WantProcessNumber: 12,
		},

		"multiple lines yield the max of each kind": {
			Text:                  FormatMaxIDs(12, 0) + "\n" + FormatMaxIDs(0, 34) + "\n",
			WantProcessNumber:     12,
			WantDeliverableNumber: 34,
		},
		"an unparsable ID is an error": {
			Text:    "X1",
			WantErr: true,
		},
		"a number without a prefix is an error": {
			Text:    "12",
			WantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotProcess, gotDeliverable, err := ParseMaxIDs(tc.Text)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("ParseMaxIDs: err = nil, want an error (got %d, %d)", gotProcess, gotDeliverable)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMaxIDs: %v", err)
			}
			if gotProcess != tc.WantProcessNumber || gotDeliverable != tc.WantDeliverableNumber {
				t.Errorf("ParseMaxIDs = (%d, %d), want (%d, %d)", gotProcess, gotDeliverable, tc.WantProcessNumber, tc.WantDeliverableNumber)
			}
		})
	}
}

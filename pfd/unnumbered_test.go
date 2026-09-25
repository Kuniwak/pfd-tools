package pfd

import "testing"

func TestNewUnnumberedNodeID(t *testing.T) {
	testCases := map[string]struct {
		Type  NodeType
		Label string
		Want  NodeID
	}{
		"an atomic process is wrapped in parentheses": {
			Type:  NodeTypeAtomicProcess,
			Label: "実装する",
			Want:  "(実装する)",
		},
		"a composite process uses the same wrapping as an atomic process": {
			Type:  NodeTypeCompositeProcess,
			Label: "実装する",
			Want:  "(実装する)",
		},
		"an atomic deliverable is wrapped in brackets": {
			Type:  NodeTypeAtomicDeliverable,
			Label: "設計書",
			Want:  "[設計書]",
		},
		"a composite deliverable uses the same wrapping as an atomic deliverable": {
			Type:  NodeTypeCompositeDeliverable,
			Label: "設計書",
			Want:  "[設計書]",
		},

		"an empty label is not wrapped": {
			Type:  NodeTypeAtomicProcess,
			Label: "",
			Want:  "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := NewUnnumberedNodeID(testCase.Type, testCase.Label); got != testCase.Want {
				t.Errorf("NewUnnumberedNodeID(%q, %q) = %q, want %q", testCase.Type, testCase.Label, got, testCase.Want)
			}
		})
	}
}

func TestNodeIDLabel(t *testing.T) {
	testCases := map[string]struct {
		ID   NodeID
		Want string
	}{
		"a process ID is unwrapped": {
			ID:   "(実装する)",
			Want: "実装する",
		},
		"a deliverable ID is unwrapped": {
			ID:   "[設計書]",
			Want: "設計書",
		},
		"an assigned ID is returned as is": {
			ID:   "P1",
			Want: "P1",
		},
		"a bare label is returned as is": {
			ID:   "実装する",
			Want: "実装する",
		},
		"an unbalanced wrapping is returned as is": {
			ID:   "(実装する",
			Want: "(実装する",
		},
		"a label containing parentheses keeps them": {
			ID:   "((実装する))",
			Want: "(実装する)",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := testCase.ID.Label(); got != testCase.Want {
				t.Errorf("NodeID(%q).Label() = %q, want %q", testCase.ID, got, testCase.Want)
			}
		})
	}
}

func TestNodeIDUnnumberedNodeType(t *testing.T) {
	testCases := map[string]struct {
		ID     NodeID
		Want   NodeType
		WantOK bool
	}{
		"a wrapped process is a process": {
			ID:     "(実装する)",
			Want:   NodeTypeAtomicProcess,
			WantOK: true,
		},
		"a wrapped deliverable is a deliverable": {
			ID:     "[設計書]",
			Want:   NodeTypeAtomicDeliverable,
			WantOK: true,
		},
		"an assigned ID has no wrapping": {
			ID:     "P1",
			WantOK: false,
		},
		"a bare label has no wrapping": {
			ID:     "実装する",
			WantOK: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			got, ok := testCase.ID.UnnumberedNodeType()
			if ok != testCase.WantOK {
				t.Fatalf("NodeID(%q).UnnumberedNodeType() ok = %v, want %v", testCase.ID, ok, testCase.WantOK)
			}
			if ok && got != testCase.Want {
				t.Errorf("NodeID(%q).UnnumberedNodeType() = %q, want %q", testCase.ID, got, testCase.Want)
			}
		})
	}
}

func TestNewUnnumberedNodeIDRoundTrip(t *testing.T) {
	testCases := map[string]struct {
		Type  NodeType
		Label string
	}{
		"a process label":                     {Type: NodeTypeAtomicProcess, Label: "実装する"},
		"a deliverable label":                 {Type: NodeTypeAtomicDeliverable, Label: "設計書"},
		"a label that looks like a process":   {Type: NodeTypeAtomicDeliverable, Label: "(実装する)"},
		"a label that looks like an assigned": {Type: NodeTypeAtomicProcess, Label: "P1"},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			id := NewUnnumberedNodeID(testCase.Type, testCase.Label)
			if got := id.Label(); got != testCase.Label {
				t.Errorf("NewUnnumberedNodeID(%q, %q).Label() = %q, want %q", testCase.Type, testCase.Label, got, testCase.Label)
			}
		})
	}
}

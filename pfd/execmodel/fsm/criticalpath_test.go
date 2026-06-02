package fsm

import (
	"log/slog"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestPrefixUntilProcessStart(t *testing.T) {
	type testCase struct {
		name     string
		ap       pfd.AtomicProcessID
		wantLen  int
		wantDesc string
	}
	testCases := []testCase{
		{
			name:     "process in first transition",
			ap:       "P1",
			wantLen:  0,
			wantDesc: "prefix should be empty when target is in first allocation",
		},
		{
			name:     "process not in plan",
			ap:       "P_unknown",
			wantLen:  -1, // special: same as full plan
			wantDesc: "prefix should be full plan when target is not found",
		},
	}

	// Build a simple plan with known transitions for testing
	// [D1] -> (P1) -> [D2] -> (P2) -> [D3]
	p := newSafePFDByUnsafePFD(&pfd.PFD{
		Nodes: sets.New(
			(*pfd.Node).Compare,
			&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "P1", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "P2", Type: pfd.NodeTypeAtomicProcess},
		),
		Edges: sets.New(
			(*pfd.Edge).Compare,
			&pfd.Edge{Source: "D1", Target: "P1"},
			&pfd.Edge{Source: "P1", Target: "D2"},
			&pfd.Edge{Source: "D2", Target: "P2"},
			&pfd.Edge{Source: "P2", Target: "D3"},
		),
	})
	initVolume := Volume(1)
	initVolumeFunc := ConstInitialVolumeFunc(initVolume)
	logger := slog.New(slogtest.NewTestHandler(t))
	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"P1": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume}),
		"P2": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: initVolume}),
	})
	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1"),
		NewAvailableAllocationsFunc(neededResourceSetsFunc),
		initVolumeFunc,
		ExponentialReworkVolumeFunc(0, initVolumeFunc),
		ConstMaxRevisionMap(1, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		AlwaysAvailableTimeFunc(),
		logger,
	)

	plans, err := searchBestPlans(env)
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := plans.At(0)
	if !ok {
		t.Fatal("no plan found")
	}

	// P2 depends on P1, so P2 should appear after P1 in the plan
	testCases = append(testCases, testCase{
		name:     "process after first transition",
		ap:       "P2",
		wantLen:  1,
		wantDesc: "prefix should contain only the first transition (P1 allocation)",
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prefix := PrefixUntilProcessStart(plan, tc.ap)
			expectedLen := tc.wantLen
			if expectedLen == -1 {
				expectedLen = plan.Len()
			}
			if prefix.Len() != expectedLen {
				t.Errorf("%s: got len=%d, want %d", tc.wantDesc, prefix.Len(), expectedLen)
			}
		})
	}
}

// TestMaximumElasticityWithResourceConflict tests that the iterative refinement
// correctly computes total float (maximum elasticity) when resource conflicts exist.
//
// PFD structure:
//
//	[D1] → (Pa) → [D2] → (Pb) → [D3] → (Pc) → [D4]
//	[D5] → (Pd) → [D6] → (Pe) → [D7]
//
// Pa: volume=5, resource=R2
// Pb: volume=2, resource=R2
// Pc: volume=3, resource=R1  (shared with Pe)
// Pd: volume=1, resource=R3  (independent of Pa/Pb path)
// Pe: volume=3, resource=R1  (shared with Pc)
//
// Baseline schedule (leadtime=10):
//   t=0..5: Pa(R2) + Pd(R3) in parallel
//   t=1: Pd done, Pe(R1) starts. t=1..4: Pe(R1)
//   t=5..7: Pb(R2)
//   t=7..10: Pc(R1) (R1 free since t=4)
//
// Pa→Pb→Pc is the critical path (5+2+3=10). Pd→Pe has slack (1+3=4, finishes at t=4).
// Pc and Pe share R1, so Pd has total float = 3 (Pe finishes at t=4, Pc starts at t=7).
//
// With the OLD algorithm (extra=leadtime=10 for Pd):
//   Pd volume becomes 11, Pd runs t=0..11, Pe delayed to t=11..14.
//   Meanwhile Pc runs t=7..10 without R1 conflict (Pe hasn't started).
//   Leadtime=14, extension=4, float=10-4=6 (WRONG, should be 3).
//
// With the NEW iterative algorithm, it converges to the correct float=3.
func TestMaximumElasticityWithResourceConflict(t *testing.T) {
	p := newSafePFDByUnsafePFD(&pfd.PFD{
		Nodes: sets.New(
			(*pfd.Node).Compare,
			&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D5", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D6", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D7", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "Pa", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "Pb", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "Pc", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "Pd", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "Pe", Type: pfd.NodeTypeAtomicProcess},
		),
		Edges: sets.New(
			(*pfd.Edge).Compare,
			&pfd.Edge{Source: "D1", Target: "Pa"},
			&pfd.Edge{Source: "Pa", Target: "D2"},
			&pfd.Edge{Source: "D2", Target: "Pb"},
			&pfd.Edge{Source: "Pb", Target: "D3"},
			&pfd.Edge{Source: "D3", Target: "Pc"},
			&pfd.Edge{Source: "Pc", Target: "D4"},
			&pfd.Edge{Source: "D5", Target: "Pd"},
			&pfd.Edge{Source: "Pd", Target: "D6"},
			&pfd.Edge{Source: "D6", Target: "Pe"},
			&pfd.Edge{Source: "Pe", Target: "D7"},
		),
	})

	initVolumeFunc := func(ap pfd.AtomicProcessID) Volume {
		switch ap {
		case "Pa":
			return 5
		case "Pb":
			return 2
		case "Pc":
			return 3
		case "Pd":
			return 1
		case "Pe":
			return 3
		default:
			panic("unknown process: " + string(ap))
		}
	}

	logger := slog.New(slogtest.NewTestHandler(t))

	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"Pa": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 1}),
		"Pb": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 1}),
		"Pc": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
		"Pd": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: 1}),
		"Pe": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 1}),
	})

	env := NewEnv(
		p,
		sets.New(ResourceID.Compare, "R1", "R2", "R3"),
		NewAvailableAllocationsFunc(neededResourceSetsFunc),
		initVolumeFunc,
		ExponentialReworkVolumeFunc(0, initVolumeFunc),
		ConstMaxRevisionMap(1, p.FeedbackSourceDeliverables()),
		NewPreconditionMap(p.AtomicProcesses, map[pfd.AtomicProcessID]*Precondition{}),
		neededResourceSetsFunc,
		AlwaysAvailableTimeFunc(),
		logger,
	)

	searchFunc := SearchBestPlansWithPrefix()
	criticalPathInfoFunc := NewCriticalPathInfoFunc(searchFunc)
	info, err := criticalPathInfoFunc(env)
	if err != nil {
		t.Fatal(err)
	}

	// Verify baseline leadtime first
	basePlans, err := searchBestPlans(env)
	if err != nil {
		t.Fatal(err)
	}
	basePlan, _ := basePlans.At(0)
	baseLeadtime := basePlan.Leadtime()
	t.Logf("baseline leadtime: %v", baseLeadtime)
	for ap, item := range info {
		t.Logf("%s: max_elasticity=%v, min_elasticity=%v (has=%v)", ap, item.MaximumElasticity, item.MinimumElasticity, item.HasMinimumElasticity)
	}

	// Baseline: Pa(R2)+Pd(R3) in parallel from t=0.
	// Pd done at t=1, Pe(R1) starts t=1..4. Pa done at t=5, Pb(R2) t=5..7.
	// Pc(R1) starts at t=7 (R1 free since t=4). Leadtime = 10.
	if baseLeadtime != 10 {
		t.Fatalf("baseline leadtime: got %v, want 10", baseLeadtime)
	}

	// Pa is on the critical path: increasing Pa by any amount increases the leadtime.
	if info["Pa"].MaximumElasticity != 0 {
		t.Errorf("Pa total float: got %v, want 0", info["Pa"].MaximumElasticity)
	}

	// Pd has total float = 3: Pe finishes at t=4, Pc starts at t=7, gap = 3.
	if info["Pd"].MaximumElasticity != 3 {
		t.Errorf("Pd total float: got %v, want 3", info["Pd"].MaximumElasticity)
	}

	// Pe also has total float = 3 (same slack as Pd).
	if info["Pe"].MaximumElasticity != 3 {
		t.Errorf("Pe total float: got %v, want 3", info["Pe"].MaximumElasticity)
	}

	// All total floats should be >= 0 (no negative values)
	for ap, item := range info {
		if item.MaximumElasticity < 0 {
			t.Errorf("%s: total float is negative: %v", ap, item.MaximumElasticity)
		}
	}

	// Verify total float correctness by actually extending each process and re-running
	for ap, item := range info {
		totalFloat := item.MaximumElasticity
		if totalFloat <= 0 {
			continue
		}

		// Extending by exactly totalFloat should NOT increase leadtime
		envExtended := env.Clone()
		capturedAP := ap
		capturedFloat := totalFloat
		envExtended.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
			if ap2 == capturedAP {
				return initVolumeFunc(ap2) + Volume(capturedFloat)
			}
			return initVolumeFunc(ap2)
		}
		extPlans, err := searchBestPlans(envExtended)
		if err != nil {
			t.Fatalf("extending %s by %v: %v", ap, totalFloat, err)
		}
		extPlan, _ := extPlans.At(0)
		if extPlan.Leadtime() > baseLeadtime {
			t.Errorf("%s: extending by totalFloat=%v increased leadtime from %v to %v",
				ap, totalFloat, baseLeadtime, extPlan.Leadtime())
		}

		// Extending by totalFloat + 1 SHOULD increase leadtime
		envOverExtended := env.Clone()
		capturedExtraFloat := totalFloat + execmodel.Time(1)
		envOverExtended.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
			if ap2 == capturedAP {
				return initVolumeFunc(ap2) + Volume(capturedExtraFloat)
			}
			return initVolumeFunc(ap2)
		}
		overExtPlans, err := searchBestPlans(envOverExtended)
		if err != nil {
			t.Fatalf("over-extending %s by %v: %v", ap, capturedExtraFloat, err)
		}
		overExtPlan, _ := overExtPlans.At(0)
		if overExtPlan.Leadtime() <= baseLeadtime {
			t.Errorf("%s: extending by totalFloat+1=%v did NOT increase leadtime (still %v)",
				ap, capturedExtraFloat, overExtPlan.Leadtime())
		}
	}
}

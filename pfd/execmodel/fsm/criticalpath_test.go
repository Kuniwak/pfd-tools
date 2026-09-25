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
			wantLen:  -1,
			wantDesc: "prefix should be full plan when target is not found",
		},
	}

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

func TestMaximumElasticityWithResourceConflict(t *testing.T) {

	searchFuncs := map[string]SearchWithPrefixFunc{
		"poorest": SearchPoorestWithPrefix(1),
		"poor":    SearchFastestWithPrefix(1),
		"best":    SearchBestPlansWithPrefix(),
	}
	rates := map[string]Volume{
		"consumed volume 1": 1,
		"consumed volume 2": 2,
	}
	for rateName, rate := range rates {
		for searchName, searchFunc := range searchFuncs {
			t.Run(rateName+"/"+searchName, func(t *testing.T) {
				assertMaximumElasticityWithResourceConflict(t, newResourceConflictEnv(t, rate), searchFunc)
			})
		}
	}
}

func newResourceConflictEnv(t *testing.T, rate Volume) *Env {
	t.Helper()

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
			return 5 * rate
		case "Pb":
			return 2 * rate
		case "Pc":
			return 3 * rate
		case "Pd":
			return 1 * rate
		case "Pe":
			return 3 * rate
		default:
			panic("unknown process: " + string(ap))
		}
	}

	logger := slog.New(slogtest.NewTestHandler(t))

	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"Pa": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: rate}),
		"Pb": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: rate}),
		"Pc": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: rate}),
		"Pd": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R3"), ConsumedVolume: rate}),
		"Pe": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: rate}),
	})

	return NewEnv(
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
}

func assertMaximumElasticityWithResourceConflict(t *testing.T, env *Env, searchFunc SearchWithPrefixFunc) {
	t.Helper()

	info, err := NewCriticalPathInfoFunc(searchFunc)(env)
	if err != nil {
		t.Fatal(err)
	}

	basePlans, err := searchFunc(env, nil)
	if err != nil {
		t.Fatal(err)
	}
	basePlan, _ := basePlans.At(0)
	baseLeadtime := basePlan.Leadtime()
	t.Logf("baseline leadtime: %v", baseLeadtime)
	for ap, item := range info {
		t.Logf("%s: max_elasticity=%v, min_elasticity=%v (has=%v)", ap, item.MaximumElasticity, item.MinimumElasticity, item.HasMinimumElasticity)
	}

	if baseLeadtime != 10 {
		t.Fatalf("baseline leadtime: got %v, want 10", baseLeadtime)
	}

	if !info["Pa"].MaximumElasticity.ApproximateEqual(0) {
		t.Errorf("Pa total float: got %v, want 0", info["Pa"].MaximumElasticity)
	}

	if !info["Pd"].MaximumElasticity.ApproximateEqual(3) {
		t.Errorf("Pd total float: got %v, want 3", info["Pd"].MaximumElasticity)
	}

	if !info["Pe"].MaximumElasticity.ApproximateEqual(3) {
		t.Errorf("Pe total float: got %v, want 3", info["Pe"].MaximumElasticity)
	}

	for ap, item := range info {
		if item.MaximumElasticity < 0 {
			t.Errorf("%s: total float is negative: %v", ap, item.MaximumElasticity)
		}
	}

	for ap, item := range info {
		rate := MaxConsumedVolume(env.NeededResourceSetsFunc(ap))
		totalFloat := item.MaximumElasticity
		if totalFloat <= 0 || totalFloat.ApproximateEqual(0) {
			continue
		}

		leadtime := planWithExtraVolume(t, env, searchFunc, basePlan, ap, Volume(totalFloat)*rate).Leadtime()
		if leadtime > baseLeadtime && !leadtime.ApproximateEqual(baseLeadtime) {
			t.Errorf("%s: extending by totalFloat=%v increased leadtime from %v to %v",
				ap, totalFloat, baseLeadtime, leadtime)
		}

		overFloat := totalFloat + execmodel.Time(1)
		overLeadtime := planWithExtraVolume(t, env, searchFunc, basePlan, ap, Volume(overFloat)*rate).Leadtime()
		if overLeadtime <= baseLeadtime || overLeadtime.ApproximateEqual(baseLeadtime) {
			t.Errorf("%s: extending by totalFloat+1=%v did NOT increase leadtime (still %v)",
				ap, overFloat, overLeadtime)
		}
	}
}

func planWithExtraVolume(t *testing.T, env *Env, searchFunc SearchWithPrefixFunc, basePlan *Plan, ap pfd.AtomicProcessID, extraVolume Volume) *Plan {
	t.Helper()

	initVolumeFunc := env.InitialVolumeFunc
	envExtended := env.Clone()
	envExtended.InitialVolumeFunc = func(ap2 pfd.AtomicProcessID) Volume {
		if ap2 == ap {
			return initVolumeFunc(ap2) + extraVolume
		}
		return initVolumeFunc(ap2)
	}

	plans, err := searchFunc(envExtended, PrefixUntilProcessStart(basePlan, ap))
	if err != nil {
		t.Fatalf("extending %s by %v: %v", ap, extraVolume, err)
	}
	plan, ok := plans.At(0)
	if !ok {
		t.Fatalf("extending %s by %v: no plan found", ap, extraVolume)
	}
	return plan
}

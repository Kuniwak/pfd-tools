package fsm

import (
	"log/slog"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
)

func TestMaximumElasticityWithMultipleConsumedVolumes(t *testing.T) {

	pcResourceSets := map[string]*sets.Set[AllocationElement]{

		"fast candidate does not contend": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1", "R3"), ConsumedVolume: 2},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 1},
		),

		"fast candidate contends with the critical path": sets.New(
			AllocationElement.Compare,
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 4},
			AllocationElement{Resources: sets.New(ResourceID.Compare, "R2"), ConsumedVolume: 1},
		),
	}
	searchFuncs := map[string]SearchWithPrefixFunc{
		"poorest": SearchPoorestWithPrefix(1),
		"poor":    SearchFastestWithPrefix(1),
		"best":    SearchBestPlansWithPrefix(),
	}
	for setName, pcResourceSet := range pcResourceSets {
		for searchName, searchFunc := range searchFuncs {
			t.Run(setName+"/"+searchName, func(t *testing.T) {
				assertMaximumElasticityWithMultipleConsumedVolumes(t, newMultipleConsumedVolumesEnv(t, pcResourceSet), searchFunc)
			})
		}
	}
}

func newMultipleConsumedVolumesEnv(t *testing.T, pcResourceSet *sets.Set[AllocationElement]) *Env {
	t.Helper()

	p := newSafePFDByUnsafePFD(&pfd.PFD{
		Nodes: sets.New(
			(*pfd.Node).Compare,
			&pfd.Node{ID: "D1", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D2", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D3", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D4", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "D5", Type: pfd.NodeTypeAtomicDeliverable},
			&pfd.Node{ID: "Pa", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "Pb", Type: pfd.NodeTypeAtomicProcess},
			&pfd.Node{ID: "Pc", Type: pfd.NodeTypeAtomicProcess},
		),
		Edges: sets.New(
			(*pfd.Edge).Compare,
			&pfd.Edge{Source: "D1", Target: "Pa"},
			&pfd.Edge{Source: "Pa", Target: "D2"},
			&pfd.Edge{Source: "D2", Target: "Pb"},
			&pfd.Edge{Source: "Pb", Target: "D3"},
			&pfd.Edge{Source: "D4", Target: "Pc"},
			&pfd.Edge{Source: "Pc", Target: "D5"},
		),
	})

	initVolumeFunc := func(ap pfd.AtomicProcessID) Volume {
		switch ap {
		case "Pa", "Pb":
			return 10
		case "Pc":
			return 4
		default:
			panic("unknown process: " + string(ap))
		}
	}

	neededResourceSetsFunc := NeededResourceSetsFuncByMap(map[pfd.AtomicProcessID]*sets.Set[AllocationElement]{
		"Pa": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 2}),
		"Pb": sets.New(AllocationElement.Compare, AllocationElement{Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: 2}),
		"Pc": pcResourceSet,
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
		slog.New(slogtest.NewTestHandler(t)),
	)
}

func assertMaximumElasticityWithMultipleConsumedVolumes(t *testing.T, env *Env, searchFunc SearchWithPrefixFunc) {
	t.Helper()

	basePlans, err := searchFunc(env, nil)
	if err != nil {
		t.Fatal(err)
	}
	basePlan, ok := basePlans.At(0)
	if !ok {
		t.Fatal("no base plan found")
	}
	t.Logf("baseline leadtime: %v, Pc runs for %v", basePlan.Leadtime(), basePlan.ProcessTime("Pc"))

	info, err := NewCriticalPathInfoFunc(searchFunc)(env)
	if err != nil {
		t.Fatal(err)
	}

	for _, ap := range []pfd.AtomicProcessID{"Pa", "Pb"} {
		if got := info[ap].MaximumElasticity; !got.ApproximateEqual(0) {
			t.Errorf("%s total float: got %v, want 0", ap, got)
		}
	}

	step := Volume(0.5)
	want := maximumElasticityByScan(t, env, searchFunc, basePlan, "Pc", step)
	got := info["Pc"].MaximumElasticity
	t.Logf("Pc total float: got %v, want %v (by scan with step %v)", got, want, step)
	if diff := got - want; diff < -execmodel.Time(step) || execmodel.Time(step) < diff {
		t.Errorf("Pc total float: got %v, want %v±%v", got, want, step)
	}
}

func maximumElasticityByScan(t *testing.T, env *Env, searchFunc SearchWithPrefixFunc, basePlan *Plan, ap pfd.AtomicProcessID, step Volume) execmodel.Time {
	t.Helper()

	baseLeadtime := basePlan.Leadtime()
	baseProcessTime := basePlan.ProcessTime(ap)
	maxExtraVolume := 2 * Volume(baseLeadtime) * MaxConsumedVolume(env.NeededResourceSetsFunc(ap))

	elasticity := execmodel.Time(0)
	for extraVolume := Volume(0); extraVolume <= maxExtraVolume; extraVolume += step {
		plan := planWithExtraVolume(t, env, searchFunc, basePlan, ap, extraVolume)
		if plan.Leadtime() > baseLeadtime && !plan.Leadtime().ApproximateEqual(baseLeadtime) {
			continue
		}
		elasticity = max(elasticity, plan.ProcessTime(ap)-baseProcessTime)
	}
	return elasticity
}

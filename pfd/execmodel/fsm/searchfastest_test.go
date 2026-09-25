package fsm

import (
	"math/rand"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

func allocVol(key pfd.AtomicProcessID, vol Volume) Allocation {
	return Allocation{key: {Resources: sets.New(ResourceID.Compare, "R1"), ConsumedVolume: vol}}
}

func transVol(key pfd.AtomicProcessID, vol Volume) *Trans {
	return &Trans{Allocation: allocVol(key, vol)}
}

func TestPickFastestAllocation(t *testing.T) {
	tests := map[string]struct {
		trs *sets.Set[*Trans]

		wantVol Volume

		wantChoices []Allocation
	}{
		"unique-max": {
			trs:         sets.New(CompareTrans, transVol("P1", 1), transVol("P2", 3), transVol("P3", 2)),
			wantVol:     3,
			wantChoices: []Allocation{allocVol("P2", 3)},
		},
		"tie-two": {
			trs:         sets.New(CompareTrans, transVol("P1", 3), transVol("P2", 3), transVol("P3", 1)),
			wantVol:     3,
			wantChoices: []Allocation{allocVol("P1", 3), allocVol("P2", 3)},
		},
		"tie-three-all": {
			trs:         sets.New(CompareTrans, transVol("P1", 2), transVol("P2", 2), transVol("P3", 2)),
			wantVol:     2,
			wantChoices: []Allocation{allocVol("P1", 2), allocVol("P2", 2), allocVol("P3", 2)},
		},
		"all-zero": {
			trs:         sets.New(CompareTrans, &Trans{Allocation: Allocation{}}),
			wantVol:     0,
			wantChoices: nil,
		},
		"empty": {
			trs:         sets.NewWithCapacity[*Trans](0),
			wantVol:     0,
			wantChoices: nil,
		},
	}

	const numSeeds = 200
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			seen := make([]bool, len(tt.wantChoices))
			for seed := int64(0); seed < numSeeds; seed++ {
				rng := rand.New(rand.NewSource(seed))
				got := PickFastestAllocation(tt.trs, rng)

				if got.TotalConsumedVolume() != tt.wantVol {
					t.Fatalf("seed=%d: TotalConsumedVolume() = %v, want %v", seed, got.TotalConsumedVolume(), tt.wantVol)
				}

				if len(tt.wantChoices) == 0 {
					if len(got) != 0 {
						t.Fatalf("seed=%d: got non-empty allocation %v, want empty", seed, got)
					}
					continue
				}

				idx := -1
				for i, c := range tt.wantChoices {
					if got.Equals(c) {
						idx = i
						break
					}
				}
				if idx < 0 {
					t.Fatalf("seed=%d: got %v, not in allowed tied choices", seed, got)
				}
				seen[idx] = true
			}

			for i, s := range seen {
				if !s {
					t.Errorf("choice %d was never selected across %d seeds (not choosing uniformly among tied transitions)", i, numSeeds)
				}
			}
		})
	}
}

func TestPickFastestAllocationReproducible(t *testing.T) {
	trs := sets.New(CompareTrans, transVol("P1", 3), transVol("P2", 3), transVol("P3", 3))

	const seed = int64(42)
	first := PickFastestAllocation(trs, rand.New(rand.NewSource(seed)))
	second := PickFastestAllocation(trs, rand.New(rand.NewSource(seed)))

	if !first.Equals(second) {
		t.Errorf("same seed produced different results: %v vs %v", first, second)
	}
}

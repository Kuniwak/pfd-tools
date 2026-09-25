package pfdfuzz

import (
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/google/go-cmp/cmp"
)

func TestCollectAtomicProcess2DeliverablePaths(t *testing.T) {
	t.Run("simplest", func(t *testing.T) {
		p := pfd.PresetSmallest
		edgeMap, _ := pfd.NewEdgeMap(p.Edges)

		aps := OnlyAtomicProcesses(p.Nodes)
		actual := CollectAtomicProcess2DeliverablePaths(edgeMap, aps)

		expected := sets.New(
			cmp2.CompareSlice[[]pfd.NodeID](pfd.NodeID.Compare),
			[]pfd.NodeID{"P1", "D2"},
		)

		if !reflect.DeepEqual(actual, expected) {
			t.Error(cmp.Diff(expected, actual))
		}
	})

	t.Run("sequential", func(t *testing.T) {
		p := pfd.PresetSequential
		edgeMap, _ := pfd.NewEdgeMap(p.Edges)

		aps := OnlyAtomicProcesses(p.Nodes)
		actual := CollectAtomicProcess2DeliverablePaths(edgeMap, aps)

		expected := sets.New(
			cmp2.CompareSlice[[]pfd.NodeID](pfd.NodeID.Compare),
			[]pfd.NodeID{"P1", "D2"},
			[]pfd.NodeID{"P1", "D2", "P2", "D3"},
			[]pfd.NodeID{"P2", "D3"},
		)

		if !reflect.DeepEqual(actual, expected) {
			t.Error(cmp.Diff(expected, actual))
		}
	})

	t.Run("bigger counterclockwise rotated y shape", func(t *testing.T) {
		p := pfd.PresetBiggerCounterclockwiseRotatedYShape
		edgeMap, _ := pfd.NewEdgeMap(p.Edges)

		aps := OnlyAtomicProcesses(p.Nodes)
		actual := CollectAtomicProcess2DeliverablePaths(edgeMap, aps)

		expected := sets.New(
			cmp2.CompareSlice[[]pfd.NodeID](pfd.NodeID.Compare),
			[]pfd.NodeID{"P1", "D3"},
			[]pfd.NodeID{"P1", "D3", "P3", "D5"},
			[]pfd.NodeID{"P2", "D4"},
			[]pfd.NodeID{"P2", "D4", "P3", "D5"},
			[]pfd.NodeID{"P3", "D5"},
		)

		if !reflect.DeepEqual(actual, expected) {
			t.Error(cmp.Diff(expected, actual))
		}
	})

	t.Run("bigger clockwise rotated y shape", func(t *testing.T) {
		p := pfd.PresetBiggerClockwiseRotatedYShape
		edgeMap, _ := pfd.NewEdgeMap(p.Edges)

		aps := OnlyAtomicProcesses(p.Nodes)
		actual := CollectAtomicProcess2DeliverablePaths(edgeMap, aps)

		expected := sets.New(
			cmp2.CompareSlice[[]pfd.NodeID](pfd.NodeID.Compare),
			[]pfd.NodeID{"P1", "D2"},
			[]pfd.NodeID{"P1", "D2", "P2", "D3"},
			[]pfd.NodeID{"P1", "D2", "P3", "D4"},
			[]pfd.NodeID{"P2", "D3"},
			[]pfd.NodeID{"P3", "D4"},
		)

		if !reflect.DeepEqual(actual, expected) {
			t.Error(cmp.Diff(expected, actual))
		}
	})
}

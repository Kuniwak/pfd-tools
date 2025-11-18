package graph

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/Kuniwak/pfd-tools/cmp2"
	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
	cmp3 "github.com/google/go-cmp/cmp"
	"pgregory.net/rapid"
)

func TestMinimals(t *testing.T) {
	testCases := map[string]struct {
		Graph    *Graph
		Expected *sets.Set[Node]
	}{
		"empty": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(Node.Compare),
		},
		"single": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A"),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(Node.Compare, "A"),
		},
		"single edge": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
				),
			},
			Expected: sets.New(Node.Compare, "A"),
		},
		"like reversed Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("A", "C"),
				),
			},
			Expected: sets.New(Node.Compare, "A"),
		},
		"like Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "C"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(Node.Compare, "A", "B"),
		},
		"sequential": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(Node.Compare, "A"),
		},
		"loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "A"),
				),
			},
			Expected: sets.New(Node.Compare),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Graph.Minimals()
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp3.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestMaximals(t *testing.T) {
	testCases := map[string]struct {
		Graph    *Graph
		Expected *sets.Set[Node]
	}{
		"empty": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(Node.Compare),
		},
		"single": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A"),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(Node.Compare, "A"),
		},
		"single edge": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
				),
			},
			Expected: sets.New(Node.Compare, "B"),
		},
		"like reversed Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("A", "C"),
				),
			},
			Expected: sets.New(Node.Compare, "B", "C"),
		},
		"like Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "C"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(Node.Compare, "C"),
		},
		"sequential": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(Node.Compare, "C"),
		},
		"loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "A"),
				),
			},
			Expected: sets.New(Node.Compare),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Graph.Maximals()
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp3.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestCycles(t *testing.T) {
	testCases := map[string]struct {
		Graph    *Graph
		Expected *sets.Set[[]Node]
	}{
		"empty": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(cmp2.CompareSlice[[]Node](Node.Compare)),
		},
		"no-loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(cmp2.CompareSlice[[]Node](Node.Compare)),
		},
		"shortest loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "A"),
				),
			},
			Expected: sets.New(
				cmp2.CompareSlice[[]Node](Node.Compare),
				[]Node{"A"},
			),
		},
		"long loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
					pairs.New[Node, Node]("C", "A"),
				),
			},
			Expected: sets.New(
				cmp2.CompareSlice[[]Node](Node.Compare),
				[]Node{"A", "B", "C"},
			),
		},
		"butterfly-like loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B1", "B2"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B1"),
					pairs.New[Node, Node]("A", "B2"),
					pairs.New[Node, Node]("B1", "A"),
					pairs.New[Node, Node]("B2", "A"),
				),
			},
			Expected: sets.New(
				cmp2.CompareSlice[[]Node](Node.Compare),
				[]Node{"A", "B1"},
				[]Node{"A", "B2"},
			),
		},
		"nested loops": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C", "D"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
					pairs.New[Node, Node]("B", "A"),
					pairs.New[Node, Node]("C", "D"),
					pairs.New[Node, Node]("D", "A"),
				),
			},
			Expected: sets.New(
				cmp2.CompareSlice[[]Node](Node.Compare),
				[]Node{"A", "B"},
				[]Node{"A", "B", "C", "D"},
			),
		},
		"branch loops": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B1", "B2", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B1"),
					pairs.New[Node, Node]("A", "B2"),
					pairs.New[Node, Node]("B1", "C"),
					pairs.New[Node, Node]("B2", "C"),
					pairs.New[Node, Node]("C", "A"),
				),
			},
			Expected: sets.New(
				cmp2.CompareSlice[[]Node](Node.Compare),
				[]Node{"A", "B1", "C"},
				[]Node{"A", "B2", "C"},
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Graph.Cycles()
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp3.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("dag", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			dag := DAG(t, 10)
			actual := dag.Cycles()
			if actual.Len() != 0 {
				t.Errorf("actual %v", actual)
			}
		})
	})
}

func TestWeakConn(t *testing.T) {
	testCases := map[string]struct {
		Graph    *Graph
		Expected *sets.Set[*sets.Set[Node]]
	}{
		"empty": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(sets.Compare(Node.Compare)),
		},
		"single": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A"),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(sets.Compare(Node.Compare), sets.New(Node.Compare, "A")),
		},
		"single edge": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
				),
			},
			Expected: sets.New(sets.Compare(Node.Compare), sets.New(Node.Compare, "A", "B")),
		},
		"like reversed Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("A", "C"),
				),
			},
			Expected: sets.New(sets.Compare(Node.Compare), sets.New(Node.Compare, "A", "B", "C")),
		},
		"like Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "C"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(sets.Compare(Node.Compare), sets.New(Node.Compare, "A", "B", "C")),
		},
		"sequential": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			Expected: sets.New(sets.Compare(Node.Compare), sets.New(Node.Compare, "A", "B", "C")),
		},
		"loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "A"),
				),
			},
			Expected: sets.New(sets.Compare(Node.Compare), sets.New(Node.Compare, "A", "B")),
		},
		"dangling": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			Expected: sets.New(
				sets.Compare(Node.Compare),
				sets.New(Node.Compare, "A"),
				sets.New(Node.Compare, "B"),
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Graph.WeaklyConnectedComponents()
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp3.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("weekly connected", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			dag := WeeklyConnectedDAG(t, 10)
			actual := dag.WeaklyConnectedComponents()
			if actual.Len() != 1 {
				t.Errorf("actual %v", actual)
			}
		})
	})
}

func TestTopologicalSort(t *testing.T) {
	testCases := map[string]struct {
		Graph         *Graph
		ExpectedOneOf [][]Node
	}{
		"empty": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			ExpectedOneOf: [][]Node{{}},
		},
		"single": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A"),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			ExpectedOneOf: [][]Node{{"A"}},
		},
		"single edge": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
				),
			},
			ExpectedOneOf: [][]Node{{"A", "B"}},
		},
		"like reversed Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("A", "C"),
				),
			},
			ExpectedOneOf: [][]Node{{"A", "B", "C"}, {"A", "C", "B"}},
		},
		"like Y shape": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "C"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			ExpectedOneOf: [][]Node{{"A", "B", "C"}, {"B", "A", "C"}},
		},
		"sequential": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B", "C"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "C"),
				),
			},
			ExpectedOneOf: [][]Node{{"A", "B", "C"}},
		},
		"loop": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(
					pairs.Compare(Node.Compare, Node.Compare),
					pairs.New[Node, Node]("A", "B"),
					pairs.New[Node, Node]("B", "A"),
				),
			},
			ExpectedOneOf: [][]Node{nil},
		},
		"not connected": {
			Graph: &Graph{
				Nodes: sets.New(Node.Compare, "A", "B"),
				Edges: sets.New(pairs.Compare(Node.Compare, Node.Compare)),
			},
			ExpectedOneOf: [][]Node{{"A", "B"}, {"B", "A"}},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Graph.TopologicalSort()
			for _, expected := range testCase.ExpectedOneOf {
				if reflect.DeepEqual(actual, expected) {
					return
				}
				t.Log(cmp3.Diff(expected, actual))
			}
			t.Error("expected one of the sorts")
		})
	}

	t.Run("dag", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			dag := DAG(t, 10)
			actual := dag.TopologicalSort()

			for i := 0; i < len(actual); i++ {
				for j := i + 1; j < len(actual); j++ {
					if dag.IsReachable(actual[j], actual[i]) {
						t.Errorf("actual %v", actual)
					}
				}
			}
		})
	})
}

func DAG(t *rapid.T, n int) *Graph {
	cur := 0
	nodes := sets.New(Node.Compare)
	edges := sets.New(pairs.Compare(Node.Compare, Node.Compare))
	for i := 0; i < n; i++ {
		node1 := Node(strconv.Itoa(cur))
		cur++
		for j, node2 := range nodes.Iter() {
			if rapid.Bool().Draw(t, fmt.Sprintf("edge %d %d", i, j)) {
				edges.Add(pairs.Compare(Node.Compare, Node.Compare), pairs.New(node1, node2))
			}
		}
		nodes.Add(Node.Compare, node1)
	}
	return &Graph{
		Nodes: nodes,
		Edges: edges,
	}
}

func WeeklyConnectedDAG(t *rapid.T, n int) *Graph {
	cur := 0
	nodes := sets.New(Node.Compare)
	edges := sets.New(pairs.Compare(Node.Compare, Node.Compare))
	for i := 0; i < n; i++ {
		node1 := Node(strconv.Itoa(cur))
		cur++

		if nodes.Len() > 0 {
			idx := rapid.IntRange(0, nodes.Len()-1).Draw(t, fmt.Sprintf("idx %d", i))
			node2, ok := nodes.At(idx)
			if !ok {
				t.Fatalf("node2 not found")
			}
			edges.Add(pairs.Compare(Node.Compare, Node.Compare), pairs.New(node1, node2))
		}

		for j, node2 := range nodes.Iter() {
			if rapid.Bool().Draw(t, fmt.Sprintf("edge %d %d", i, j)) {
				edges.Add(pairs.Compare(Node.Compare, Node.Compare), pairs.New(node1, node2))
			}
		}
		nodes.Add(Node.Compare, node1)
	}
	return &Graph{
		Nodes: nodes,
		Edges: edges,
	}
}

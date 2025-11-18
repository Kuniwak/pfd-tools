package graph

import (
	"slices"
	"strings"

	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

type Node string

func (a Node) Compare(b Node) int {
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return strings.Compare(string(a), string(b))
}

type Graph struct {
	Nodes *sets.Set[Node]
	Edges *sets.Set[*pairs.Pair[Node, Node]]
}

func (g *Graph) Clone() *Graph {
	edges := sets.NewWithCapacity[*pairs.Pair[Node, Node]](g.Edges.Len())
	for _, pair := range g.Edges.Iter() {
		edges.Add(pairs.Compare(Node.Compare, Node.Compare), pair.Clone())
	}
	return &Graph{
		Nodes: g.Nodes.Clone(),
		Edges: edges,
	}
}

func (g *Graph) IsReachable(src, dst Node) bool {
	if !g.Nodes.Contains(Node.Compare, src) || !g.Nodes.Contains(Node.Compare, dst) {
		return false
	}

	if src == dst {
		return true
	}

	adj := make(map[Node][]Node, g.Nodes.Len())
	for _, e := range g.Edges.Iter() {
		adj[e.First] = append(adj[e.First], e.Second)
	}

	visited := sets.NewWithCapacity[Node](g.Nodes.Len())
	stack := []Node{src}

	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited.Contains(Node.Compare, v) {
			continue
		}
		visited.Add(Node.Compare, v)

		for _, nb := range adj[v] {
			if nb == dst {
				return true
			}
			if !visited.Contains(Node.Compare, nb) {
				stack = append(stack, nb)
			}
		}
	}

	return false
}

func (g *Graph) Maximals() *sets.Set[Node] {
	state := make(map[Node]bool)
	for _, node := range g.Nodes.Iter() {
		state[node] = true
	}
	for _, edge := range g.Edges.Iter() {
		src := edge.First
		state[src] = false
	}
	maximals := sets.NewWithCapacity[Node](g.Nodes.Len())
	for _, node := range g.Nodes.Iter() {
		if state[node] {
			maximals.Add(Node.Compare, node)
		}
	}
	return maximals
}

func (g *Graph) Minimals() *sets.Set[Node] {
	state := make(map[Node]bool)
	for _, node := range g.Nodes.Iter() {
		state[node] = true
	}
	for _, edge := range g.Edges.Iter() {
		dst := edge.Second
		state[dst] = false
	}
	minimals := sets.NewWithCapacity[Node](g.Nodes.Len())
	for _, node := range g.Nodes.Iter() {
		if state[node] {
			minimals.Add(Node.Compare, node)
		}
	}
	return minimals
}

func (g *Graph) Cycles() *sets.Set[[]Node] {
	// Build adjacency list for directed graph (outgoing edges only)
	adj := make(map[Node][]Node, g.Nodes.Len())
	for _, u := range g.Nodes.Iter() {
		adj[u] = nil
	}
	for _, e := range g.Edges.Iter() {
		u := e.First
		v := e.Second
		adj[u] = append(adj[u], v)
	}
	// Sort adjacent nodes for deterministic enumeration
	for u := range adj {
		slices.SortFunc(adj[u], Node.Compare)
	}

	// Nodes themselves are also in deterministic order
	nodes := g.Nodes.Slice()
	slices.SortFunc(nodes, Node.Compare)

	cycles := make([][]Node, 0)

	blocked := make(map[Node]bool)          // blocked set for Johnson's algorithm
	B := make(map[Node]map[Node]bool)       // back-link set
	stack := make([]Node, 0, g.Nodes.Len()) // current path
	var s Node                              // current start node (explore only nodes >= this node)

	var unblock func(u Node)
	unblock = func(u Node) {
		if !blocked[u] {
			return
		}
		blocked[u] = false
		if B[u] != nil {
			for w := range B[u] {
				delete(B[u], w)
				unblock(w)
			}
		}
	}

	var circuit func(v Node) bool
	circuit = func(v Node) bool {
		found := false
		stack = append(stack, v)
		blocked[v] = true

		for _, w := range adj[v] {
			// Do not explore nodes smaller than s (subgraph induction)
			if Node.Compare(w, s) < 0 {
				continue
			}
			if w == s {
				// Record simple cycle only once (stack contains s..v)
				c := make([]Node, len(stack))
				copy(c, stack)
				cycles = append(cycles, c)
				found = true
			} else if !blocked[w] {
				if circuit(w) {
					found = true
				}
			}
		}

		if found {
			unblock(v)
		} else {
			for _, w := range adj[v] {
				if Node.Compare(w, s) < 0 {
					continue
				}
				if B[w] == nil {
					B[w] = make(map[Node]bool)
				}
				B[w][v] = true
			}
		}

		stack = stack[:len(stack)-1]
		return found
	}

	for _, start := range nodes {
		s = start
		// Initialize blocked and B (for nodes >= s)
		for _, u := range nodes {
			if Node.Compare(u, s) >= 0 {
				blocked[u] = false
				B[u] = make(map[Node]bool)
			}
		}
		_ = circuit(s)
	}

	// Fix output order (length first → lexicographic)
	cmpCycle := func(a, b []Node) int {
		if len(a) < len(b) {
			return -1
		}
		if len(a) > len(b) {
			return 1
		}
		return slices.CompareFunc(a, b, Node.Compare)
	}
	slices.SortFunc(cycles, cmpCycle)

	res := sets.Set[[]Node](cycles)
	return &res
}

// WeaklyConnectedComponents returns the weakly connected components when g is viewed as undirected.
// The return value is a slice of Node sets for each component. The order of weakly connected components is undefined.
func (g *Graph) WeaklyConnectedComponents() *sets.Set[*sets.Set[Node]] {
	adj := make(map[Node][]Node, g.Nodes.Len())
	for _, u := range g.Nodes.Iter() {
		adj[u] = adj[u][:0]
	}
	for _, edge := range g.Edges.Iter() {
		u := edge.First
		v := edge.Second
		adj[u] = append(adj[u], v)
		adj[v] = append(adj[v], u)
	}

	visited := make(map[Node]bool, len(adj))
	comps := sets.NewWithCapacity[*sets.Set[Node]](len(adj))

	for u := range adj {
		if visited[u] {
			continue
		}
		comp := BFSComponent(u, adj, visited)
		comps.Add(sets.Compare(Node.Compare), comp)
	}

	return comps
}

func BFSComponent(start Node, adj map[Node][]Node, visited map[Node]bool) *sets.Set[Node] {
	queue := []Node{start}
	visited[start] = true
	comp := sets.New(Node.Compare)

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		comp.Add(Node.Compare, u)
		for _, v := range adj[u] {
			if !visited[v] {
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}
	return comp
}

func (g *Graph) TopologicalSort() []Node {
	// Build adjacency list and in-degrees
	adj := make(map[Node][]Node, g.Nodes.Len())
	inDeg := make(map[Node]int, g.Nodes.Len())
	for _, u := range g.Nodes.Iter() {
		adj[u] = nil
		inDeg[u] = 0
	}
	for _, e := range g.Edges.Iter() {
		u := e.First
		v := e.Second
		adj[u] = append(adj[u], v)
		inDeg[v]++
	}

	// Maintain set of nodes with in-degree 0 (in Node.Compare order)
	zero := make([]Node, 0)
	for u := range adj {
		if inDeg[u] == 0 {
			zero = append(zero, u)
		}
	}
	slices.SortFunc(zero, Node.Compare)

	pushZero := func(v Node) {
		i, _ := slices.BinarySearchFunc(zero, v, Node.Compare)
		zero = append(zero, v)
		copy(zero[i+1:], zero[i:])
		zero[i] = v
	}

	order := make([]Node, 0, g.Nodes.Len())
	for len(zero) > 0 {
		u := zero[0]
		zero = zero[1:]
		order = append(order, u)
		for _, v := range adj[u] {
			inDeg[v]--
			if inDeg[v] == 0 {
				pushZero(v)
			}
		}
	}

	// If there are cycles, no topological order exists
	if len(order) != g.Nodes.Len() {
		return nil
	}
	return order
}

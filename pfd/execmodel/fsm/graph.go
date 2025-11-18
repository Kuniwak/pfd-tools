package fsm

import (
	"cmp"
	"fmt"
	"hash/maphash"
	"log/slog"
	"slices"

	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

type StateID int
type StateHash uint64

func (a StateID) Compare(b StateID) int {
	return cmp.Compare(a, b)
}

type StateTransitionGraph struct {
	InitialState StateID
	Nodes        map[StateID]State
	Edges        map[StateID]map[StateID]*sets.Set[Allocation]
}

type StateTransitionGraphEdge struct {
	Source     StateID
	Target     StateID
	Allocation Allocation
}

func (a StateTransitionGraphEdge) Compare(b StateTransitionGraphEdge) int {
	c := cmp.Compare(a.Source, b.Source)
	if c != 0 {
		return c
	}
	c = cmp.Compare(a.Target, b.Target)
	if c != 0 {
		return c
	}
	return a.Allocation.Compare(b.Allocation)
}

func (g StateTransitionGraph) FindDeadlocks(isCompleted func(State) bool, ch chan<- Deadlock) {
	g.findDeadlocks(g.InitialState, []*pairs.Pair[StateID, Allocation]{}, isCompleted, ch)
}

type Deadlock struct {
	StateID StateID
	Path    []*pairs.Pair[StateID, Allocation]
}

func (d Deadlock) Compare(b Deadlock) int {
	c := cmp.Compare(d.StateID, b.StateID)
	if c != 0 {
		return c
	}
	return slices.CompareFunc(d.Path, b.Path, pairs.Compare(StateID.Compare, Allocation.Compare))
}

func (g StateTransitionGraph) findDeadlocks(stateID StateID, path []*pairs.Pair[StateID, Allocation], isCompleted func(State) bool, ch chan<- Deadlock) {
	if g.IsDeadlock(stateID, isCompleted) {
		ch <- Deadlock{StateID: stateID, Path: path}
		return
	}

	m, ok := g.Edges[stateID]
	if !ok {
		return
	}

	for dst, as := range m {
		for _, allocation := range as.Iter() {
			path = append(path, pairs.New(stateID, allocation))
			g.findDeadlocks(dst, path, isCompleted, ch)
			path = path[:len(path)-1]
		}
	}
}

func (g StateTransitionGraph) IsDeadlock(stateID StateID, isCompleted func(State) bool) bool {
	s, ok := g.Nodes[stateID]
	if !ok {
		panic("StateTransitionGraph.IsDeadlock: state not found")
	}
	if isCompleted(s) {
		return false
	}

	m, ok := g.Edges[stateID]
	if !ok {
		return true
	}

	for _, as := range m {
		if as.Len() > 0 {
			return false
		}
	}
	return true
}

func Graph(e *Env, maxDepth int, logger *slog.Logger) (StateTransitionGraph, error) {
	res := &StateTransitionGraph{
		InitialState: StateID(0),
		Nodes:        make(map[StateID]State),
		Edges:        make(map[StateID]map[StateID]*sets.Set[Allocation]),
	}

	h := &maphash.Hash{}
	seen := make(map[StateHash]StateID)

	var dfs func(state State, depth int) (StateID, error)
	dfs = func(state State, depth int) (StateID, error) {
		nodeID := StateID(len(seen))

		h.Reset()
		HashState(state, h)
		seen[StateHash(h.Sum64())] = nodeID

		if len(seen)%10000 == 0 {
			logger.Debug("Graph: Seen states", "num of states", len(seen))
		}

		res.Nodes[nodeID] = state

		if depth > maxDepth {
			return nodeID, nil
		}

		trans := e.Transitions(state)
		for _, trans := range trans.Iter() {
			h.Reset()
			if err := HashState(trans.NextState, h); err != nil {
				return -1, fmt.Errorf("Graph: HashState: %w", err)
			}
			sh := StateHash(h.Sum64())
			nextNodeID, ok := seen[sh]
			if !ok {
				var err error
				nextNodeID, err = dfs(trans.NextState, depth+1)
				if err != nil {
					return -1, fmt.Errorf("Graph: dfs: %w", err)
				}
			}

			if s, ok := res.Edges[nodeID]; ok {
				if _, ok := s[nextNodeID]; ok {
					s[nextNodeID].Add(Allocation.Compare, trans.Allocation)
				} else {
					s[nextNodeID] = sets.New(Allocation.Compare, trans.Allocation)
				}
			} else {
				res.Edges[nodeID] = map[StateID]*sets.Set[Allocation]{
					nextNodeID: sets.New(Allocation.Compare, trans.Allocation),
				}
			}
		}

		return nodeID, nil
	}

	logger.Debug("Graph: starting...")

	dfs(e.InitialState(), 1)

	logger.Debug("Graph: completed")

	return *res, nil
}

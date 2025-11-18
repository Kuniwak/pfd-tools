package fsm

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
)

// Plan is an execution plan.
// The elements of an execution plan consist of a post-state and the allocation that leads to that post-state.
type Plan struct {
	InitialState State    `json:"initial_state"`
	Transitions  []*Trans `json:"transitions"`
}

func ParsePlan(reader io.Reader) (*Plan, error) {
	var plan Plan
	if err := json.NewDecoder(reader).Decode(&plan); err != nil {
		return nil, fmt.Errorf("fsm.ParsePlan: %w", err)
	}
	return &plan, nil
}

func NewEmptyPlan(initialState State) *Plan {
	return &Plan{
		InitialState: initialState,
		Transitions:  make([]*Trans, 0),
	}
}

func (c *Plan) Clone() *Plan {
	return &Plan{
		InitialState: c.InitialState,
		Transitions:  slices.Clone(c.Transitions),
	}
}

func (c *Plan) Add(tr *Trans) {
	c.Transitions = append(c.Transitions, tr)
}

func (c *Plan) Len() int {
	return len(c.Transitions)
}

func (c *Plan) States() []State {
	states := make([]State, 0, len(c.Transitions)+1)
	states = append(states, c.InitialState)
	for _, tr := range c.Transitions {
		states = append(states, tr.NextState)
	}
	return states
}

func (c *Plan) Leadtime() execmodel.Time {
	states := c.States()
	return states[len(states)-1].Time
}

func (p *Plan) Compare(b *Plan) int {
	c := cmp.Compare(p.InitialState.Time, b.InitialState.Time)
	if c != 0 {
		return c
	}
	return slices.CompareFunc(p.Transitions, b.Transitions, CompareTrans)
}

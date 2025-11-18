package fsm

import "github.com/Kuniwak/pfd-tools/sets"

type SearchFunc func(e *Env) (*sets.Set[*Plan], error)

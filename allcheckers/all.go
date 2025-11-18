package allcheckers

import (
	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmchecker"
	"github.com/Kuniwak/pfd-tools/pfd/pfdcheckers"
)

var PFDCheckers = checkers.NewParallelChecker(
	pfdcheckers.ConsistentDesc,
	pfdcheckers.InField,
	pfdcheckers.ExInput,
	pfdcheckers.ExOutput,
	pfdcheckers.NoD2D,
	pfdcheckers.NoP2P,
	pfdcheckers.NoP2DFB,
	pfdcheckers.SingleSrc,
	pfdcheckers.AcyclicExceptFB,
	pfdcheckers.WeakConn,
	pfdcheckers.Finite,
	pfdcheckers.DisjOrProperSubsetComp,
	pfdcheckers.ConsistentInputComp,
	pfdcheckers.ConsistentOutputComp,
	pfdcheckers.NoDesc,
	pfdcheckers.ConsistentAPTable,
	pfdcheckers.ConsistentCPTable,
	pfdcheckers.ConsistentDTable,
)

var FSMCheckers = checkers.NewParallelChecker(
	fsmchecker.ConsistentGTable,
	fsmchecker.ConsistentMTable,
	fsmchecker.ConsistentResourceTable,
	fsmchecker.ValidAvailableTime,
	fsmchecker.ValidInitVolume,
	fsmchecker.ValidMaxRevision,
	fsmchecker.ValidResourcesSet,
	fsmchecker.ValidPrecondition,
)

package fsm

import (
	"fmt"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

type ResourceAspect struct {
	AvailableResources     *sets.Set[ResourceID]
	NeededResourceSetsFunc NeededResourceSetsFunc
}

func UnlimitedResourceAspect() ResourceAspect {
	return ResourceAspect{
		AvailableResources:     sets.New(ResourceID.Compare),
		NeededResourceSetsFunc: UnlimitedNeededResourceSetsFunc(),
	}
}

type FeedbackAspect struct {
	ReworkVolumeFunc          ReworkVolumeFunc
	FeedbackSourceMaxRevision map[pfd.AtomicDeliverableID]int
}

func NoFeedbackAspect(initialVolumeFunc InitialVolumeFunc) FeedbackAspect {
	return FeedbackAspect{
		ReworkVolumeFunc:          NoReworkVolumeFunc(initialVolumeFunc),
		FeedbackSourceMaxRevision: map[pfd.AtomicDeliverableID]int{},
	}
}

func ValidateNoFeedbackEdges(p *pfd.ValidPFD) error {
	edges := p.FeedbackEdges()
	if len(edges) == 0 {
		return nil
	}

	texts := make([]string, 0, len(edges))
	for _, e := range edges {
		texts = append(texts, fmt.Sprintf("%s -> %s", e.AtomicDeliverable, e.AtomicProcess))
	}

	return fmt.Errorf("fsm.ValidateNoFeedbackEdges: feedback edges are not available in this exec model: %s", strings.Join(texts, ", "))
}

package fsm

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Kuniwak/pfd-tools/pfd"
)

// Volume is the work volume.
type Volume float64

const MinimumVolume = 0.001 // Approximately equivalent to 30 people (calculated as 9 hours per business day)

func (v Volume) String() string {
	if v.IsZero() {
		return "0"
	}
	return strconv.FormatFloat(float64(v), 'f', 3, 64)
}

func (v Volume) IsZero() bool {
	return v.ApproximateEqual(0)
}

func (v Volume) ApproximateEqual(b Volume) bool {
	return math.Abs(float64(v-b)) < MinimumVolume
}

// InitialVolumeFunc returns the required work volume for an atomic process when given that atomic process.
// Behavior is undefined when given an ID of an element that is not an atomic process.
type InitialVolumeFunc func(pfd.AtomicProcessID) Volume

func InitialVolumeByMap(m map[pfd.AtomicProcessID]Volume) InitialVolumeFunc {
	return func(ap pfd.AtomicProcessID) Volume {
		volume, ok := m[ap]
		if !ok {
			panic(fmt.Sprintf("InitialVolumeByMap: missing volume: %q", ap))
		}
		return volume
	}
}

func ConstInitialVolumeFunc(volume Volume) InitialVolumeFunc {
	return func(pfd.AtomicProcessID) Volume {
		return volume
	}
}

// ReworkVolumeFunc returns the work volume that is recovered when feedback edge deliverables are
// created or recreated, given an atomic process that receives feedback edges and the number of reworks for that atomic process.
// Behavior is undefined when given an element that is not an atomic process receiving feedback edges, or when given a non-positive numOfRework.
type ReworkVolumeFunc func(ap pfd.AtomicProcessID, numOfRework int) Volume

func ReworkVolumeByMaxReworksMap(m map[pfd.AtomicProcessID]ReworkVolumeFunc) ReworkVolumeFunc {
	return func(ap pfd.AtomicProcessID, numOfRework int) Volume {
		f, ok := m[ap]
		if !ok {
			panic(fmt.Sprintf("fsm.ReworkVolumeByMaxReworksMap: missing rework volume: %q", ap))
		}

		return f(ap, numOfRework)
	}
}

func ExponentialReworkVolumeFunc(reworkVolumeRatio float64, init InitialVolumeFunc) ReworkVolumeFunc {
	return func(ap pfd.AtomicProcessID, numOfRework int) Volume {
		if reworkVolumeRatio < 0 {
			panic("fsm.ExponentialReworkVolumeFunc: rework volume ratio must be positive")
		}
		if reworkVolumeRatio > 1 {
			panic("fsm.ExponentialReworkVolumeFunc: rework volume ratio must be less than 1")
		}
		vol := Volume(float64(init(ap)) * math.Pow(float64(reworkVolumeRatio), float64(numOfRework)))
		return max(vol, MinimumVolume)
	}
}

package fsm

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Kuniwak/pfd-tools/pfd"
)

type Volume float64

const MinimumVolume = 0.001

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

func NoReworkVolumeFunc(init InitialVolumeFunc) ReworkVolumeFunc {
	return func(ap pfd.AtomicProcessID, _ int) Volume {
		return init(ap)
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

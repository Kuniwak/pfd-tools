package fsmtable

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

const (
	InitialVolumeColumnHeaderJa = "予想作業量"
	InitialVolumeColumnHeaderEn = "Est. Work Volume"
)

var DefaultInitialVolumeColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	InitialVolumeColumnHeaderJa,
	InitialVolumeColumnHeaderEn,
))

func RawInitialVolumeMap(t *pfd.AtomicProcessTable, selectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]string, error) {
	m := make(map[pfd.AtomicProcessID]string, len(t.Rows))

	idx := selectFunc(t.ExtraHeaders)
	if idx < 0 {
		return nil, fmt.Errorf("fsm.InitialVolumeByTableFunc: missing initial volume column")
	}

	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[idx]
	}
	return m, nil
}

func ValidateInitialVolume(initialVolumeText string) (fsm.Volume, error) {
	initialVolume, err := strconv.ParseFloat(initialVolumeText, 32)
	if err != nil {
		return 0, fmt.Errorf("fsm.InitialVolumeByTableFunc: initial volume is not a number")
	}
	if initialVolume < 0 {
		return 0, fmt.Errorf("fsm.InitialVolumeByTableFunc: initial volume cannot be negative")
	}
	return max(fsm.Volume(initialVolume), fsm.MinimumVolume), nil
}

func ValidateInitialVolumeMap(m map[pfd.AtomicProcessID]string) (map[pfd.AtomicProcessID]fsm.Volume, error) {
	m2 := make(map[pfd.AtomicProcessID]fsm.Volume, len(m))

	for ap, row := range m {
		initialVolumeText := row
		initialVolume, err := ValidateInitialVolume(initialVolumeText)
		if err != nil {
			return nil, fmt.Errorf("fsm.InitialVolumeByTableFunc: initial volume is not a number")
		}
		m2[ap] = initialVolume
	}

	return m2, nil
}

func InitialVolumeByTableFunc(t *pfd.AtomicProcessTable, selectFunc pfd.ColumnSelectFunc) (fsm.InitialVolumeFunc, error) {
	m, err := RawInitialVolumeMap(t, selectFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.InitialVolumeByTableFunc: %w", err)
	}
	m2, err := ValidateInitialVolumeMap(m)
	if err != nil {
		return nil, fmt.Errorf("fsm.InitialVolumeByTableFunc: %w", err)
	}
	return fsm.InitialVolumeByMap(m2), nil
}

type ReworkVolumeColumnSelectFunc func([]string) int

const (
	ReworkVolumeRatioColumnHeaderJa = "予想手戻り作業量割合"
	ReworkVolumeRatioColumnHeaderEn = "Est. Rework Volume Ratio"
)

var DefaultReworkVolumeRatioColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	ReworkVolumeRatioColumnHeaderJa,
	ReworkVolumeRatioColumnHeaderEn,
))

func RawReworkVolumeRatioMap(t *pfd.AtomicProcessTable, selectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicProcessID]string, error) {
	m := make(map[pfd.AtomicProcessID]string, len(t.Rows))

	idx := selectFunc(t.ExtraHeaders)
	if idx < 0 {
		return nil, fmt.Errorf("fsm.ReworkVolumeRatioByTableFunc: missing rework volume ratio column")
	}

	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[idx]
	}
	return m, nil
}

func ValidateReworkVolumeRatio(reworkVolumeRatioText string) (float64, error) {
	reworkVolumeRatio, err := strconv.ParseFloat(reworkVolumeRatioText, 64)
	if err != nil {
		return 0, fmt.Errorf("fsm.ValidateReworkVolumeRatio: rework volume ratio is not a number")
	}
	if reworkVolumeRatio < 0 {
		return 0, fmt.Errorf("fsm.ValidateReworkVolumeRatio: rework volume ratio must be zero or positive")
	}
	if reworkVolumeRatio > 1 {
		return 0, fmt.Errorf("fsm.ValidateReworkVolumeRatio: rework volume ratio must be less than 1")
	}

	return reworkVolumeRatio, nil
}

func ValidateReworkVolumeRatioMap(
	reworkVolumeRatioTextMap map[pfd.AtomicProcessID]string,
	initVolumeFunc fsm.InitialVolumeFunc,
) (map[pfd.AtomicProcessID]fsm.ReworkVolumeFunc, error) {
	m := make(map[pfd.AtomicProcessID]fsm.ReworkVolumeFunc, len(reworkVolumeRatioTextMap))
	for ap, reworkVolumeRatioText := range reworkVolumeRatioTextMap {
		reworkVolumeRatio, err := ValidateReworkVolumeRatio(reworkVolumeRatioText)
		if err != nil {
			return nil, fmt.Errorf("fsm.ValidateReworkVolumeRatioMap: rework volume ratio is not a number")
		}
		m[ap] = fsm.ExponentialReworkVolumeFunc(reworkVolumeRatio, initVolumeFunc)
	}
	return m, nil
}

func ReworkVolumeRatioByTableFunc(t *pfd.AtomicProcessTable, selectFunc pfd.ColumnSelectFunc, initVolumeFunc fsm.InitialVolumeFunc) (fsm.ReworkVolumeFunc, error) {
	m1, err := RawReworkVolumeRatioMap(t, selectFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.ReworkVolumeRatioByTableFunc: %w", err)
	}
	m2, err := ValidateReworkVolumeRatioMap(m1, initVolumeFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.ReworkVolumeRatioByTableFunc: %w", err)
	}
	return fsm.ReworkVolumeByMaxReworksMap(m2), nil
}

func ReworkVolumeFuncByTableFunc(
	t *pfd.AtomicProcessTable,
	reworkVolumeRatioColumnSelectFunc pfd.ColumnSelectFunc,
	initVolumeFunc fsm.InitialVolumeFunc,
) (fsm.ReworkVolumeFunc, error) {
	reworkVolumeRatioColumnIdx := reworkVolumeRatioColumnSelectFunc(t.ExtraHeaders)

	if reworkVolumeRatioColumnIdx >= 0 {
		f, err := ReworkVolumeRatioByTableFunc(t, reworkVolumeRatioColumnSelectFunc, initVolumeFunc)
		if err != nil {
			return nil, fmt.Errorf("fsm.ReworkVolumeFuncByTableFunc: %w", err)
		}
		return f, nil
	}

	return nil, fmt.Errorf("fsm.ReworkVolumeFuncByTableFunc: missing rework volume ratio column")
}

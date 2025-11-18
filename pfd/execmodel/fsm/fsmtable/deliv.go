package fsmtable

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

const (
	AvailableTimeHeaderJa = "利用可能時刻"
	AvailableTimeHeaderEn = "Available Time"
)

var DefaultAvailableTimeColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	AvailableTimeHeaderJa,
	AvailableTimeHeaderEn,
))

func RawAvailableTimeMap(t *pfd.AtomicDeliverableTable, selectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicDeliverableID]string, error) {
	m := make(map[pfd.AtomicDeliverableID]string, len(t.Rows))

	idx := selectFunc(t.ExtraHeaders)
	if idx < 0 {
		return nil, fmt.Errorf("fsm.AvailableTimeFuncByTable: missing available time header in fsmtable")
	}

	for _, row := range t.Rows {
		availableTimeText := row.ExtraCells[idx]
		m[row.ID] = availableTimeText
	}

	return m, nil
}

func ValidateAvailableTime(availableTimeText string) (execmodel.Time, error) {
	if availableTimeText == "" {
		return execmodel.Time(0), nil
	}

	availableTime, err := strconv.ParseFloat(availableTimeText, 64)
	if err != nil {
		return 0, fmt.Errorf("fsm.ValidateAvailableTime: failed to parse available time header in fsmtable: %w", err)
	}
	if availableTime < 0 {
		return 0, fmt.Errorf("fsm.ValidateAvailableTime: available time cannot be negative")
	}
	return execmodel.Time(availableTime), nil
}

func ValidateAvailableTimeMap(m map[pfd.AtomicDeliverableID]string, ids *sets.Set[pfd.AtomicDeliverableID]) (map[pfd.AtomicDeliverableID]execmodel.Time, error) {
	m2 := make(map[pfd.AtomicDeliverableID]execmodel.Time, len(m))
	for d, availableTimeText := range m {
		if !ids.Contains(pfd.AtomicDeliverableID.Compare, d) {
			continue
		}

		availableTime, err := ValidateAvailableTime(availableTimeText)
		if err != nil {
			return nil, fmt.Errorf("fsm.ValidateAvailableTimeMap: deliverable: %q: %w", d, err)
		}
		m2[d] = availableTime
	}
	return m2, nil
}

func AvailableTimeFuncByTable(t *pfd.AtomicDeliverableTable, selectFunc pfd.ColumnSelectFunc, ids *sets.Set[pfd.AtomicDeliverableID]) (fsm.DeliverableAvailableTimeFunc, error) {
	m, err := RawAvailableTimeMap(t, selectFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.AvailableTimeFuncByTable: %w", err)
	}

	m2, err := ValidateAvailableTimeMap(m, ids)
	if err != nil {
		return nil, fmt.Errorf("fsm.AvailableTimeFuncByTable: %w", err)
	}
	return fsm.AvailableTimeFuncByMap(m2), nil
}

const (
	MaxRevisionHeaderJa = "最大版"
	MaxRevisionHeaderEn = "Max Revision"
)

var DefaultMaxRevisionColumnMatchFunc = pfd.ColumnMatchFunc(sets.New(
	strings.Compare,
	MaxRevisionHeaderJa,
	MaxRevisionHeaderEn,
))

func RawMaxRevisionMap(t *pfd.AtomicDeliverableTable, selectFunc pfd.ColumnSelectFunc) (map[pfd.AtomicDeliverableID]string, error) {
	m := make(map[pfd.AtomicDeliverableID]string, len(t.Rows))

	idx := selectFunc(t.ExtraHeaders)
	if idx < 0 {
		return nil, fmt.Errorf("fsm.MaxRevisionMapByTableFunc: missing max revision header in fsmtable")
	}

	for _, row := range t.Rows {
		m[row.ID] = row.ExtraCells[idx]
	}
	return m, nil
}

func ValidateMaxRevision(maxRevisionText string, isFeedbackSource bool) (int, error) {
	maxRevisionText = strings.TrimSpace(maxRevisionText)
	if isFeedbackSource {
		maxRevision, err := strconv.Atoi(maxRevisionText)
		if err != nil {
			return 0, fmt.Errorf("fsm.ValidateMaxRevision: failed to parse max revision header in table: %w", err)
		}

		if maxRevision < 1 {
			return 0, fmt.Errorf("fsm.ValidateMaxRevision: max revision cannot be less than 1")
		}

		return maxRevision, nil
	}

	if maxRevisionText != "" && maxRevisionText != "-" {
		return 0, fmt.Errorf("fsm.ValidateMaxRevision: must be empty or '-': %q", maxRevisionText)
	}

	return 0, nil
}

func ValidateMaxRevisionMap(m map[pfd.AtomicDeliverableID]string, feedbackSources *sets.Set[pfd.AtomicDeliverableID]) (map[pfd.AtomicDeliverableID]int, error) {
	m2 := make(map[pfd.AtomicDeliverableID]int, len(m))
	for d, maxRevisionText := range m {
		isFeedbackSource := feedbackSources.Contains(pfd.AtomicDeliverableID.Compare, d)
		maxRevision, err := ValidateMaxRevision(maxRevisionText, isFeedbackSource)
		if err != nil {
			return nil, fmt.Errorf("fsm.ValidateMaxRevisionMap: deliverable: %q: %w", d, err)
		}
		if isFeedbackSource {
			m2[d] = maxRevision
		}
	}
	return m2, nil
}

func MaxRevisionMapByTableFunc(t *pfd.AtomicDeliverableTable, selectFunc pfd.ColumnSelectFunc, feedbackSources *sets.Set[pfd.AtomicDeliverableID]) (map[pfd.AtomicDeliverableID]int, error) {
	m, err := RawMaxRevisionMap(t, selectFunc)
	if err != nil {
		return nil, fmt.Errorf("fsm.MaxRevisionMapByTableFunc: %w", err)
	}
	m2, err := ValidateMaxRevisionMap(m, feedbackSources)
	if err != nil {
		return nil, fmt.Errorf("fsm.MaxRevisionMapByTableFunc: %w", err)
	}
	return m2, nil
}

package pfdtsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"slices"

	"github.com/Kuniwak/pfd-tools/pfd"
)

func RenumberPlanHeader() []string {
	return []string{"Key", "ID"}
}

func WriteRenumberPlan(w io.Writer, plan pfd.RenumberPlan) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write(RenumberPlanHeader()); err != nil {
		return fmt.Errorf("pfdtsv.WriteRenumberPlan: %w", err)
	}

	keys := make([]string, 0, len(plan))
	for key := range plan {
		keys = append(keys, string(key))
	}
	slices.Sort(keys)

	for _, key := range keys {
		if err := csvWriter.Write([]string{key, string(plan[pfd.NodeID(key)].ID)}); err != nil {
			return fmt.Errorf("pfdtsv.WriteRenumberPlan: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("pfdtsv.WriteRenumberPlan: %w", err)
	}
	return nil
}

func ParseRenumberPlan(r io.Reader) (pfd.RenumberPlan, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: cannot read the header %v: %w", RenumberPlanHeader(), err)
	}
	if !slices.Equal(header, RenumberPlanHeader()) {
		return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: unexpected header: %v (want %v)", header, RenumberPlanHeader())
	}

	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: %w", err)
	}

	plan := make(pfd.RenumberPlan, len(rows))
	keyByID := make(map[pfd.NodeID]pfd.NodeID, len(rows))
	for _, row := range rows {
		key, id := pfd.NodeID(row[0]), pfd.NodeID(row[1])

		if _, err := pfd.ParseNodeID(id); err != nil {
			return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: %w", err)
		}

		if _, err := pfd.ParseNodeID(key); err == nil {
			return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: key must not be an ID: %q", key)
		}

		keyType, hasKind := key.UnnumberedNodeType()
		if !hasKind && key != "" {
			return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: 旧形式の採番計画です（キー %q に種別がありません）。プロセスは %q、成果物は %q のように種別で囲むか、pfdrenum で作り直してください",
				key,
				pfd.NewUnnumberedNodeID(pfd.NodeTypeAtomicProcess, key.Label()),
				pfd.NewUnnumberedNodeID(pfd.NodeTypeAtomicDeliverable, key.Label()))
		}

		if hasKind {
			if keyType.IsProcess() {
				if _, err := pfd.ParseProcessID(id); err != nil {
					return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: the process key %q has a non-process ID %q", key, id)
				}
			} else {
				if _, err := pfd.ParseDeliverableID(id); err != nil {
					return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: the deliverable key %q has a non-deliverable ID %q", key, id)
				}
			}
		}

		if _, dup := plan[key]; dup {
			return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: duplicate key: %q", key)
		}

		if dup, ok := keyByID[id]; ok {
			return nil, fmt.Errorf("pfdtsv.ParseRenumberPlan: duplicate ID %q for keys %q and %q", id, dup, key)
		}
		keyByID[id] = key

		plan[key] = &pfd.Node{ID: id, Description: key.Label()}
	}
	return plan, nil
}

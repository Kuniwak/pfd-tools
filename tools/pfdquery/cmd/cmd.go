package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmmasterschedule"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/version"
)

var ErrNotFound = errors.New("query not found")

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	opts, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(opts, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	return 0
}

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	w := csv.NewWriter(inout.Stdout)
	defer w.Flush()
	w.Comma = '\t'

	w.Write([]string{"QUERY", "SOURCE", "RESULT"})
	sb := &strings.Builder{}
	found := false
	for _, query := range opts.Queries {
		ok, err := respondToQuery(query, opts, w, sb)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		if ok {
			found = true
		}
	}

	if !found {
		return ErrNotFound
	}

	return nil
}

func respondToQuery(query string, opts *Options, w *csv.Writer, sb *strings.Builder) (bool, error) {
	found := false
	opts.CommonOptions.Logger.Debug("respondToQuery", "query", query, "opts", opts)

	if opts.PFD != nil {
		nodeMap := pfd.NewNodeMap(opts.PFD.Nodes, opts.CommonOptions.Logger)
		node, ok := nodeMap[pfd.NodeID(query)]
		if ok {
			found = true
			w.Write([]string{query, "PFD[desc]", escapeString(node.Description)})
			w.Write([]string{query, "PFD[type]", string(node.Type)})
			sb.Reset()
			for i, input := range opts.PFD.InputsExceptFeedback(node.ID).Iter() {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(string(input))
			}
			w.Write([]string{query, "PFD[input]", sb.String()})
			sb.Reset()
			for i, fbInput := range opts.PFD.InputsOnlyFeedback(node.ID).Iter() {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(string(fbInput))
			}
			w.Write([]string{query, "PFD[fb_input]", sb.String()})
			sb.Reset()
			for i, output := range opts.PFD.OutputsExceptFeedback(node.ID).Iter() {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(string(output))
			}
			w.Write([]string{query, "PFD[output]", sb.String()})
			sb.Reset()
			for i, fbOutput := range opts.PFD.OutputsOnlyFeedback(node.ID).Iter() {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(string(fbOutput))
			}
			w.Write([]string{query, "PFD[fb_output]", sb.String()})
		}

		p, err := pfd.NewSafePFDByUnsafePFD(opts.PFD)
		if err == nil {
			if opts.BackwardReachable {
				backwardReachable := sets.New(pfd.AtomicProcessID.Compare)
				p.CollectBackwardReachableAtomicProcessesExceptFeedback(pfd.AtomicProcessID(query), backwardReachable, opts.CommonOptions.Logger)
				sb.Reset()
				for i, ap := range backwardReachable.Iter() {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(string(ap))
				}
				w.Write([]string{query, "BACKWARD_REACHABLE", sb.String()})
			}
			if opts.BackwardReachableFeedbackDestination {
				backwardReachableFeedbackDestination := sets.New(pfd.AtomicProcessID.Compare)
				p.CollectBackwardReachableAtomicProcessesExceptFeedback(pfd.AtomicProcessID(query), backwardReachableFeedbackDestination, opts.CommonOptions.Logger)
				sb.Reset()
				for i, ap := range backwardReachableFeedbackDestination.Iter() {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(string(ap))
				}
				w.Write([]string{query, "BACKWARD_REACHABLE_FB", sb.String()})
			}
			if opts.Reachable {
				reachable := sets.New(pfd.AtomicProcessID.Compare)
				p.CollectReachableAtomicProcessesExceptFeedback(pfd.AtomicProcessID(query), reachable, opts.CommonOptions.Logger)
				sb.Reset()
				for i, ap := range reachable.Iter() {
					if i > 0 {
						sb.WriteString(",")
					}
					sb.WriteString(string(ap))
				}
				w.Write([]string{query, "REACHABLE", sb.String()})
			}
		} else {
			opts.CommonOptions.Logger.Warn("respondToQuery", "error", err.Error())
		}
	}

	if opts.AtomicProcessTable != nil {
		for _, row := range opts.AtomicProcessTable.Rows {
			if string(row.ID) == query {
				found = true
				w.Write([]string{query, "AP_TABLE[desc]", escapeString(row.Description)})
				for i, extraCell := range row.ExtraCells {
					w.Write([]string{query, fmt.Sprintf("AP_TABLE[%s]", opts.AtomicProcessTable.ExtraHeaders[i]), escapeString(extraCell)})
				}
			}
		}
	}

	if opts.AtomicDeliverableTable != nil {
		for _, row := range opts.AtomicDeliverableTable.Rows {
			if string(row.ID) == query {
				found = true
				w.Write([]string{query, "AD_TABLE[desc]", escapeString(row.Description)})
				for i, extraCell := range row.ExtraCells {
					w.Write([]string{query, fmt.Sprintf("AD_TABLE[%s]", opts.AtomicDeliverableTable.ExtraHeaders[i]), escapeString(extraCell)})
				}
			}
		}
	}

	if opts.CompositeProcessTable != nil {
		for _, row := range opts.CompositeProcessTable.Rows {
			if string(row.ID) == query {
				found = true
				w.Write([]string{query, "CP_TABLE[desc]", escapeString(row.Description)})
				for i, extraCell := range row.ExtraCells {
					w.Write([]string{query, fmt.Sprintf("CP_TABLE[%s]", opts.CompositeProcessTable.ExtraHeaders[i]), escapeString(extraCell)})
				}
			}
		}
	}

	if opts.CompositeDeliverableTable != nil {
		for _, row := range opts.CompositeDeliverableTable.Rows {
			if string(row.ID) == query {
				found = true
				w.Write([]string{query, "CD_TABLE[desc]", escapeString(row.Description)})
				for _, deliverable := range row.Deliverables {
					w.Write([]string{query, "CD_TABLE[deliverable]", string(deliverable)})
				}
				for i, extraCell := range row.ExtraCells {
					w.Write([]string{query, fmt.Sprintf("CD_TABLE[%s]", opts.CompositeDeliverableTable.ExtraHeaders[i]), escapeString(extraCell)})
				}
			}
		}
	}

	if opts.MilestoneTable != nil {
		ms1 := sets.New(fsmmasterschedule.Milestone.Compare)
		ms2 := sets.New(fsmmasterschedule.Milestone.Compare)
		for _, row := range opts.MilestoneTable.Rows {
			if string(row.MilestoneID) == query {
				found = true
				w.Write([]string{query, "MILESTONE_TABLE[desc]", escapeString(row.Description)})
			}
			if gs, err := fsmtable.ParseGroups(row.GroupIDs); err == nil {
				w.Write([]string{query, "MILESTONE_TABLE[groups]", row.GroupIDs})
				for _, groupID := range gs.Iter() {
					if string(groupID) == query {
						found = true
						ms1.Add(fsmmasterschedule.Milestone.Compare, row.MilestoneID)
					}
				}
			}
			if successors, err := fsmtable.ParseSuccessors(row.Successors); err == nil {
				w.Write([]string{query, "MILESTONE_TABLE[successors]", row.Successors})
				for _, successor := range successors.Iter() {
					if string(successor) == query {
						found = true
						ms2.Add(fsmmasterschedule.Milestone.Compare, row.MilestoneID)
					}
				}
			}
		}
		sb.Reset()
		for i, milestone := range ms1.Iter() {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(string(milestone))
		}
		w.Write([]string{query, "MILESTONE_TABLE[group_member]", sb.String()})
		sb.Reset()
		for i, milestone := range ms2.Iter() {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(string(milestone))
		}
		w.Write([]string{query, "MILESTONE_TABLE[predecessor]", sb.String()})
	}

	if opts.GroupTable != nil {
		for _, row := range opts.GroupTable.Rows {
			if string(row.ID) == query {
				found = true
				w.Write([]string{query, "GROUP_TABLE[desc]", escapeString(row.Description)})
			}
		}
	}

	if opts.PFD != nil && opts.AtomicProcessTable != nil && opts.AtomicDeliverableTable != nil && opts.CompositeDeliverableTable != nil && opts.ResourceTable != nil {
		fsmEnvSeed := &tools.FSMEnvSeed{
			PFD:                       opts.PFD,
			AtomicProcessTable:        opts.AtomicProcessTable,
			AtomicDeliverableTable:    opts.AtomicDeliverableTable,
			CompositeDeliverableTable: opts.CompositeDeliverableTable,
			ResourceTable:             opts.ResourceTable,
		}
		e, err := tools.FSMPrepare(fsmEnvSeed, opts.CommonOptions.Locale, opts.CommonOptions.Logger)
		if err == nil {
			if e.PFD.AtomicProcesses.Contains(pfd.AtomicProcessID.Compare, pfd.AtomicProcessID(query)) {
				ap := pfd.AtomicProcessID(query)
				precondition, ok := e.PreconditionMap[ap]
				if ok {
					found = true
					p := precondition.Compile(e.PFD, opts.CommonOptions.Logger)
					sb := &strings.Builder{}
					p.Write(sb)
					w.Write([]string{query, "PRECONDITION", escapeString(sb.String())})
				} else {
					w.Write([]string{query, "PRECONDITION", "not found"})
				}
			}
		} else {
			opts.CommonOptions.Logger.Warn("respondToQuery", "error", err.Error())
		}
	}

	return found, nil
}

func escapeString(s string) string {
	return strings.Trim(fmt.Sprintf("%q", s), `"`)
}

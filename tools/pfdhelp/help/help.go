package help

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/cli"
	bizdaycmd "github.com/Kuniwak/pfd-tools/tools/bizday/cmd"
	criticalpathcmd "github.com/Kuniwak/pfd-tools/tools/criticalpath/cmd"
	pfddeadlockcmd "github.com/Kuniwak/pfd-tools/tools/debug/pfddeadlock/cmd"
	pfddotcmd "github.com/Kuniwak/pfd-tools/tools/debug/pfddot/cmd"
	pfdparsecmd "github.com/Kuniwak/pfd-tools/tools/debug/pfdparse/cmd"
	pfdruncmd "github.com/Kuniwak/pfd-tools/tools/debug/pfdrun/cmd"
	pfdrungraphcmd "github.com/Kuniwak/pfd-tools/tools/debug/pfdrungraph/cmd"
	holidayscmd "github.com/Kuniwak/pfd-tools/tools/holidays/cmd"
	pfdcpcompcmd "github.com/Kuniwak/pfd-tools/tools/pfdcpcomp/cmd"
	pfddiffcmd "github.com/Kuniwak/pfd-tools/tools/pfddiff/cmd"
	pfddupmarkcmd "github.com/Kuniwak/pfd-tools/tools/pfddupmark/cmd"
	pfdestcalloutcmd "github.com/Kuniwak/pfd-tools/tools/pfdestcallout/cmd"
	pfdesttablecmd "github.com/Kuniwak/pfd-tools/tools/pfdesttable/cmd"
	pfdfixcmd "github.com/Kuniwak/pfd-tools/tools/pfdfix/cmd"
	pfdlintcmd "github.com/Kuniwak/pfd-tools/tools/pfdlint/cmd"
	pfdplancmd "github.com/Kuniwak/pfd-tools/tools/pfdplan/cmd"
	pfdquerycmd "github.com/Kuniwak/pfd-tools/tools/pfdquery/cmd"
	pfdrenumcmd "github.com/Kuniwak/pfd-tools/tools/pfdrenum/cmd"
	pfdrescmd "github.com/Kuniwak/pfd-tools/tools/pfdres/cmd"
	pfdsortcmd "github.com/Kuniwak/pfd-tools/tools/pfdsort/cmd"
	pfdtablecmd "github.com/Kuniwak/pfd-tools/tools/pfdtable/cmd"
	pfdticketcmd "github.com/Kuniwak/pfd-tools/tools/pfdticket/cmd"
	planmastercmd "github.com/Kuniwak/pfd-tools/tools/planmaster/cmd"
	plantimelinecmd "github.com/Kuniwak/pfd-tools/tools/plantimeline/cmd"
)

type ToolHelp struct {
	Name string

	Command cli.Command

	Short string
}

var Tools = []ToolHelp{
	{"bizday", bizdaycmd.MainCommandByArgs, bizdaycmd.ShortHelp},
	{"criticalpath", criticalpathcmd.MainCommandByArgs, criticalpathcmd.ShortHelp},
	{"holidays", holidayscmd.MainCommandByArgs, holidayscmd.ShortHelp},
	{"pfdcpcomp", pfdcpcompcmd.MainCommandByArgs, pfdcpcompcmd.ShortHelp},
	{"pfddeadlock", pfddeadlockcmd.MainCommandByArgs, pfddeadlockcmd.ShortHelp},
	{"pfddiff", pfddiffcmd.MainCommandByArgs, pfddiffcmd.ShortHelp},
	{"pfddot", pfddotcmd.MainCommandByArgs, pfddotcmd.ShortHelp},
	{"pfddupmark", pfddupmarkcmd.MainCommandByArgs, pfddupmarkcmd.ShortHelp},
	{"pfdestcallout", pfdestcalloutcmd.MainCommandByArgs, pfdestcalloutcmd.ShortHelp},
	{"pfdesttable", pfdesttablecmd.MainCommandByArgs, pfdesttablecmd.ShortHelp},
	{"pfdfix", pfdfixcmd.MainCommandByArgs, pfdfixcmd.ShortHelp},
	{"pfdlint", pfdlintcmd.MainCommandByArgs, pfdlintcmd.ShortHelp},
	{"pfdparse", pfdparsecmd.MainCommandByArgs, pfdparsecmd.ShortHelp},
	{"pfdplan", pfdplancmd.MainCommandByArgs, pfdplancmd.ShortHelp},
	{"pfdquery", pfdquerycmd.MainCommandByArgs, pfdquerycmd.ShortHelp},
	{"pfdrenum", pfdrenumcmd.MainCommandByArgs, pfdrenumcmd.ShortHelp},
	{"pfdres", pfdrescmd.MainCommandByArgs, pfdrescmd.ShortHelp},
	{"pfdrun", pfdruncmd.MainCommandByArgs, pfdruncmd.ShortHelp},
	{"pfdrungraph", pfdrungraphcmd.MainCommandByArgs, pfdrungraphcmd.ShortHelp},
	{"pfdsort", pfdsortcmd.MainCommandByArgs, pfdsortcmd.ShortHelp},
	{"pfdtable", pfdtablecmd.MainCommandByArgs, pfdtablecmd.ShortHelp},
	{"pfdticket", pfdticketcmd.MainCommandByArgs, pfdticketcmd.ShortHelp},
	{"planmaster", planmastercmd.MainCommandByArgs, planmastercmd.ShortHelp},
	{"plantimeline", plantimelinecmd.MainCommandByArgs, plantimelinecmd.ShortHelp},
}

func FindTool(name string) (ToolHelp, bool) {
	for _, tool := range Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolHelp{}, false
}

func SelectTools(names []string) ([]ToolHelp, error) {
	if len(names) == 0 {
		return Tools, nil
	}
	selected := make([]ToolHelp, 0, len(names))
	for _, name := range names {
		tool, ok := FindTool(name)
		if !ok {
			return nil, fmt.Errorf("help.SelectTools: unknown tool: %q", name)
		}
		selected = append(selected, tool)
	}
	return selected, nil
}

func WriteHelp(w io.Writer, tools []ToolHelp) []string {
	toolInout := cli.StubProcInout()
	toolInout.Stdout = w
	toolInout.Stderr = w
	var failed []string
	for i, tool := range tools {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "# %s\n\n", tool.Name)
		if code := tool.Command([]string{"-h"}, toolInout); code != 0 {
			failed = append(failed, tool.Name)
		}
	}
	return failed
}

func WriteShortHelp(w io.Writer, tools []ToolHelp) error {
	cw := csv.NewWriter(w)
	cw.Comma = '\t'
	for _, tool := range tools {
		if err := cw.Write([]string{tool.Name, tool.Short}); err != nil {
			return fmt.Errorf("help.WriteShortHelp: %w", err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("help.WriteShortHelp: %w", err)
	}
	return nil
}

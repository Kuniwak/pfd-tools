package cmd

import (
	"fmt"
	"slices"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/holidays"
	"github.com/Kuniwak/pfd-tools/version"
)

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

func MainCommandByOptions(options *Options, inout *cli.ProcInout) error {
	if options.CommonOptions.Help {
		return nil
	}

	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	hs := make([]string, 0, holidays.JPHolidays().Len())

	for _, holiday := range holidays.JPHolidays().Iter() {
		hs = append(hs, holiday.Format("2006-01-02"))
	}

	slices.Sort(hs)

	for _, holiday := range hs {
		fmt.Fprintln(inout.Stdout, holiday)
	}
	return nil
}

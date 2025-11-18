package cmd

import (
	"fmt"
	"time"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	options, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(options, inout); err != nil {
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

	unit := 1.0 / options.BusinessTimeFuncOptions.Duration.Hours()
	for t := 0.0; t < 100000.0; t += unit {
		d := options.BusinessTimeFuncOptions.BusinessTimeFunc(options.BusinessTimeFuncOptions.StartDay, t)
		if time.Time.Compare(d, options.Time) >= 0 {
			fmt.Fprintf(inout.Stdout, "%f\n", t)
			return nil
		}
	}
	return fmt.Errorf("cmd.MainCommandByOptions: time is not in the business time range")
}

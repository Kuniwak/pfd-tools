package cmd

import (
	"fmt"
	"strings"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools/pfdhelp/help"
	"github.com/Kuniwak/pfd-tools/version"
)

const ShortHelp = "pfd-tools 各コマンドのヘルプをまとめて表示します。"

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

	if opts.CommonOptions.ShortHelp {
		fmt.Fprintln(inout.Stdout, ShortHelp)
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	if opts.Short {
		if err := help.WriteShortHelp(inout.Stdout, opts.Tools); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		return nil
	}

	if failed := help.WriteHelp(inout.Stdout, opts.Tools); len(failed) > 0 {
		fmt.Fprintf(inout.Stderr, "pfdhelp: warning: -h failed for: %s\n", strings.Join(failed, ", "))
	}
	return nil
}

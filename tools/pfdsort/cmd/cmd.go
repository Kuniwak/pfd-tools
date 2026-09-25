package cmd

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/version"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const ShortHelp = "PFD を読みやすく自動整列します。"

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

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	input, err := io.ReadAll(opts.Reader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	sortOpts := pfddrawio.SortOptions{
		DupRankSpan:    opts.DupRankSpan,
		DupRowSpan:     opts.DupRowSpan,
		HGap:           opts.HGap,
		VGap:           opts.VGap,
		OnlyPages:      opts.OnlyPages,
		OnlyNodes:      opts.OnlyNodes,
		PosRestriction: opts.PosRestriction,
		BreakCycles:    opts.BreakCycles,
	}

	out, err := pfdfmt.TransformDrawioBytes(input, func(r io.Reader, l *slog.Logger) ([]*xmldom.Node, error) {
		return pfddrawio.Sort(r, sortOpts, l)
	}, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	inplacePath := ""
	if opts.Inplace {
		inplacePath = opts.InputFilePath
	}
	if err := tools.WriteOutput(inout.Stdout, inplacePath, out); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	return nil
}

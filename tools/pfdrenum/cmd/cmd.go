package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
	"github.com/Kuniwak/pfd-tools/slograw"
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

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	f, r, err := pfdfmt.Detect(opts.Reader)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	logger.Debug("detected format", "format", f)

	switch f {
	case pfdfmt.FormatDrawio:
		nodes, err := pfddrawio.Renumber(r, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		for _, node := range nodes {
			if err := node.Write(opts.Writer); err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
		}
	case pfdfmt.FormatDrawioPNG:
		pngBytes, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		xmlReader, err := pfddrawiopng.ExtractMxfile(bytes.NewReader(pngBytes))
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		nodes, err := pfddrawio.Renumber(xmlReader, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		xmlBuf := bytes.NewBuffer(nil)
		for _, node := range nodes {
			if err := node.Write(xmlBuf); err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
		}
		out, err := pfddrawiopng.ReplaceMxfile(pngBytes, xmlBuf.Bytes())
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		if _, err := opts.Writer.Write(out); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	default:
		return fmt.Errorf("cmd.MainCommandByOptions: not supported format: %q", f)
	}

	return nil
}

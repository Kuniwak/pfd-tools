package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Reader        io.Reader
	Inplace       bool
	InputFilePath string
	DeleteUnhit   bool
	ExpandX       float64
	ExpandY       float64
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdfix", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdfix [options] [path/to/pfd.drawio]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
  $ pfdfix path/to/pfd.drawio
  $ pfdfix -inplace path/to/pfd.drawio
  $ pfdfix -delete-unhit path/to/pfd.drawio
  $ pfdfix -hitbox-expand 1.1:0.9 path/to/pfd.drawio
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	inplaceFlag := flags.Bool("inplace", false, "overwrite the file in place")
	deleteUnhitFlag := flags.Bool("delete-unhit", false, "delete edges that cannot be connected: endpoint in no hitbox, or source and target resolve to the same vertex (self-loop) (default: keep and warn)")
	hitboxExpandFlag := flags.String("hitbox-expand", "1.0", `scale each hitbox around its center: "x:y" or "n" (=n:n). 1.0=100%, 0.9=90%`)

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Options{CommonOptions: &tools.CommonOptions{Help: true}}, nil
		}
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	commonOptions, err := tools.ValidateCommonOptions(&commonRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if commonOptions.ShortHelp {
		return &Options{CommonOptions: commonOptions}, nil
	}
	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	expandX, expandY, err := ParseHitboxExpand(*hitboxExpandFlag)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	shouldReadFromStdin := false
	var inputFilePath string

	if flags.NArg() < 1 {
		shouldReadFromStdin = true
	} else if flags.NArg() > 1 {
		return nil, fmt.Errorf("cmd.ParseOptions: too many arguments")
	} else {
		inputFilePath = flags.Arg(0)
		if inputFilePath == "" {
			shouldReadFromStdin = true
		}
	}

	var r io.Reader
	if shouldReadFromStdin {
		if *inplaceFlag {
			return nil, fmt.Errorf("cmd.ParseOptions: -inplace requires a file path")
		}
		r = inout.Stdin
	} else {
		content, err := os.ReadFile(inputFilePath)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		r = bytes.NewReader(content)
	}

	return &Options{
		CommonOptions: commonOptions,
		Reader:        r,
		Inplace:       *inplaceFlag,
		InputFilePath: inputFilePath,
		DeleteUnhit:   *deleteUnhitFlag,
		ExpandX:       expandX,
		ExpandY:       expandY,
	}, nil
}

func ParseHitboxExpand(s string) (float64, float64, error) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		n, err := parseNonNegativeFloat(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("cmd.ParseHitboxExpand: %w", err)
		}
		return n, n, nil
	case 2:
		x, err := parseNonNegativeFloat(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("cmd.ParseHitboxExpand: %w", err)
		}
		y, err := parseNonNegativeFloat(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("cmd.ParseHitboxExpand: %w", err)
		}
		return x, y, nil
	default:
		return 0, 0, fmt.Errorf("cmd.ParseHitboxExpand: invalid format: %q (want \"x:y\" or \"n\")", s)
	}
}

func parseNonNegativeFloat(s string) (float64, error) {
	trimmed := strings.TrimSpace(s)
	f, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %q", trimmed)
	}
	if f < 0 {
		return 0, fmt.Errorf("must be >= 0.0: %q", trimmed)
	}
	return f, nil
}

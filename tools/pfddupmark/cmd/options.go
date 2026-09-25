package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions *tools.CommonOptions
	Reader        io.Reader
	Inplace       bool
	InputFilePath string
	OnlyPages     []string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfddupmark", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfddupmark [options] [path/to/pfd.drawio]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
  $ pfddupmark path/to/pfd.drawio
  $ pfddupmark -inplace path/to/pfd.drawio
  $ pfddupmark -only-page P123,P456 path/to/pfd.drawio
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	inplaceFlag := flags.Bool("inplace", false, "overwrite the file in place")
	onlyPageFlag := flags.String("only-page", "", `mark only the given pages (comma-separated page names, e.g. "P123,P456"; empty = all pages)`)

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
		OnlyPages:     tools.ParsePageList(*onlyPageFlag),
	}, nil
}

package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/tools"
)

const OutputFormatDrawio = "drawio"

const OutputFormatTSV = "tsv"

const OutputFormatMaxID = "maxid"

type Mode string

const (
	ModeRenumber Mode = "renumber"

	ModeEmitPlan Mode = "emit-plan"

	ModeApplyPlan Mode = "apply-plan"

	ModeExtendPlan Mode = "extend-plan"

	ModeEmitMaxID Mode = "emit-maxid"
)

func ParseMode(outputFormat string, inplace bool, planPath string, minID string) (Mode, error) {
	switch outputFormat {
	case OutputFormatDrawio:
		if planPath == "" {
			return ModeRenumber, nil
		}
		if minID != "" {
			return "", fmt.Errorf("cmd.ParseMode: -min-id cannot be used with -renum-plan: applying an existing plan does not renumber")
		}
		return ModeApplyPlan, nil

	case OutputFormatTSV:
		if inplace {
			return "", fmt.Errorf("cmd.ParseMode: -out-format tsv cannot be used with -inplace: it would overwrite the drawio with a TSV")
		}
		if planPath != "" {
			return ModeExtendPlan, nil
		}
		return ModeEmitPlan, nil

	case OutputFormatMaxID:
		if inplace {
			return "", fmt.Errorf("cmd.ParseMode: -out-format maxid cannot be used with -inplace: it would overwrite the drawio with an ID")
		}
		if planPath != "" {
			return "", fmt.Errorf("cmd.ParseMode: -out-format maxid cannot be used with -renum-plan: -out-format maxid does not renumber")
		}
		if minID != "" {
			return "", fmt.Errorf("cmd.ParseMode: -out-format maxid cannot be used with -min-id: -out-format maxid does not renumber")
		}
		return ModeEmitMaxID, nil

	default:
		return "", fmt.Errorf("cmd.ParseMode: unknown output format: %q (available: %s, %s, %s)", outputFormat, OutputFormatDrawio, OutputFormatTSV, OutputFormatMaxID)
	}
}

type Options struct {
	CommonOptions *tools.CommonOptions
	Reader        io.Reader
	Inplace       bool
	InputFilePath string
	Mode          Mode

	RenumberPlan pfd.RenumberPlan

	MaxProcessNumber     int
	MaxDeliverableNumber int
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdrenum", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdrenum [options]", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Example
  $ pfdrenum path/to/pfd.drawio
  <mxfile host="65bd71144e">
    <diagram id="1ni4HEU6g7zc3-6eLzPC" name="P0">
    ...

  $ pfdrenum -inplace path/to/pfd.drawio

  $ # Renumber a PFD split into multiple files: decide the plan on the whole PFD,
  $ # then apply it to each file.
  $ drawiocat *.drawio | pfdrenum -out-format tsv >renum.tsv
  Key	ID
  (テストを書く)	P12
  $ for f in *.drawio; do pfdrenum -renum-plan renum.tsv -inplace "$f"; done
  $ drawiocat *.drawio | pfdrenum -out-format tsv   # no rows means every ID has been applied
  Key	ID

  $ # Number variations of a composite process uniquely across all of them:
  $ # grow one plan by feeding the variations through it one by one.
  $ floor=$(for f in parent.drawio variants/*/*.drawio; do pfdrenum -out-format maxid "$f"; done)
  $ printf 'Key\tID\n' >renum.tsv
  $ for v in variants/*; do
  >   drawiocat parent.drawio "$v"/*.drawio |
  >     pfdrenum -min-id "$floor" -renum-plan renum.tsv -out-format tsv >renum.tsv.new
  >   mv renum.tsv.new renum.tsv
  > done
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	inplaceFlag := flags.Bool("inplace", false, "overwrite the file in place")
	outputFormatFlag := flags.String(tools.OutFormatFlag, string(OutputFormatDrawio), "output format (available: drawio(renumbered drawio), tsv(renumber plan), maxid(the max IDs already used))")
	renumberPlanFlag := flags.String("renum-plan", "", "apply the existing renumber plan (TSV) instead of deciding a new one. with -out-format tsv, use it as the base of a new plan instead")
	minIDFlag := flags.String("min-id", "", `numbering floor: new IDs come after the given ones (e.g. "P12 D34", as printed by -out-format maxid)`)

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

	mode, err := ParseMode(*outputFormatFlag, *inplaceFlag, *renumberPlanFlag, *minIDFlag)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	maxProcessNumber, maxDeliverableNumber, err := pfd.ParseMaxIDs(*minIDFlag)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var plan pfd.RenumberPlan
	if mode == ModeApplyPlan || mode == ModeExtendPlan {
		f, err := os.Open(*renumberPlanFlag)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		defer f.Close()

		plan, err = pfdtsv.ParseRenumberPlan(f)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var r io.Reader

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
		CommonOptions:        commonOptions,
		Reader:               r,
		Inplace:              *inplaceFlag,
		InputFilePath:        inputFilePath,
		Mode:                 mode,
		RenumberPlan:         plan,
		MaxProcessNumber:     maxProcessNumber,
		MaxDeliverableNumber: maxDeliverableNumber,
	}, nil
}

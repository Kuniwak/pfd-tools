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
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/table"
	"github.com/Kuniwak/pfd-tools/tools"
)

type TableCategory string

const (
	TableCategoryPFD TableCategory = "PFD"
	TableCategoryFSM TableCategory = "FSM"
)

type Options struct {
	CommonOptions                   *tools.CommonOptions
	Writer                          io.Writer
	HasPFD                          bool
	PFDReader                       io.Reader
	AtomicProcessTableReader        io.Reader
	HasCompositeDeliverableTable    bool
	CompositeDeliverableTableReader io.Reader
	ExistingTableReader             io.Reader
	HasExistingTable                bool
	IsInplace                       bool
	TableCategory                   TableCategory
	PFDTableType                    pfd.TableType
	FSMTableType                    fsmtable.TableType
	InputFormat                     table.Format
	OutputFormat                    table.Format
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdtable", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: pfdtable [options]")
		_, _ = fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		_, _ = fmt.Fprintf(flags.Output(), `
Example
  $ pfdtable -t ad -p path/to/pfd.drawio
  ID      Description     Location
  D1      Implementation  https://example.com/1
  ...

  $ pfdtable -t ap -p path/to/pfd.drawio
  ID      Description
  P1      Implement
  ...

  $ pfdtable -t cp -p path/to/pfd.drawio
  ID      Description
  P1      Implement
  ...

  $ # Copy to clipboard as RTF (it is useful for pasting into Confluence and Microsoft Word and so on)
  $ pfdtable -t ad -o html path/to/pfd.drawio | textutil -stdin -format html -convert rtf -inputencoding UTF-8 -stdout | pbcopy

  $ # Print updated fsmtable from the existing fsmtable
  $ pfdtable -t ad -existing path/to/existing.tsv -p path/to/pfd.drawio
  ID      Description     Location
  D1      Implementation  https://example.com/1
  ...
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var pfdShortPath, pfdLongPath string
	tools.DeclarePFDOptions(flags, &pfdShortPath, &pfdLongPath)

	var atomicProcessTableShortPath, atomicProcessTableLongPath string
	tools.DeclareAtomicProcessTableOptions(flags, &atomicProcessTableShortPath, &atomicProcessTableLongPath)

	var compositeDeliverableTableShortPath, compositeDeliverableTableLongPath string
	tools.DeclareCompositeDeliverableTableOptions(flags, &compositeDeliverableTableShortPath, &compositeDeliverableTableLongPath)

	typeShortFlag := flags.String("t", "", "type of the table (available: ap(atomic-process), ad(atomic-deliverable), cp(composite-process), cd(composite-deliverable), r(resource))")
	typeFlag := flags.String("type", "", "type of the table (available: ap(atomic-process), ad(atomic-deliverable), cp(composite-process), cd(composite-deliverable), r(resource))")
	existingPathFlag := flags.String("existing", "", "path of the existing fsmtable")
	inplaceFlag := flags.Bool("inplace", false, "overwrite the file in place")
	inputFormatShortFlag := flags.String("i", "", "format of the input PFD")
	inputFormatFlag := flags.String("input-format", "", "format of the input PFD")
	outputFormatShortFlag := flags.String("o", "", "format of the output fsmtable")
	outputFormatFlag := flags.String("output-format", "", "format of the output fsmtable")

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

	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var tableTypeString string
	if *typeShortFlag != "" {
		tableTypeString = *typeShortFlag
	} else {
		tableTypeString = *typeFlag
	}

	var pfdReader io.Reader
	hasPFD := false
	if pfdShortPath != "" || pfdLongPath != "" {
		pfdReader, _, err = tools.ValidatePFDOptions(&pfdShortPath, &pfdLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		hasPFD = true
	}

	var compositeDeliverableTableReader io.Reader
	hasCompositeDeliverableTable := compositeDeliverableTableShortPath != "" || compositeDeliverableTableLongPath != ""
	if hasCompositeDeliverableTable {
		compositeDeliverableTableReader, _, err = tools.ValidateCompositeDeliverableTableOptions(&compositeDeliverableTableShortPath, &compositeDeliverableTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var apTableReader io.Reader
	var pfdTableType pfd.TableType
	var fsmTableType fsmtable.TableType
	var tableCategory TableCategory
	switch tableTypeString {
	case "ap", "atomic-process":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicProcess
	case "ad", "deliverable":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicDeliverable
	case "cp", "composite-process":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeCompositeProcess
	case "cd", "composite-deliverable":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeCompositeDeliverable
	case "r", "res", "resource":
		apTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		tableCategory = TableCategoryFSM
		fsmTableType = fsmtable.TableTypeResource
	case "m", "milestone":
		apTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		tableCategory = TableCategoryFSM
		fsmTableType = fsmtable.TableTypeMilestone
	default:
		return nil, fmt.Errorf("cmd.ParseOptions: invalid table type: %q", tableTypeString)
	}

	var outputFormat table.Format
	var outputFormatString string
	if *outputFormatShortFlag != "" {
		outputFormatString = *outputFormatShortFlag
	} else {
		outputFormatString = *outputFormatFlag
	}

	switch outputFormatString {
	case "tsv", "":
		outputFormat = table.FormatTSV
	case "html":
		outputFormat = table.FormatHTML
	default:
		return nil, fmt.Errorf("cmd.ParseOptions: invalid output format: %q", outputFormatString)
	}

	var inputFormat table.Format
	var inputFormatString string
	if *inputFormatShortFlag != "" {
		inputFormatString = *inputFormatShortFlag
	} else {
		inputFormatString = *inputFormatFlag
	}

	switch inputFormatString {
	case "tsv", "":
		inputFormat = table.FormatTSV
	case "html":
		inputFormat = table.FormatHTML
	default:
		return nil, fmt.Errorf("cmd.ParseOptions: invalid input format: %q", inputFormatString)
	}

	var existingTableReader io.Reader
	hasExistingTable := *existingPathFlag != ""
	if hasExistingTable {
		bs, err := os.ReadFile(*existingPathFlag)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		existingTableReader = bytes.NewReader(bs)
	}

	if !hasExistingTable && *inplaceFlag {
		return nil, fmt.Errorf("cmd.ParseOptions: inplace flag is only valid when existing table is specified")
	}

	var writer io.Writer
	if *inplaceFlag {
		writer, err = os.OpenFile(*existingPathFlag, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	} else {
		writer = inout.Stdout
	}

	return &Options{
		CommonOptions:                   commonOptions,
		HasPFD:                          hasPFD,
		PFDReader:                       pfdReader,
		HasCompositeDeliverableTable:    hasCompositeDeliverableTable,
		CompositeDeliverableTableReader: compositeDeliverableTableReader,
		AtomicProcessTableReader:        apTableReader,
		ExistingTableReader:             existingTableReader,
		HasExistingTable:                hasExistingTable,
		TableCategory:                   tableCategory,
		PFDTableType:                    pfdTableType,
		FSMTableType:                    fsmTableType,
		OutputFormat:                    outputFormat,
		InputFormat:                     inputFormat,
		IsInplace:                       *inplaceFlag,
		Writer:                          writer,
	}, nil
}

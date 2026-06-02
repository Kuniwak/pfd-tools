package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable"
	"github.com/Kuniwak/pfd-tools/table"
	"github.com/Kuniwak/pfd-tools/tools"
)

type TableCategory string

const (
	TableCategoryPFD TableCategory = "PFD"
	TableCategoryFSM TableCategory = "FSM"
	TableCategoryAll TableCategory = "ALL"
)

// ProjectFile holds decoded project.json data and its resolved location on disk.
type ProjectFile struct {
	Path     string
	BasePath string
	Raw      tools.FSMRawOptions
}

// existingTablePath returns the absolute path of the existing table for the given type.
// Returns "" if the corresponding field is absent in project.json.
func (p *ProjectFile) existingTablePath(category TableCategory, pfdType pfd.TableType, fsmType fsmtable.TableType) string {
	var rel string
	switch category {
	case TableCategoryPFD:
		switch pfdType {
		case pfd.TableTypeAtomicProcess:
			rel = p.Raw.AtomicProcessTablePath
		case pfd.TableTypeAtomicDeliverable:
			rel = p.Raw.AtomicDeliverableTablePath
		case pfd.TableTypeCompositeProcess:
			rel = p.Raw.CompositeProcessTablePath
		case pfd.TableTypeCompositeDeliverable:
			rel = p.Raw.CompositeDeliverableTablePath
		}
	case TableCategoryFSM:
		switch fsmType {
		case fsmtable.TableTypeResource:
			rel = p.Raw.ResourceTablePath
		case fsmtable.TableTypeMilestone:
			rel = p.Raw.MilestoneTablePath
		case fsmtable.TableTypeGroup:
			rel = p.Raw.GroupTablePath
		}
	}
	if rel == "" {
		return ""
	}
	return projResolve(p.BasePath, rel)
}

// projResolve resolves a path from a project.json relative to basePath.
// Absolute paths are returned as-is.
func projResolve(basePath, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(basePath, rel)
}

// AllExistingTables holds existing table readers for -t all + project.json refresh.
// A nil field means that table is absent in project.json and will be generated fresh from PFD.
// M and G are always derived fresh from AP and are not loaded here.
type AllExistingTables struct {
	AP io.Reader
	AD io.Reader
	CD io.Reader
	R  io.Reader
}

// InplaceTargets maps table identifiers to the absolute paths to overwrite on -inplace.
type InplaceTargets map[string]string

type Options struct {
	CommonOptions                   *tools.CommonOptions
	Writer                          io.Writer
	HasPFD                          bool
	PFDPath                         string
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
	Mode                            pfdtable.Mode
	InputFormat                     table.Format
	OutputFormat                    table.Format
	OutDir                          string
	ProjectFile                     *ProjectFile
	AllExisting                     *AllExistingTables
	InplaceTargets                  InplaceTargets
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

  $ # Print updated fsmtable using project.json (supplies PFD, existing table, and CD automatically)
  $ pfdtable -t ad -f path/to/project.json
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

	typeShortFlag := flags.String("t", "", "type of the table (available: ap(atomic-process), ad(atomic-deliverable), cp(composite-process), cd(composite-deliverable), r(resource), m(milestone), g(group), a(all); suffix '-plan' or '-plan-master' on ap/ad/a/all bootstraps required columns for pfdplan/planmaster (applies to AP/AD only; cp/cd are unaffected))")
	typeFlag := flags.String("type", "", "type of the table (available: ap(atomic-process), ad(atomic-deliverable), cp(composite-process), cd(composite-deliverable), r(resource), m(milestone), g(group), a(all); suffix '-plan' or '-plan-master' on ap/ad/a/all bootstraps required columns for pfdplan/planmaster (applies to AP/AD only; cp/cd are unaffected))")
	existingShortFlag := flags.String("f", "", "path of the existing fsmtable or project.json (alias of -existing)")
	existingPathFlag := flags.String("existing", "", "path of the existing fsmtable or project.json (same as -f)")
	inplaceFlag := flags.Bool("inplace", false, "overwrite the file in place")
	outDirFlag := flags.String("out-dir", "", "output directory (required for -t all)")
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

	existingPath := *existingPathFlag
	cli.OverrideIfSet(&existingPath, *existingShortFlag)

	// Detect and decode project.json early to enable fallback path resolution.
	var projectFile *ProjectFile
	isProjectJSON := existingPath != "" && strings.EqualFold(filepath.Ext(existingPath), ".json")
	if isProjectJSON {
		absProjPath, absErr := filepath.Abs(existingPath)
		if absErr != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", absErr)
		}
		projBasePath := filepath.Dir(absProjPath)

		f, openErr := os.Open(absProjPath)
		if openErr != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", openErr)
		}
		var raw tools.FSMRawOptions
		if decodeErr := json.NewDecoder(f).Decode(&raw); decodeErr != nil {
			f.Close()
			return nil, fmt.Errorf("cmd.ParseOptions: parsing project.json: %w", decodeErr)
		}
		f.Close()

		projectFile = &ProjectFile{Path: absProjPath, BasePath: projBasePath, Raw: raw}
	}

	var tableTypeString string
	if *typeShortFlag != "" {
		tableTypeString = *typeShortFlag
	} else {
		tableTypeString = *typeFlag
	}

	// Validate PFD from CLI flags, then fall back to project.json if not provided.
	var pfdReader io.Reader
	var pfdPath string
	hasPFD := false
	if pfdShortPath != "" || pfdLongPath != "" {
		pfdReader, pfdPath, err = tools.ValidatePFDOptions(&pfdShortPath, &pfdLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		hasPFD = true
	}
	if !hasPFD && isProjectJSON && projectFile.Raw.PFDPath != "" {
		resolvedPFD := projResolve(projectFile.BasePath, projectFile.Raw.PFDPath)
		shortFallback := resolvedPFD
		longFallback := ""
		pfdReader, pfdPath, err = tools.ValidatePFDOptions(&shortFallback, &longFallback, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		hasPFD = true
	}

	// Validate CD from CLI flags, then fall back to project.json if not provided.
	var compositeDeliverableTableReader io.Reader
	hasCompositeDeliverableTable := compositeDeliverableTableShortPath != "" || compositeDeliverableTableLongPath != ""
	if hasCompositeDeliverableTable {
		compositeDeliverableTableReader, _, err = tools.ValidateCompositeDeliverableTableOptions(&compositeDeliverableTableShortPath, &compositeDeliverableTableLongPath, cwd)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}
	isAllTableType := tableTypeString == "a" || tableTypeString == "all" ||
		tableTypeString == "a-plan" || tableTypeString == "all-plan" ||
		tableTypeString == "a-plan-master" || tableTypeString == "all-plan-master"
	if !hasCompositeDeliverableTable && isProjectJSON && !isAllTableType && projectFile.Raw.CompositeDeliverableTablePath != "" {
		cdPath := projResolve(projectFile.BasePath, projectFile.Raw.CompositeDeliverableTablePath)
		compositeDeliverableTableReader, err = os.OpenFile(cdPath, os.O_RDONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		hasCompositeDeliverableTable = true
	}

	// Pre-compute AP path from project.json for use inside the switch when CLI didn't supply it.
	var projAPPath string
	if isProjectJSON && atomicProcessTableShortPath == "" && atomicProcessTableLongPath == "" && projectFile.Raw.AtomicProcessTablePath != "" {
		projAPPath = projResolve(projectFile.BasePath, projectFile.Raw.AtomicProcessTablePath)
	}

	var apTableReader io.Reader
	var pfdTableType pfd.TableType
	var fsmTableType fsmtable.TableType
	var tableCategory TableCategory
	mode := pfdtable.ModeMinimal
	switch tableTypeString {
	case "ap", "atomic-process":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicProcess
	case "ap-plan":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicProcess
		mode = pfdtable.ModePlan
	case "ap-plan-master":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicProcess
		mode = pfdtable.ModePlanMaster
	case "ad", "deliverable":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicDeliverable
	case "ad-plan":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicDeliverable
		mode = pfdtable.ModePlan
	case "ad-plan-master":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeAtomicDeliverable
		mode = pfdtable.ModePlanMaster
	case "cp", "composite-process":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeCompositeProcess
	case "cd", "composite-deliverable":
		tableCategory = TableCategoryPFD
		pfdTableType = pfd.TableTypeCompositeDeliverable
	case "r", "res", "resource":
		if projAPPath != "" {
			apTableReader, err = os.OpenFile(projAPPath, os.O_RDONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		} else {
			apTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		}
		tableCategory = TableCategoryFSM
		fsmTableType = fsmtable.TableTypeResource
	case "m", "milestone":
		hasAP := atomicProcessTableShortPath != "" || atomicProcessTableLongPath != "" || projAPPath != ""
		if hasAP {
			if projAPPath != "" {
				apTableReader, err = os.OpenFile(projAPPath, os.O_RDONLY, 0644)
				if err != nil {
					return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
				}
			} else {
				apTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
				if err != nil {
					return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
				}
			}
		}
		tableCategory = TableCategoryFSM
		fsmTableType = fsmtable.TableTypeMilestone
	case "g", "group":
		hasAP := atomicProcessTableShortPath != "" || atomicProcessTableLongPath != "" || projAPPath != ""
		if hasAP {
			if projAPPath != "" {
				apTableReader, err = os.OpenFile(projAPPath, os.O_RDONLY, 0644)
				if err != nil {
					return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
				}
			} else {
				apTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
				if err != nil {
					return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
				}
			}
		}
		tableCategory = TableCategoryFSM
		fsmTableType = fsmtable.TableTypeGroup
	case "a", "all", "a-plan", "all-plan", "a-plan-master", "all-plan-master":
		if isProjectJSON {
			// -t all + project.json: requires exactly one of -out-dir or -inplace
			if *outDirFlag == "" && !*inplaceFlag {
				return nil, fmt.Errorf("cmd.ParseOptions: -t all with -existing project.json requires either -out-dir or -inplace")
			}
			if *outDirFlag != "" && *inplaceFlag {
				return nil, fmt.Errorf("cmd.ParseOptions: -out-dir and -inplace are mutually exclusive")
			}
		} else {
			if *outDirFlag == "" {
				return nil, fmt.Errorf("cmd.ParseOptions: -out-dir is required for -t all")
			}
			if existingPath != "" {
				return nil, fmt.Errorf("cmd.ParseOptions: -existing is not supported with -t all")
			}
			if *inplaceFlag {
				return nil, fmt.Errorf("cmd.ParseOptions: -inplace is not supported with -t all")
			}
		}
		if atomicProcessTableShortPath != "" || atomicProcessTableLongPath != "" {
			apTableReader, _, err = tools.ValidateAtomicProcessTableOptions(&atomicProcessTableShortPath, &atomicProcessTableLongPath, cwd)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		} else if projAPPath != "" {
			apTableReader, err = os.OpenFile(projAPPath, os.O_RDONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
			}
		}
		tableCategory = TableCategoryAll
		switch tableTypeString {
		case "a-plan", "all-plan":
			mode = pfdtable.ModePlan
		case "a-plan-master", "all-plan-master":
			mode = pfdtable.ModePlanMaster
		}
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

	// Resolve the existing table path (from project.json) or use the TSV path directly.
	var projExistingTablePath string
	if isProjectJSON && tableCategory != TableCategoryAll {
		projExistingTablePath = projectFile.existingTablePath(tableCategory, pfdTableType, fsmTableType)
		if projExistingTablePath == "" {
			return nil, fmt.Errorf("cmd.ParseOptions: project.json has no entry for -t %q", tableTypeString)
		}
	}

	// For -t all + project.json: build AllExisting readers and InplaceTargets.
	var allExisting *AllExistingTables
	var inplaceTargets InplaceTargets
	if isProjectJSON && tableCategory == TableCategoryAll {
		ae := &AllExistingTables{}
		if *inplaceFlag {
			inplaceTargets = make(InplaceTargets)
		}
		raw := projectFile.Raw
		bp := projectFile.BasePath
		loadAllTable := func(key, relPath string, dst *io.Reader) error {
			if relPath == "" {
				return nil
			}
			absPath := projResolve(bp, relPath)
			bs, readErr := os.ReadFile(absPath)
			if readErr != nil {
				return fmt.Errorf("cmd.ParseOptions: %w", readErr)
			}
			*dst = bytes.NewReader(bs)
			if *inplaceFlag {
				inplaceTargets[key] = absPath
			}
			return nil
		}
		if err := loadAllTable("ap", raw.AtomicProcessTablePath, &ae.AP); err != nil {
			return nil, err
		}
		if err := loadAllTable("ad", raw.AtomicDeliverableTablePath, &ae.AD); err != nil {
			return nil, err
		}
		// CD: only register the inplace target; do not load for refresh because
		// composite_deliverable_table in project.json points to the OUTPUT format
		// (cd.tsv), which lacks the Deliverables column and cannot be re-parsed.
		// Fresh generation from PFD is used instead (ae.CD stays nil).
		if *inplaceFlag && raw.CompositeDeliverableTablePath != "" {
			inplaceTargets["cd"] = projResolve(bp, raw.CompositeDeliverableTablePath)
		}
		if err := loadAllTable("r", raw.ResourceTablePath, &ae.R); err != nil {
			return nil, err
		}
		// M and G are always derived fresh from AP; loading them from project.json is unnecessary.
		if *inplaceFlag && raw.MilestoneTablePath != "" {
			inplaceTargets["m"] = projResolve(bp, raw.MilestoneTablePath)
		}
		if *inplaceFlag && raw.GroupTablePath != "" {
			inplaceTargets["g"] = projResolve(bp, raw.GroupTablePath)
		}
		allExisting = ae
	}

	var existingTableReader io.Reader
	hasExistingTable := existingPath != ""
	if hasExistingTable && tableCategory != TableCategoryAll {
		readPath := existingPath
		if isProjectJSON {
			readPath = projExistingTablePath
		}
		bs, readErr := os.ReadFile(readPath)
		if readErr != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", readErr)
		}
		existingTableReader = bytes.NewReader(bs)
	}

	if !hasExistingTable && *inplaceFlag && tableCategory != TableCategoryAll {
		return nil, fmt.Errorf("cmd.ParseOptions: inplace flag is only valid when existing table is specified")
	}

	var writer io.Writer
	if *inplaceFlag && tableCategory != TableCategoryAll {
		writePath := existingPath
		if isProjectJSON && projExistingTablePath != "" {
			writePath = projExistingTablePath
		}
		writer, err = os.OpenFile(writePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	} else {
		writer = inout.Stdout
	}

	return &Options{
		CommonOptions:                   commonOptions,
		HasPFD:                          hasPFD,
		PFDPath:                         pfdPath,
		PFDReader:                       pfdReader,
		HasCompositeDeliverableTable:    hasCompositeDeliverableTable,
		CompositeDeliverableTableReader: compositeDeliverableTableReader,
		AtomicProcessTableReader:        apTableReader,
		ExistingTableReader:             existingTableReader,
		HasExistingTable:                hasExistingTable,
		TableCategory:                   tableCategory,
		PFDTableType:                    pfdTableType,
		FSMTableType:                    fsmTableType,
		Mode:                            mode,
		OutputFormat:                    outputFormat,
		InputFormat:                     inputFormat,
		IsInplace:                       *inplaceFlag,
		Writer:                          writer,
		OutDir:                          *outDirFlag,
		ProjectFile:                     projectFile,
		AllExisting:                     allExisting,
		InplaceTargets:                  inplaceTargets,
	}, nil
}

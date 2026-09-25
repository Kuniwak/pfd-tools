package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/pfd/pfdticket"
	"github.com/Kuniwak/pfd-tools/tools"
)

const (
	OutputFormatJSON = "json"
	OutputFormatTSV  = "tsv"
)

type Options struct {
	CommonOptions          *tools.CommonOptions
	PFD                    *pfd.PFD
	AtomicDeliverableTable *pfd.AtomicDeliverableTable
	OutputFormat           string
	Config                 pfdticket.Config
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("pfdticket", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: pfdticket [options]", ShortHelp)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var pfdShortPath, pfdLongPath string
	tools.DeclarePFDOptions(flags, &pfdShortPath, &pfdLongPath)

	var adShortPath, adLongPath string
	tools.DeclareAtomicDeliverableTableOptions(flags, &adShortPath, &adLongPath)

	var cdShortPath, cdLongPath string
	tools.DeclareCompositeDeliverableTableOptions(flags, &cdShortPath, &cdLongPath)

	var configShortPath, configLongPath string
	tools.DeclareConfigOptions(flags, &configShortPath, &configLongPath)

	var outputFormat string
	flags.StringVar(&outputFormat, "out-format", OutputFormatJSON, "output format (available: json, tsv)")

	var formatColumn, reviewCriteriaColumn, reviewerColumn, urlColumn string
	flags.StringVar(&formatColumn, "format-column", "", "ad.tsv のフォーマット列見出し (default: -locale に応じた既定名)")
	flags.StringVar(&reviewCriteriaColumn, "review-criteria-column", "", "ad.tsv の品質基準（レビュー基準）列見出し (default: -locale に応じた既定名)")
	flags.StringVar(&reviewerColumn, "reviewer-column", "", "ad.tsv のレビューア列見出し (default: -locale に応じた既定名)")
	flags.StringVar(&urlColumn, "url-column", "", "ad.tsv の URL 列見出し (default: -locale に応じた既定名)")

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

	if outputFormat != OutputFormatJSON && outputFormat != OutputFormatTSV {
		return nil, fmt.Errorf("cmd.ParseOptions: invalid output format: %q", outputFormat)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	cliOptions := tools.FSMRawOptions{
		ShortPFDPath:                       pfdShortPath,
		PFDPath:                            pfdLongPath,
		ShortAtomicDeliverableTablePath:    adShortPath,
		AtomicDeliverableTablePath:         adLongPath,
		ShortCompositeDeliverableTablePath: cdShortPath,
		CompositeDeliverableTablePath:      cdLongPath,
	}

	mergedOptions, basePath, err := tools.ReadFSMRawOptions(&configShortPath, &configLongPath, cliOptions, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var atomicDeliverableTable *pfd.AtomicDeliverableTable
	if mergedOptions.AtomicDeliverableTablePath != "" || mergedOptions.ShortAtomicDeliverableTablePath != "" {
		adReader, _, err := tools.ValidateAtomicDeliverableTableOptions(&mergedOptions.ShortAtomicDeliverableTablePath, &mergedOptions.AtomicDeliverableTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		atomicDeliverableTable, err = pfdtsv.ParseAtomicDeliverableTable(adReader)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	var compositeDeliverableTableReader io.Reader
	if mergedOptions.CompositeDeliverableTablePath != "" || mergedOptions.ShortCompositeDeliverableTablePath != "" {
		compositeDeliverableTableReader, _, err = tools.ValidateCompositeDeliverableTableOptions(&mergedOptions.ShortCompositeDeliverableTablePath, &mergedOptions.CompositeDeliverableTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	compositeDeliverableTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(compositeDeliverableTableReader)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var p *pfd.PFD
	if mergedOptions.PFDPath != "" || mergedOptions.ShortPFDPath != "" {
		pfdReader, _, err := tools.ValidatePFDOptions(&mergedOptions.ShortPFDPath, &mergedOptions.PFDPath, basePath)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
		p, err = pfdfmt.Parse("", pfdReader, &pfdfmt.ParseOptions{CompositeDeliverableTable: compositeDeliverableTable}, commonOptions.Logger)
		if err != nil {
			return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
		}
	}

	config := pfdticket.DefaultConfig(commonOptions.Locale)
	if formatColumn != "" {
		config.FormatColumn = formatColumn
	}
	if reviewCriteriaColumn != "" {
		config.ReviewCriteriaColumn = reviewCriteriaColumn
	}
	if reviewerColumn != "" {
		config.ReviewerColumn = reviewerColumn
	}
	if urlColumn != "" {
		config.URLColumn = urlColumn
	}

	return &Options{
		CommonOptions:          commonOptions,
		PFD:                    p,
		AtomicDeliverableTable: atomicDeliverableTable,
		OutputFormat:           outputFormat,
		Config:                 config,
	}, nil
}

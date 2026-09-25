package cmd

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"syscall"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/perf"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/version"
)

const ShortHelp = "PFD からクリティカルパスを検出します。"

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

	if options.CommonOptions.ShortHelp {
		fmt.Fprintln(inout.Stdout, ShortHelp)
		return nil
	}
	if options.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	prof, err := perf.StartWithSignalHandler(options.CPUProfilePath, options.MemProfilePath, func() { os.Exit(130) }, os.Interrupt, syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	defer func() { _ = prof.Stop() }()

	logger := slog.New(slograw.NewHandler(inout.Stderr, options.CommonOptions.LogLevel))

	fsmEnvSeed, err := tools.ParseFSMEnvSeed(options.FSMOptions, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	if err := tools.ValidateFSMEnvSeed(fsmEnvSeed, logger, options.CommonOptions.Locale); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	env, err := tools.FSMPrepare(fsmEnvSeed, options.CommonOptions.Locale, logger)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	criticalPathInfoFunc := fsm.NewCriticalPathInfoFunc(options.SearchWithPrefixFunc)

	criticalPathInfo, err := criticalPathInfoFunc(env)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	w := csv.NewWriter(inout.Stdout)
	w.Comma = '\t'
	defer w.Flush()
	w.Write(Headers(options.CommonOptions.Locale))

	for _, ap := range env.PFD.AtomicProcesses.Iter() {
		info, ok := criticalPathInfo[ap]
		if !ok {
			panic(fmt.Sprintf("cmd.MainCommandByOptions: critical path info not found for atomic process %s", ap))
		}

		maximumElasticityString := strconv.FormatFloat(float64(info.MaximumElasticity), 'f', 2, 64)

		var minimumElasticityString string
		if info.HasMinimumElasticity {
			minimumElasticityString = strconv.FormatFloat(float64(info.MinimumElasticity), 'f', 2, 64)
		} else {
			minimumElasticityString = "-"
		}

		w.Write([]string{string(ap), maximumElasticityString, minimumElasticityString})
	}

	return nil
}

const (
	MaximumElasticityColumnHeaderJa = "最大弾性値（全余裕）"
	MaximumElasticityColumnHeaderEn = "Maximum Elasticity (Total Float)"

	MinimumElasticityColumnHeaderJa = "最小弾性値"
	MinimumElasticityColumnHeaderEn = "Minimum Elasticity"
)

var (
	MaximumElasticityColumn = pfdtable.ColumnAliases{Ja: MaximumElasticityColumnHeaderJa, En: MaximumElasticityColumnHeaderEn}
	MinimumElasticityColumn = pfdtable.ColumnAliases{Ja: MinimumElasticityColumnHeaderJa, En: MinimumElasticityColumnHeaderEn}
)

func Columns() []pfdtable.ColumnAliases {
	return []pfdtable.ColumnAliases{pfdtable.IDColumn, MaximumElasticityColumn, MinimumElasticityColumn}
}

func Headers(l locale.Locale) []string {
	return pfdtable.CanonicalNames(Columns(), l)
}

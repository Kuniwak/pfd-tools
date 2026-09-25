package tools

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kuniwak/pfd-tools/bizday"
	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/emphasis"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmreporter"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slograw"
)

var (
	PFDShortFlag                       = "p"
	PFDLongFlag                        = "pfd"
	AtomicProcessTableShortFlag        = "ap"
	AtomicProcessTableLongFlag         = "atomic-process"
	AtomicDeliverableTableShortFlag    = "ad"
	AtomicDeliverableTableLongFlag     = "atomic-deliverable"
	CompositeProcessTableShortFlag     = "cp"
	CompositeProcessTableLongFlag      = "composite-process"
	CompositeDeliverableTableShortFlag = "cd"
	CompositeDeliverableTableLongFlag  = "composite-deliverable"
	ResourceTableShortFlag             = "r"
	ResourceTableLongFlag              = "resource"
	MilestoneTableShortFlag            = "m"
	MilestoneTableLongFlag             = "milestone"
	GroupTableShortFlag                = "g"
	GroupTableLongFlag                 = "group"
	ConfigShortFlag                    = "f"
	ConfigLongFlag                     = "config"
	ResourceModeFlag                   = "res"
	FeedbackModeFlag                   = "fb"
	OutFormatFlag                      = "out-format"
	EmphasisTSVFlag                    = "em-tsv"
	OutDirFlag                         = "out-dir"
	InFormatFlag                       = "in-format"
)

type CommonOptions struct {
	Help      bool          `json:"help"`
	ShortHelp bool          `json:"short_help"`
	Version   bool          `json:"version"`
	LogLevel  slog.Level    `json:"log_level"`
	Logger    *slog.Logger  `json:"logger"`
	Locale    locale.Locale `json:"locale"`
}

type CommonRawOptions struct {
	Help         bool
	ShortHelp    bool
	ShortVersion bool
	Version      bool
	Silent       bool
	Debug        bool
	Locale       string
}

func DeclareCommonOptions(flags *flag.FlagSet, options *CommonRawOptions) {
	flags.BoolVar(&options.ShortVersion, "v", false, "show version")
	flags.BoolVar(&options.Version, "version", false, "show version")
	flags.BoolVar(&options.Silent, "silent", false, "silent mode")
	flags.BoolVar(&options.Debug, "debug", false, "debug mode")
	flags.StringVar(&options.Locale, "locale", "ja", "locale of the fsmreporter")
	flags.BoolVar(&options.ShortHelp, "short-help", false, "ツールの短い説明を表示して終了する")
}

func ParsePageList(s string) []string {
	var pages []string
	for _, part := range strings.Split(s, ",") {
		if name := strings.TrimSpace(part); name != "" {
			pages = append(pages, name)
		}
	}
	return pages
}

func PrintUsageHeader(flags *flag.FlagSet, usageLine, shortHelp string) {
	w := flags.Output()
	fmt.Fprintln(w, usageLine)
	fmt.Fprintln(w, "\n"+shortHelp)
	fmt.Fprintln(w, "\nOptions")
	flags.PrintDefaults()
}

func ValidateCommonOptions(options *CommonRawOptions) (*CommonOptions, error) {
	if options.ShortHelp {
		return &CommonOptions{ShortHelp: true}, nil
	}
	if options.ShortVersion || options.Version {
		return &CommonOptions{Version: true}, nil
	}

	var logLevel slog.Level
	if options.Debug {
		logLevel = slog.LevelDebug
	} else if options.Silent {
		logLevel = slog.LevelError
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slograw.NewHandler(os.Stderr, logLevel))

	var l locale.Locale
	if options.Locale != "" {
		var err error
		if l, err = locale.Parse(options.Locale); err != nil {
			return nil, fmt.Errorf("cmd.ValidateCommonOptions: %w", err)
		}
	}

	return &CommonOptions{
		LogLevel: logLevel,
		Logger:   logger,
		Locale:   l,
	}, nil
}

type PFDOptions struct {
	PFDReader                       io.Reader
	CompositeDeliverableTableReader io.Reader
}

func DeclarePFDOptions(flags *flag.FlagSet, pfdShortPath *string, pfdLongPath *string) {
	DeclarePFDOptionsWithFlagNames(PFDShortFlag, PFDLongFlag, flags, pfdShortPath, pfdLongPath)
}

func DeclarePFDOptionsWithFlagNames(shortName string, longName string, flags *flag.FlagSet, pfdShortFlag *string, pfdLongFlag *string) {
	flags.StringVar(pfdShortFlag, shortName, "", fmt.Sprintf("path to the PFD (same as -%s)", longName))
	flags.StringVar(pfdLongFlag, longName, "", fmt.Sprintf("path to the PFD (same as -%s)", shortName))
}

func ResolvePath(basePath, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(basePath, path)
}

func ValidatePFDOptions(pfdShortPath *string, pfdLongPath *string, basePath string) (io.Reader, string, error) {
	pfdRelPath := *pfdLongPath
	if *pfdShortPath != "" {
		pfdRelPath = *pfdShortPath
	}

	pfdPath := ResolvePath(basePath, pfdRelPath)

	r, err := os.OpenFile(pfdPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidatePFDOptions: %w", err)
	}
	return r, pfdPath, nil
}

type FSMOptions struct {
	PFDReader                            io.Reader       `json:"-"`
	AtomicProcessTableReader             io.Reader       `json:"-"`
	AtomicDeliverableTableReader         io.Reader       `json:"-"`
	CompositeProcessTableReader          io.Reader       `json:"-"`
	CompositeDeliverableTableReader      io.Reader       `json:"-"`
	ResourceTableReader                  io.Reader       `json:"-"`
	MilestoneTableReader                 io.Reader       `json:"-"`
	GroupTableReader                     io.Reader       `json:"-"`
	MaximalAvailableAllocationsThreshold int             `json:"maximal_available_allocations_threshold"`
	Model                                execmodel.Model `json:"-"`
}

type FSMRawOptions struct {
	PFDPath                              string `json:"pfd,omitempty"`
	ShortPFDPath                         string `json:"-"`
	AtomicProcessTablePath               string `json:"atomic_process_table,omitempty"`
	ShortAtomicProcessTablePath          string `json:"-"`
	AtomicDeliverableTablePath           string `json:"atomic_deliverable_table,omitempty"`
	ShortAtomicDeliverableTablePath      string `json:"-"`
	CompositeProcessTablePath            string `json:"composite_process_table,omitempty"`
	ShortCompositeProcessTablePath       string `json:"-"`
	CompositeDeliverableTablePath        string `json:"composite_deliverable_table,omitempty"`
	ShortCompositeDeliverableTablePath   string `json:"-"`
	ResourceTablePath                    string `json:"resource_table,omitempty"`
	ShortResourceTablePath               string `json:"-"`
	MilestoneTablePath                   string `json:"milestone_table,omitempty"`
	ShortMilestoneTablePath              string `json:"-"`
	GroupTablePath                       string `json:"group_table,omitempty"`
	ShortGroupTablePath                  string `json:"-"`
	MaximalAvailableAllocationsThreshold int    `json:"maximal_available_allocations_threshold,omitempty"`
	ResourceMode                         string `json:"resource_mode,omitempty"`
	FeedbackMode                         string `json:"feedback_mode,omitempty"`
}

func ResolveModel(options FSMRawOptions, resourceModeFlag string, feedbackModeFlag string) (execmodel.Model, error) {
	resourceMode, feedbackMode := options.ResourceMode, options.FeedbackMode
	cli.OverrideIfSet(&resourceMode, resourceModeFlag)
	cli.OverrideIfSet(&feedbackMode, feedbackModeFlag)
	model, err := execmodel.ParseModel(resourceMode, feedbackMode)
	if err != nil {
		return execmodel.Model{}, fmt.Errorf("tools.ResolveModel: %w", err)
	}
	return model, nil
}

type ProjectLayout struct {
	PFDPath                   string
	AtomicProcessTable        string
	AtomicDeliverableTable    string
	CompositeDeliverableTable string
	ResourceTable             string
	MilestoneTable            string
	GroupTable                string
}

func DefaultProjectLayout(pfdPath string) ProjectLayout {
	return ProjectLayout{
		PFDPath:                   pfdPath,
		AtomicProcessTable:        "ap.tsv",
		AtomicDeliverableTable:    "ad.tsv",
		CompositeDeliverableTable: "cd.tsv",
		ResourceTable:             "r.tsv",
		MilestoneTable:            "m.tsv",
		GroupTable:                "g.tsv",
	}
}

func (l ProjectLayout) FSMRawOptions(model execmodel.Model) FSMRawOptions {
	options := FSMRawOptions{
		PFDPath:                       l.PFDPath,
		AtomicProcessTablePath:        l.AtomicProcessTable,
		AtomicDeliverableTablePath:    l.AtomicDeliverableTable,
		CompositeDeliverableTablePath: l.CompositeDeliverableTable,
		MilestoneTablePath:            l.MilestoneTable,
		GroupTablePath:                l.GroupTable,
		ResourceMode:                  model.Resource.String(),
		FeedbackMode:                  model.Feedback.String(),
	}
	if model.Resource == execmodel.ResourceModeFinite {
		options.ResourceTablePath = l.ResourceTable
	}
	return options
}

func WriteProjectJSON(w io.Writer, options FSMRawOptions) error {
	bs, err := json.MarshalIndent(options, "", "\t")
	if err != nil {
		return fmt.Errorf("tools.WriteProjectJSON: %w", err)
	}
	if _, err := w.Write(append(bs, '\n')); err != nil {
		return fmt.Errorf("tools.WriteProjectJSON: %w", err)
	}
	return nil
}

func DeclareAtomicProcessTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, AtomicProcessTableShortFlag, "", "path to the atomic process fsmtable")
	flags.StringVar(path, AtomicProcessTableLongFlag, "", "path to the atomic process fsmtable")
}

func ValidateAtomicProcessTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	atomicProcessTableRelPath := *path
	if *shortPath != "" {
		atomicProcessTableRelPath = *shortPath
	}

	atomicProcessTablePath := ResolvePath(basePath, atomicProcessTableRelPath)

	r, err := os.OpenFile(atomicProcessTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateAtomicProcessTableOptions: %w", err)
	}
	return r, atomicProcessTablePath, nil
}

func DeclareAtomicDeliverableTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, AtomicDeliverableTableShortFlag, "", "path to the atomic deliverable fsmtable")
	flags.StringVar(path, AtomicDeliverableTableLongFlag, "", "path to the atomic deliverable fsmtable")
}

func ValidateAtomicDeliverableTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	atomicDeliverableTableRelPath := *path
	if *shortPath != "" {
		atomicDeliverableTableRelPath = *shortPath
	}

	atomicDeliverableTablePath := ResolvePath(basePath, atomicDeliverableTableRelPath)

	r, err := os.OpenFile(atomicDeliverableTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateAtomicDeliverableTableOptions: %w", err)
	}
	return r, atomicDeliverableTablePath, nil
}

func DeclareResourceTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, ResourceTableShortFlag, "", "path to the resource fsmtable")
	flags.StringVar(path, ResourceTableLongFlag, "", "path to the resource fsmtable")
}

func ValidateResourceTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	resourceTableRelPath := *path
	if *shortPath != "" {
		resourceTableRelPath = *shortPath
	}

	resourceTablePath := ResolvePath(basePath, resourceTableRelPath)

	r, err := os.OpenFile(resourceTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateResourceTableOptions: %w", err)
	}
	return r, resourceTablePath, nil
}

func DeclareCompositeProcessTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, CompositeProcessTableShortFlag, "", "path to the composite process fsmtable")
	flags.StringVar(path, CompositeProcessTableLongFlag, "", "path to the composite process fsmtable")
}

func ValidateCompositeProcessTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	compositeProcessTableRelPath := *path
	if *shortPath != "" {
		compositeProcessTableRelPath = *shortPath
	}

	compositeProcessTablePath := ResolvePath(basePath, compositeProcessTableRelPath)

	r, err := os.OpenFile(compositeProcessTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateCompositeProcessTableOptions: %w", err)
	}
	return r, compositeProcessTablePath, nil
}

func DeclareCompositeDeliverableTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	DeclareCompositeDeliverableTableOptionsWithFlagNames(CompositeDeliverableTableShortFlag, CompositeDeliverableTableLongFlag, flags, shortPath, path)
}

func DeclareCompositeDeliverableTableOptionsWithFlagNames(shortName string, longName string, flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, shortName, "", fmt.Sprintf("path to the composite deliverable fsmtable (same as -%s)", longName))
	flags.StringVar(path, longName, "", fmt.Sprintf("path to the composite deliverable fsmtable (same as -%s)", shortName))
}

func ValidateCompositeDeliverableTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	compositeDeliverableTableRelPath := *path
	if *shortPath != "" {
		compositeDeliverableTableRelPath = *shortPath
	}

	compositeDeliverableTablePath := ResolvePath(basePath, compositeDeliverableTableRelPath)

	r, err := os.OpenFile(compositeDeliverableTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateCompositeDeliverableTableOptions: %w", err)
	}
	return r, compositeDeliverableTablePath, nil
}

func DeclareMilestoneTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, MilestoneTableShortFlag, "", "path to the milestone table")
	flags.StringVar(path, MilestoneTableLongFlag, "", "path to the milestone table")
}

func ValidateMilestoneTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	milestoneTableRelPath := *path
	if *shortPath != "" {
		milestoneTableRelPath = *shortPath
	}

	milestoneTablePath := ResolvePath(basePath, milestoneTableRelPath)

	r, err := os.OpenFile(milestoneTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateMilestoneTableOptions: %w", err)
	}
	return r, milestoneTablePath, nil
}

func DeclareGroupTableOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, GroupTableShortFlag, "", "path to the group table")
	flags.StringVar(path, GroupTableLongFlag, "", "path to the group table")
}

func ValidateGroupTableOptions(shortPath *string, path *string, basePath string) (io.Reader, string, error) {
	groupTableRelPath := *path
	if *shortPath != "" {
		groupTableRelPath = *shortPath
	}
	groupTablePath := ResolvePath(basePath, groupTableRelPath)

	r, err := os.OpenFile(groupTablePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidateGroupTableOptions: %w", err)
	}
	return r, groupTablePath, nil
}

func DeclareConfigOptions(flags *flag.FlagSet, shortPath *string, path *string) {
	flags.StringVar(shortPath, ConfigShortFlag, "", "path to the run config file")
	flags.StringVar(path, ConfigLongFlag, "", "path to the run config file")
}

func DeclareExecModelOptions(flags *flag.FlagSet, resourceMode *string, feedbackMode *string) {
	flags.StringVar(resourceMode, ResourceModeFlag, "", fmt.Sprintf("how the exec model handles resources (available: %s, %s; default: %s)", execmodel.ResourceModeFinite, execmodel.ResourceModeInfinite, execmodel.DefaultModel().Resource))
	flags.StringVar(feedbackMode, FeedbackModeFlag, "", fmt.Sprintf("how the exec model handles feedback edges (available: %s, %s; default: %s)", execmodel.FeedbackModeEnabled, execmodel.FeedbackModeDisabled, execmodel.DefaultModel().Feedback))
}

func DeclareFSMOptions(flags *flag.FlagSet, options *FSMRawOptions, configLongPath *string, configShortPath *string) {
	DeclareExecModelOptions(flags, &options.ResourceMode, &options.FeedbackMode)
	flags.StringVar(&options.ShortPFDPath, PFDShortFlag, "", "path to the PFD")
	flags.StringVar(&options.PFDPath, PFDLongFlag, "", "path to the PFD")
	flags.IntVar(&options.MaximalAvailableAllocationsThreshold, "maximal-available-allocations-threshold", 10, "use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive")
	DeclareAtomicProcessTableOptions(flags, &options.ShortAtomicProcessTablePath, &options.AtomicProcessTablePath)
	DeclareAtomicDeliverableTableOptions(flags, &options.ShortAtomicDeliverableTablePath, &options.AtomicDeliverableTablePath)
	DeclareCompositeProcessTableOptions(flags, &options.ShortCompositeProcessTablePath, &options.CompositeProcessTablePath)
	DeclareCompositeDeliverableTableOptions(flags, &options.ShortCompositeDeliverableTablePath, &options.CompositeDeliverableTablePath)
	DeclareResourceTableOptions(flags, &options.ShortResourceTablePath, &options.ResourceTablePath)
	DeclareMilestoneTableOptions(flags, &options.ShortMilestoneTablePath, &options.MilestoneTablePath)
	DeclareGroupTableOptions(flags, &options.ShortGroupTablePath, &options.GroupTablePath)
	DeclareConfigOptions(flags, configShortPath, configLongPath)
}

func (options FSMRawOptions) Absolutize(basePath string) FSMRawOptions {
	for _, path := range []*string{
		&options.PFDPath, &options.ShortPFDPath,
		&options.AtomicProcessTablePath, &options.ShortAtomicProcessTablePath,
		&options.AtomicDeliverableTablePath, &options.ShortAtomicDeliverableTablePath,
		&options.CompositeProcessTablePath, &options.ShortCompositeProcessTablePath,
		&options.CompositeDeliverableTablePath, &options.ShortCompositeDeliverableTablePath,
		&options.ResourceTablePath, &options.ShortResourceTablePath,
		&options.MilestoneTablePath, &options.ShortMilestoneTablePath,
		&options.GroupTablePath, &options.ShortGroupTablePath,
	} {
		if *path == "" {
			continue
		}
		*path = ResolvePath(basePath, *path)
	}
	return options
}

type ProjectConfig struct {
	FSMRawOptions
	PlanOutputFormatRawOptions

	PlanPath string `json:"plan,omitempty"`

	ExplicitBusinessTimeFlags []string `json:"-"`

	ExplicitEmphasisTSVFlag bool `json:"-"`
}

var businessTimeSettingNames = []struct {
	Flag string
	Key  string
}{
	{"start", "start_day"},
	{"start-time", "start_time"},
	{"duration", "duration"},
	{"weekdays", "weekdays"},
	{"not-biz-days", "not_biz_days"},
}

func configPath(configShortPath *string, configLongPath *string) string {
	if *configShortPath != "" {
		return *configShortPath
	}
	return *configLongPath
}

func NoFlags() *flag.FlagSet {
	return flag.NewFlagSet("", flag.ContinueOnError)
}

func explicitFlagNames(flags *flag.FlagSet) map[string]struct{} {
	res := make(map[string]struct{})
	flags.Visit(func(f *flag.Flag) {
		res[f.Name] = struct{}{}
	})
	return res
}

func ReadProjectConfig(
	configShortPath *string,
	configLongPath *string,
	cliFSMOptions FSMRawOptions,
	cliPlanOutputOptions PlanOutputFormatRawOptions,
	cwd string,
	flags *flag.FlagSet,
) (ProjectConfig, string, error) {
	explicit := explicitFlagNames(flags)

	path := configPath(configShortPath, configLongPath)
	if path == "" {
		return ProjectConfig{
			FSMRawOptions:              cliFSMOptions,
			PlanOutputFormatRawOptions: cliPlanOutputOptions,
			ExplicitBusinessTimeFlags:  explicitBusinessTimeFlags(explicit),
			ExplicitEmphasisTSVFlag:    isExplicitFlag(explicit, EmphasisTSVFlag),
		}, cwd, nil
	}

	basePath := filepath.Dir(path)

	r, err := os.Open(path)
	if err != nil {
		return ProjectConfig{}, "", fmt.Errorf("tools.ReadProjectConfig: %w", err)
	}
	defer r.Close()

	var result ProjectConfig
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return ProjectConfig{}, "", fmt.Errorf("tools.ReadProjectConfig: %w", err)
	}

	cliFSMOptions = cliFSMOptions.Absolutize(cwd)

	cli.OverrideIfSet(&result.PFDPath, cliFSMOptions.ShortPFDPath, cliFSMOptions.PFDPath)
	cli.OverrideIfSet(&result.AtomicProcessTablePath, cliFSMOptions.ShortAtomicProcessTablePath, cliFSMOptions.AtomicProcessTablePath)
	cli.OverrideIfSet(&result.AtomicDeliverableTablePath, cliFSMOptions.ShortAtomicDeliverableTablePath, cliFSMOptions.AtomicDeliverableTablePath)
	cli.OverrideIfSet(&result.CompositeProcessTablePath, cliFSMOptions.ShortCompositeProcessTablePath, cliFSMOptions.CompositeProcessTablePath)
	cli.OverrideIfSet(&result.CompositeDeliverableTablePath, cliFSMOptions.ShortCompositeDeliverableTablePath, cliFSMOptions.CompositeDeliverableTablePath)
	cli.OverrideIfSet(&result.ResourceTablePath, cliFSMOptions.ShortResourceTablePath, cliFSMOptions.ResourceTablePath)
	cli.OverrideIfSet(&result.MilestoneTablePath, cliFSMOptions.ShortMilestoneTablePath, cliFSMOptions.MilestoneTablePath)
	cli.OverrideIfSet(&result.GroupTablePath, cliFSMOptions.ShortGroupTablePath, cliFSMOptions.GroupTablePath)
	cli.OverrideIfSet(&result.ResourceMode, cliFSMOptions.ResourceMode)
	cli.OverrideIfSet(&result.FeedbackMode, cliFSMOptions.FeedbackMode)

	if cliFSMOptions.MaximalAvailableAllocationsThreshold != 0 {
		result.MaximalAvailableAllocationsThreshold = cliFSMOptions.MaximalAvailableAllocationsThreshold
	}

	result.ExplicitBusinessTimeFlags = explicitBusinessTimeFlags(explicit)
	result.ExplicitEmphasisTSVFlag = isExplicitFlag(explicit, EmphasisTSVFlag)

	if result.AdditionalNotBusinessDays != "" {
		result.AdditionalNotBusinessDays = ResolvePath(basePath, result.AdditionalNotBusinessDays)
	}
	if cliPlanOutputOptions.AdditionalNotBusinessDays != "" {
		cliPlanOutputOptions.AdditionalNotBusinessDays = ResolvePath(cwd, cliPlanOutputOptions.AdditionalNotBusinessDays)
	}
	if result.PlanPath != "" {
		result.PlanPath = ResolvePath(basePath, result.PlanPath)
	}
	if result.EmphasisTSVPath != "" {
		result.EmphasisTSVPath = ResolvePath(basePath, result.EmphasisTSVPath)
	}
	if cliPlanOutputOptions.EmphasisTSVPath != "" {
		cliPlanOutputOptions.EmphasisTSVPath = ResolvePath(cwd, cliPlanOutputOptions.EmphasisTSVPath)
	}

	result.PlanOutputFormatRawOptions = mergePlanOutputFormatRawOptions(result.PlanOutputFormatRawOptions, cliPlanOutputOptions, explicit)

	return result, basePath, nil
}

func isExplicitFlag(explicitFlags map[string]struct{}, flagName string) bool {
	_, ok := explicitFlags[flagName]
	return ok
}

func explicitBusinessTimeFlags(explicitFlags map[string]struct{}) []string {
	var res []string
	for _, name := range businessTimeSettingNames {
		if _, ok := explicitFlags[name.Flag]; ok {
			res = append(res, "-"+name.Flag)
		}
	}
	return res
}

func mergePlanOutputFormatRawOptions(fileOptions PlanOutputFormatRawOptions, cliOptions PlanOutputFormatRawOptions, explicitFlags map[string]struct{}) PlanOutputFormatRawOptions {
	res := cliOptions

	isExplicit := func(flagName string) bool {
		return isExplicitFlag(explicitFlags, flagName)
	}

	if !isExplicit(OutFormatFlag) && fileOptions.OutputFormat != "" {
		res.OutputFormat = fileOptions.OutputFormat
	}
	if !isExplicit(EmphasisTSVFlag) && fileOptions.EmphasisTSVPath != "" {
		res.EmphasisTSVPath = fileOptions.EmphasisTSVPath
	}
	if !isExplicit("start") && fileOptions.StartDay != "" {
		res.StartDay = fileOptions.StartDay
	}
	if !isExplicit("start-time") && fileOptions.StartTime != "" {
		res.StartTime = fileOptions.StartTime
	}
	if !isExplicit("duration") && fileOptions.Duration != 0 {
		res.Duration = fileOptions.Duration
	}
	if !isExplicit("weekdays") && fileOptions.Weekdays != "" {
		res.Weekdays = fileOptions.Weekdays
	}
	if !isExplicit("not-biz-days") && fileOptions.AdditionalNotBusinessDays != "" {
		res.AdditionalNotBusinessDays = fileOptions.AdditionalNotBusinessDays
	}

	return res
}

func ReadFSMRawOptions(configShortPath *string, configLongPath *string, cliOptions FSMRawOptions, cwd string) (FSMRawOptions, string, error) {
	config, basePath, err := ReadProjectConfig(configShortPath, configLongPath, cliOptions, PlanOutputFormatRawOptions{}, cwd, NoFlags())
	if err != nil {
		return FSMRawOptions{}, "", fmt.Errorf("tools.ReadFSMRawOptions: %w", err)
	}
	return config.FSMRawOptions, basePath, nil
}

func RequireTablePath(flagName string, path string, shortPath string) error {
	if path == "" && shortPath == "" {
		return fmt.Errorf("-%s is required", flagName)
	}
	return nil
}

func ValidateAllFSMOptions(options *FSMRawOptions, basePath string) (*FSMOptions, error) {

	model, err := ResolveModel(*options, "", "")
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
	}

	for _, required := range []struct {
		Flag      string
		Path      string
		ShortPath string
	}{
		{PFDShortFlag, options.PFDPath, options.ShortPFDPath},
		{AtomicProcessTableShortFlag, options.AtomicProcessTablePath, options.ShortAtomicProcessTablePath},
		{AtomicDeliverableTableShortFlag, options.AtomicDeliverableTablePath, options.ShortAtomicDeliverableTablePath},
	} {
		if err := RequireTablePath(required.Flag, required.Path, required.ShortPath); err != nil {
			return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
		}
	}

	pfdReader, _, err := ValidatePFDOptions(&options.ShortPFDPath, &options.PFDPath, basePath)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
	}
	atomicProcessTableReader, _, err := ValidateAtomicProcessTableOptions(&options.ShortAtomicProcessTablePath, &options.AtomicProcessTablePath, basePath)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
	}

	var resourceTableReader io.Reader
	hasResourceTablePath := options.ResourceTablePath != "" || options.ShortResourceTablePath != ""
	if model.Resource == execmodel.ResourceModeFinite && !hasResourceTablePath {
		return nil, fmt.Errorf("tools.ValidateAllFSMOptions: resource table (-%s) is required for the %s resource mode", ResourceTableShortFlag, model.Resource)
	}
	if hasResourceTablePath {
		resourceTableReader, _, err = ValidateResourceTableOptions(&options.ShortResourceTablePath, &options.ResourceTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
		}
	}
	atomicDeliverableTableReader, _, err := ValidateAtomicDeliverableTableOptions(&options.ShortAtomicDeliverableTablePath, &options.AtomicDeliverableTablePath, basePath)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
	}

	var compositeDeliverableTableReader io.Reader
	if options.CompositeDeliverableTablePath != "" || options.ShortCompositeDeliverableTablePath != "" {
		compositeDeliverableTableReader, _, err = ValidateCompositeDeliverableTableOptions(&options.ShortCompositeDeliverableTablePath, &options.CompositeDeliverableTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
		}
	}

	var compositeProcessTableReader io.Reader
	if options.CompositeProcessTablePath != "" {
		compositeProcessTableReader, _, err = ValidateCompositeProcessTableOptions(&options.ShortCompositeProcessTablePath, &options.CompositeProcessTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
		}
	}

	var milestoneTableReader io.Reader
	if options.MilestoneTablePath != "" {
		milestoneTableReader, _, err = ValidateMilestoneTableOptions(&options.ShortMilestoneTablePath, &options.MilestoneTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
		}
	}

	var groupTableReader io.Reader
	if options.GroupTablePath != "" {
		groupTableReader, _, err = ValidateGroupTableOptions(&options.ShortGroupTablePath, &options.GroupTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateAllFSMOptions: %w", err)
		}
	}

	return &FSMOptions{
		PFDReader:                            pfdReader,
		AtomicProcessTableReader:             atomicProcessTableReader,
		AtomicDeliverableTableReader:         atomicDeliverableTableReader,
		CompositeProcessTableReader:          compositeProcessTableReader,
		CompositeDeliverableTableReader:      compositeDeliverableTableReader,
		ResourceTableReader:                  resourceTableReader,
		MilestoneTableReader:                 milestoneTableReader,
		GroupTableReader:                     groupTableReader,
		MaximalAvailableAllocationsThreshold: options.MaximalAvailableAllocationsThreshold,
		Model:                                model,
	}, nil
}

func ValidatePossibleFSMOptions(options *FSMRawOptions, basePath string) (*FSMOptions, error) {
	model, err := ResolveModel(*options, "", "")
	if err != nil {
		return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
	}

	var pfdReader io.Reader
	if options.PFDPath != "" || options.ShortPFDPath != "" {
		pfdReader, _, err = ValidatePFDOptions(&options.ShortPFDPath, &options.PFDPath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var atomicProcessTableReader io.Reader
	if options.AtomicProcessTablePath != "" || options.ShortAtomicProcessTablePath != "" {
		atomicProcessTableReader, _, err = ValidateAtomicProcessTableOptions(&options.ShortAtomicProcessTablePath, &options.AtomicProcessTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var resourceTableReader io.Reader
	if options.ResourceTablePath != "" || options.ShortResourceTablePath != "" {
		resourceTableReader, _, err = ValidateResourceTableOptions(&options.ShortResourceTablePath, &options.ResourceTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var atomicDeliverableTableReader io.Reader
	if options.AtomicDeliverableTablePath != "" || options.ShortAtomicDeliverableTablePath != "" {
		atomicDeliverableTableReader, _, err = ValidateAtomicDeliverableTableOptions(&options.ShortAtomicDeliverableTablePath, &options.AtomicDeliverableTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var compositeProcessTableReader io.Reader
	if options.CompositeProcessTablePath != "" || options.ShortCompositeProcessTablePath != "" {
		compositeProcessTableReader, _, err = ValidateCompositeProcessTableOptions(&options.ShortCompositeProcessTablePath, &options.CompositeProcessTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var compositeDeliverableTableReader io.Reader
	if options.CompositeDeliverableTablePath != "" || options.ShortCompositeDeliverableTablePath != "" {
		compositeDeliverableTableReader, _, err = ValidateCompositeDeliverableTableOptions(&options.ShortCompositeDeliverableTablePath, &options.CompositeDeliverableTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var milestoneTableReader io.Reader
	if options.MilestoneTablePath != "" || options.ShortMilestoneTablePath != "" {
		milestoneTableReader, _, err = ValidateMilestoneTableOptions(&options.ShortMilestoneTablePath, &options.MilestoneTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	var groupTableReader io.Reader
	if options.GroupTablePath != "" || options.ShortGroupTablePath != "" {
		groupTableReader, _, err = ValidateGroupTableOptions(&options.ShortGroupTablePath, &options.GroupTablePath, basePath)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidatePossibleFSMOptions: %w", err)
		}
	}

	return &FSMOptions{
		PFDReader:                            pfdReader,
		AtomicProcessTableReader:             atomicProcessTableReader,
		AtomicDeliverableTableReader:         atomicDeliverableTableReader,
		CompositeProcessTableReader:          compositeProcessTableReader,
		CompositeDeliverableTableReader:      compositeDeliverableTableReader,
		ResourceTableReader:                  resourceTableReader,
		MilestoneTableReader:                 milestoneTableReader,
		GroupTableReader:                     groupTableReader,
		MaximalAvailableAllocationsThreshold: options.MaximalAvailableAllocationsThreshold,
		Model:                                model,
	}, nil
}

type SearchQualityPreset string

const (
	SearchQualityPresetSmall   SearchQualityPreset = "s"
	SearchQualityPresetMedium  SearchQualityPreset = "m"
	SearchQualityPresetLarge   SearchQualityPreset = "l"
	SearchQualityPresetXLarge  SearchQualityPreset = "xl"
	SearchQualityPresetXXLarge SearchQualityPreset = "xxl"
	SearchQualityPresetCustom  SearchQualityPreset = "custom"
	SearchQualityPresetDefault SearchQualityPreset = "default"
)

func (q SearchQualityPreset) String() string {
	return strings.ToUpper(string(q))
}

func (q SearchQualityPreset) Quality(randomSeed int64) fsm.Quality {
	switch q {
	case SearchQualityPresetSmall:
		return fsm.Quality{NodeBudget: 10_000, TopKPerState: 128, Weight: 1.2, MaxResults: 3, RandomSeed: randomSeed, Restarts: 0}
	case SearchQualityPresetMedium:
		return fsm.Quality{NodeBudget: 30_000, TopKPerState: 64, Weight: 1.5, MaxResults: 3, RandomSeed: randomSeed, Restarts: 0}
	case SearchQualityPresetLarge:
		return fsm.Quality{NodeBudget: 80_000, TopKPerState: 32, Weight: 2.0, MaxResults: 3, RandomSeed: randomSeed, Restarts: 1}
	case SearchQualityPresetXLarge:
		return fsm.Quality{NodeBudget: 200_000, TopKPerState: 16, Weight: 2.5, MaxResults: 3, RandomSeed: randomSeed, Restarts: 2}
	case SearchQualityPresetXXLarge:
		return fsm.Quality{NodeBudget: 500_000, TopKPerState: 8, Weight: 3.0, MaxResults: 3, RandomSeed: randomSeed, Restarts: 3}
	case SearchQualityPresetDefault:
		return fsm.Quality{NodeBudget: 10_000, TopKPerState: 128, Weight: 2.0, MaxResults: 3, RandomSeed: randomSeed, Restarts: 0}
	default:
		panic(fmt.Sprintf("cmd.SearchQualityPreset.Quality: invalid quality preset: %q", q))
	}
}

type SearchRawOptions struct {
	Best    bool
	Better  bool
	Poor    bool
	Poorest bool

	QualityPreset string

	Quality fsm.Quality
}

func DeclareSearchOptions(flags *flag.FlagSet, options *SearchRawOptions, randomSeed int64) {
	flags.BoolVar(&options.Best, "best", false, "search best plan")
	flags.BoolVar(&options.Better, "better", false, "search better plan")
	flags.BoolVar(&options.Poor, "poor", false, "search plan by greedy algorithm (faster than best and better; ties broken pseudo-randomly by -random-seed)")
	flags.BoolVar(&options.Poorest, "poorest", false, "search plan by fast greedy allocation without enumerating allocations (fastest and deterministic; completes even where -poor does not)")
	flags.StringVar(&options.QualityPreset, "quality", "small", "quality preset (available: s, m, l, xl, xxl)")
	flags.Int64Var(&options.Quality.RandomSeed, "random-seed", randomSeed, "random seed")

	defaultQuality := SearchQualityPresetDefault.Quality(randomSeed)
	flags.IntVar(&options.Quality.NodeBudget, "node-budget", defaultQuality.NodeBudget, "upper bound of the number of nodes to expand >= 1")
	flags.IntVar(&options.Quality.TopKPerState, "top-k-per-state", defaultQuality.TopKPerState, "upper bound of the number of transitions to consider per state >= 1")
	flags.Float64Var(&options.Quality.Weight, "weight", defaultQuality.Weight, "weight >= 1.0 of Weighted A*. closer to 1.0 means closer to A*, greater than 1.0 means closer to greedy")
	flags.IntVar(&options.Quality.MaxResults, "max-results", defaultQuality.MaxResults, "upper bound of the number of results to return >= 1")
	flags.IntVar(&options.Quality.Restarts, "restarts", defaultQuality.Restarts, "number >= 0 of restarts for diversity")
}

func ValidateSearchOptions(searchRawOptions *SearchRawOptions) (fsm.SearchFunc, error) {
	if !searchRawOptions.Best && !searchRawOptions.Better && !searchRawOptions.Poor && !searchRawOptions.Poorest {
		return nil, fmt.Errorf("cmd.ValidateSearchOptions: either best or better or poor or poorest must be true")
	}

	if searchRawOptions.Poorest {
		return fsm.SearchPoorest(searchRawOptions.Quality.RandomSeed), nil
	}

	if searchRawOptions.Poor {
		return fsm.SearchFastest(searchRawOptions.Quality.RandomSeed), nil
	}

	if searchRawOptions.Best {
		return fsm.SearchBestPlans(), nil
	}

	if searchRawOptions.Better {
		var searchQuality fsm.Quality

		switch strings.ToLower(searchRawOptions.QualityPreset) {
		case "s", "small":
			searchQuality = SearchQualityPresetSmall.Quality(searchRawOptions.Quality.RandomSeed)
		case "m", "medium":
			searchQuality = SearchQualityPresetMedium.Quality(searchRawOptions.Quality.RandomSeed)
		case "l", "large":
			searchQuality = SearchQualityPresetLarge.Quality(searchRawOptions.Quality.RandomSeed)
		case "xl", "xlarge":
			searchQuality = SearchQualityPresetXLarge.Quality(searchRawOptions.Quality.RandomSeed)
		case "xxl", "xxlarge":
			searchQuality = SearchQualityPresetXXLarge.Quality(searchRawOptions.Quality.RandomSeed)
		case "d", "default":
			searchQuality = SearchQualityPresetDefault.Quality(searchRawOptions.Quality.RandomSeed)
		case "custom":
			if searchRawOptions.Quality.Weight < 1.0 {
				return nil, fmt.Errorf("cmd.ValidateSearchOptions: weight must be >= 1.0")
			}

			if searchRawOptions.Quality.NodeBudget < 1 {
				return nil, fmt.Errorf("cmd.ValidateSearchOptions: node-budget must be >= 1")
			}

			if searchRawOptions.Quality.TopKPerState < 0 {
				return nil, fmt.Errorf("cmd.ValidateSearchOptions: top-k-per-state must be >= 0")
			}

			if searchRawOptions.Quality.MaxResults < 1 {
				return nil, fmt.Errorf("cmd.ValidateSearchOptions: max-results must be >= 1")
			}

			if searchRawOptions.Quality.Restarts < 0 {
				return nil, fmt.Errorf("cmd.ValidateSearchOptions: restarts must be >= 0")
			}

			searchQuality = searchRawOptions.Quality
		default:
			return nil, fmt.Errorf("cmd.ValidateSearchOptions: invalid quality preset: %q", searchRawOptions.QualityPreset)
		}

		return fsm.SearchBetterPlans(searchQuality), nil
	}

	return nil, fmt.Errorf("cmd.ValidateSearchOptions: either best or better must be true")
}

func ValidateSearchWithPrefixOptions(searchRawOptions *SearchRawOptions) (fsm.SearchWithPrefixFunc, error) {
	if !searchRawOptions.Best && !searchRawOptions.Better && !searchRawOptions.Poor && !searchRawOptions.Poorest {
		return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: either best or better or poor or poorest must be true")
	}

	if searchRawOptions.Poorest {
		return fsm.SearchPoorestWithPrefix(searchRawOptions.Quality.RandomSeed), nil
	}

	if searchRawOptions.Poor {
		return fsm.SearchFastestWithPrefix(searchRawOptions.Quality.RandomSeed), nil
	}

	if searchRawOptions.Best {
		return fsm.SearchBestPlansWithPrefix(), nil
	}

	if searchRawOptions.Better {
		var searchQuality fsm.Quality

		switch strings.ToLower(searchRawOptions.QualityPreset) {
		case "s", "small":
			searchQuality = SearchQualityPresetSmall.Quality(searchRawOptions.Quality.RandomSeed)
		case "m", "medium":
			searchQuality = SearchQualityPresetMedium.Quality(searchRawOptions.Quality.RandomSeed)
		case "l", "large":
			searchQuality = SearchQualityPresetLarge.Quality(searchRawOptions.Quality.RandomSeed)
		case "xl", "xlarge":
			searchQuality = SearchQualityPresetXLarge.Quality(searchRawOptions.Quality.RandomSeed)
		case "xxl", "xxlarge":
			searchQuality = SearchQualityPresetXXLarge.Quality(searchRawOptions.Quality.RandomSeed)
		case "d", "default":
			searchQuality = SearchQualityPresetDefault.Quality(searchRawOptions.Quality.RandomSeed)
		case "custom":
			if searchRawOptions.Quality.Weight < 1.0 {
				return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: weight must be >= 1.0")
			}
			if searchRawOptions.Quality.NodeBudget < 1 {
				return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: node-budget must be >= 1")
			}
			if searchRawOptions.Quality.TopKPerState < 0 {
				return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: top-k-per-state must be >= 0")
			}
			if searchRawOptions.Quality.MaxResults < 1 {
				return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: max-results must be >= 1")
			}
			if searchRawOptions.Quality.Restarts < 0 {
				return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: restarts must be >= 0")
			}
			searchQuality = searchRawOptions.Quality
		default:
			return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: invalid quality preset: %q", searchRawOptions.QualityPreset)
		}

		return fsm.SearchBetterPlansWithPrefix(searchQuality), nil
	}

	return nil, fmt.Errorf("cmd.ValidateSearchWithPrefixOptions: either best or better must be true")
}

type BusinessTimeFuncRawOptions struct {
	StartDay                  string  `json:"start_day,omitempty"`
	StartTime                 string  `json:"start_time,omitempty"`
	Duration                  float64 `json:"duration,omitempty"`
	Weekdays                  string  `json:"weekdays,omitempty"`
	AdditionalNotBusinessDays string  `json:"not_biz_days,omitempty"`
}

func DeclareBusinessTimeFuncOptions(flags *flag.FlagSet, options *BusinessTimeFuncRawOptions) {
	flags.StringVar(&options.StartDay, "start", "", "start day")
	flags.StringVar(&options.StartTime, "start-time", "10:00", "start time")
	flags.Float64Var(&options.Duration, "duration", 9, "duration")
	flags.StringVar(&options.Weekdays, "weekdays", "mon,tue,wed,thu,fri", "comma separated weekdays (available: sun,mon,tue,wed,thu,fri,sat)")
	flags.StringVar(&options.AdditionalNotBusinessDays, "not-biz-days", "", "not business days except weekdays (comma separated dates. e.g. 2025-01-01,2025-01-02)")
}

type BusinessTimeFuncOptions struct {
	StartDay         bizday.Day
	BusinessTimeFunc bizday.BusinessTimeFunc
	Duration         time.Duration
}

func ValidateBusinessTimeFuncOptions(options *BusinessTimeFuncRawOptions) (*BusinessTimeFuncOptions, error) {
	var startDay bizday.Day
	if options.StartDay == "" {
		startDate := time.Now()
		startDay = bizday.NewDayByTime(startDate)
	} else {
		startDate, err := time.ParseInLocation("2006-01-02", options.StartDay, time.Local)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
		}
		startDay = bizday.NewDayByTime(startDate)
	}
	var startTime bizday.Time
	startTimeRaw, err := time.ParseInLocation("15:04", options.StartTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
	}
	startTime = bizday.NewTimeByTime(startTimeRaw)

	if options.Duration < 0 {
		return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: invalid duration: %.0f", options.Duration)
	}
	duration := time.Duration(options.Duration) * time.Hour

	_, err = startTime.Add(duration)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
	}

	var weekdays []time.Weekday
	if options.Weekdays != "" {
		weekdayTexts := strings.SplitSeq(options.Weekdays, ",")
		for weekdayText := range weekdayTexts {
			switch strings.ToLower(weekdayText) {
			case "sun", "sunday":
				weekdays = append(weekdays, time.Sunday)
			case "mon", "monday":
				weekdays = append(weekdays, time.Monday)
			case "tue", "tuesday":
				weekdays = append(weekdays, time.Tuesday)
			case "wed", "wednesday":
				weekdays = append(weekdays, time.Wednesday)
			case "thu", "thursday":
				weekdays = append(weekdays, time.Thursday)
			case "fri", "friday":
				weekdays = append(weekdays, time.Friday)
			case "sat", "saturday":
				weekdays = append(weekdays, time.Saturday)
			default:
				return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: invalid weekday: %q", weekdayText)
			}
		}
	}

	additionalNotBusinessDays := sets.NewWithCapacity[bizday.Day](0)
	if options.AdditionalNotBusinessDays != "" {
		f, err := os.Open(options.AdditionalNotBusinessDays)
		if err != nil {
			return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			additionalNotBusinessDaysText := strings.TrimSpace(scanner.Text())
			if additionalNotBusinessDaysText == "" {
				continue
			}

			additionalNotBusinessDay, err := time.ParseInLocation("2006-01-02", additionalNotBusinessDaysText, time.Local)
			if err != nil {
				return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
			}
			additionalNotBusinessDays.Add(bizday.Day.Compare, bizday.NewDayByTime(additionalNotBusinessDay))
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
		}
		f.Close()
	}

	businessHoursFunc, err := bizday.NewBusinessHoursFunc(startTime, duration)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateBusinessTimeFuncOptions: %w", err)
	}
	isBusinessDayFunc := bizday.NewIsBusinessDayFunc(weekdays, additionalNotBusinessDays)
	businessTimeFunc := bizday.NewBusinessTime(businessHoursFunc, isBusinessDayFunc)

	return &BusinessTimeFuncOptions{
		StartDay:         startDay,
		BusinessTimeFunc: businessTimeFunc,
		Duration:         duration,
	}, nil
}

type PlanOutputFormatRawOptions struct {
	OutputFormat    string `json:"output_format,omitempty"`
	EmphasisTSVPath string `json:"emphasis_tsv_path,omitempty"`

	BusinessTimeFuncRawOptions
}

type PlanOutputFormat string

const (
	PlanOutputFormatGoogleSpreadsheetTSV PlanOutputFormat = "google-spreadsheet-tsv"
	PlanOutputFormatPlanJSON             PlanOutputFormat = "plan-json"
	PlanOutputFormatTimelineJSON         PlanOutputFormat = "timeline-json"
	PlanOutputFormatMermaid              PlanOutputFormat = "mermaid"
	PlanOutputFormatPlantUML             PlanOutputFormat = "plantuml"
)

func DeclarePlanOutputFormatOptions(flags *flag.FlagSet, options *PlanOutputFormatRawOptions) {
	DeclareBusinessTimeFuncOptions(flags, &options.BusinessTimeFuncRawOptions)
	flags.StringVar(&options.OutputFormat, OutFormatFlag, "", "output format (available: google-spreadsheet-tsv, plan-json, timeline-json, mermaid, plantuml)")
	DeclareEmphasisTSVOptions(flags, &options.EmphasisTSVPath)
}

func DeclareEmphasisTSVOptions(flags *flag.FlagSet, path *string) {
	flags.StringVar(path, EmphasisTSVFlag, "", "path to the emphasis ID table (TSV with an ID column. e.g. the criticalpath output filtered by qhs)")
}

func ValidateEmphasisTSVOptions(path string) (*emphasis.Set, error) {
	if path == "" {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateEmphasisTSVOptions: %w", err)
	}
	defer f.Close()

	em, err := emphasis.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("tools.ValidateEmphasisTSVOptions: %w", err)
	}
	return em, nil
}

func ValidatePlanOutputFormat(options *PlanOutputFormatRawOptions, config ProjectConfig, logger *slog.Logger) (fsmreporter.PlanReporter, PlanOutputFormat, error) {
	explicitBusinessTimeFlags := config.ExplicitBusinessTimeFlags

	em, err := ValidateEmphasisTSVOptions(options.EmphasisTSVPath)
	if err != nil {
		return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: %w", err)
	}

	switch options.OutputFormat {
	case "", "google-spreadsheet-tsv":
		businessTimeFuncOptions, err := ValidateBusinessTimeFuncOptions(&options.BusinessTimeFuncRawOptions)
		if err != nil {
			return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: %w", err)
		}
		return fsmreporter.NewGoogleSpreadsheetTimelineTSVReporter(businessTimeFuncOptions.StartDay, businessTimeFuncOptions.BusinessTimeFunc, em, logger), PlanOutputFormatGoogleSpreadsheetTSV, nil

	case "plan-json":
		if len(explicitBusinessTimeFlags) > 0 {
			return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: business time flags cannot be used with plan-json output format: %v", explicitBusinessTimeFlags)
		}
		if config.ExplicitEmphasisTSVFlag {
			return nil, "", errEmphasisNotAvailable("plan-json")
		}
		return fsmreporter.NewPlanJSONReporter(), PlanOutputFormatPlanJSON, nil

	case "timeline-json":
		if len(explicitBusinessTimeFlags) > 0 {
			return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: business time flags cannot be used with timeline-json output format: %v", explicitBusinessTimeFlags)
		}
		if config.ExplicitEmphasisTSVFlag {
			return nil, "", errEmphasisNotAvailable("timeline-json")
		}
		return fsmreporter.NewTimelineJSONReporter(logger), PlanOutputFormatTimelineJSON, nil

	case "mermaid":
		businessTimeFuncOptions, err := ValidateBusinessTimeFuncOptions(&options.BusinessTimeFuncRawOptions)
		if err != nil {
			return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: %w", err)
		}
		return fsmreporter.NewMermaidGanttReporter(businessTimeFuncOptions.StartDay, businessTimeFuncOptions.BusinessTimeFunc, em, logger), PlanOutputFormatMermaid, nil

	case "plantuml":
		businessTimeFuncOptions, err := ValidateBusinessTimeFuncOptions(&options.BusinessTimeFuncRawOptions)
		if err != nil {
			return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: %w", err)
		}
		return fsmreporter.NewPlantUMLGanttReporter(businessTimeFuncOptions.StartDay, businessTimeFuncOptions.BusinessTimeFunc, em, logger), PlanOutputFormatPlantUML, nil

	default:
		return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: invalid output format: %q", options.OutputFormat)
	}
}

func errEmphasisNotAvailable(outputFormat string) error {
	return fmt.Errorf("tools.ValidatePlanOutputFormat: -%s is not available for the %s output format (available: google-spreadsheet-tsv, mermaid, plantuml)", EmphasisTSVFlag, outputFormat)
}

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
	"github.com/Kuniwak/pfd-tools/locale"
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
)

type CommonOptions struct {
	Help     bool          `json:"help"`
	Version  bool          `json:"version"`
	LogLevel slog.Level    `json:"log_level"`
	Logger   *slog.Logger  `json:"logger"`
	Locale   locale.Locale `json:"locale"`
}

type CommonRawOptions struct {
	Help         bool
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
	flags.StringVar(&options.Locale, "locale", "en", "locale of the project")
}

func ValidateCommonOptions(options *CommonRawOptions) (*CommonOptions, error) {
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
	} else {
		l = locale.LocaleEn
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

func ValidatePFDOptions(pfdShortPath *string, pfdLongPath *string, basePath string) (io.Reader, string, error) {
	pfdRelPath := *pfdLongPath
	if *pfdShortPath != "" {
		pfdRelPath = *pfdShortPath
	}

	var pfdPath string
	if filepath.IsAbs(pfdRelPath) {
		pfdPath = pfdRelPath
	} else {
		pfdPath = filepath.Join(basePath, pfdRelPath)
	}

	r, err := os.OpenFile(pfdPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("cmd.ValidatePFDOptions: %w", err)
	}
	return r, pfdPath, nil
}

type FSMOptions struct {
	PFDReader                            io.Reader `json:"-"`
	AtomicProcessTableReader             io.Reader `json:"-"`
	AtomicDeliverableTableReader         io.Reader `json:"-"`
	CompositeDeliverableTableReader      io.Reader `json:"-"`
	ResourceTableReader                  io.Reader `json:"-"`
	MilestoneTableReader                 io.Reader `json:"-"`
	GroupTableReader                     io.Reader `json:"-"`
	MaximalAvailableAllocationsThreshold int       `json:"maximal_available_allocations_threshold"`
}

type FSMRawOptions struct {
	PFDPath                              string `json:"pfd"`
	ShortPFDPath                         string `json:"-"`
	AtomicProcessTablePath               string `json:"atomic_process_table"`
	ShortAtomicProcessTablePath          string `json:"-"`
	AtomicDeliverableTablePath           string `json:"atomic_deliverable_table"`
	ShortAtomicDeliverableTablePath      string `json:"-"`
	CompositeDeliverableTablePath        string `json:"composite_deliverable_table"`
	ShortCompositeDeliverableTablePath   string `json:"-"`
	ResourceTablePath                    string `json:"resource_table"`
	ShortResourceTablePath               string `json:"-"`
	MilestoneTablePath                   string `json:"milestone_table"`
	ShortMilestoneTablePath              string `json:"-"`
	GroupTablePath                       string `json:"group_table"`
	ShortGroupTablePath                  string `json:"-"`
	MaximalAvailableAllocationsThreshold int    `json:"maximal_available_allocations_threshold"`
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

	atomicProcessTablePath := filepath.Join(basePath, atomicProcessTableRelPath)

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

	atomicDeliverableTablePath := filepath.Join(basePath, atomicDeliverableTableRelPath)

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

	resourceTablePath := filepath.Join(basePath, resourceTableRelPath)

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

	compositeProcessTablePath := filepath.Join(basePath, compositeProcessTableRelPath)

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

	compositeDeliverableTablePath := filepath.Join(basePath, compositeDeliverableTableRelPath)

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

	milestoneTablePath := filepath.Join(basePath, milestoneTableRelPath)

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
	groupTablePath := filepath.Join(basePath, groupTableRelPath)

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

func DeclareFSMOptions(flags *flag.FlagSet, options *FSMRawOptions, configLongPath *string, configShortPath *string) {
	flags.StringVar(&options.ShortPFDPath, PFDShortFlag, "", "path to the PFD")
	flags.StringVar(&options.PFDPath, PFDLongFlag, "", "path to the PFD")
	flags.IntVar(&options.MaximalAvailableAllocationsThreshold, "maximal-available-allocations-threshold", 10, "use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive")
	DeclareAtomicProcessTableOptions(flags, &options.ShortAtomicProcessTablePath, &options.AtomicProcessTablePath)
	DeclareAtomicDeliverableTableOptions(flags, &options.ShortAtomicDeliverableTablePath, &options.AtomicDeliverableTablePath)
	DeclareCompositeDeliverableTableOptions(flags, &options.ShortCompositeDeliverableTablePath, &options.CompositeDeliverableTablePath)
	DeclareResourceTableOptions(flags, &options.ShortResourceTablePath, &options.ResourceTablePath)
	DeclareMilestoneTableOptions(flags, &options.ShortMilestoneTablePath, &options.MilestoneTablePath)
	DeclareGroupTableOptions(flags, &options.ShortGroupTablePath, &options.GroupTablePath)
	DeclareConfigOptions(flags, configShortPath, configLongPath)
}

func ValidateFSMOptionsOrConfig(rawOptions *FSMRawOptions, configShortPath *string, configLongPath *string, cwd string) (*FSMOptions, error) {
	if *configShortPath != "" || *configLongPath != "" {
		options, err := ValidateFSMOptionsJSON(configShortPath, configLongPath, *rawOptions)
		if err != nil {
			return nil, fmt.Errorf("cmd.ValidateFSMOptionsOrConfig: %w", err)
		}
		return options, nil
	}
	options, err := ValidateFSMOptions(rawOptions, cwd)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptionsOrConfig: %w", err)
	}
	return options, nil
}

func ValidateFSMOptions(options *FSMRawOptions, basePath string) (*FSMOptions, error) {
	pfdReader, _, err := ValidatePFDOptions(&options.ShortPFDPath, &options.PFDPath, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptions: %w", err)
	}
	atomicProcessTableReader, _, err := ValidateAtomicProcessTableOptions(&options.ShortAtomicProcessTablePath, &options.AtomicProcessTablePath, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptions: %w", err)
	}
	resourceTableReader, _, err := ValidateResourceTableOptions(&options.ShortResourceTablePath, &options.ResourceTablePath, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptions: %w", err)
	}
	atomicDeliverableTableReader, _, err := ValidateAtomicDeliverableTableOptions(&options.ShortAtomicDeliverableTablePath, &options.AtomicDeliverableTablePath, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptions: %w", err)
	}
	compositeDeliverableTableReader, _, err := ValidateCompositeDeliverableTableOptions(&options.ShortCompositeDeliverableTablePath, &options.CompositeDeliverableTablePath, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptions: %w", err)
	}

	return &FSMOptions{
		PFDReader:                            pfdReader,
		AtomicProcessTableReader:             atomicProcessTableReader,
		AtomicDeliverableTableReader:         atomicDeliverableTableReader,
		CompositeDeliverableTableReader:      compositeDeliverableTableReader,
		ResourceTableReader:                  resourceTableReader,
		MaximalAvailableAllocationsThreshold: options.MaximalAvailableAllocationsThreshold,
	}, nil
}

func ValidateFSMOptionsJSON(shortPath *string, longPath *string, rawOptions FSMRawOptions) (*FSMOptions, error) {
	path := *longPath
	if *shortPath != "" {
		path = *shortPath
	}

	basePath := filepath.Dir(path)

	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptionsJSON: %w", err)
	}
	defer r.Close()

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&rawOptions); err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptionsJSON: %w", err)
	}

	options, err := ValidateFSMOptions(&rawOptions, basePath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ValidateFSMOptionsJSON: %w", err)
	}

	return options, nil
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
	// Mode is the search mode. best searches for optimal solutions, better searches for better solutions. poor returns approximate solutions using greedy method.
	Best   bool
	Better bool
	Poor   bool

	// QualityPreset is the quality preset. Ignored when Mode is best.
	QualityPreset string

	// Quality is the quality. Ignored when QualityPreset is not "custom" or when Mode is best.
	Quality fsm.Quality
}

func DeclareSearchOptions(flags *flag.FlagSet, options *SearchRawOptions, randomSeed int64) {
	flags.BoolVar(&options.Best, "best", false, "search best plan")
	flags.BoolVar(&options.Better, "better", false, "search better plan")
	flags.BoolVar(&options.Poor, "poor", false, "search plan by greedy algorithm (faster than best and better)")
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
	if !searchRawOptions.Best && !searchRawOptions.Better && !searchRawOptions.Poor {
		return nil, fmt.Errorf("cmd.ValidateSearchOptions: either best or better or poor must be true")
	}

	if searchRawOptions.Poor {
		return fsm.SearchFastest(), nil
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

type BusinessTimeFuncRawOptions struct {
	StartDay                  string
	StartTime                 string
	Duration                  float64
	Weekdays                  string
	AdditionalNotBusinessDays string
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
	OutputFormat               string
	BusinessTimeFuncRawOptions BusinessTimeFuncRawOptions
}

type PlanOutputFormat string

const (
	PlanOutputFormatGoogleSpreadsheetTSV PlanOutputFormat = "google-spreadsheet-tsv"
	PlanOutputFormatPlanJSON             PlanOutputFormat = "plan-json"
	PlanOutputFormatTimelineJSON         PlanOutputFormat = "timeline-json"
)

func DeclarePlanOutputFormatOptions(flags *flag.FlagSet, options *PlanOutputFormatRawOptions) {
	DeclareBusinessTimeFuncOptions(flags, &options.BusinessTimeFuncRawOptions)
	flags.StringVar(&options.OutputFormat, "out-format", "", "output format (available: google-spreadsheet-tsv, plan-json, timeline-json)")
}

func ValidatePlanOutputFormat(options *PlanOutputFormatRawOptions, logger *slog.Logger) (fsmreporter.PlanReporter, PlanOutputFormat, error) {
	switch options.OutputFormat {
	case "", "google-spreadsheet-tsv":
		businessTimeFuncOptions, err := ValidateBusinessTimeFuncOptions(&options.BusinessTimeFuncRawOptions)
		if err != nil {
			return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: %w", err)
		}
		return fsmreporter.NewGoogleSpreadsheetTimelineTSVReporter(businessTimeFuncOptions.StartDay, businessTimeFuncOptions.BusinessTimeFunc, logger), PlanOutputFormatGoogleSpreadsheetTSV, nil

	case "plan-json":
		return fsmreporter.NewPlanJSONReporter(), PlanOutputFormatPlanJSON, nil

	case "timeline-json":
		return fsmreporter.NewTimelineJSONReporter(logger), PlanOutputFormatTimelineJSON, nil

	default:
		return nil, "", fmt.Errorf("tools.ValidatePlanOutputFormat: invalid output format: %q", options.OutputFormat)
	}
}

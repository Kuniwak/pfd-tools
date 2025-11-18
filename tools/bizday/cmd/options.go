package cmd

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions           *tools.CommonOptions
	BusinessTimeFuncOptions *tools.BusinessTimeFuncOptions
	Time                    time.Time
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("bizday", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Usage: bizday [-start <start-day>] [-start-time <start-time>] [-duration <duration>] [-weekdays <weekdays>] [-not-biz-days <not-biz-days>] -time <time>")
		fmt.Fprintln(flags.Output(), "\nOptions")
		flags.PrintDefaults()
		fmt.Fprintf(flags.Output(), `
Example
  $ bizday -time '2016-01-02 15:00'
  1234

  $ bizday -start 2025-01-01 -start-time 10:00 -duration 9 -not-biz-days <(holidays -locale ja) -weekdays mon,tue,wed,thu,fri -time '2025-01-01 10:00'
  1234
`)
	}
	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var businessTimeFuncRawOptions tools.BusinessTimeFuncRawOptions
	tools.DeclareBusinessTimeFuncOptions(flags, &businessTimeFuncRawOptions)

	var timeShortFlag, timeLongFlag string
	flags.StringVar(&timeShortFlag, "t", "", "time")
	flags.StringVar(&timeLongFlag, "time", "", "time")

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

	businessTimeFuncOptions, err := tools.ValidateBusinessTimeFuncOptions(&businessTimeFuncRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var timeFlag string
	if timeShortFlag != "" {
		timeFlag = timeShortFlag
	} else {
		timeFlag = timeLongFlag
	}

	time, err := time.ParseInLocation("2006-01-02 15:04", timeFlag, time.Local)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	return &Options{
		CommonOptions:           commonOptions,
		BusinessTimeFuncOptions: businessTimeFuncOptions,
		Time:                    time,
	}, nil
}

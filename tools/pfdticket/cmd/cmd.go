package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd/pfdticket"
	"github.com/Kuniwak/pfd-tools/version"
)

const ShortHelp = "PFD と成果物表から、原子プロセスごとのチケット本文を生成します。"

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	opts, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(opts, inout); err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	return 0
}

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.ShortHelp {
		fmt.Fprintln(inout.Stdout, ShortHelp)
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	if opts.PFD == nil {
		return fmt.Errorf("cmd.MainCommandByOptions: pfd is required")
	}

	tickets := pfdticket.GenerateTickets(opts.PFD, opts.AtomicDeliverableTable, opts.Config, opts.CommonOptions.Logger)

	switch opts.OutputFormat {
	case OutputFormatJSON:
		return writeJSON(inout.Stdout, tickets)
	case OutputFormatTSV:
		return writeTSV(inout.Stdout, tickets)
	default:
		return fmt.Errorf("cmd.MainCommandByOptions: invalid output format: %q", opts.OutputFormat)
	}
}

func writeJSON(w io.Writer, tickets []pfdticket.Ticket) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tickets); err != nil {
		return fmt.Errorf("cmd.writeJSON: %w", err)
	}
	return nil
}

func writeTSV(w io.Writer, tickets []pfdticket.Ticket) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	if err := csvWriter.Write([]string{"ID", "Summary", "Description"}); err != nil {
		return fmt.Errorf("cmd.writeTSV: %w", err)
	}
	for _, ticket := range tickets {
		if err := csvWriter.Write([]string{ticket.ID, ticket.Summary, ticket.Description}); err != nil {
			return fmt.Errorf("cmd.writeTSV: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("cmd.writeTSV: %w", err)
	}
	return nil
}

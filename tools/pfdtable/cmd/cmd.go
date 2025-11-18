package cmd

import (
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	fsmtableencoding "github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	pfdtableencoding "github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	opts, err := ParseOptions(args, inout)
	if err != nil {
		_, _ = fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(opts, inout); err != nil {
		_, _ = fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	return 0
}

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.Version {
		_, _ = fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	switch opts.TableCategory {
	case TableCategoryPFD:
		switch opts.PFDTableType {
		case pfd.TableTypeAtomicProcess:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			var cdTable *pfd.CompositeDeliverableTable
			if opts.HasCompositeDeliverableTable {
				var err error
				cdTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}

			p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
				CompositeDeliverableTable: cdTable,
			}, logger)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			nodeMap := pfd.NewNodeMap(p.Nodes, logger)

			tableWriter, err := pfdtableencoding.NewAtomicProcessTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if opts.HasExistingTable {
				tableParser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				processTable, err := tableParser(opts.ExistingTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				processTable.Refresh(p, nodeMap)

				if err := tableWriter(opts.Writer, processTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				processTable := pfd.NewAtomicProcessTable(p, nodeMap)
				if err := tableWriter(opts.Writer, processTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case pfd.TableTypeAtomicDeliverable:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			var cdTable *pfd.CompositeDeliverableTable
			if opts.HasCompositeDeliverableTable {
				var err error
				cdTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}

			p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
				CompositeDeliverableTable: cdTable,
			}, logger)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			nodeMap := pfd.NewNodeMap(p.Nodes, logger)

			tableWriter, err := pfdtableencoding.NewAtomicDeliverableTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if opts.HasExistingTable {
				tableParser, err := pfdtableencoding.NewAtomicDeliverableTableParser(opts.InputFormat)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				deliverableTable, err := tableParser(opts.ExistingTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				deliverableTable.Refresh(p, nodeMap)

				if err := tableWriter(opts.Writer, deliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				deliverableTable := pfd.NewAtomicDeliverableTable(p, nodeMap)
				if err := tableWriter(opts.Writer, deliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case pfd.TableTypeCompositeProcess:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			var cdTable *pfd.CompositeDeliverableTable
			if opts.HasCompositeDeliverableTable {
				var err error
				cdTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}

			p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
				CompositeDeliverableTable: cdTable,
			}, logger)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			nodeMap := pfd.NewNodeMap(p.Nodes, logger)

			tableWriter, err := pfdtableencoding.NewCompositeProcessTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if opts.HasExistingTable {
				tableParser, err := pfdtableencoding.NewCompositeProcessTableParser(opts.InputFormat)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				compositeProcessTable, err := tableParser(opts.ExistingTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				compositeProcessTable.Refresh(p, nodeMap)

				if err := tableWriter(opts.Writer, compositeProcessTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				compositeProcessTable := pfd.NewCompositeProcessTable(p, nodeMap)
				if err := tableWriter(opts.Writer, compositeProcessTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case pfd.TableTypeCompositeDeliverable:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			var cdTable *pfd.CompositeDeliverableTable
			if opts.HasCompositeDeliverableTable {
				var err error
				cdTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
				CompositeDeliverableTable: cdTable,
			}, logger)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			nodeMap := pfd.NewNodeMap(p.Nodes, logger)
			tableWriter, err := pfdtableencoding.NewCompositeDeliverableTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if opts.HasExistingTable {
				tableParser, err := pfdtableencoding.NewCompositeDeliverableTableParser(opts.InputFormat)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				compositeDeliverableTable, err := tableParser(opts.ExistingTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				compositeDeliverableTable.Refresh(p, nodeMap)

				if err := tableWriter(opts.Writer, compositeDeliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				compositeDeliverableTable := pfd.NewCompositeDeliverableTable(p, nodeMap)
				if err := tableWriter(opts.Writer, compositeDeliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil
		default:
			panic(fmt.Sprintf("cmd.MainCommandByOptions: unknown pfdtable type: %q", opts.PFDTableType))
		}

	case TableCategoryFSM:
		switch opts.FSMTableType {
		case fsmtable.TableTypeResource:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			var cdTable *pfd.CompositeDeliverableTable
			if opts.HasCompositeDeliverableTable {
				var err error
				cdTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}

			p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
				CompositeDeliverableTable: cdTable,
			}, logger)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			nodeMap := pfd.NewNodeMap(p.Nodes, logger)

			apTableParser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			apTable, err := apTableParser(opts.AtomicProcessTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			neededResourceSetsFunc, err := fsmtable.NeededResourcesSetFuncByTable(apTable, fsmtable.DefaultNeededResourceSetsColumnSelectFunc)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			tableWriter, err := fsmtableencoding.NewResourceTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if opts.HasExistingTable {
				tableParser, err := fsmtableencoding.NewResourceTableParser(opts.InputFormat)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				resourceTable, err := tableParser(opts.ExistingTableReader)
				if err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}

				resourceTable.Refresh(p.AtomicProcesses(), nodeMap, neededResourceSetsFunc)

				if err := tableWriter(opts.Writer, resourceTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				resourceTable := fsmtable.NewResourceTableByAtomicProcessTable(apTable, neededResourceSetsFunc)
				if err := tableWriter(opts.Writer, resourceTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case fsmtable.TableTypeMilestone:
			return fmt.Errorf("cmd.MainCommandByOptions: not implemented")

		default:
			return fmt.Errorf("cmd.MainCommandByOptions: unknown fsmtable type: %q", opts.FSMTableType)
		}
	default:
		panic(fmt.Sprintf("cmd.MainCommandByOptions: unknown table category: %q", opts.TableCategory))
	}
}

package cmd

import (
	"bytes"
	"fmt"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	fsmtableencoding "github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable"
	pfdtableencoding "github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/tools"
	"github.com/Kuniwak/pfd-tools/version"
)

const ShortHelp = "PFD から要素表を作成します。要素表の更新も可能です。"

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

	if opts.CommonOptions.ShortHelp {
		fmt.Fprintln(inout.Stdout, ShortHelp)
		return nil
	}

	if opts.CommonOptions.Version {
		_, _ = fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, opts.CommonOptions.LogLevel))

	if opts.InplaceOutputPath == "" {
		return mainCommandTable(opts, inout.Stdout, logger)
	}
	buf := bytes.NewBuffer(nil)
	if err := mainCommandTable(opts, buf, logger); err != nil {
		return err
	}
	if err := tools.WriteOutput(inout.Stdout, opts.InplaceOutputPath, buf.Bytes()); err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}
	return nil
}

func mainCommandTable(opts *Options, w io.Writer, logger *slog.Logger) error {
	switch opts.TableCategory {
	case TableCategoryPFD:
		switch opts.PFDTableType {
		case pfd.TableTypeAtomicProcess:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			cdTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(opts.CompositeDeliverableTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
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
				pfdtable.EnsureAPExtraHeaders(processTable, opts.Mode, opts.Model, opts.CommonOptions.Locale)

				if err := tableWriter(w, processTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				processTable := pfd.NewAtomicProcessTable(p, nodeMap)
				pfdtable.ApplyAPMode(processTable, opts.Mode, opts.Model, opts.CommonOptions.Locale)
				if err := tableWriter(w, processTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case pfd.TableTypeAtomicDeliverable:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			cdTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(opts.CompositeDeliverableTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
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
				pfdtable.EnsureADExtraHeaders(deliverableTable, opts.Mode, opts.Model, opts.CommonOptions.Locale)

				if err := tableWriter(w, deliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				deliverableTable := pfd.NewAtomicDeliverableTable(p, nodeMap)
				pfdtable.ApplyADMode(deliverableTable, opts.Mode, opts.Model, opts.CommonOptions.Locale)
				if err := tableWriter(w, deliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case pfd.TableTypeCompositeProcess:
			if !opts.HasPFD {
				return fmt.Errorf("cmd.MainCommandByOptions: missing pfd")
			}

			cdTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(opts.CompositeDeliverableTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
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

				if err := tableWriter(w, compositeProcessTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				compositeProcessTable := pfd.NewCompositeProcessTable(p, nodeMap)
				if err := tableWriter(w, compositeProcessTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case pfd.TableTypeCompositeDeliverable:
			tableWriter, err := pfdtableencoding.NewCompositeDeliverableTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if !opts.HasPFD {
				emptyTable := &pfd.CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: []*pfd.CompositeDeliverableRow{}}
				if err := tableWriter(w, emptyTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
				return nil
			}

			cdTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(opts.CompositeDeliverableTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
				CompositeDeliverableTable: cdTable,
			}, logger)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			nodeMap := pfd.NewNodeMap(p.Nodes, logger)

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

				if err := tableWriter(w, compositeDeliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				compositeDeliverableTable := pfd.NewCompositeDeliverableTable(p, nodeMap)
				if err := tableWriter(w, compositeDeliverableTable); err != nil {
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

			cdTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(opts.CompositeDeliverableTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
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

				if err := tableWriter(w, resourceTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				resourceTable := fsmtable.NewResourceTableByAtomicProcessTable(apTable, neededResourceSetsFunc)
				if err := tableWriter(w, resourceTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			}
			return nil

		case fsmtable.TableTypeMilestone:
			tableWriter, err := fsmtableencoding.NewMilestoneTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			if opts.AtomicProcessTableReader == nil {
				if err := tableWriter(w, &fsmtable.MilestoneTable{ExtraHeaders: []string{}, Rows: []*fsmtable.MilestoneTableRow{}}); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
				return nil
			}
			apTableParser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			apTable, err := apTableParser(opts.AtomicProcessTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			if err := tableWriter(w, fsmtable.NewMilestoneTableByAtomicProcessTable(apTable)); err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			return nil

		case fsmtable.TableTypeGroup:
			tableWriter, err := fsmtableencoding.NewGroupTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			if opts.AtomicProcessTableReader == nil {
				if err := tableWriter(w, &fsmtable.GroupTable{ExtraHeaders: []string{}, Rows: []*fsmtable.GroupTableRow{}}); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
				return nil
			}
			apTableParser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			apTable, err := apTableParser(opts.AtomicProcessTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			if err := tableWriter(w, fsmtable.NewGroupTableByAtomicProcessTable(apTable)); err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			return nil

		default:
			return fmt.Errorf("cmd.MainCommandByOptions: unknown fsmtable type: %q", opts.FSMTableType)
		}
	case TableCategoryAll:
		if !opts.HasPFD {
			return fmt.Errorf("cmd.MainCommandByOptions: missing pfd for -t all")
		}

		cdTable, err := pfdtsv.ParseCompositeDeliverableTableOrEmpty(opts.CompositeDeliverableTableReader)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
			CompositeDeliverableTable: cdTable,
		}, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		nodeMap := pfd.NewNodeMap(p.Nodes, logger)

		if opts.AllExisting != nil {
			if err := mainCommandAllRefresh(opts, p, nodeMap, logger); err != nil {
				return err
			}
			return nil
		}

		type tableFile struct {
			name  string
			write func(out io.Writer) error
		}

		var apTable *pfd.AtomicProcessTable
		if opts.AtomicProcessTableReader != nil {
			apTableParser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			apTable, err = apTableParser(opts.AtomicProcessTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
		}

		layout := tools.DefaultProjectLayout("")
		tables := []tableFile{
			{layout.AtomicProcessTable, func(out io.Writer) error {
				w, err := pfdtableencoding.NewAtomicProcessTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				t := pfd.NewAtomicProcessTable(p, nodeMap)
				pfdtable.ApplyAPMode(t, opts.Mode, opts.Model, opts.CommonOptions.Locale)
				return w(out, t)
			}},
			{layout.AtomicDeliverableTable, func(out io.Writer) error {
				w, err := pfdtableencoding.NewAtomicDeliverableTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				t := pfd.NewAtomicDeliverableTable(p, nodeMap)
				pfdtable.ApplyADMode(t, opts.Mode, opts.Model, opts.CommonOptions.Locale)
				return w(out, t)
			}},
			{layout.CompositeDeliverableTable, func(out io.Writer) error {
				w, err := pfdtableencoding.NewCompositeDeliverableTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				return w(out, pfd.NewCompositeDeliverableTable(p, nodeMap))
			}},
			{layout.ResourceTable, func(out io.Writer) error {
				w, err := fsmtableencoding.NewResourceTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				if apTable == nil {
					return w(out, &fsmtable.ResourceTable{ExtraHeaders: []string{}, Rows: []*fsmtable.ResourceTableRow{}})
				}
				neededResourceSetsFunc, err := fsmtable.NeededResourcesSetFuncByTable(apTable, fsmtable.DefaultNeededResourceSetsColumnSelectFunc)
				if err != nil {
					return err
				}
				return w(out, fsmtable.NewResourceTableByAtomicProcessTable(apTable, neededResourceSetsFunc))
			}},
			{layout.MilestoneTable, func(out io.Writer) error {
				w, err := fsmtableencoding.NewMilestoneTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				if apTable == nil {
					return w(out, &fsmtable.MilestoneTable{ExtraHeaders: []string{}, Rows: []*fsmtable.MilestoneTableRow{}})
				}
				return w(out, fsmtable.NewMilestoneTableByAtomicProcessTable(apTable))
			}},
			{layout.GroupTable, func(out io.Writer) error {
				w, err := fsmtableencoding.NewGroupTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				if apTable == nil {
					return w(out, &fsmtable.GroupTable{ExtraHeaders: []string{}, Rows: []*fsmtable.GroupTableRow{}})
				}
				return w(out, fsmtable.NewGroupTableByAtomicProcessTable(apTable))
			}},
		}

		if opts.Model.Resource == execmodel.ResourceModeInfinite {
			kept := make([]tableFile, 0, len(tables))
			for _, tf := range tables {
				if tf.name == layout.ResourceTable {
					continue
				}
				kept = append(kept, tf)
			}
			tables = kept
		}

		outputs := make([]tools.FileOutput, 0, len(tables)+1)
		for _, tf := range tables {
			output, err := tools.BuildFileOutput(filepath.Join(opts.OutDir, tf.name), tf.write)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			outputs = append(outputs, output)
		}

		projectJSON, err := buildProjectJSON(opts.OutDir, opts.PFDPath, opts.Model)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		outputs = append(outputs, projectJSON)

		if err := tools.WriteFiles(outputs); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		return nil

	default:
		panic(fmt.Sprintf("cmd.MainCommandByOptions: unknown table category: %q", opts.TableCategory))
	}
}

func mainCommandAllRefresh(opts *Options, p *pfd.PFD, nodeMap map[pfd.NodeID]*pfd.Node, logger *slog.Logger) error {
	layout := tools.DefaultProjectLayout("")
	ae := opts.AllExisting
	outPath := func(key, filename string) string {
		if opts.IsInplace {
			if path, ok := opts.InplaceTargets[key]; ok {
				return path
			}
			return ""
		}
		return filepath.Join(opts.OutDir, filename)
	}

	outputs := make([]tools.FileOutput, 0, len(opts.InplaceTargets)+1)
	buildToPath := func(path string, write func(out io.Writer) error) error {
		output, err := tools.BuildFileOutput(path, write)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
		outputs = append(outputs, output)
		return nil
	}

	var apTable *pfd.AtomicProcessTable
	if opts.AtomicProcessTableReader != nil {
		apTableParser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
		apTable, err = apTableParser(opts.AtomicProcessTableReader)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
	}

	if err := buildToPath(outPath("ap", layout.AtomicProcessTable), func(out io.Writer) error {
		w, err := pfdtableencoding.NewAtomicProcessTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if ae.AP != nil {
			parser, err := pfdtableencoding.NewAtomicProcessTableParser(opts.InputFormat)
			if err != nil {
				return err
			}
			t, err := parser(ae.AP)
			if err != nil {
				return err
			}
			t.Refresh(p, nodeMap)
			pfdtable.EnsureAPExtraHeaders(t, opts.Mode, opts.Model, opts.CommonOptions.Locale)
			return w(out, t)
		}
		t := pfd.NewAtomicProcessTable(p, nodeMap)
		pfdtable.ApplyAPMode(t, opts.Mode, opts.Model, opts.CommonOptions.Locale)
		return w(out, t)
	}); err != nil {
		return err
	}

	if err := buildToPath(outPath("ad", layout.AtomicDeliverableTable), func(out io.Writer) error {
		w, err := pfdtableencoding.NewAtomicDeliverableTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if ae.AD != nil {
			parser, err := pfdtableencoding.NewAtomicDeliverableTableParser(opts.InputFormat)
			if err != nil {
				return err
			}
			t, err := parser(ae.AD)
			if err != nil {
				return err
			}
			t.Refresh(p, nodeMap)
			pfdtable.EnsureADExtraHeaders(t, opts.Mode, opts.Model, opts.CommonOptions.Locale)
			return w(out, t)
		}
		t := pfd.NewAtomicDeliverableTable(p, nodeMap)
		pfdtable.ApplyADMode(t, opts.Mode, opts.Model, opts.CommonOptions.Locale)
		return w(out, t)
	}); err != nil {
		return err
	}

	if err := buildToPath(outPath("cd", layout.CompositeDeliverableTable), func(out io.Writer) error {
		w, err := pfdtableencoding.NewCompositeDeliverableTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if ae.CD != nil {
			parser, err := pfdtableencoding.NewCompositeDeliverableTableParser(opts.InputFormat)
			if err != nil {
				return err
			}
			t, err := parser(ae.CD)
			if err != nil {
				return err
			}
			t.Refresh(p, nodeMap)
			return w(out, t)
		}
		return w(out, pfd.NewCompositeDeliverableTable(p, nodeMap))
	}); err != nil {
		return err
	}

	if opts.Model.Resource == execmodel.ResourceModeFinite || ae.R != nil {
		if err := buildToPath(outPath("r", layout.ResourceTable), func(out io.Writer) error {
			w, err := fsmtableencoding.NewResourceTableWriter(opts.OutputFormat)
			if err != nil {
				return err
			}
			empty := &fsmtable.ResourceTable{ExtraHeaders: []string{}, Rows: []*fsmtable.ResourceTableRow{}}
			if apTable == nil {
				return w(out, empty)
			}
			hasResourcesCol := fsmtable.DefaultNeededResourceSetsColumnSelectFunc(apTable.ExtraHeaders) >= 0
			if !hasResourcesCol {

				if ae.R != nil {
					parser, err := fsmtableencoding.NewResourceTableParser(opts.InputFormat)
					if err != nil {
						return err
					}
					t, err := parser(ae.R)
					if err != nil {
						return err
					}
					return w(out, t)
				}
				return w(out, empty)
			}
			neededResourceSetsFunc, err := fsmtable.NeededResourcesSetFuncByTable(apTable, fsmtable.DefaultNeededResourceSetsColumnSelectFunc)
			if err != nil {
				return err
			}
			if ae.R != nil {
				parser, err := fsmtableencoding.NewResourceTableParser(opts.InputFormat)
				if err != nil {
					return err
				}
				t, err := parser(ae.R)
				if err != nil {
					return err
				}
				t.Refresh(p.AtomicProcesses(), nodeMap, neededResourceSetsFunc)
				return w(out, t)
			}
			return w(out, fsmtable.NewResourceTableByAtomicProcessTable(apTable, neededResourceSetsFunc))
		}); err != nil {
			return err
		}
	}

	if err := buildToPath(outPath("m", layout.MilestoneTable), func(out io.Writer) error {
		w, err := fsmtableencoding.NewMilestoneTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if apTable == nil {
			return w(out, &fsmtable.MilestoneTable{ExtraHeaders: []string{}, Rows: []*fsmtable.MilestoneTableRow{}})
		}
		return w(out, fsmtable.NewMilestoneTableByAtomicProcessTable(apTable))
	}); err != nil {
		return err
	}

	if err := buildToPath(outPath("g", layout.GroupTable), func(out io.Writer) error {
		w, err := fsmtableencoding.NewGroupTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if apTable == nil {
			return w(out, &fsmtable.GroupTable{ExtraHeaders: []string{}, Rows: []*fsmtable.GroupTableRow{}})
		}
		return w(out, fsmtable.NewGroupTableByAtomicProcessTable(apTable))
	}); err != nil {
		return err
	}

	if !opts.IsInplace {
		projectJSON, err := buildProjectJSON(opts.OutDir, opts.PFDPath, opts.Model)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
		outputs = append(outputs, projectJSON)
	}

	if err := tools.WriteFiles(outputs); err != nil {
		return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
	}

	return nil
}

func buildProjectJSON(outDir string, pfdPath string, model execmodel.Model) (tools.FileOutput, error) {
	absOutDir, err := filepath.Abs(outDir)
	if err != nil {
		return tools.FileOutput{}, fmt.Errorf("cmd.buildProjectJSON: %w", err)
	}
	relPFDPath, err := filepath.Rel(absOutDir, pfdPath)
	if err != nil {
		return tools.FileOutput{}, fmt.Errorf("cmd.buildProjectJSON: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := tools.WriteProjectJSON(buf, tools.DefaultProjectLayout(relPFDPath).FSMRawOptions(model)); err != nil {
		return tools.FileOutput{}, fmt.Errorf("cmd.buildProjectJSON: %w", err)
	}

	return tools.FileOutput{Path: filepath.Join(outDir, "project.json"), Content: buf.Bytes()}, nil
}

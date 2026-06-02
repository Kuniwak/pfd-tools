package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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
				pfdtable.EnsureAPExtraHeaders(processTable, opts.Mode, opts.CommonOptions.Locale)

				if err := tableWriter(opts.Writer, processTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				processTable := pfd.NewAtomicProcessTable(p, nodeMap)
				pfdtable.ApplyAPMode(processTable, opts.Mode, opts.CommonOptions.Locale)
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
				pfdtable.EnsureADExtraHeaders(deliverableTable, opts.Mode, opts.CommonOptions.Locale)

				if err := tableWriter(opts.Writer, deliverableTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
			} else {
				deliverableTable := pfd.NewAtomicDeliverableTable(p, nodeMap)
				pfdtable.ApplyADMode(deliverableTable, opts.Mode, opts.CommonOptions.Locale)
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
			tableWriter, err := pfdtableencoding.NewCompositeDeliverableTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}

			if !opts.HasPFD {
				emptyTable := &pfd.CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: []*pfd.CompositeDeliverableRow{}}
				if err := tableWriter(opts.Writer, emptyTable); err != nil {
					return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
				}
				return nil
			}

			var cdTable *pfd.CompositeDeliverableTable
			if opts.HasCompositeDeliverableTable {
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
			tableWriter, err := fsmtableencoding.NewMilestoneTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			if opts.AtomicProcessTableReader == nil {
				if err := tableWriter(opts.Writer, &fsmtable.MilestoneTable{ExtraHeaders: []string{}, Rows: []*fsmtable.MilestoneTableRow{}}); err != nil {
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
			if err := tableWriter(opts.Writer, fsmtable.NewMilestoneTableByAtomicProcessTable(apTable)); err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			return nil

		case fsmtable.TableTypeGroup:
			tableWriter, err := fsmtableencoding.NewGroupTableWriter(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
			if opts.AtomicProcessTableReader == nil {
				if err := tableWriter(opts.Writer, &fsmtable.GroupTable{ExtraHeaders: []string{}, Rows: []*fsmtable.GroupTableRow{}}); err != nil {
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
			if err := tableWriter(opts.Writer, fsmtable.NewGroupTableByAtomicProcessTable(apTable)); err != nil {
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

		var cdTable *pfd.CompositeDeliverableTable
		if opts.HasCompositeDeliverableTable {
			var err error
			cdTable, err = pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
			}
		} else {
			cdTable = &pfd.CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: []*pfd.CompositeDeliverableRow{}}
		}

		p, err := pfdfmt.Parse("", opts.PFDReader, &pfdfmt.ParseOptions{
			CompositeDeliverableTable: cdTable,
		}, logger)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		nodeMap := pfd.NewNodeMap(p.Nodes, logger)

		// When existing tables are provided via project.json, refresh each table in place
		// or write to the output directory. Milestone and group tables have no Refresh method
		// and are always derived fresh from the AP table.
		if opts.AllExisting != nil {
			if err := mainCommandAllRefresh(opts, p, nodeMap, logger); err != nil {
				return err
			}
			return nil
		}

		type tableFile struct {
			name  string
			write func(f *os.File) error
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

		tables := []tableFile{
			{"ap.tsv", func(f *os.File) error {
				w, err := pfdtableencoding.NewAtomicProcessTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				t := pfd.NewAtomicProcessTable(p, nodeMap)
				pfdtable.ApplyAPMode(t, opts.Mode, opts.CommonOptions.Locale)
				return w(f, t)
			}},
			{"ad.tsv", func(f *os.File) error {
				w, err := pfdtableencoding.NewAtomicDeliverableTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				t := pfd.NewAtomicDeliverableTable(p, nodeMap)
				pfdtable.ApplyADMode(t, opts.Mode, opts.CommonOptions.Locale)
				return w(f, t)
			}},
			{"cd.tsv", func(f *os.File) error {
				w, err := pfdtableencoding.NewCompositeDeliverableTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				return w(f, pfd.NewCompositeDeliverableTable(p, nodeMap))
			}},
			{"r.tsv", func(f *os.File) error {
				w, err := fsmtableencoding.NewResourceTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				if apTable == nil {
					return w(f, &fsmtable.ResourceTable{ExtraHeaders: []string{}, Rows: []*fsmtable.ResourceTableRow{}})
				}
				neededResourceSetsFunc, err := fsmtable.NeededResourcesSetFuncByTable(apTable, fsmtable.DefaultNeededResourceSetsColumnSelectFunc)
				if err != nil {
					return err
				}
				return w(f, fsmtable.NewResourceTableByAtomicProcessTable(apTable, neededResourceSetsFunc))
			}},
			{"m.tsv", func(f *os.File) error {
				w, err := fsmtableencoding.NewMilestoneTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				if apTable == nil {
					return w(f, &fsmtable.MilestoneTable{ExtraHeaders: []string{}, Rows: []*fsmtable.MilestoneTableRow{}})
				}
				return w(f, fsmtable.NewMilestoneTableByAtomicProcessTable(apTable))
			}},
			{"g.tsv", func(f *os.File) error {
				w, err := fsmtableencoding.NewGroupTableWriter(opts.OutputFormat)
				if err != nil {
					return err
				}
				if apTable == nil {
					return w(f, &fsmtable.GroupTable{ExtraHeaders: []string{}, Rows: []*fsmtable.GroupTableRow{}})
				}
				return w(f, fsmtable.NewGroupTableByAtomicProcessTable(apTable))
			}},
		}

		for _, tf := range tables {
			outPath := filepath.Join(opts.OutDir, tf.name)
			f, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("cmd.MainCommandByOptions: creating %s: %w", tf.name, err)
			}
			if err := tf.write(f); err != nil {
				f.Close()
				return fmt.Errorf("cmd.MainCommandByOptions: writing %s: %w", tf.name, err)
			}
			f.Close()
		}

		absOutDir, err := filepath.Abs(opts.OutDir)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		relPFDPath, err := filepath.Rel(absOutDir, opts.PFDPath)
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		projectConfig := struct {
			PFD                       string `json:"pfd"`
			AtomicProcessTable        string `json:"atomic_process_table"`
			AtomicDeliverableTable    string `json:"atomic_deliverable_table"`
			CompositeDeliverableTable string `json:"composite_deliverable_table"`
			ResourceTable             string `json:"resource_table"`
			MilestoneTable            string `json:"milestone_table"`
			GroupTable                string `json:"group_table"`
		}{
			PFD:                       relPFDPath,
			AtomicProcessTable:        "ap.tsv",
			AtomicDeliverableTable:    "ad.tsv",
			CompositeDeliverableTable: "cd.tsv",
			ResourceTable:             "r.tsv",
			MilestoneTable:            "m.tsv",
			GroupTable:                "g.tsv",
		}
		configJSON, err := json.MarshalIndent(projectConfig, "", "\t")
		if err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
		configPath := filepath.Join(opts.OutDir, "project.json")
		if err := os.WriteFile(configPath, append(configJSON, '\n'), 0644); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: writing project.json: %w", err)
		}

		return nil

	default:
		panic(fmt.Sprintf("cmd.MainCommandByOptions: unknown table category: %q", opts.TableCategory))
	}
}

// mainCommandAllRefresh handles -t all with an existing project.json.
// Each table is refreshed from the existing data (or generated fresh if absent),
// then written to the inplace path or the output directory.
// project.json is regenerated only when writing to -out-dir, not on -inplace.
func mainCommandAllRefresh(opts *Options, p *pfd.PFD, nodeMap map[pfd.NodeID]*pfd.Node, logger *slog.Logger) error {
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

	writeToPath := func(path string, write func(f *os.File) error) error {
		if path == "" {
			return nil
		}
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: creating %s: %w", path, err)
		}
		if err := write(f); err != nil {
			f.Close()
			return fmt.Errorf("cmd.mainCommandAllRefresh: writing %s: %w", path, err)
		}
		return f.Close()
	}

	// Parse the AP table used for resource/milestone/group derivation.
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

	// AP
	if err := writeToPath(outPath("ap", "ap.tsv"), func(f *os.File) error {
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
			pfdtable.EnsureAPExtraHeaders(t, opts.Mode, opts.CommonOptions.Locale)
			return w(f, t)
		}
		t := pfd.NewAtomicProcessTable(p, nodeMap)
		pfdtable.ApplyAPMode(t, opts.Mode, opts.CommonOptions.Locale)
		return w(f, t)
	}); err != nil {
		return err
	}

	// AD
	if err := writeToPath(outPath("ad", "ad.tsv"), func(f *os.File) error {
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
			pfdtable.EnsureADExtraHeaders(t, opts.Mode, opts.CommonOptions.Locale)
			return w(f, t)
		}
		t := pfd.NewAtomicDeliverableTable(p, nodeMap)
		pfdtable.ApplyADMode(t, opts.Mode, opts.CommonOptions.Locale)
		return w(f, t)
	}); err != nil {
		return err
	}

	// CD
	if err := writeToPath(outPath("cd", "cd.tsv"), func(f *os.File) error {
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
			return w(f, t)
		}
		return w(f, pfd.NewCompositeDeliverableTable(p, nodeMap))
	}); err != nil {
		return err
	}

	// R (resource)
	if err := writeToPath(outPath("r", "r.tsv"), func(f *os.File) error {
		w, err := fsmtableencoding.NewResourceTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		empty := &fsmtable.ResourceTable{ExtraHeaders: []string{}, Rows: []*fsmtable.ResourceTableRow{}}
		if apTable == nil {
			return w(f, empty)
		}
		hasResourcesCol := fsmtable.DefaultNeededResourceSetsColumnSelectFunc(apTable.ExtraHeaders) >= 0
		if !hasResourcesCol {
			// AP is OUTPUT format (no needed-resources column); preserve existing R or write empty.
			if ae.R != nil {
				parser, err := fsmtableencoding.NewResourceTableParser(opts.InputFormat)
				if err != nil {
					return err
				}
				t, err := parser(ae.R)
				if err != nil {
					return err
				}
				return w(f, t)
			}
			return w(f, empty)
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
			return w(f, t)
		}
		return w(f, fsmtable.NewResourceTableByAtomicProcessTable(apTable, neededResourceSetsFunc))
	}); err != nil {
		return err
	}

	// M (milestone) — always derived fresh from AP; no Refresh method.
	if err := writeToPath(outPath("m", "m.tsv"), func(f *os.File) error {
		w, err := fsmtableencoding.NewMilestoneTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if apTable == nil {
			return w(f, &fsmtable.MilestoneTable{ExtraHeaders: []string{}, Rows: []*fsmtable.MilestoneTableRow{}})
		}
		return w(f, fsmtable.NewMilestoneTableByAtomicProcessTable(apTable))
	}); err != nil {
		return err
	}

	// G (group) — always derived fresh from AP; no Refresh method.
	if err := writeToPath(outPath("g", "g.tsv"), func(f *os.File) error {
		w, err := fsmtableencoding.NewGroupTableWriter(opts.OutputFormat)
		if err != nil {
			return err
		}
		if apTable == nil {
			return w(f, &fsmtable.GroupTable{ExtraHeaders: []string{}, Rows: []*fsmtable.GroupTableRow{}})
		}
		return w(f, fsmtable.NewGroupTableByAtomicProcessTable(apTable))
	}); err != nil {
		return err
	}

	// Regenerate project.json only when writing to -out-dir (not -inplace).
	if !opts.IsInplace {
		absOutDir, err := filepath.Abs(opts.OutDir)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
		relPFDPath, err := filepath.Rel(absOutDir, opts.PFDPath)
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
		projectConfig := struct {
			PFD                       string `json:"pfd"`
			AtomicProcessTable        string `json:"atomic_process_table"`
			AtomicDeliverableTable    string `json:"atomic_deliverable_table"`
			CompositeDeliverableTable string `json:"composite_deliverable_table"`
			ResourceTable             string `json:"resource_table"`
			MilestoneTable            string `json:"milestone_table"`
			GroupTable                string `json:"group_table"`
		}{
			PFD:                       relPFDPath,
			AtomicProcessTable:        "ap.tsv",
			AtomicDeliverableTable:    "ad.tsv",
			CompositeDeliverableTable: "cd.tsv",
			ResourceTable:             "r.tsv",
			MilestoneTable:            "m.tsv",
			GroupTable:                "g.tsv",
		}
		configJSON, err := json.MarshalIndent(projectConfig, "", "\t")
		if err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: %w", err)
		}
		configPath := filepath.Join(opts.OutDir, "project.json")
		if err := os.WriteFile(configPath, append(configJSON, '\n'), 0644); err != nil {
			return fmt.Errorf("cmd.mainCommandAllRefresh: writing project.json: %w", err)
		}
	}

	return nil
}

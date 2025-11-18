package encoding

import (
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding/fsmhtml"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm/fsmtable/encoding/fsmtsv"
	"github.com/Kuniwak/pfd-tools/table"
)

func NewResourceTableParser(format table.Format) (func(r io.Reader) (*fsmtable.ResourceTable, error), error) {
	switch format {
	case table.FormatTSV:
		return fsmtsv.ParseResourceTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewResourceTableParser: not supported: %q", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewResourceTableParser: unknown format: %q", format)
	}
}

func NewResourceTableWriter(format table.Format) (func(w io.Writer, table *fsmtable.ResourceTable) error, error) {
	switch format {
	case table.FormatTSV:
		return fsmtsv.WriteResourceTable, nil
	case table.FormatHTML:
		return fsmhtml.WriteResourceTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewResourceTableWriter: unknown format: %q", format)
	}
}

func NewMilestoneTableParser(format table.Format) (func(r io.Reader) (*fsmtable.MilestoneTable, error), error) {
	switch format {
	case table.FormatTSV:
		return fsmtsv.ParseMilestoneTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewMilestoneTableParser: not supported: %q", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewMilestoneTableParser: unknown format: %q", format)
	}
}

func NewMilestoneTableWriter(format table.Format) (func(w io.Writer, table *fsmtable.MilestoneTable) error, error) {
	switch format {
	case table.FormatTSV:
		return fsmtsv.WriteMilestoneTable, nil
	case table.FormatHTML:
		return fsmhtml.WriteMilestoneTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewMilestoneTableWriter: unknown format: %q", format)
	}
}

func NewGroupTableParser(format table.Format) (func(r io.Reader) (*fsmtable.GroupTable, error), error) {
	switch format {
	case table.FormatTSV:
		return fsmtsv.ParseGroupTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewGroupTableParser: not supported: %q", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewGroupTableParser: unknown format: %q", format)
	}
}

func NewGroupTableWriter(format table.Format) (func(w io.Writer, table *fsmtable.GroupTable) error, error) {
	switch format {
	case table.FormatTSV:
		return fsmtsv.WriteGroupTable, nil
	case table.FormatHTML:
		return fsmhtml.WriteGroupTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewGroupTableWriter: unknown format: %q", format)
	}
}

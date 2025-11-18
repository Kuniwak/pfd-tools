package encoding

import (
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdhtml"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/table"
)

func NewAtomicProcessTableParser(format table.Format) (func(r io.Reader) (*pfd.AtomicProcessTable, error), error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.ParseAtomicProcessTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewAtomicProcessTableParser: %q format is not supported", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewAtomicProcessTableParser: unknown format: %q", format)
	}
}

func NewAtomicDeliverableTableParser(format table.Format) (func(r io.Reader) (*pfd.AtomicDeliverableTable, error), error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.ParseAtomicDeliverableTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewDeliverableTableParser: %q format is not supported", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewDeliverableTableParser: unknown format: %q", format)
	}
}

func NewCompositeProcessTableParser(format table.Format) (func(r io.Reader) (*pfd.CompositeProcessTable, error), error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.ParseCompositeProcessTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewCompositeProcessTableParser: %q format is not supported", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewCompositeProcessTableParser: unknown format: %q", format)
	}
}

func NewAtomicProcessTableWriter(format table.Format) (func(w io.Writer, table *pfd.AtomicProcessTable) error, error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.WriteAtomicProcessTable, nil
	case table.FormatHTML:
		return pfdhtml.WriteAtomicProcessTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewAtomicProcessWriter: unknown format: %q", format)
	}
}

func NewAtomicDeliverableTableWriter(format table.Format) (func(w io.Writer, table *pfd.AtomicDeliverableTable) error, error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.WriteAtomicDeliverableTable, nil
	case table.FormatHTML:
		return pfdhtml.WriteAtomicDeliverableTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewDeliverableTableWriter: unknown format: %q", format)
	}
}

func NewCompositeProcessTableWriter(format table.Format) (func(w io.Writer, table *pfd.CompositeProcessTable) error, error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.WriteCompositeProcessTable, nil
	case table.FormatHTML:
		return pfdhtml.WriteCompositeProcessTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewCompositeProcessTableWriter: unknown format: %q", format)
	}
}

func NewCompositeDeliverableTableWriter(format table.Format) (func(w io.Writer, table *pfd.CompositeDeliverableTable) error, error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.WriteCompositeDeliverableTable, nil
	case table.FormatHTML:
		return pfdhtml.WriteCompositeDeliverableTable, nil
	default:
		return nil, fmt.Errorf("pfdencoding.NewCompositeDeliverableTableWriter: unknown format: %q", format)
	}
}

func NewCompositeDeliverableTableParser(format table.Format) (func(r io.Reader) (*pfd.CompositeDeliverableTable, error), error) {
	switch format {
	case table.FormatTSV:
		return pfdtsv.ParseCompositeDeliverableTable, nil
	case table.FormatHTML:
		return nil, fmt.Errorf("pfdencoding.NewCompositeDeliverableTableParser: %q format is not supported", format)
	default:
		return nil, fmt.Errorf("pfdencoding.NewCompositeDeliverableTableParser: unknown format: %q", format)
	}
}

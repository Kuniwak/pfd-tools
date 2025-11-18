package pfdfmt

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
)

type ParseOptions struct {
	CompositeDeliverableTable *pfd.CompositeDeliverableTable
}

func Parse(title string, r io.Reader, opts *ParseOptions, logger *slog.Logger) (*pfd.PFD, error) {
	format, r2, err := Detect(r)
	if err != nil {
		return nil, err
	}
	logger.Debug("detected format", "format", format)

	switch format {
	case FormatDrawio:
		if opts == nil || opts.CompositeDeliverableTable == nil {
			return nil, fmt.Errorf("pfdfmt.Parse: missing composite deliverable table")
		}

		p, _, err := pfddrawio.Parse(title, r2, opts.CompositeDeliverableTable, logger)
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.Parse: %w", err)
		}

		return p, nil
	default:
		return nil, fmt.Errorf("pfdfmt.Parse: unknown pfdfmt")
	}
}

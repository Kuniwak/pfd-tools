package pfdfmt

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

func ReadDrawioNodes(input []byte, logger *slog.Logger) ([]*xmldom.Node, error) {
	f, r, err := Detect(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("pfdfmt.ReadDrawioNodes: %w", err)
	}

	logger.Debug("detected format", "format", f)

	switch f {
	case FormatDrawio:
		nodes, err := xmldom.ParseXML(r)
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.ReadDrawioNodes: %w", err)
		}
		return nodes, nil

	case FormatDrawioPNG:
		xmlReader, err := pfddrawiopng.ExtractMxfile(bytes.NewReader(input))
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.ReadDrawioNodes: %w", err)
		}
		nodes, err := xmldom.ParseXML(xmlReader)
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.ReadDrawioNodes: %w", err)
		}
		return nodes, nil

	default:
		return nil, fmt.Errorf("pfdfmt.ReadDrawioNodes: not supported format: %q", f)
	}
}

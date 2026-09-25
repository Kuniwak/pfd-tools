package pfdfmt

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

func TransformDrawioBytes(input []byte, transform func(io.Reader, *slog.Logger) ([]*xmldom.Node, error), logger *slog.Logger) ([]byte, error) {
	f, r, err := Detect(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("pfdfmt.TransformDrawioBytes: %w", err)
	}

	logger.Debug("detected format", "format", f)

	switch f {
	case FormatDrawio:
		nodes, err := transform(r, logger)
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.TransformDrawioBytes: %w", err)
		}
		return xmldom.Marshal(nodes)

	case FormatDrawioPNG:
		xmlReader, err := pfddrawiopng.ExtractMxfile(bytes.NewReader(input))
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.TransformDrawioBytes: %w", err)
		}
		nodes, err := transform(xmlReader, logger)
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.TransformDrawioBytes: %w", err)
		}
		xmlBytes, err := xmldom.Marshal(nodes)
		if err != nil {
			return nil, err
		}
		out, err := pfddrawiopng.ReplaceMxfile(input, xmlBytes)
		if err != nil {
			return nil, fmt.Errorf("pfdfmt.TransformDrawioBytes: %w", err)
		}
		return out, nil

	default:
		return nil, fmt.Errorf("pfdfmt.TransformDrawioBytes: not supported format: %q", f)
	}
}

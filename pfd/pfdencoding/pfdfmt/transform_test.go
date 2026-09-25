package pfdfmt

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawiopng"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

func markMxfile(r io.Reader, _ *slog.Logger) ([]*xmldom.Node, error) {
	nodes, err := xmldom.ParseXML(r)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Kind == xmldom.ElementNode && n.Start.Name.Local == "mxfile" {
			n.SetAttr("marked", "1")
		}
	}
	return nodes, nil
}

func TestTransformDrawioBytes(t *testing.T) {
	const xml = `<mxfile host="test"><diagram id="a" name="P0"></diagram></mxfile>`
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("drawio in, drawio out", func(t *testing.T) {
		out, err := TransformDrawioBytes([]byte(xml), markMxfile, logger)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(out, []byte(`marked="1"`)) {
			t.Errorf("transform not applied; got %q", out)
		}
	})

	t.Run("png in, png out with transformed mxfile", func(t *testing.T) {
		png := buildPNGWithMxfile(t, xml)

		out, err := TransformDrawioBytes(png, markMxfile, logger)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(out, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
			t.Fatalf("output is not a PNG")
		}
		xmlReader, err := pfddrawiopng.ExtractMxfile(bytes.NewReader(out))
		if err != nil {
			t.Fatal(err)
		}
		gotXML, err := io.ReadAll(xmlReader)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(gotXML, []byte(`marked="1"`)) {
			t.Errorf("transform not applied to embedded mxfile; got %q", gotXML)
		}
	})
}

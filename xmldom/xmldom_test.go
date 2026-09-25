package xmldom

import (
	"encoding/xml"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseXML(t *testing.T) {
	testCases := map[string]struct {
		Reader   io.Reader
		Expected []*Node
	}{
		"example": {
			Reader: openFile("./testdata/example.drawio"),
			Expected: []*Node{
				{
					Kind:  0,
					Start: xml.StartElement{Name: xml.Name{Local: "mxfile"}, Attr: []xml.Attr{{Name: xml.Name{Local: "host"}, Value: "65bd71144e"}}},
					End:   xml.EndElement{Name: xml.Name{Local: "mxfile"}},
					Children: []*Node{
						{Kind: 1, Data: []byte("\n    ")},
						{
							Kind:  0,
							Start: xml.StartElement{Name: xml.Name{Local: "diagram"}, Attr: []xml.Attr{{Name: xml.Name{Local: "id"}, Value: "_v1VNxwC-s0niwqlnLhB"}, {Name: xml.Name{Local: "name"}, Value: "P0"}}},
							End:   xml.EndElement{Name: xml.Name{Local: "diagram"}},
							Children: []*Node{
								{Kind: 1, Data: []byte("\n        ")},
								{
									Kind: 0,
									Start: xml.StartElement{
										Name: xml.Name{Local: "mxGraphModel"},
										Attr: []xml.Attr{
											{Name: xml.Name{Local: "dx"}, Value: "734"},
											{Name: xml.Name{Local: "dy"}, Value: "536"},
											{Name: xml.Name{Local: "grid"}, Value: "1"},
											{Name: xml.Name{Local: "gridSize"}, Value: "10"},
											{Name: xml.Name{Local: "guides"}, Value: "1"},
											{Name: xml.Name{Local: "tooltips"}, Value: "1"},
											{Name: xml.Name{Local: "connect"}, Value: "1"},
											{Name: xml.Name{Local: "arrows"}, Value: "1"},
											{Name: xml.Name{Local: "fold"}, Value: "1"},
											{Name: xml.Name{Local: "page"}, Value: "1"},
											{Name: xml.Name{Local: "pageScale"}, Value: "1"},
											{Name: xml.Name{Local: "pageWidth"}, Value: "827"},
											{Name: xml.Name{Local: "pageHeight"}, Value: "1169"},
											{Name: xml.Name{Local: "math"}, Value: "0"},
											{Name: xml.Name{Local: "shadow"}, Value: "0"},
										},
									},
									End: xml.EndElement{Name: xml.Name{Local: "mxGraphModel"}},
									Children: []*Node{
										{Kind: 1, Data: []byte("\n            ")},
										{
											Kind:  0,
											Start: xml.StartElement{Name: xml.Name{Local: "root"}, Attr: []xml.Attr{}},
											End:   xml.EndElement{Name: xml.Name{Local: "root"}},
											Children: []*Node{
												{Kind: 1, Data: []byte("\n                ")},
												{Start: xml.StartElement{Name: xml.Name{Local: "mxCell"}, Attr: []xml.Attr{{Name: xml.Name{Local: "id"}, Value: "0"}}}, End: xml.EndElement{Name: xml.Name{Local: "mxCell"}}},
												{Kind: 1, Data: []byte("\n                ")},
												{Start: xml.StartElement{Name: xml.Name{Local: "mxCell"}, Attr: []xml.Attr{{Name: xml.Name{Local: "id"}, Value: "1"}, {Name: xml.Name{Local: "parent"}, Value: "0"}}}, End: xml.EndElement{Name: xml.Name{Local: "mxCell"}}},
												{Kind: 1, Data: []byte("\n                ")},
												{
													Kind: 0,
													Start: xml.StartElement{
														Name: xml.Name{Local: "mxCell"},
														Attr: []xml.Attr{
															{Name: xml.Name{Local: "id"}, Value: "4"},
															{Name: xml.Name{Local: "value"}},
															{Name: xml.Name{Local: "style"}, Value: "edgeStyle=none;html=1;"},
															{Name: xml.Name{Local: "parent"}, Value: "1"},
															{Name: xml.Name{Local: "source"}, Value: "2"},
															{Name: xml.Name{Local: "target"}, Value: "3"},
															{Name: xml.Name{Local: "edge"}, Value: "1"},
														},
													},
													End: xml.EndElement{Name: xml.Name{Local: "mxCell"}},
													Children: []*Node{
														{Kind: 1, Data: []byte("\n                    ")},
														{
															Kind: 0,
															Start: xml.StartElement{
																Name: xml.Name{Local: "mxGeometry"},
																Attr: []xml.Attr{
																	{
																		Name: xml.Name{
																			Space: "",
																			Local: "relative",
																		},
																		Value: "1",
																	},
																	{
																		Name: xml.Name{
																			Space: "",
																			Local: "as",
																		},
																		Value: "geometry",
																	},
																},
															},
															End:      xml.EndElement{Name: xml.Name{Local: "mxGeometry"}},
															Children: nil,
															Data:     nil,
															PI:       nil,
														},
														{Kind: 1, Data: []byte("\n                ")},
													},
													Data: nil,
													PI:   nil,
												},
												{Kind: 1, Data: []byte("\n                ")},
												{
													Kind: 0,
													Start: xml.StartElement{
														Name: xml.Name{Local: "mxCell"},
														Attr: []xml.Attr{
															{
																Name: xml.Name{
																	Space: "",
																	Local: "id",
																},
																Value: "2",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "value",
																},
																Value: "入力成果物",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "style",
																},
																Value: "rounded=0;whiteSpace=wrap;html=1;",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "parent",
																},
																Value: "1",
															},
															{Name: xml.Name{Local: "vertex"}, Value: "1"},
														},
													},
													End: xml.EndElement{Name: xml.Name{Local: "mxCell"}},
													Children: []*Node{
														{Kind: 1, Data: []byte("\n                    ")},
														{
															Kind: 0,
															Start: xml.StartElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
																Attr: []xml.Attr{
																	{Name: xml.Name{Local: "x"}, Value: "160"},
																	{Name: xml.Name{Local: "y"}, Value: "240"},
																	{Name: xml.Name{Local: "width"}, Value: "120"},
																	{Name: xml.Name{Local: "height"}, Value: "80"},
																	{Name: xml.Name{Local: "as"}, Value: "geometry"},
																},
															},
															End: xml.EndElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
															},
															Children: nil,
															Data:     nil,
															PI:       nil,
														},
														{Kind: 1, Data: []byte("\n                ")},
													},
													Data: nil,
													PI:   nil,
												},
												{Kind: 1, Data: []byte("\n                ")},
												{
													Kind: 0,
													Start: xml.StartElement{
														Name: xml.Name{Local: "mxCell"},
														Attr: []xml.Attr{
															{
																Name: xml.Name{
																	Space: "",
																	Local: "id",
																},
																Value: "6",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "value",
																},
																Value: "",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "style",
																},
																Value: "edgeStyle=none;html=1;",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "parent",
																},
																Value: "1",
															},
															{Name: xml.Name{Local: "source"}, Value: "3"},
															{Name: xml.Name{Local: "target"}, Value: "5"},
															{Name: xml.Name{Local: "edge"}, Value: "1"},
														},
													},
													End: xml.EndElement{Name: xml.Name{Local: "mxCell"}},
													Children: []*Node{
														{Kind: 1, Data: []byte("\n                    ")},
														{
															Kind: 0,
															Start: xml.StartElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
																Attr: []xml.Attr{
																	{Name: xml.Name{Local: "relative"}, Value: "1"},
																	{Name: xml.Name{Local: "as"}, Value: "geometry"},
																},
															},
															End: xml.EndElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
															},
															Children: nil,
															Data:     nil,
															PI:       nil,
														},
														{Kind: 1, Data: []byte("\n                ")},
													},
													Data: nil,
													PI:   nil,
												},
												{Kind: 1, Data: []byte("\n                ")},
												{
													Kind: 0,
													Start: xml.StartElement{
														Name: xml.Name{Local: "mxCell"},
														Attr: []xml.Attr{
															{
																Name: xml.Name{
																	Space: "",
																	Local: "id",
																},
																Value: "3",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "value",
																},
																Value: "プロセス",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "style",
																},
																Value: "ellipse;whiteSpace=wrap;html=1;",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "parent",
																},
																Value: "1",
															},
															{Name: xml.Name{Local: "vertex"}, Value: "1"},
														},
													},
													End: xml.EndElement{Name: xml.Name{Local: "mxCell"}},
													Children: []*Node{
														{Kind: 1, Data: []byte("\n                    ")},
														{
															Kind: 0,
															Start: xml.StartElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
																Attr: []xml.Attr{
																	{Name: xml.Name{Local: "x"}, Value: "320"},
																	{Name: xml.Name{Local: "y"}, Value: "240"},
																	{Name: xml.Name{Local: "width"}, Value: "120"},
																	{Name: xml.Name{Local: "height"}, Value: "80"},
																	{Name: xml.Name{Local: "as"}, Value: "geometry"},
																},
															},
															End: xml.EndElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
															},
															Children: nil,
															Data:     nil,
															PI:       nil,
														},
														{Kind: 1, Data: []byte("\n                ")},
													},
													Data: nil,
													PI:   nil,
												},
												{Kind: 1, Data: []byte("\n                ")},
												{
													Kind: 0,
													Start: xml.StartElement{
														Name: xml.Name{Local: "mxCell"},
														Attr: []xml.Attr{
															{
																Name: xml.Name{
																	Space: "",
																	Local: "id",
																},
																Value: "5",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "value",
																},
																Value: "出力成果物",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "style",
																},
																Value: "rounded=0;whiteSpace=wrap;html=1;",
															},
															{
																Name: xml.Name{
																	Space: "",
																	Local: "parent",
																},
																Value: "1",
															},
															{Name: xml.Name{Local: "vertex"}, Value: "1"},
														},
													},
													End: xml.EndElement{Name: xml.Name{Local: "mxCell"}},
													Children: []*Node{
														{Kind: 1, Data: []byte("\n                    ")},
														{
															Kind: 0,
															Start: xml.StartElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
																Attr: []xml.Attr{
																	{Name: xml.Name{Local: "x"}, Value: "480"},
																	{Name: xml.Name{Local: "y"}, Value: "240"},
																	{Name: xml.Name{Local: "width"}, Value: "120"},
																	{Name: xml.Name{Local: "height"}, Value: "80"},
																	{Name: xml.Name{Local: "as"}, Value: "geometry"},
																},
															},
															End: xml.EndElement{
																Name: xml.Name{
																	Space: "",
																	Local: "mxGeometry",
																},
															},
															Children: nil,
															Data:     nil,
															PI:       nil,
														},
														{Kind: 1, Data: []byte("\n                ")},
													},
													Data: nil,
													PI:   nil,
												},
												{Kind: 1, Data: []byte("\n            ")},
											},
											Data: nil,
											PI:   nil,
										},
										{Kind: 1, Data: []byte("\n        ")},
									},
									Data: nil,
									PI:   nil,
								},
								{Kind: 1, Data: []byte("\n    ")},
							},
							Data: nil,
							PI:   nil,
						},
						{Kind: 1, Data: []byte("\n")},
					},
					Data: nil,
					PI:   nil,
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual, err := ParseXML(testCase.Reader)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func openFile(filePath string) io.Reader {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	return f
}

func TestNodeSetAttr(t *testing.T) {
	nodes, err := ParseXML(strings.NewReader(`<a x="1"><b/></a>`))
	if err != nil {
		t.Fatal(err)
	}
	a := nodes[0]

	a.SetAttr("x", "2")
	if v, _ := a.GetAttr("x", ""); v != "2" {
		t.Errorf("SetAttr update: GetAttr(x) = %q, want 2", v)
	}

	a.SetAttr("y", "3")
	if v, ok := a.GetAttr("y", ""); !ok || v != "3" {
		t.Errorf("SetAttr append: GetAttr(y) = %q, ok=%v, want 3, true", v, ok)
	}
}

func TestNodeFloatAttr(t *testing.T) {
	nodes, err := ParseXML(strings.NewReader(`<g x="12.5" bad="nope"/>`))
	if err != nil {
		t.Fatal(err)
	}
	g := nodes[0]

	if f, ok := g.FloatAttr("x"); !ok || f != 12.5 {
		t.Errorf("FloatAttr(x) = %v, %v; want 12.5, true", f, ok)
	}
	if _, ok := g.FloatAttr("bad"); ok {
		t.Error("FloatAttr(bad) ok = true; want false")
	}
	if _, ok := g.FloatAttr("missing"); ok {
		t.Error("FloatAttr(missing) ok = true; want false")
	}
}

func TestNodeChildHelpers(t *testing.T) {
	nodes, err := ParseXML(strings.NewReader("<root>\n  <child a=\"1\"/>\n</root>"))
	if err != nil {
		t.Fatal(err)
	}
	root := nodes[0]

	if c := root.FirstChildElement("child"); c == nil {
		t.Fatal("FirstChildElement(child) = nil, want the child element")
	}
	if c := root.FirstChildElement("nope"); c != nil {
		t.Errorf("FirstChildElement(nope) = %v, want nil", c)
	}
	if !root.HasChildElements() {
		t.Error("HasChildElements() = false, want true")
	}

	sawBlank := false
	for _, ch := range root.Children {
		if ch.Kind == TextNode {
			sawBlank = true
			if !ch.IsBlankText() {
				t.Errorf("IsBlankText() = false for whitespace text %q", string(ch.Data))
			}
		}
	}
	if !sawBlank {
		t.Error("expected at least one whitespace text node")
	}

	leaf, err := ParseXML(strings.NewReader("<leaf>text</leaf>"))
	if err != nil {
		t.Fatal(err)
	}
	if leaf[0].HasChildElements() {
		t.Error("leaf HasChildElements() = true, want false (only a text child)")
	}
	for _, ch := range leaf[0].Children {
		if ch.Kind == TextNode && ch.IsBlankText() {
			t.Errorf("IsBlankText() = true for non-blank text %q", string(ch.Data))
		}
	}
}

func TestMarshal(t *testing.T) {
	const src = `<mxfile><diagram id="d1"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram></mxfile>`
	nodes, err := ParseXML(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ParseXML: %v", err)
	}
	out, err := Marshal(nodes)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), `<diagram id="d1">`) {
		t.Errorf("Marshal output missing diagram element: %s", out)
	}

	renodes, err := ParseXML(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-ParseXML: %v", err)
	}
	out2, err := Marshal(renodes)
	if err != nil {
		t.Fatalf("re-Marshal: %v", err)
	}
	if string(out) != string(out2) {
		t.Errorf("Marshal not idempotent:\nfirst:  %s\nsecond: %s", out, out2)
	}
}

func TestNewElementAndNewText(t *testing.T) {
	el := NewElement("foo")
	el.SetAttr("a", "1")
	el.Children = append(el.Children, NewText("hello & <world>"))

	child := NewElement("bar")
	el.Children = append(el.Children, child)

	buf := &strings.Builder{}
	if err := el.Write(buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	want := `<foo a="1">hello &amp; &lt;world&gt;<bar></bar></foo>`
	if got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestAppendChildIndented(t *testing.T) {
	roots, err := ParseXML(strings.NewReader("<p>\n    <a></a>\n</p>"))
	if err != nil {
		t.Fatal(err)
	}
	p := roots[0]
	p.AppendChildIndented("\n    ", NewElement("b"))

	buf := &strings.Builder{}
	if err := p.Write(buf); err != nil {
		t.Fatal(err)
	}
	want := "<p>\n    <a></a>\n    <b></b>\n</p>"
	if buf.String() != want {
		t.Errorf("Write() = %q, want %q", buf.String(), want)
	}
}

func TestRemoveChildElements(t *testing.T) {
	testCases := map[string]struct {
		src    string
		remove string
		want   string
	}{
		"複数の子から 1 つだけ除く": {
			src:    "<g>\n  <a/>\n  <b/>\n</g>",
			remove: "a",
			want:   "<g>\n  <b/>\n</g>",
		},
		"同名の子をすべて除く": {
			src:    "<g>\n  <a/>\n  <b/>\n  <a/>\n</g>",
			remove: "a",
			want:   "<g>\n  <b/>\n</g>",
		},
		"要素が残らなければ中身を空にする": {
			src:    "<g>\n  <a/>\n</g>",
			remove: "a",
			want:   "<g></g>",
		},
		"一致しなければ変えない": {
			src:    "<g>\n  <a/>\n</g>",
			remove: "z",
			want:   "<g>\n  <a/>\n</g>",
		},
		"子がなければ変えない": {
			src:    "<g></g>",
			remove: "a",
			want:   "<g></g>",
		},
		"要素でないテキストは残す": {
			src:    "<g>hello</g>",
			remove: "a",
			want:   "<g>hello</g>",
		},
		"除去後に残るテキストも残す": {
			src:    "<g>hello<a/></g>",
			remove: "a",
			want:   "<g>hello</g>",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			roots, err := ParseXML(strings.NewReader(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			g := roots[0]

			g.RemoveChildElements(func(n *Node) bool {
				return n.Start.Name.Local == tc.remove
			})

			buf := &strings.Builder{}
			if err := g.Write(buf); err != nil {
				t.Fatal(err)
			}

			want := strings.NewReplacer("<a/>", "<a></a>", "<b/>", "<b></b>").Replace(tc.want)
			if buf.String() != want {
				t.Errorf("Write() = %q, want %q", buf.String(), want)
			}
		})
	}
}

func TestNodeCloneIsolatesAttrs(t *testing.T) {
	roots, err := ParseXML(strings.NewReader(`<a id="1" v="x"><b/></a>`))
	if err != nil {
		t.Fatal(err)
	}
	src := roots[0]
	clone := src.Clone()

	clone.SetAttr("id", "2")
	if got, _ := src.GetAttr("id", ""); got != "1" {
		t.Errorf("source id changed to %q; clone must not alias source attrs", got)
	}
	if got, _ := clone.GetAttr("id", ""); got != "2" {
		t.Errorf("clone id = %q, want 2", got)
	}
}

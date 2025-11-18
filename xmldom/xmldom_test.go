package xmldom

import (
	"encoding/xml"
	"io"
	"os"
	"reflect"
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
																Value: "Input deliverable",
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
																Value: "Process",
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
																Value: "Output deliverable",
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

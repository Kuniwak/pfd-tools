package pfddrawio

import (
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		FilePath string
		Expected []Diagram
	}{
		"without XML declaration": {
			FilePath: "testdata/example.drawio",
			Expected: exampleFile,
		},
		"with XML declaration": {
			FilePath: "testdata/example_with_xml_decl.drawio",
			Expected: exampleFile,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			exampleFile, err := os.Open(testCase.FilePath)
			if err != nil {
				t.Fatal(err)
			}
			defer sugar.IgnoreError(exampleFile.Close)
			logger := slog.New(slogtest.NewTestHandler(t))
			actual, err := ParseDiagrams(exampleFile, logger)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, testCase.Expected) {
				t.Error(cmp.Diff(testCase.Expected, actual))
			}
		})
	}
}

func TestParseDiagramsError(t *testing.T) {
	testCases := map[string]struct {
		XML string
	}{

		"diagram without id attribute": {
			XML: `<mxfile><diagram name="P0"><mxGraphModel><root><mxCell id="0"/></root></mxGraphModel></diagram></mxfile>`,
		},

		"diagram without name attribute": {
			XML: `<mxfile><diagram id="d0"><mxGraphModel><root><mxCell id="0"/></root></mxGraphModel></diagram></mxfile>`,
		},

		"compressed diagram has no mxGraphModel": {
			XML: `<mxfile><diagram id="d0" name="P0">7VjLbtswEPwaH3MgKcmyj0mTHnpKgQI9BwzFSGwoUqBox+7Xd0lRtmwnbYo0KAr4YHK5s7PkPkR6ll+p7SdD2vpB1Fxl+</diagram></mxfile>`,
		},

		"mxGraphModel without root": {
			XML: `<mxfile><diagram id="d0" name="P0"><mxGraphModel/></diagram></mxfile>`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			_, err := ParseDiagrams(strings.NewReader(testCase.XML), logger)
			if err == nil {
				t.Fatal("expected an error for the dropped page, got nil")
			}
		})
	}
}

func TestParseVertexValue(t *testing.T) {
	testCases := map[string]struct {
		Input        string
		ExpectedID   pfd.NodeID
		ExpectedDesc string
	}{
		"no descriptions": {
			Input:        "D1",
			ExpectedID:   "D1",
			ExpectedDesc: "",
		},
		"description": {
			Input:        "D1: 実装",
			ExpectedID:   "D1",
			ExpectedDesc: "実装",
		},
		"empty description": {
			Input:        "D1: ",
			ExpectedID:   "D1",
			ExpectedDesc: "",
		},
		"spaces before colon": {
			Input:        "D1 : 実装",
			ExpectedID:   "D1",
			ExpectedDesc: "実装",
		},
		"description with colons": {
			Input:        "D1: 実装: 実装",
			ExpectedID:   "D1",
			ExpectedDesc: "実装: 実装",
		},
		"duplicate mark (full width) is transparent": {
			Input:        "D1＊: 実装",
			ExpectedID:   "D1",
			ExpectedDesc: "実装",
		},
		"duplicate mark (half width) is transparent": {
			Input:        "D1*: 実装",
			ExpectedID:   "D1",
			ExpectedDesc: "実装",
		},
		"duplicate mark without description": {
			Input:        "D1＊",
			ExpectedID:   "D1",
			ExpectedDesc: "",
		},
		"asterisk in an unnumbered label is kept": {
			Input:        "実装＊",
			ExpectedID:   "実装＊",
			ExpectedDesc: "",
		},
		"asterisk in a process ID is kept": {
			Input:        "P1＊: 実装する",
			ExpectedID:   "P1＊",
			ExpectedDesc: "実装する",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actualID, actualDesc, err := ParseVertexValue(testCase.Input)
			if err != nil {
				t.Fatal(err)
			}
			if actualID != testCase.ExpectedID {
				t.Errorf("expected ID: %q, actual: %q", testCase.ExpectedID, actualID)
			}
			if actualDesc != testCase.ExpectedDesc {
				t.Errorf("expected description: %q, actual: %q", testCase.ExpectedDesc, actualDesc)
			}
		})
	}
}

func TestParseValueHTML(t *testing.T) {
	testCases := map[string]struct {
		Input    ValueHTML
		Expected string
	}{
		"no tags": {
			Input:    "D1: 実装",
			Expected: "D1: 実装",
		},
		"wrapping": {
			Input:    "D1: <b>実装</b>",
			Expected: "D1: 実装",
		},
		"embedded tags": {
			Input:    "D1: 実<br>装",
			Expected: "D1: 実装",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			sb := strings.Builder{}
			err := ParseValueHTML(testCase.Input, &sb)
			if err != nil {
				t.Fatal(err)
			}
			if sb.String() != testCase.Expected {
				t.Errorf("expected: %q, actual: %q", testCase.Expected, sb.String())
			}
		})
	}
}

func TestParseExpandsCompositeDeliverables(t *testing.T) {
	const xml = `<mxfile><diagram id="d0" name="P0"><mxGraphModel><root>
<mxCell id="0"/>
<mxCell id="1" parent="0"/>
<mxCell id="n1" value="D0: 初期成果物" style="rounded=0;whiteSpace=wrap;html=1;" vertex="1" parent="1"><mxGeometry x="0" y="0" width="120" height="60" as="geometry"/></mxCell>
<mxCell id="n2" value="P1: プロセス1" style="ellipse;whiteSpace=wrap;html=1;" vertex="1" parent="1"><mxGeometry x="160" y="0" width="120" height="60" as="geometry"/></mxCell>
<mxCell id="n3" value="D3: 複合成果物" style="rounded=0;whiteSpace=wrap;html=1;strokeWidth=3;" vertex="1" parent="1"><mxGeometry x="320" y="0" width="120" height="60" as="geometry"/></mxCell>
<mxCell id="n4" value="P2: プロセス2" style="ellipse;whiteSpace=wrap;html=1;" vertex="1" parent="1"><mxGeometry x="480" y="0" width="120" height="60" as="geometry"/></mxCell>
<mxCell id="n5" value="D5: 最終成果物" style="rounded=0;whiteSpace=wrap;html=1;" vertex="1" parent="1"><mxGeometry x="640" y="0" width="120" height="60" as="geometry"/></mxCell>
<mxCell id="e1" style="html=1;" edge="1" parent="1" source="n1" target="n2"><mxGeometry relative="1" as="geometry"/></mxCell>
<mxCell id="e2" style="html=1;" edge="1" parent="1" source="n2" target="n3"><mxGeometry relative="1" as="geometry"/></mxCell>
<mxCell id="e3" style="html=1;" edge="1" parent="1" source="n3" target="n4"><mxGeometry relative="1" as="geometry"/></mxCell>
<mxCell id="e4" style="html=1;" edge="1" parent="1" source="n4" target="n5"><mxGeometry relative="1" as="geometry"/></mxCell>
</root></mxGraphModel></diagram></mxfile>`

	expandedEdges := sets.New(
		(*pfd.Edge).Compare,
		&pfd.Edge{Source: "D0", Target: "P1"},
		&pfd.Edge{Source: "P1", Target: "D3"},
		&pfd.Edge{Source: "P1", Target: "D1"},
		&pfd.Edge{Source: "P1", Target: "D2"},
		&pfd.Edge{Source: "D3", Target: "P2"},
		&pfd.Edge{Source: "D1", Target: "P2"},
		&pfd.Edge{Source: "D2", Target: "P2"},
		&pfd.Edge{Source: "P2", Target: "D5"},
	)
	expandedComposition := map[pfd.NodeID]*sets.Set[pfd.NodeID]{
		"D3": sets.New(pfd.NodeID.Compare, "D1", "D2"),
	}

	testCases := map[string]struct {
		Rows                []*pfd.CompositeDeliverableRow
		ExpectedEdges       *sets.Set[*pfd.Edge]
		ExpectedComposition map[pfd.NodeID]*sets.Set[pfd.NodeID]
		WantErr             bool
	}{
		"both ends of an edge are expanded": {
			Rows: []*pfd.CompositeDeliverableRow{
				{ID: "D3", Description: "複合成果物", Deliverables: []pfd.NodeID{"D1", "D2"}, ExtraCells: []string{}},
			},
			ExpectedEdges:       expandedEdges,
			ExpectedComposition: expandedComposition,
		},

		"nested composition is flattened": {
			Rows: []*pfd.CompositeDeliverableRow{
				{ID: "D3", Description: "複合成果物", Deliverables: []pfd.NodeID{"D6"}, ExtraCells: []string{}},
				{ID: "D6", Description: "複合成果物6", Deliverables: []pfd.NodeID{"D1", "D2"}, ExtraCells: []string{}},
			},
			ExpectedEdges:       expandedEdges,
			ExpectedComposition: expandedComposition,
		},
		"cyclic nesting is an error": {
			Rows: []*pfd.CompositeDeliverableRow{
				{ID: "D3", Description: "複合成果物", Deliverables: []pfd.NodeID{"D6"}, ExtraCells: []string{}},
				{ID: "D6", Description: "複合成果物6", Deliverables: []pfd.NodeID{"D3"}, ExtraCells: []string{}},
			},
			WantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cdt := &pfd.CompositeDeliverableTable{ExtraHeaders: []string{}, Rows: tc.Rows}
			logger := slog.New(slogtest.NewTestHandler(t))

			p, _, err := Parse("", strings.NewReader(xml), cdt, logger)
			if tc.WantErr {
				if err == nil {
					t.Fatal("expected an error for the cyclic nesting, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if !sets.IsEqual((*pfd.Edge).Compare, p.Edges, tc.ExpectedEdges) {
				t.Error(cmp.Diff(tc.ExpectedEdges.Slice(), p.Edges.Slice()))
			}
			if !reflect.DeepEqual(p.DeliverableComposition, tc.ExpectedComposition) {
				t.Error(cmp.Diff(tc.ExpectedComposition, p.DeliverableComposition))
			}
		})
	}
}

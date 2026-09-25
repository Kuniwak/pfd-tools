package pfddrawio

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

const sameLabelDrawio = `<mxfile host="x">
    <diagram id="d0" name="P0">
        <mxGraphModel><root>
            <mxCell id="0"/>
            <mxCell id="1" parent="0"/>
            <mxCell id="2" value="実装する" style="ellipse;whiteSpace=wrap;html=1;" vertex="1" parent="1"/>
            <mxCell id="3" value="実装する" style="rounded=0;whiteSpace=wrap;html=1;" vertex="1" parent="1"/>
        </root></mxGraphModel>
    </diagram>
</mxfile>
`

func TestApplyRenumberPlanIsKindQualified(t *testing.T) {
	testCases := map[string]struct {
		Plan     pfd.RenumberPlan
		Contains []string
		Excludes []string
	}{
		"a deliverable key rewrites only the rectangle": {
			Plan: pfd.RenumberPlan{
				"[実装する]": {ID: "D1", Description: "実装する", Type: pfd.NodeTypeAtomicDeliverable},
			},
			Contains: []string{`value="D1: 実装する"`, `value="実装する"`},
		},
		"a process key rewrites only the ellipse": {
			Plan: pfd.RenumberPlan{
				"(実装する)": {ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
			},
			Contains: []string{`value="P1: 実装する"`, `value="実装する"`},
		},
		"both keys rewrite both shapes": {
			Plan: pfd.RenumberPlan{
				"(実装する)": {ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
				"[実装する]": {ID: "D1", Description: "実装する", Type: pfd.NodeTypeAtomicDeliverable},
			},
			Contains: []string{`value="P1: 実装する"`, `value="D1: 実装する"`},
			Excludes: []string{`value="実装する"`},
		},

		"a plain label key rewrites nothing": {
			Plan: pfd.RenumberPlan{
				"実装する": {ID: "P1", Description: "実装する", Type: pfd.NodeTypeAtomicProcess},
			},
			Excludes: []string{`value="P1: 実装する"`, `value="D1: 実装する"`},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodes, err := xmldom.ParseXML(strings.NewReader(sameLabelDrawio))
			if err != nil {
				t.Fatal(err)
			}

			ApplyRenumberPlan(nodes, testCase.Plan, logger)

			w := bytes.NewBuffer(nil)
			for _, node := range nodes {
				if err := node.Write(w); err != nil {
					t.Fatal(err)
				}
			}

			got := w.String()
			for _, want := range testCase.Contains {
				if !strings.Contains(got, want) {
					t.Errorf("the result does not contain %s:\n%s", want, got)
				}
			}
			for _, unwanted := range testCase.Excludes {
				if strings.Contains(got, unwanted) {
					t.Errorf("the result must not contain %s:\n%s", unwanted, got)
				}
			}
		})
	}
}

func TestMaxNumberedIDs(t *testing.T) {
	testCases := map[string]struct {
		Drawio          string
		WantProcess     int
		WantDeliverable int
	}{
		"unnumbered labels are not counted": {
			Drawio:          sameLabelDrawio,
			WantProcess:     0,
			WantDeliverable: 0,
		},
		"a bare assigned ID with no description still counts": {
			Drawio: strings.NewReplacer(
				`value="実装する" style="ellipse`, `value="P3" style="ellipse`,
				`value="実装する" style="rounded`, `value="D4" style="rounded`,
			).Replace(sameLabelDrawio),
			WantProcess:     3,
			WantDeliverable: 4,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodes, err := xmldom.ParseXML(strings.NewReader(testCase.Drawio))
			if err != nil {
				t.Fatal(err)
			}

			gotProcess, gotDeliverable := MaxNumberedIDs(nodes, logger)
			if gotProcess != testCase.WantProcess || gotDeliverable != testCase.WantDeliverable {
				t.Errorf("MaxNumberedIDs = (%d, %d), want (%d, %d)",
					gotProcess, gotDeliverable, testCase.WantProcess, testCase.WantDeliverable)
			}
		})
	}
}

func TestRenumber(t *testing.T) {
	tests := []struct {
		FilePath     string
		ExpectedPath string
	}{
		{
			FilePath:     "testdata/sequential_without_id.drawio",
			ExpectedPath: "testdata/sequential_with_id.drawio",
		},
		{
			FilePath:     "testdata/default_page_name_en_without_id.drawio",
			ExpectedPath: "testdata/default_page_name_en_with_id.drawio",
		},
		{
			FilePath:     "testdata/default_page_name_ja_without_id.drawio",
			ExpectedPath: "testdata/default_page_name_ja_with_id.drawio",
		},
	}
	for _, test := range tests {
		t.Run(test.FilePath, func(t *testing.T) {
			f1, err := os.OpenFile(test.FilePath, os.O_RDONLY, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer f1.Close()

			bs, err := os.ReadFile(test.ExpectedPath)
			if err != nil {
				t.Fatal(err)
			}

			nodes, err := Renumber(f1, slog.New(slogtest.NewTestHandler(t)))
			if err != nil {
				t.Fatal(err)
			}

			w := bytes.NewBuffer(nil)
			for _, node := range nodes {
				if err := node.Write(w); err != nil {
					t.Fatal(err)
				}
			}

			expected := string(bs)
			if expected != w.String() {
				t.Logf("got: %s", w.String())
				t.Error(cmp.Diff(expected, w.String()))
			}
		})
	}
}

func TestRenumberFollowsPageNames(t *testing.T) {
	page := func(id, name, cells string) string {
		return `<diagram id="` + id + `" name="` + name + `"><mxGraphModel><root>` +
			`<mxCell id="0"/><mxCell id="1" parent="0"/>` + cells +
			`</root></mxGraphModel></diagram>`
	}
	composite := func(id, value string) string {
		return `<mxCell id="` + id + `" value="` + value + `" style="ellipse;whiteSpace=wrap;html=1;strokeWidth=2;" parent="1" vertex="1"><mxGeometry x="0" y="0" width="120" height="80" as="geometry"/></mxCell>`
	}
	atomic := func(id, value string) string {
		return `<mxCell id="` + id + `" value="` + value + `" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="0" y="0" width="120" height="80" as="geometry"/></mxCell>`
	}

	testCases := map[string]struct {
		XML string

		Plan pfd.RenumberPlan
		Want []string
	}{
		"未採番のラベル名のページは採番後の `ID: 説明` になる": {
			XML: `<mxfile>` +
				page("g0", "P0", composite("c1", "実装する")) +
				page("g1", "実装する", atomic("c2", "コーディングする")) +
				`</mxfile>`,
			Want: []string{"P0", "P2: 実装する"},
		},
		"素の ID のページには説明が付く": {
			XML: `<mxfile>` +
				page("g0", "P0", composite("c1", "P2: 実装する")) +
				page("g1", "P2", atomic("c2", "P1: コーディングする")) +
				`</mxfile>`,
			Want: []string{"P0", "P2: 実装する"},
		},
		"コンテキスト図のページ名は変わらない": {
			XML: `<mxfile>` +
				page("g0", "ページ-1", composite("c1", "P2: 実装する")) +
				page("g1", "P2: 実装する", atomic("c2", "P1: コーディングする")) +
				`</mxfile>`,
			Want: []string{"ページ-1", "P2: 実装する"},
		},

		"計画にあるラベル名のページは追従する": {
			XML:  `<mxfile>` + page("g1", "実装する", atomic("c2", "コーディングする")) + `</mxfile>`,
			Plan: pfd.RenumberPlan{"(実装する)": &pfd.Node{ID: "P7", Description: "実装する", Type: pfd.NodeTypeCompositeProcess}},
			Want: []string{"P7: 実装する"},
		},
		"計画に載っていないページ名は変わらない": {
			XML:  `<mxfile>` + page("g1", "P2", atomic("c2", "コーディングする")) + `</mxfile>`,
			Plan: pfd.RenumberPlan{"(コーディングする)": &pfd.Node{ID: "P1", Description: "コーディングする", Type: pfd.NodeTypeAtomicProcess}},
			Want: []string{"P2"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			renumber := Renumber
			if testCase.Plan != nil {
				renumber = RenumberByPlan(testCase.Plan)
			}
			nodes, err := renumber(strings.NewReader(testCase.XML), logger)
			if err != nil {
				t.Fatal(err)
			}

			diagrams, err := ParseNodes(nodes, logger)
			if err != nil {
				t.Fatal(err)
			}

			got := make([]string, 0, len(diagrams))
			for _, diagram := range diagrams {
				got = append(got, diagram.Name)
			}
			if diff := cmp.Diff(testCase.Want, got); diff != "" {
				t.Errorf("page names mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRenumberMarksDuplicates(t *testing.T) {
	testCases := map[string]struct {
		XML  string
		Want map[string]string
	}{
		"複製表示に採番するとマークが付く": {
			XML: `<mxfile><diagram id="g0" name="P0"><mxGraphModel><root>` +
				`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
				`<mxCell id="p1" value="作る" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="0" y="0" width="120" height="80" as="geometry"/></mxCell>` +
				`<mxCell id="d1a" value="出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="0" width="120" height="80" as="geometry"/></mxCell>` +
				`<mxCell id="d1b" value="出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="200" width="120" height="80" as="geometry"/></mxCell>` +
				`<mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d1b" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>` +
				`</root></mxGraphModel></diagram></mxfile>`,
			Want: map[string]string{"d1a": "D1＊: 出力", "d1b": "D1: 出力"},
		},
		"既にあるマークは採番後も保たれる": {
			XML: `<mxfile><diagram id="g0" name="P0"><mxGraphModel><root>` +
				`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
				`<mxCell id="p1" value="P9: 作る" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="0" y="0" width="120" height="80" as="geometry"/></mxCell>` +
				`<mxCell id="d1a" value="D9＊: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="0" width="120" height="80" as="geometry"/></mxCell>` +
				`<mxCell id="d1b" value="D9: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="200" width="120" height="80" as="geometry"/></mxCell>` +
				`<mxCell id="e1" style="edgeStyle=none;html=1;" parent="1" source="p1" target="d1b" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>` +
				`</root></mxGraphModel></diagram></mxfile>`,
			Want: map[string]string{"d1a": "D9＊: 出力", "d1b": "D9: 出力"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodes, err := Renumber(strings.NewReader(tc.XML), logger)
			if err != nil {
				t.Fatalf("Renumber: %v", err)
			}
			dom, ok := CollectDiagramDOMs(nodes)["g0"]
			if !ok {
				t.Fatal("diagram g0 not found")
			}
			for cellID, want := range tc.Want {
				got, _ := dom.CellsByID[CellID(cellID)].GetAttr("value", "")
				if got != want {
					t.Errorf("cell %s value = %q, want %q", cellID, got, want)
				}
			}
		})
	}
}

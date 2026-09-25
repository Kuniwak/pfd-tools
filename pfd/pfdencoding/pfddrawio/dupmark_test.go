package pfddrawio_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

func TestSplitDuplicateMark(t *testing.T) {
	testCases := map[string]struct {
		In       pfd.NodeID
		WantBase pfd.NodeID
		WantMark string
	}{
		"全角マークを剥がす":                    {In: "D10＊", WantBase: "D10", WantMark: "＊"},
		"半角マークを剥がす":                    {In: "D10*", WantBase: "D10", WantMark: "*"},
		"マークなしはそのまま":                   {In: "D10", WantBase: "D10", WantMark: ""},
		"未採番のラベルは剥がさない":                {In: "ほげ＊", WantBase: "ほげ＊", WantMark: ""},
		"プロセス ID は剥がさない":               {In: "P3＊", WantBase: "P3＊", WantMark: ""},
		"重ね付けはまとめて剥がす":                 {In: "D10＊＊", WantBase: "D10", WantMark: "＊"},
		"全角と半角の混在もまとめて剥がす":             {In: "D10*＊", WantBase: "D10", WantMark: "*"},
		"空文字はそのまま":                     {In: "", WantBase: "", WantMark: ""},
		"マークだけのラベルは剥がさない":              {In: "＊", WantBase: "＊", WantMark: ""},
		"成果物 ID でない接頭辞は剥がさない":          {In: "X10＊", WantBase: "X10＊", WantMark: ""},
		"マークを剥がしても成果物 ID にならないなら剥がさない": {In: "D＊", WantBase: "D＊", WantMark: ""},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotBase, gotMark := pfddrawio.SplitDuplicateMark(tc.In)
			if gotBase != tc.WantBase || gotMark != tc.WantMark {
				t.Errorf("SplitDuplicateMark(%q) = (%q, %q), want (%q, %q)", tc.In, gotBase, gotMark, tc.WantBase, tc.WantMark)
			}
		})
	}
}

func TestRewriteDuplicateMark(t *testing.T) {
	testCases := map[string]struct {
		In     pfddrawio.ValueHTML
		Dup    bool
		Want   pfddrawio.ValueHTML
		WantOK bool
	}{
		"平文にマークを付ける": {
			In: "D10: ほげ", Dup: true, Want: "D10＊: ほげ", WantOK: true,
		},
		"平文からマークを外す": {
			In: "D10＊: ほげ", Dup: false, Want: "D10: ほげ", WantOK: true,
		},
		"説明のないラベルにマークを付ける": {
			In: "D10", Dup: true, Want: "D10＊", WantOK: true,
		},
		"既にマークがあれば文字を変えない": {
			In: "D10*: ほげ", Dup: true, Want: "D10*: ほげ", WantOK: true,
		},
		"重ね付けは 1 個へ正規化する": {
			In: "D10＊＊: ほげ", Dup: true, Want: "D10＊: ほげ", WantOK: true,
		},
		"マークのない正本はそのまま": {
			In: "D10: ほげ", Dup: false, Want: "D10: ほげ", WantOK: true,
		},
		"HTML のタグは壊さない": {
			In: "<b>D10</b>: ほげ", Dup: true, Want: "<b>D10＊</b>: ほげ", WantOK: true,
		},
		"属性のコロンを ID 部と間違えない": {
			In: `<span style="color:red">D10</span>: ほげ`, Dup: true, Want: `<span style="color:red">D10＊</span>: ほげ`, WantOK: true,
		},
		"改行タグを挟んだラベル": {
			In: "D10:<br>ほげ", Dup: true, Want: "D10＊:<br>ほげ", WantOK: true,
		},
		"ID 部の前後の空白は保つ": {
			In: "D10 : ほげ", Dup: true, Want: "D10＊ : ほげ", WantOK: true,
		},
		"ID 部がタグで分断されていてもマークを付けられる": {
			In: "<b>D</b>10: ほげ", Dup: true, Want: "<b>D</b>10＊: ほげ", WantOK: true,
		},
		"タグで分断された ID 部からマークを外せる": {
			In: "<b>D</b>10＊: ほげ", Dup: false, Want: "<b>D</b>10: ほげ", WantOK: true,
		},
		"未採番のラベルは対象外": {
			In: "ほげ", Dup: true, Want: "ほげ", WantOK: false,
		},
		"プロセスのラベルは対象外": {
			In: "P3: ほげする", Dup: true, Want: "P3: ほげする", WantOK: false,
		},
		"空のラベルは対象外": {
			In: "", Dup: true, Want: "", WantOK: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, ok := pfddrawio.RewriteDuplicateMark(tc.In, tc.Dup)
			if got != tc.Want || ok != tc.WantOK {
				t.Errorf("RewriteDuplicateMark(%q, %v) = (%q, %v), want (%q, %v)", tc.In, tc.Dup, got, ok, tc.Want, tc.WantOK)
			}

			again, _ := pfddrawio.RewriteDuplicateMark(got, tc.Dup)
			if again != got {
				t.Errorf("RewriteDuplicateMark は冪等でない: %q -> %q", got, again)
			}
		})
	}
}

func dupVertex(id pfddrawio.CellID, nodeID pfd.NodeID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, NodeID: nodeID, Kind: pfddrawio.VertexDeliverable}
}

func dupProcess(id pfddrawio.CellID, nodeID pfd.NodeID) pfddrawio.LayoutVertex {
	return pfddrawio.LayoutVertex{ID: id, NodeID: nodeID, Kind: pfddrawio.VertexProcess}
}

func TestDuplicateMarks(t *testing.T) {
	testCases := map[string]struct {
		Vertices []pfddrawio.LayoutVertex
		Edges    []pfddrawio.LayoutEdge
		Want     map[pfddrawio.CellID]bool
	}{
		"単独のセルは無印": {
			Vertices: []pfddrawio.LayoutVertex{dupVertex("d1", "D1")},
			Want:     map[pfddrawio.CellID]bool{"d1": false},
		},
		"生産辺を持つセルが正本": {
			Vertices: []pfddrawio.LayoutVertex{
				dupProcess("p1", "P1"),
				dupVertex("d1", "D1"),
				dupVertex("d2", "D1"),
			},
			Edges: []pfddrawio.LayoutEdge{{ID: "e1", Source: "p1", Target: "d2"}},
			Want:  map[pfddrawio.CellID]bool{"d1": true, "d2": false},
		},
		"生産辺がなければ最小 CellID が正本": {
			Vertices: []pfddrawio.LayoutVertex{
				dupVertex("d1", "D1"),
				dupVertex("d1-dup1", "D1"),
				dupVertex("d1-dup2", "D1"),
			},
			Want: map[pfddrawio.CellID]bool{"d1": false, "d1-dup1": true, "d1-dup2": true},
		},
		"破線の入辺は生産辺として数えない": {
			Vertices: []pfddrawio.LayoutVertex{
				dupProcess("p1", "P1"),
				dupVertex("d1", "D1"),
				dupVertex("d2", "D1"),
			},
			Edges: []pfddrawio.LayoutEdge{{ID: "e1", Source: "p1", Target: "d2", IsFeedback: true}},
			Want:  map[pfddrawio.CellID]bool{"d1": false, "d2": true},
		},
		"別の成果物は別グループ": {
			Vertices: []pfddrawio.LayoutVertex{
				dupVertex("d1", "D1"),
				dupVertex("d2", "D2"),
			},
			Want: map[pfddrawio.CellID]bool{"d1": false, "d2": false},
		},
		"プロセスは同名でも判定しない": {
			Vertices: []pfddrawio.LayoutVertex{
				dupProcess("p1", "P1"),
				dupProcess("p2", "P1"),
			},
			Want: map[pfddrawio.CellID]bool{},
		},
		"ID を読めないセルは判定しない": {
			Vertices: []pfddrawio.LayoutVertex{
				dupVertex("d1", ""),
				dupVertex("d2", ""),
			},
			Want: map[pfddrawio.CellID]bool{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.DuplicateMarks(tc.Vertices, tc.Edges)
			if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Errorf("DuplicateMarks mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func labelsOf(t *testing.T, nodes []*xmldom.Node, diagramID pfddrawio.DiagramID) map[pfddrawio.CellID]string {
	t.Helper()
	dom, ok := pfddrawio.CollectDiagramDOMs(nodes)[diagramID]
	if !ok {
		t.Fatalf("diagram %q not found", diagramID)
	}
	labels := make(map[pfddrawio.CellID]string)
	for id, cell := range dom.CellsByID {
		if v, _ := cell.GetAttr("vertex", ""); v != "1" {
			continue
		}
		value, _ := cell.GetAttr("value", "")
		labels[id] = value
	}
	return labels
}

func TestMarkDuplicates(t *testing.T) {

	dupPage := diagramPage("g0", "P0",
		rectCell("d0", "D9: 入力", 0, 0)+
			ellipseCell("p1", "P1: 作る", 200, 0)+
			rectCell("d1a", "D1: 出力", 400, 0)+
			rectCell("d1b", "D1: 出力", 400, 200)+
			rectCell("d1c", "D1: 出力", 400, 400)+
			edgeCell("e0", "d0", "p1", false)+
			edgeCell("e1", "p1", "d1b", false))

	otherPage := diagramPage("g1", "P7", rectCell("d5", "D5: 別ページ", 0, 0))

	testCases := map[string]struct {
		XML       string
		OnlyPages []string
		Want      map[pfddrawio.CellID]string
		Diagram   pfddrawio.DiagramID
	}{
		"生産辺を持つセル以外にマークが付く": {
			XML:     `<mxfile>` + dupPage + `</mxfile>`,
			Diagram: "g0",
			Want: map[pfddrawio.CellID]string{
				"d0":  "D9: 入力",
				"p1":  "P1: 作る",
				"d1a": "D1＊: 出力",
				"d1b": "D1: 出力",
				"d1c": "D1＊: 出力",
			},
		},
		"正本が消えていれば最小 CellID のマークが外れる": {
			XML: `<mxfile>` + diagramPage("g0", "P0",
				rectCell("d1a", "D1＊: 出力", 400, 0)+
					rectCell("d1b", "D1＊: 出力", 400, 200)) + `</mxfile>`,
			Diagram: "g0",
			Want: map[pfddrawio.CellID]string{
				"d1a": "D1: 出力",
				"d1b": "D1＊: 出力",
			},
		},
		"複製でなくなったセルのマークは外れる": {
			XML:     `<mxfile>` + diagramPage("g0", "P0", rectCell("d1a", "D1＊: 出力", 0, 0)) + `</mxfile>`,
			Diagram: "g0",
			Want:    map[pfddrawio.CellID]string{"d1a": "D1: 出力"},
		},
		"未採番のラベルはマークしない": {
			XML: `<mxfile>` + diagramPage("g0", "P0",
				rectCell("d1a", "出力", 0, 0)+
					rectCell("d1b", "出力", 0, 200)) + `</mxfile>`,
			Diagram: "g0",
			Want: map[pfddrawio.CellID]string{
				"d1a": "出力",
				"d1b": "出力",
			},
		},
		"-only-page の対象外のページは触らない": {
			XML:       `<mxfile>` + dupPage + otherPage + `</mxfile>`,
			OnlyPages: []string{"P7"},
			Diagram:   "g0",
			Want: map[pfddrawio.CellID]string{
				"d0":  "D9: 入力",
				"p1":  "P1: 作る",
				"d1a": "D1: 出力",
				"d1b": "D1: 出力",
				"d1c": "D1: 出力",
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slogtest.NewTestHandler(t))
			nodes, err := pfddrawio.MarkDuplicates(strings.NewReader(tc.XML), pfddrawio.DupMarkOptions{OnlyPages: tc.OnlyPages}, logger)
			if err != nil {
				t.Fatalf("MarkDuplicates: %v", err)
			}
			if diff := cmp.Diff(tc.Want, labelsOf(t, nodes, tc.Diagram)); diff != "" {
				t.Errorf("labels mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarkDuplicatesIsIdempotent(t *testing.T) {
	xmlText := `<mxfile>` + diagramPage("g0", "P0",
		ellipseCell("p1", "P1: 作る", 200, 0)+
			rectCell("d1a", "D1: 出力", 400, 0)+
			rectCell("d1b", "D1: 出力", 400, 200)+
			edgeCell("e1", "p1", "d1b", false)) + `</mxfile>`

	logger := slog.New(slogtest.NewTestHandler(t))
	once, err := pfddrawio.MarkDuplicates(strings.NewReader(xmlText), pfddrawio.DupMarkOptions{}, logger)
	if err != nil {
		t.Fatalf("MarkDuplicates (1st): %v", err)
	}
	onceBytes, err := xmldom.Marshal(once)
	if err != nil {
		t.Fatalf("Marshal (1st): %v", err)
	}

	twice, err := pfddrawio.MarkDuplicates(bytes.NewReader(onceBytes), pfddrawio.DupMarkOptions{}, logger)
	if err != nil {
		t.Fatalf("MarkDuplicates (2nd): %v", err)
	}
	twiceBytes, err := xmldom.Marshal(twice)
	if err != nil {
		t.Fatalf("Marshal (2nd): %v", err)
	}

	if diff := cmp.Diff(string(onceBytes), string(twiceBytes)); diff != "" {
		t.Errorf("MarkDuplicates は冪等でない (-1st +2nd):\n%s", diff)
	}
}

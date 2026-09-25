package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const messyPFD = `<mxfile><diagram id="d0" name="P0"><mxGraphModel><root>` +
	`<mxCell id="0"/><mxCell id="1" parent="0"/>` +
	`<mxCell id="2" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="900" y="90" width="120" height="80" as="geometry"/></mxCell>` +
	`<mxCell id="3" value="P1: 加工" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="100" y="400" width="120" height="80" as="geometry"/></mxCell>` +
	`<mxCell id="5" value="D2: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="500" y="250" width="120" height="80" as="geometry"/></mxCell>` +
	`<mxCell id="4" style="edgeStyle=none;html=1;" parent="1" source="2" target="3" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>` +
	`<mxCell id="6" style="edgeStyle=none;html=1;" parent="1" source="3" target="5" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>` +
	`</root></mxGraphModel></diagram></mxfile>`

func vertexX(t *testing.T, xmlBytes []byte) map[pfddrawio.CellID]float64 {
	t.Helper()
	nodes, err := xmldom.ParseXML(bytes.NewReader(xmlBytes))
	if err != nil {
		t.Fatalf("ParseXML: %v", err)
	}
	xs := make(map[pfddrawio.CellID]float64)
	for _, dom := range pfddrawio.CollectDiagramDOMs(nodes) {
		for id, cell := range dom.CellsByID {
			if v, _ := cell.GetAttr("vertex", ""); v != "1" {
				continue
			}
			if r, ok := pfddrawio.GeometryRect(cell); ok {
				xs[id] = r.X
			}
		}
	}
	return xs
}

func TestMainCommandByArgs_StdinSortsLeftToRight(t *testing.T) {
	spy := cli.SpyProcInout(messyPFD)
	exit := MainCommandByArgs(nil, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}

	x := vertexX(t, spy.Stdout.Bytes())
	if !(x["2"] < x["3"] && x["3"] < x["5"]) {
		t.Errorf("want left-to-right by rank: x[D1]=%v < x[P1]=%v < x[D2]=%v", x["2"], x["3"], x["5"])
	}
}

func TestMainCommandByArgs_Inplace(t *testing.T) {
	tmpPath := filepath.Join(t.TempDir(), "pfd.drawio")
	if err := os.WriteFile(tmpPath, []byte(messyPFD), 0644); err != nil {
		t.Fatal(err)
	}
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"-inplace", tmpPath}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	if spy.Stdout.Len() != 0 {
		t.Errorf("stdout should be empty in -inplace mode, got %q", spy.Stdout.String())
	}
	out, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(out, []byte(messyPFD)) {
		t.Errorf("file should be rewritten in place")
	}
}

func TestMainCommandByArgs_ShortHelp(t *testing.T) {
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"--short-help"}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0", exit)
	}
	if got := spy.Stdout.String(); got != ShortHelp+"\n" {
		t.Errorf("stdout = %q, want %q", got, ShortHelp+"\n")
	}
}

func TestParseOptionsRejectsNegativeGaps(t *testing.T) {

	testCases := map[string]struct {
		Args    []string
		WantErr bool
	}{
		"zero gaps are allowed":     {Args: []string{"-hgap", "0", "-vgap", "0"}, WantErr: false},
		"negative hgap is an error": {Args: []string{"-hgap", "-1"}, WantErr: true},
		"negative vgap is an error": {Args: []string{"-vgap", "-100"}, WantErr: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(messyPFD)
			_, err := ParseOptions(tc.Args, spy.NewProcInout())
			if tc.WantErr && err == nil {
				t.Errorf("ParseOptions(%v) = nil error, want an error", tc.Args)
			}
			if !tc.WantErr && err != nil {
				t.Errorf("ParseOptions(%v): %v", tc.Args, err)
			}
		})
	}
}

func TestParseOptionsDupSpans(t *testing.T) {
	defaults := pfddrawio.DefaultSortOptions()

	testCases := map[string]struct {
		Args         []string
		WantRankSpan int
		WantRowSpan  int
	}{
		"defaults come from DefaultSortOptions": {
			Args: nil, WantRankSpan: defaults.DupRankSpan, WantRowSpan: defaults.DupRowSpan,
		},
		"row span is overridable": {
			Args: []string{"-dup-row-span", "12"}, WantRankSpan: defaults.DupRankSpan, WantRowSpan: 12,
		},
		"rank span is overridable": {
			Args: []string{"-dup-rank-span", "3"}, WantRankSpan: 3, WantRowSpan: defaults.DupRowSpan,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(messyPFD)
			opts, err := ParseOptions(tc.Args, spy.NewProcInout())
			if err != nil {
				t.Fatalf("ParseOptions: %v", err)
			}
			if opts.DupRankSpan != tc.WantRankSpan {
				t.Errorf("DupRankSpan = %d, want %d", opts.DupRankSpan, tc.WantRankSpan)
			}
			if opts.DupRowSpan != tc.WantRowSpan {
				t.Errorf("DupRowSpan = %d, want %d", opts.DupRowSpan, tc.WantRowSpan)
			}
		})
	}
}

func TestMainCommandByArgs_OnlyPageLeavesOtherPagesUntouched(t *testing.T) {

	spy := cli.SpyProcInout(messyPFD)
	exit := MainCommandByArgs([]string{"-only-page", "P999"}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	before := vertexX(t, []byte(messyPFD))
	after := vertexX(t, spy.Stdout.Bytes())
	for id, x := range before {
		if after[id] != x {
			t.Errorf("page not selected by -only-page should be untouched: cell %s x %v -> %v", id, x, after[id])
		}
	}
}

func TestMainCommandByArgs_TooManyArgs(t *testing.T) {
	spy := cli.SpyProcInout()
	exit := MainCommandByArgs([]string{"a.drawio", "b.drawio"}, spy.NewProcInout())
	if exit == 0 {
		t.Errorf("want non-zero exit for too many arguments")
	}
}

func TestParseOptionsPosRestriction(t *testing.T) {
	testCases := map[string]struct {
		Args    []string
		Want    pfddrawio.PosRestriction
		WantErr bool
	}{
		"default is lock":  {nil, pfddrawio.PosLock, false},
		"free-v":           {[]string{"-pos-restriction", "free-v"}, pfddrawio.PosFreeV, false},
		"free-h":           {[]string{"-pos-restriction", "free-h"}, pfddrawio.PosFreeH, false},
		"free":             {[]string{"-pos-restriction", "free"}, pfddrawio.PosFree, false},
		"invalid is error": {[]string{"-pos-restriction", "diagonal"}, 0, true},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(messyPFD)
			opts, err := ParseOptions(tc.Args, spy.NewProcInout())
			if tc.WantErr {
				if err == nil {
					t.Errorf("ParseOptions(%v) = nil error, want error", tc.Args)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseOptions(%v): %v", tc.Args, err)
			}
			if opts.PosRestriction != tc.Want {
				t.Errorf("PosRestriction = %v, want %v", opts.PosRestriction, tc.Want)
			}
		})
	}
}

func TestParseOptionsOnlyNode(t *testing.T) {
	spy := cli.SpyProcInout(messyPFD)
	opts, err := ParseOptions([]string{"-only-node", "D1, P1 ,D2"}, spy.NewProcInout())
	if err != nil {
		t.Fatalf("ParseOptions: %v", err)
	}
	want := []pfd.NodeID{"D1", "P1", "D2"}
	if len(opts.OnlyNodes) != len(want) {
		t.Fatalf("OnlyNodes = %v, want %v", opts.OnlyNodes, want)
	}
	for i := range want {
		if opts.OnlyNodes[i] != want[i] {
			t.Errorf("OnlyNodes[%d] = %q, want %q", i, opts.OnlyNodes[i], want[i])
		}
	}
}

func TestMainCommandByArgs_OnlyNodeLeavesOtherNodesUntouched(t *testing.T) {

	spy := cli.SpyProcInout(messyPFD)
	exit := MainCommandByArgs([]string{"-only-node", "P1"}, spy.NewProcInout())
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr: %s", exit, spy.Stderr.String())
	}
	before := vertexX(t, []byte(messyPFD))
	after := vertexX(t, spy.Stdout.Bytes())
	if after["2"] != before["2"] || after["5"] != before["5"] {
		t.Errorf("非選択ノードは不動のはず: D1 %v->%v, D2 %v->%v", before["2"], after["2"], before["5"], after["5"])
	}
	if after["3"] == before["3"] {
		t.Errorf("選択ノード P1 は動くはず: %v", after["3"])
	}
}

func TestParseOptionsBreakCycles(t *testing.T) {
	testCases := map[string]struct {
		Args []string
		Want bool
	}{
		"off by default":        {Args: nil, Want: false},
		"turned on by the flag": {Args: []string{"-break-cycles"}, Want: true},
		"explicitly turned off": {Args: []string{"-break-cycles=false"}, Want: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			spy := cli.SpyProcInout(messyPFD)
			opts, err := ParseOptions(tc.Args, spy.NewProcInout())
			if err != nil {
				t.Fatalf("ParseOptions: %v", err)
			}
			if opts.BreakCycles != tc.Want {
				t.Errorf("BreakCycles = %v, want %v", opts.BreakCycles, tc.Want)
			}
		})
	}
}

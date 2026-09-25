package pfddrawio_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/sugar"
	"github.com/Kuniwak/pfd-tools/xmldom"
	"github.com/google/go-cmp/cmp"
)

func TestEstimateBoxesByProcess(t *testing.T) {
	t.Run("filled boxes map to the atomic processes above them", func(t *testing.T) {
		src := sugar.Must(os.ReadFile("../../../testdata/callout/pfd.drawio"))
		nodes := sugar.Must(xmldom.ParseXML(bytes.NewReader(src)))
		logger := slog.New(slogtest.NewTestHandler(t))

		got := sugar.Must(pfddrawio.EstimateBoxesByProcess(nodes, logger))

		want := map[pfd.NodeID][]pfddrawio.EstimateBox{
			"P1": {{DiagramID: "wRU_aafd9vpDkhm-03GV", CellID: "18", Value: "楽観: 2d 悲観: 3d", Rect: geom.Rect{X: 485, Y: 320, Width: 110, Height: 30}}},
			"P2": {{DiagramID: "wRU_aafd9vpDkhm-03GV", CellID: "20", Value: "楽観: 1d 悲観: 2d", Rect: geom.Rect{X: 485, Y: 440, Width: 110, Height: 30}}},
			"P3": {{DiagramID: "wRU_aafd9vpDkhm-03GV", CellID: "21", Value: "楽観: 1d 悲観: 5d", Rect: geom.Rect{X: 805, Y: 320, Width: 110, Height: 30}}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("EstimateBoxesByProcess mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("multi-page composite: box on the detail page maps to the atomic process", func(t *testing.T) {
		logger := slog.New(slogtest.NewTestHandler(t))
		nodes := callout(t, []byte(compositeMultiPageXML))

		got := sugar.Must(pfddrawio.EstimateBoxesByProcess(nodes, logger))

		want := map[pfd.NodeID][]pfddrawio.EstimateBox{
			"P2": {{DiagramID: "pg1", CellID: "15", Value: pfddrawio.EstimateBoxPlaceholder, Rect: geom.Rect{X: 205, Y: 120, Width: 110, Height: 30}}},
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("EstimateBoxesByProcess mismatch (-want +got):\n%s", diff)
		}
	})
}

func twoProcessChainXML(extraCells string) []byte {
	return []byte(`<mxfile host="test">
    <diagram id="pg0" name="P0">
        <mxGraphModel>
            <root>
                <mxCell id="0"/>
                <mxCell id="1" parent="0"/>
                <mxCell id="2" value="D1: 入力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="40" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="3" value="P1: 原子1" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="200" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="4" value="D2: 中間" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="360" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="5" value="P2: 原子2" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="520" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="6" value="D3: 出力" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1"><mxGeometry x="680" y="40" width="120" height="80" as="geometry"/></mxCell>
                <mxCell id="7" style="edgeStyle=none;html=1;" parent="1" source="2" target="3" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>
                <mxCell id="8" style="edgeStyle=none;html=1;" parent="1" source="3" target="4" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>
                <mxCell id="9" style="edgeStyle=none;html=1;" parent="1" source="4" target="5" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>
                <mxCell id="10" style="edgeStyle=none;html=1;" parent="1" source="5" target="6" edge="1"><mxGeometry relative="1" as="geometry"/></mxCell>
                ` + extraCells + `
            </root>
        </mxGraphModel>
    </diagram>
</mxfile>`)
}

func estimateBoxCellXML(id, value string, x, y int) string {
	return textBoxCellXML(id, "1", value, x, y)
}

func textBoxCellXML(id, parent, value string, x, y int) string {
	return fmt.Sprintf(`<mxCell id=%q value=%q style=%q parent=%q vertex="1"><mxGeometry x="%d" y="%d" width="110" height="30" as="geometry"/></mxCell>`,
		id, value, pfddrawio.EstimateBoxStyle, parent, x, y)
}

func estimateTable(t *testing.T, xmlBytes []byte) *pfd.AtomicProcessTable {
	t.Helper()
	logger := slog.New(slogtest.NewTestHandler(t))
	nodes := sugar.Must(xmldom.ParseXML(bytes.NewReader(xmlBytes)))
	return sugar.Must(pfddrawio.EstimateTable(nodes, logger))
}

func TestEstimateTable(t *testing.T) {
	testCases := map[string]struct {
		Input []byte
		Want  *pfd.AtomicProcessTable
	}{
		"two points filled (testdata/callout)": {
			Input: sugar.Must(os.ReadFile("../../../testdata/callout/pfd.drawio")),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "プロセス1", ExtraCells: []string{"2", "3"}},
					{ID: "P2", Description: "プロセス2", ExtraCells: []string{"1", "2"}},
					{ID: "P3", Description: "プロセス3", ExtraCells: []string{"1", "5"}},
				},
			},
		},
		"placeholder (unfilled) yields empty cells": {
			Input: sugar.Must(xmldom.Marshal(callout(t, []byte(compositeMultiPageXML)))),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P2", Description: "原子", ExtraCells: []string{"", ""}},
				},
			},
		},
		"three-point and two-point mixed (N-point generic)": {
			Input: twoProcessChainXML(
				estimateBoxCellXML("11", "楽観: 2d 最頻: 3d 悲観: 5d", 205, 120) +
					estimateBoxCellXML("12", "楽観: 1d 悲観: 2d", 525, 120),
			),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "最頻作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "原子1", ExtraCells: []string{"2", "3", "5"}},
					{ID: "P2", Description: "原子2", ExtraCells: []string{"1", "", "2"}},
				},
			},
		},
		"process without an estimate box yields an empty row": {
			Input: twoProcessChainXML(
				estimateBoxCellXML("11", "楽観: 2d 悲観: 3d", 205, 120),
			),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "原子1", ExtraCells: []string{"2", "3"}},
					{ID: "P2", Description: "原子2", ExtraCells: []string{"", ""}},
				},
			},
		},
		"comment layer memo below a process is ignored": {
			Input: twoProcessChainXML(
				`<mxCell id="90" value="コメント" parent="0"/>` +
					textBoxCellXML("91", "90", "楽観: 9d 悲観: 9d", 205, 150) +
					estimateBoxCellXML("11", "楽観: 2d 悲観: 3d", 205, 120),
			),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "原子1", ExtraCells: []string{"2", "3"}},
					{ID: "P2", Description: "原子2", ExtraCells: []string{"", ""}},
				},
			},
		},
		"non-vertex cell with estimate-like text is ignored": {
			Input: twoProcessChainXML(
				estimateBoxCellXML("11", "楽観: 2d 悲観: 3d", 205, 120) +
					fmt.Sprintf(`<mxCell id="92" value="楽観: 9d 悲観: 9d" style=%q parent="1" edge="1"><mxGeometry x="205" y="150" width="110" height="30" as="geometry"/></mxCell>`, pfddrawio.EstimateBoxStyle),
			),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "原子1", ExtraCells: []string{"2", "3"}},
					{ID: "P2", Description: "原子2", ExtraCells: []string{"", ""}},
				},
			},
		},
		"conflicting values across boxes yield an empty cell for that label only": {
			Input: twoProcessChainXML(
				estimateBoxCellXML("11", "楽観: 2d 悲観: 3d", 205, 120) +
					estimateBoxCellXML("12", "楽観: 4d 悲観: 3d", 205, 150),
			),
			Want: &pfd.AtomicProcessTable{
				ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
				Rows: []*pfd.AtomicProcessRow{
					{ID: "P1", Description: "原子1", ExtraCells: []string{"", "3"}},
					{ID: "P2", Description: "原子2", ExtraCells: []string{"", ""}},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := estimateTable(t, tc.Input)
			if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Errorf("EstimateTable mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEstimateTableWarnsOrphanBox(t *testing.T) {
	input := twoProcessChainXML(
		estimateBoxCellXML("11", "楽観: 2d 悲観: 3d", 205, 120) +
			estimateBoxCellXML("93", "楽観: 7d 悲観: 8d", 205, 400),
	)
	var buf bytes.Buffer
	logger := slog.New(slograw.NewHandler(&buf, slog.LevelWarn))
	nodes := sugar.Must(xmldom.ParseXML(bytes.NewReader(input)))

	got := sugar.Must(pfddrawio.EstimateTable(nodes, logger))

	want := &pfd.AtomicProcessTable{
		ExtraHeaders: []string{"楽観作業量", "悲観作業量"},
		Rows: []*pfd.AtomicProcessRow{
			{ID: "P1", Description: "原子1", ExtraCells: []string{"2", "3"}},
			{ID: "P2", Description: "原子2", ExtraCells: []string{"", ""}},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("EstimateTable mismatch (-want +got):\n%s", diff)
	}
	if !strings.Contains(buf.String(), "not below any atomic process") {
		t.Errorf("expected a warning about the orphan estimate box, got: %s", buf.String())
	}
}

func TestParseEstimateBoxValue(t *testing.T) {
	testCases := map[string]struct {
		Input   string
		Want    []pfddrawio.EstimatePoint
		WantErr bool
	}{
		"two points (testdata/callout format)": {
			Input: "楽観: 2d 悲観: 3d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "2"},
				{Label: "悲観", Value: "3"},
			},
		},
		"three points with an arbitrary label (N-point generic)": {
			Input: "楽観: 2d 最頻: 3d 悲観: 5d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "2"},
				{Label: "最頻", Value: "3"},
				{Label: "悲観", Value: "5"},
			},
		},
		"placeholder (unfilled)": {
			Input: pfddrawio.EstimateBoxPlaceholder,
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: ""},
				{Label: "悲観", Value: ""},
			},
		},
		"partially filled": {
			Input: "楽観: 2d 悲観: d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "2"},
				{Label: "悲観", Value: ""},
			},
		},
		"full-width colon and float value": {
			Input: "楽観：0.5d 悲観：1.5d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "0.5"},
				{Label: "悲観", Value: "1.5"},
			},
		},
		"html break between points (drawio html=1 value)": {
			Input: "楽観: 2d<br>悲観: 3d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "2"},
				{Label: "悲観", Value: "3"},
			},
		},
		"html entity nbsp between points": {
			Input: "楽観: 2d&nbsp;悲観: 3d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "2"},
				{Label: "悲観", Value: "3"},
			},
		},
		"ideographic space between points": {
			Input: "楽観: 2d　悲観: 3d",
			Want: []pfddrawio.EstimatePoint{
				{Label: "楽観", Value: "2"},
				{Label: "悲観", Value: "3"},
			},
		},
		"duplicated label": {
			Input:   "楽観: 2d 楽観: 3d",
			WantErr: true,
		},
		"negative value": {
			Input:   "楽観: -2d 悲観: 3d",
			WantErr: true,
		},
		"missing unit d": {
			Input:   "楽観: 2 悲観: 3d",
			WantErr: true,
		},
		"empty": {
			Input:   "",
			WantErr: true,
		},
		"no estimate points": {
			Input:   "ただのメモ",
			WantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := pfddrawio.ParseEstimateBoxValue(tc.Input)
			if tc.WantErr {
				if err == nil {
					t.Fatalf("ParseEstimateBoxValue(%q) = %v, want error", tc.Input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseEstimateBoxValue(%q): %v", tc.Input, err)
			}
			if diff := cmp.Diff(tc.Want, got); diff != "" {
				t.Errorf("ParseEstimateBoxValue(%q) mismatch (-want +got):\n%s", tc.Input, diff)
			}
		})
	}
}

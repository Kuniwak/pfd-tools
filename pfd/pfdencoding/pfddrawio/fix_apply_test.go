package pfddrawio_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/Kuniwak/pfd-tools/slogtest"
	"github.com/Kuniwak/pfd-tools/xmldom"
)

const fixSampleXML = `<mxfile host="test">
  <diagram id="d1" name="P0">
    <mxGraphModel>
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
        <mxCell id="4" style="edgeStyle=orthogonalEdgeStyle;html=1;" parent="1" source="2" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="440" y="280" as="sourcePoint"/>
            <mxPoint x="480" y="280" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
        <mxCell id="2" value="D1" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="320" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="6" style="endArrow=classic;html=1;" parent="1" target="5" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="600" y="280" as="sourcePoint"/>
            <mxPoint x="640" y="280" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
        <mxCell id="3" value="P1" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="480" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="7" style="dashed=1;html=1;" parent="1" edge="1">
          <mxGeometry relative="1" as="geometry">
            <Array as="points">
              <mxPoint x="700" y="200"/>
              <mxPoint x="540" y="200"/>
            </Array>
            <mxPoint x="700" y="240" as="sourcePoint"/>
            <mxPoint x="540" y="240" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
        <mxCell id="5" value="D2" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="640" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="99" style="html=1;" parent="1" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="9999" y="9999" as="sourcePoint"/>
            <mxPoint x="8888" y="8888" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`

const selfLoopXML = `<mxfile host="test">
  <diagram id="d1" name="P0">
    <mxGraphModel>
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
        <mxCell id="2" value="D1" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="320" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="4" style="edgeStyle=orthogonalEdgeStyle;html=1;" parent="1" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="360" y="280" as="sourcePoint"/>
            <mxPoint x="400" y="280" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`

const connectorFixXML = `<mxfile host="test">
  <diagram id="d1" name="P0">
    <mxGraphModel>
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
        <mxCell id="2" value="D1: 初期成果物" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="0" y="60" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="16" value="" style="ellipse;whiteSpace=wrap;html=1;aspect=fixed;" parent="1" vertex="1">
          <mxGeometry x="200" y="90" width="20" height="20" as="geometry"/>
        </mxCell>
        <mxCell id="3" value="P1: プロセス1" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="320" y="0" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="4" value="P2: プロセス2" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="320" y="120" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="20" style="edgeStyle=none;html=1;" parent="1" source="2" target="16" edge="1">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="21" style="edgeStyle=none;html=1;" parent="1" source="16" target="3" edge="1">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
        <mxCell id="22" style="edgeStyle=none;html=1;" parent="1" source="16" target="4" edge="1">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`

const commentLayerXML = `<mxfile host="test">
  <diagram id="d1" name="P0">
    <mxGraphModel>
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
        <mxCell id="cmt" value="Comment" parent="0"/>
        <mxCell id="2" value="D1" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="320" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="3" value="P1" style="ellipse;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="480" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="4" style="edgeStyle=orthogonalEdgeStyle;html=1;" parent="1" source="2" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="440" y="280" as="sourcePoint"/>
            <mxPoint x="480" y="280" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
        <mxCell id="8" style="edgeStyle=orthogonalEdgeStyle;html=1;" parent="1" source="2" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="440" y="280" as="sourcePoint"/>
            <mxPoint x="1060" y="1040" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
        <mxCell id="c1" value="コメント枠" style="rounded=0;whiteSpace=wrap;html=1;" parent="cmt" vertex="1">
          <mxGeometry x="1000" y="1000" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="c9" style="html=1;" parent="cmt" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="7000" y="7000" as="sourcePoint"/>
            <mxPoint x="8000" y="8000" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`

const groupedCommentLayerXML = `<mxfile host="test">
  <diagram id="d1" name="P0">
    <mxGraphModel>
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
        <mxCell id="cmt" value="Comment" parent="0"/>
        <mxCell id="2" value="D1" style="rounded=0;whiteSpace=wrap;html=1;" parent="1" vertex="1">
          <mxGeometry x="320" y="240" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="re" style="edgeStyle=orthogonalEdgeStyle;html=1;" parent="1" source="2" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="440" y="280" as="sourcePoint"/>
            <mxPoint x="1060" y="1040" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
        <mxCell id="grp" value="" style="group" parent="cmt" vertex="1" connectable="0">
          <mxGeometry x="1000" y="1000" width="200" height="200" as="geometry"/>
        </mxCell>
        <mxCell id="gc1" value="コメント枠" style="rounded=0;whiteSpace=wrap;html=1;" parent="grp" vertex="1">
          <mxGeometry x="1000" y="1000" width="120" height="80" as="geometry"/>
        </mxCell>
        <mxCell id="gc9" style="html=1;" parent="grp" edge="1">
          <mxGeometry relative="1" as="geometry">
            <mxPoint x="7000" y="7000" as="sourcePoint"/>
            <mxPoint x="8000" y="8000" as="targetPoint"/>
          </mxGeometry>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>`

func collectVertexIDs(nodes []*xmldom.Node) map[string]bool {
	result := make(map[string]bool)
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			if v, _ := n.GetAttr("vertex", ""); v != "1" {
				return
			}
			id, _ := n.GetAttr("id", "")
			result[id] = true
		}, nil)
	}
	return result
}

type edgeInfo struct {
	Source, Target string
	HasSourcePoint bool
	HasTargetPoint bool
	HasArray       bool
	Present        bool
}

func collectEdgeInfo(nodes []*xmldom.Node) map[string]edgeInfo {
	result := make(map[string]edgeInfo)
	for _, root := range nodes {
		root.Traverse(func(n *xmldom.Node) {
			if n.Kind != xmldom.ElementNode || n.Start.Name.Local != "mxCell" {
				return
			}
			if e, _ := n.GetAttr("edge", ""); e != "1" {
				return
			}
			id, _ := n.GetAttr("id", "")
			src, _ := n.GetAttr("source", "")
			tgt, _ := n.GetAttr("target", "")
			info := edgeInfo{Source: src, Target: tgt, Present: true}
			for _, geo := range n.Children {
				if geo.Kind != xmldom.ElementNode || geo.Start.Name.Local != "mxGeometry" {
					continue
				}
				for _, gch := range geo.Children {
					if gch.Kind != xmldom.ElementNode {
						continue
					}
					switch gch.Start.Name.Local {
					case "mxPoint":
						switch as, _ := gch.GetAttr("as", ""); as {
						case "sourcePoint":
							info.HasSourcePoint = true
						case "targetPoint":
							info.HasTargetPoint = true
						}
					case "Array":
						info.HasArray = true
					}
				}
			}
			result[id] = info
		}, nil)
	}
	return result
}

func TestFix(t *testing.T) {
	logger := slog.New(slogtest.NewTestHandler(t))

	t.Run("connects touching edges, strips points, keeps waypoints, keeps unhit", func(t *testing.T) {
		nodes, err := pfddrawio.Fix(strings.NewReader(fixSampleXML), pfddrawio.FixOptions{ExpandX: 1, ExpandY: 1}, logger)
		if err != nil {
			t.Fatal(err)
		}
		edges := collectEdgeInfo(nodes)

		if e := edges["4"]; e.Source != "2" || e.Target != "3" || e.HasSourcePoint || e.HasTargetPoint {
			t.Errorf("edge 4 = %+v, want source=2 target=3 with no fixed points", e)
		}
		if e := edges["6"]; e.Source != "3" || e.Target != "5" || e.HasSourcePoint || e.HasTargetPoint {
			t.Errorf("edge 6 = %+v, want source=3 target=5 with no fixed points", e)
		}
		if e := edges["7"]; e.Source != "5" || e.Target != "3" || e.HasSourcePoint || e.HasTargetPoint || !e.HasArray {
			t.Errorf("edge 7 = %+v, want source=5 target=3 no fixed points but Array kept", e)
		}
		if e := edges["99"]; !e.Present || e.Source != "" || e.Target != "" || !e.HasSourcePoint || !e.HasTargetPoint {
			t.Errorf("edge 99 = %+v, want unchanged unhit edge retaining its points", e)
		}
	})

	t.Run("delete-unhit removes unhit edges but keeps connected ones", func(t *testing.T) {
		nodes, err := pfddrawio.Fix(strings.NewReader(fixSampleXML), pfddrawio.FixOptions{DeleteUnhit: true, ExpandX: 1, ExpandY: 1}, logger)
		if err != nil {
			t.Fatal(err)
		}
		edges := collectEdgeInfo(nodes)
		if e, ok := edges["99"]; ok {
			t.Errorf("edge 99 should be deleted, but present: %+v", e)
		}
		if e := edges["4"]; e.Source != "2" || e.Target != "3" {
			t.Errorf("edge 4 = %+v, want still connected", e)
		}
	})

	t.Run("self-loop edge is left unchanged (not connected) by default", func(t *testing.T) {
		nodes, err := pfddrawio.Fix(strings.NewReader(selfLoopXML), pfddrawio.FixOptions{ExpandX: 1, ExpandY: 1}, logger)
		if err != nil {
			t.Fatal(err)
		}
		edges := collectEdgeInfo(nodes)
		if e := edges["4"]; !e.Present || e.Source != "" || e.Target != "" || !e.HasSourcePoint || !e.HasTargetPoint {
			t.Errorf("edge 4 = %+v, want unchanged self-loop edge retaining its points", e)
		}
	})

	t.Run("delete-unhit removes the self-loop edge", func(t *testing.T) {
		nodes, err := pfddrawio.Fix(strings.NewReader(selfLoopXML), pfddrawio.FixOptions{DeleteUnhit: true, ExpandX: 1, ExpandY: 1}, logger)
		if err != nil {
			t.Fatal(err)
		}
		if e, ok := collectEdgeInfo(nodes)["4"]; ok {
			t.Errorf("edge 4 should be deleted, but present: %+v", e)
		}
	})

	t.Run("delete-unhit keeps connector vertex and all its connected edges", func(t *testing.T) {
		nodes, err := pfddrawio.Fix(strings.NewReader(connectorFixXML), pfddrawio.FixOptions{DeleteUnhit: true, ExpandX: 1, ExpandY: 1}, logger)
		if err != nil {
			t.Fatal(err)
		}

		if !collectVertexIDs(nodes)["16"] {
			t.Error("connector vertex 16 was removed, want kept")
		}

		edges := collectEdgeInfo(nodes)
		for _, want := range []struct {
			id, source, target string
		}{
			{"20", "2", "16"},
			{"21", "16", "3"},
			{"22", "16", "4"},
		} {
			if e := edges[want.id]; !e.Present || e.Source != want.source || e.Target != want.target {
				t.Errorf("edge %s = %+v, want present with source=%s target=%s", want.id, e, want.source, want.target)
			}
		}
	})

	t.Run("hitbox-expand 0.9 shrinks hitbox so a boundary endpoint becomes unhit", func(t *testing.T) {
		nodes, err := pfddrawio.Fix(strings.NewReader(fixSampleXML), pfddrawio.FixOptions{ExpandX: 0.9, ExpandY: 0.9}, logger)
		if err != nil {
			t.Fatal(err)
		}
		edges := collectEdgeInfo(nodes)

		if e := edges["4"]; e.Target != "" || !e.HasTargetPoint {
			t.Errorf("edge 4 = %+v, want unchanged (unhit) under 0.9 expand", e)
		}
	})
}

type connectedEdge struct {
	id     string
	source string
	target string
}

func TestFixIgnoresCommentLayer(t *testing.T) {
	logger := slog.New(slogtest.NewTestHandler(t))

	testCases := map[string]struct {
		XML string

		CommentEdgeID string

		UnhitEdgeID string

		Connected *connectedEdge
	}{
		"direct children of comment layer": {
			XML:           commentLayerXML,
			CommentEdgeID: "c9",
			UnhitEdgeID:   "8",
			Connected:     &connectedEdge{id: "4", source: "2", target: "3"},
		},
		"grouped descendants of comment layer": {
			XML:           groupedCommentLayerXML,
			CommentEdgeID: "gc9",
			UnhitEdgeID:   "re",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Run("comment cells are ignored; real edges still fixed", func(t *testing.T) {
				nodes, err := pfddrawio.Fix(strings.NewReader(testCase.XML), pfddrawio.FixOptions{ExpandX: 1, ExpandY: 1}, logger)
				if err != nil {
					t.Fatal(err)
				}
				edges := collectEdgeInfo(nodes)

				if e := edges[testCase.UnhitEdgeID]; !e.Present || e.Source != "2" || e.Target != "" || !e.HasTargetPoint {
					t.Errorf("edge %s = %+v, want unchanged (unhit; not attracted to comment vertex)", testCase.UnhitEdgeID, e)
				}

				if e := edges[testCase.CommentEdgeID]; !e.Present || e.Source != "" || e.Target != "" || !e.HasSourcePoint || !e.HasTargetPoint {
					t.Errorf("edge %s = %+v, want unchanged comment edge retaining its points", testCase.CommentEdgeID, e)
				}

				if c := testCase.Connected; c != nil {
					if e := edges[c.id]; e.Source != c.source || e.Target != c.target || e.HasSourcePoint || e.HasTargetPoint {
						t.Errorf("edge %s = %+v, want source=%s target=%s with no fixed points", c.id, e, c.source, c.target)
					}
				}
			})

			t.Run("delete-unhit keeps comment edge but deletes the real unhit edge", func(t *testing.T) {
				nodes, err := pfddrawio.Fix(strings.NewReader(testCase.XML), pfddrawio.FixOptions{DeleteUnhit: true, ExpandX: 1, ExpandY: 1}, logger)
				if err != nil {
					t.Fatal(err)
				}
				edges := collectEdgeInfo(nodes)

				if e, ok := edges[testCase.CommentEdgeID]; !ok || !e.Present {
					t.Errorf("edge %s should be kept (comment ignored), but missing: %+v", testCase.CommentEdgeID, e)
				}

				if e, ok := edges[testCase.UnhitEdgeID]; ok {
					t.Errorf("edge %s (real unhit edge) should be deleted, but present: %+v", testCase.UnhitEdgeID, e)
				}

				if c := testCase.Connected; c != nil {
					if e := edges[c.id]; e.Source != c.source || e.Target != c.target {
						t.Errorf("edge %s = %+v, want still connected", c.id, e)
					}
				}
			})
		})
	}
}

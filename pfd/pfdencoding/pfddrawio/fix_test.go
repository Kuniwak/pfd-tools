package pfddrawio_test

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/geom"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/google/go-cmp/cmp"
)

func pt(x, y float64) *geom.Point {
	return &geom.Point{X: x, Y: y}
}

func TestPlanFix(t *testing.T) {

	boxA := pfddrawio.Hitbox{ID: "2", Rect: geom.Rect{X: 320, Y: 240, Width: 120, Height: 80}}
	boxB := pfddrawio.Hitbox{ID: "3", Rect: geom.Rect{X: 480, Y: 240, Width: 120, Height: 80}}
	boxD := pfddrawio.Hitbox{ID: "5", Rect: geom.Rect{X: 640, Y: 240, Width: 120, Height: 80}}

	testCases := map[string]struct {
		Hitboxes []pfddrawio.Hitbox
		Edges    []pfddrawio.EdgeGeom
		Expected pfddrawio.FixPlan
	}{
		"resolve missing target (source already set)": {
			Hitboxes: []pfddrawio.Hitbox{boxA, boxB},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "4", Source: "2", TargetPoint: pt(480, 280)},
			},
			Expected: pfddrawio.FixPlan{
				Connections: []pfddrawio.Connection{{ID: "4", Source: "2", Target: "3"}},
			},
		},
		"resolve both sides (fully unlinked)": {
			Hitboxes: []pfddrawio.Hitbox{boxB, boxD},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "7", SourcePoint: pt(700, 240), TargetPoint: pt(540, 240)},
			},
			Expected: pfddrawio.FixPlan{
				Connections: []pfddrawio.Connection{{ID: "7", Source: "5", Target: "3"}},
			},
		},
		"unhit: target point in no hitbox": {
			Hitboxes: []pfddrawio.Hitbox{boxA},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "9", Source: "2", TargetPoint: pt(9999, 9999)},
			},
			Expected: pfddrawio.FixPlan{Unhit: []pfddrawio.CellID{"9"}},
		},
		"unhit: missing side has no point": {
			Hitboxes: []pfddrawio.Hitbox{boxA},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "11", Source: "2", TargetPoint: nil},
			},
			Expected: pfddrawio.FixPlan{Unhit: []pfddrawio.CellID{"11"}},
		},
		"ambiguous: point inside two overlapping hitboxes": {
			Hitboxes: []pfddrawio.Hitbox{
				{ID: "x", Rect: geom.Rect{X: 0, Y: 0, Width: 100, Height: 100}},
				{ID: "y", Rect: geom.Rect{X: 50, Y: 50, Width: 100, Height: 100}},
			},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "10", SourcePoint: pt(75, 75), TargetPoint: pt(75, 75)},
			},
			Expected: pfddrawio.FixPlan{Ambiguous: []pfddrawio.CellID{"10"}},
		},
		"self-loop: both endpoints resolve to the same hitbox": {
			Hitboxes: []pfddrawio.Hitbox{boxA},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "4", SourcePoint: pt(360, 280), TargetPoint: pt(400, 280)},
			},
			Expected: pfddrawio.FixPlan{SelfLoop: []pfddrawio.CellID{"4"}},
		},
		"self-loop: existing source equals resolved target": {
			Hitboxes: []pfddrawio.Hitbox{boxA},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "5", Source: "2", TargetPoint: pt(400, 280)},
			},
			Expected: pfddrawio.FixPlan{SelfLoop: []pfddrawio.CellID{"5"}},
		},
		"already fully linked is ignored": {
			Hitboxes: []pfddrawio.Hitbox{boxA, boxB},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "4", Source: "2", Target: "3"},
			},
			Expected: pfddrawio.FixPlan{},
		},
		"ambiguous takes precedence over unhit, order preserved": {
			Hitboxes: []pfddrawio.Hitbox{
				boxA,
				{ID: "x", Rect: geom.Rect{X: 0, Y: 0, Width: 100, Height: 100}},
				{ID: "y", Rect: geom.Rect{X: 50, Y: 50, Width: 100, Height: 100}},
			},
			Edges: []pfddrawio.EdgeGeom{
				{ID: "e1", Source: "2", TargetPoint: pt(9999, 9999)},
				{ID: "e2", SourcePoint: pt(75, 75), TargetPoint: pt(9999, 9999)},
			},
			Expected: pfddrawio.FixPlan{
				Unhit:     []pfddrawio.CellID{"e1"},
				Ambiguous: []pfddrawio.CellID{"e2"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.PlanFix(tc.Hitboxes, tc.Edges)
			if diff := cmp.Diff(tc.Expected, got); diff != "" {
				t.Errorf("PlanFix() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

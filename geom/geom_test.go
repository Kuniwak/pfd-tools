package geom_test

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/geom"
)

func TestRectContains(t *testing.T) {

	r := geom.Rect{X: 320, Y: 240, Width: 120, Height: 80}
	testCases := map[string]struct {
		Point    geom.Point
		Expected bool
	}{
		"center inside":         {geom.Point{X: 380, Y: 280}, true},
		"left edge inclusive":   {geom.Point{X: 320, Y: 280}, true},
		"right edge inclusive":  {geom.Point{X: 440, Y: 280}, true},
		"top edge inclusive":    {geom.Point{X: 380, Y: 240}, true},
		"bottom edge inclusive": {geom.Point{X: 380, Y: 320}, true},
		"corner inclusive":      {geom.Point{X: 440, Y: 320}, true},
		"outside left":          {geom.Point{X: 319.9, Y: 280}, false},
		"outside right":         {geom.Point{X: 440.1, Y: 280}, false},
		"outside above":         {geom.Point{X: 380, Y: 239.9}, false},
		"outside below":         {geom.Point{X: 380, Y: 320.1}, false},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := r.Contains(tc.Point); got != tc.Expected {
				t.Errorf("Rect%v.Contains(%v) = %v, want %v", r, tc.Point, got, tc.Expected)
			}
		})
	}
}

func TestRectOverlaps(t *testing.T) {

	r := geom.Rect{X: 100, Y: 100, Width: 120, Height: 80}
	testCases := map[string]struct {
		Other    geom.Rect
		Expected bool
	}{
		"identical":              {geom.Rect{X: 100, Y: 100, Width: 120, Height: 80}, true},
		"partial corner overlap": {geom.Rect{X: 200, Y: 160, Width: 120, Height: 80}, true},
		"contained":              {geom.Rect{X: 130, Y: 120, Width: 20, Height: 20}, true},
		"contains":               {geom.Rect{X: 50, Y: 50, Width: 300, Height: 300}, true},
		"touch right edge only":  {geom.Rect{X: 220, Y: 100, Width: 120, Height: 80}, false},
		"touch bottom edge only": {geom.Rect{X: 100, Y: 180, Width: 120, Height: 80}, false},
		"touch corner only":      {geom.Rect{X: 220, Y: 180, Width: 50, Height: 50}, false},
		"disjoint right":         {geom.Rect{X: 240, Y: 100, Width: 120, Height: 80}, false},
		"disjoint below":         {geom.Rect{X: 100, Y: 200, Width: 120, Height: 80}, false},
		"overlap x disjoint y":   {geom.Rect{X: 150, Y: 200, Width: 120, Height: 80}, false},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := r.Overlaps(tc.Other); got != tc.Expected {
				t.Errorf("Rect%v.Overlaps(%v) = %v, want %v", r, tc.Other, got, tc.Expected)
			}

			if got := tc.Other.Overlaps(r); got != tc.Expected {
				t.Errorf("Rect%v.Overlaps(%v) = %v, want %v (symmetry)", tc.Other, r, got, tc.Expected)
			}
		})
	}
}

func TestRectScale(t *testing.T) {

	r := geom.Rect{X: 320, Y: 240, Width: 120, Height: 80}
	testCases := map[string]struct {
		SX, SY   float64
		Expected geom.Rect
	}{
		"identity":    {1.0, 1.0, geom.Rect{X: 320, Y: 240, Width: 120, Height: 80}},
		"shrink 0.5":  {0.5, 0.5, geom.Rect{X: 350, Y: 260, Width: 60, Height: 40}},
		"expand 2.0":  {2.0, 2.0, geom.Rect{X: 260, Y: 200, Width: 240, Height: 160}},
		"anisotropic": {2.0, 0.5, geom.Rect{X: 260, Y: 260, Width: 240, Height: 40}},
		"zero":        {0.0, 0.0, geom.Rect{X: 380, Y: 280, Width: 0, Height: 0}},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := r.Scale(tc.SX, tc.SY)
			if got != tc.Expected {
				t.Errorf("Rect%v.Scale(%v, %v) = %v, want %v", r, tc.SX, tc.SY, got, tc.Expected)
			}
		})
	}
}

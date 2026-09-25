package geom

type Point struct {
	X, Y float64
}

type Rect struct {
	X, Y, Width, Height float64
}

func (r Rect) Contains(p Point) bool {
	return p.X >= r.X && p.X <= r.X+r.Width &&
		p.Y >= r.Y && p.Y <= r.Y+r.Height
}

func (r Rect) Overlaps(o Rect) bool {
	return r.X < o.X+o.Width && o.X < r.X+r.Width &&
		r.Y < o.Y+o.Height && o.Y < r.Y+r.Height
}

func (r Rect) Scale(sx, sy float64) Rect {
	cx := r.X + r.Width/2
	cy := r.Y + r.Height/2
	w := r.Width * sx
	h := r.Height * sy
	return Rect{X: cx - w/2, Y: cy - h/2, Width: w, Height: h}
}

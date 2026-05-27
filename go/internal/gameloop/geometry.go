package gameloop

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func (r *Rect) SetPosition(x, y float64) {
	r.X = x
	r.Y = y
}

func (r1 *Rect) Collides(r2 *Rect) bool {
	if r1.X > r2.X+r2.Width || r2.X > r1.X+r1.Width {
		return false
	}
	if r1.Y > r2.Y+r2.Height || r2.Y > r1.Y+r1.Height {
		return false
	}
	return true
}

type Vector2D = [2]float64

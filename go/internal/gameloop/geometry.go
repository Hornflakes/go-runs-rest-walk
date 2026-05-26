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

type Vector2D = [2]float64

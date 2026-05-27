package gameloop_test

import (
	"testing"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
)

func TestRectCollides(t *testing.T) {
	a := gameloop.Rect{
		X:      3,
		Y:      7,
		Width:  17,
		Height: 23,
	}

	b := gameloop.Rect{
		X:      3 + 17 + 1,
		Y:      7,
		Width:  17,
		Height: 23,
	}

	c := gameloop.Rect{
		X:      3 + 17,
		Y:      7,
		Width:  17,
		Height: 23,
	}

	d := gameloop.Rect{
		X:      3,
		Y:      7 + 23 + 1,
		Width:  17,
		Height: 23,
	}

	e := gameloop.Rect{
		X:      3,
		Y:      7 + 23,
		Width:  17,
		Height: 23,
	}

	if got := a.Collides(&b); got != false {
		t.Errorf("%+v.Collides(%+v) = %t, want false", a, b, got)
	}

	if got := a.Collides(&c); got != true {
		t.Errorf("%+v.Collides(%+v) = %t, want true", a, c, got)
	}

	if got := a.Collides(&d); got != false {
		t.Errorf("%+v.Collides(%+v) = %t, want false", a, d, got)
	}

	if got := a.Collides(&e); got != true {
		t.Errorf("%+v.Collides(%+v) = %t, want true", a, e, got)
	}
}

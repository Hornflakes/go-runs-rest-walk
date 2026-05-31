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

	if got, want := a.Collides(&b), false; got != want {
		t.Errorf("%+v.Collides(%+v) = %t, want %t", a, b, got, want)
	}

	if got, want := a.Collides(&c), true; got != want {
		t.Errorf("%+v.Collides(%+v) = %t, want %t", a, c, got, want)
	}

	if got, want := a.Collides(&d), false; got != want {
		t.Errorf("%+v.Collides(%+v) = %t, want %t", a, d, got, want)
	}

	if got, want := a.Collides(&e), true; got != want {
		t.Errorf("%+v.Collides(%+v) = %t, want %t", a, e, got, want)
	}
}

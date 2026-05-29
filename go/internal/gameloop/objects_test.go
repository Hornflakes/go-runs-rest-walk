package gameloop_test

import (
	"testing"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
)

func TestPlayerFire(t *testing.T) {
	pos := gameloop.Vector2D{0, 0}
	dir := gameloop.Vector2D{1, 0}
	fireRate := int64(128)

	clock := &gameloop.SyntheticClock{}
	t0 := time.Unix(1, 0)
	clock.SetNow(t0)
	player := gameloop.NewPlayerWithClock(pos, dir, fireRate, clock)

	if got := player.Fire(); got != true {
		t.Fatalf("first player.Fire() = %t, want true", got)
	}
	if got := player.Fire(); got != false {
		t.Errorf("instant second player.Fire() = %t, want false", got)
	}

	clock.SetNow(t0.Add((time.Duration(fireRate + 1)) * time.Millisecond))
	if got := player.Fire(); got != true {
		t.Errorf("after cooldown player.Fire() = %t, want true", got)
	}
}

func TestCreateBulletFromPlayer(t *testing.T) {
	pos := gameloop.Vector2D{0, 0}
	dir := gameloop.Vector2D{1, 0}
	fireRate := int64(64)
	player := gameloop.NewPlayer(pos, dir, fireRate)
	bullet := gameloop.CreateBulletFromPlayer(player, 64)

	if bullet.Rect.X != player.Rect.X+player.Rect.Width+1 {
		t.Errorf("expected bullet to be on the positive side of the player")
	}

	want := 64.0
	if got := bullet.Velocity[0]; got != want {
		t.Errorf("bullet.Velocity[0] = %f, want %f", got, want)
	}

	player.Dir[0] = -1
	bullet = gameloop.CreateBulletFromPlayer(player, 36)

	if bullet.Rect.X != player.Rect.X-gameloop.BulletWidth-1 {
		t.Errorf("expected bullet to be on the negative side of the player")
	}

	want = -36.0
	if got := bullet.Velocity[0]; got != want {
		t.Errorf("got %f, want %f", got, want)
	}
}

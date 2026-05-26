package gameloop_test

import (
	"testing"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
)

func TestCreateBulletFromPlayer(t *testing.T) {
	pos := gameloop.Vector2D{0, 0}
	dir := gameloop.Vector2D{1, 0}
	fireRate := int64(64)
	player := gameloop.NewPlayer(pos, dir, fireRate)
	bullet := gameloop.CreateBulletFromPlayer(player, 64)

	if bullet.Rect.X != player.Rect.X+player.Rect.Width+1 {
		t.Errorf("expected bullet to be on the positive side of the player")
	}
	if bullet.Velocity[0] != 64 {
		t.Errorf("got %f, want %f velocity", bullet.Velocity[0], 64.0)
	}

	player.Dir[0] = -1
	bullet = gameloop.CreateBulletFromPlayer(player, 36)

	if bullet.Rect.X != player.Rect.X-gameloop.BulletWidth-1 {
		t.Errorf("expected bullet to be on the negative side of the player")
	}
	if bullet.Velocity[0] != -36 {
		t.Errorf("got %f, want %f velocity", bullet.Velocity[0], -36.0)
	}
}

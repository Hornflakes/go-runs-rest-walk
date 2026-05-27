package gameloop_test

import (
	"testing"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func TestGameStart(t *testing.T) {
	game, _ := newGameAndSockets()
	gameloop.GameStart(game)
	defer gameloop.GameStop(game)

	if got := gameloop.GameQueue(game); got == nil {
		t.Fatalf("game.queue = nil, want non-nil")
	}
}

func TestGameStop(t *testing.T) {
	game, _ := newGameAndSockets()
	gameloop.GameStart(game)

	done := make(chan struct{})
	go func() {
		gameloop.GameStop(game)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("game did not stop within timeout (possible deadlock)")
	}
}

func TestGameUpdateStateFromMessageQueue(t *testing.T) {
	game, sockets := newGameAndSockets()
	gameloop.GameStart(game)
	defer gameloop.GameStop(game)

	shoot := server.CreateSocketMessage(server.Shoot)
	sockets[0].in <- shoot
	sockets[1].in <- shoot

	time.Sleep(time.Millisecond)

	gameloop.GameUpdateStateFromMessageQueue(game)

	if got := len(gameloop.GameBullets(game)); got != 2 {
		t.Errorf("got %d bullets, want 2", got)
	}
}

func TestGameUpdateStateFromMessageQueueRateLimit(t *testing.T) {
	game, sockets := newGameAndSockets()
	gameloop.GameStart(game)
	defer gameloop.GameStop(game)

	shoot := server.CreateSocketMessage(server.Shoot)
	sockets[0].in <- shoot

	time.Sleep(time.Millisecond)

	gameloop.GameUpdateStateFromMessageQueue(game)

	sockets[0].in <- shoot
	gameloop.GameUpdateStateFromMessageQueue(game)

	if got := len(gameloop.GameBullets(game)); got != 1 {
		t.Errorf("got %d bullets, want 1", got)
	}
}

func TestGameUpdateBulletsPositions(t *testing.T) {
	game, sockets := newGameAndSockets()
	gameloop.GameStart(game)
	defer gameloop.GameStop(game)

	shoot := server.CreateSocketMessage(server.Shoot)
	sockets[0].in <- shoot
	sockets[1].in <- shoot

	time.Sleep(time.Millisecond)

	gameloop.GameUpdateStateFromMessageQueue(game)

	if got := len(gameloop.GameBullets(game)); got != 2 {
		t.Fatalf("got %d bullets, want 2", got)
	}

	bullets := gameloop.GameBullets(game)
	x0 := bullets[0].Rect.X
	y0 := bullets[0].Rect.Y
	x1 := bullets[1].Rect.X
	y1 := bullets[1].Rect.Y

	gameloop.GameUpdateBulletsPositions(game, 0)

	if bullets[0].Rect.X != x0 || bullets[0].Rect.Y != y0 ||
		bullets[1].Rect.X != x1 || bullets[1].Rect.Y != y1 {

		t.Error("expected no bullets positions updates")
	}
	gameloop.GameUpdateBulletsPositions(game, 10_000)

	if bullets[0].Rect.X != x0+10*bullets[0].Velocity[0] || bullets[0].Rect.Y != y0 ||
		bullets[1].Rect.X != x1+10*bullets[1].Velocity[0] || bullets[1].Rect.Y != y1 {

		t.Error("expected bullets to be updated, but they were either updated wrong or not updated")
	}

}

func TestGameCheckBulletBulletCollisions(t *testing.T) {
	game, sockets := newGameAndSockets()
	gameloop.GameStart(game)
	defer gameloop.GameStop(game)

	shoot := server.CreateSocketMessage(server.Shoot)
	sockets[0].in <- shoot
	sockets[1].in <- shoot

	time.Sleep(time.Millisecond)

	gameloop.GameUpdateStateFromMessageQueue(game)

	if got := len(gameloop.GameBullets(game)); got != 2 {
		t.Fatalf("got %d bullets, want 2", got)
	}

	gameloop.GameUpdateBulletsPositions(game, 983_000)

	if got := len(gameloop.GameBullets(game)); got != 2 {
		t.Fatalf("after move: got %d bullets, want 2", got)
	}

	gameloop.GameCheckBulletBulletCollisions(game)

	if got := len(gameloop.GameBullets(game)); got != 0 {
		t.Errorf("got %d bullets, want 0", got)
	}
}

func TestGameCheckBulletPlayerCollisions(t *testing.T) {
	game, sockets := newGameAndSockets()
	gameloop.GameStart(game)
	defer gameloop.GameStop(game)

	shoot := server.CreateSocketMessage(server.Shoot)
	sockets[0].in <- shoot
	sheriff := gameloop.GamePlayers(game)[1]

	time.Sleep(time.Millisecond)

	gameloop.GameUpdateStateFromMessageQueue(game)

	if got := len(gameloop.GameBullets(game)); got != 1 {
		t.Fatalf("got %d bullets, want 1", got)
	}

	bullets := gameloop.GameBullets(game)
	player := gameloop.GameCheckBulletPlayerCollisions(game)

	if got := player; got != nil {
		t.Errorf("bullet %+v: hit player = %+v, want nil", bullets[0], got)
	}

	gameloop.GameUpdateBulletsPositions(game, 2_039_000+1_000)

	player = gameloop.GameCheckBulletPlayerCollisions(game)

	if got := player; got == nil {
		t.Fatalf("bullet: %+v hit player = nil, want sheriff %+v", bullets[0], sheriff)
	}

	if got := player; got != sheriff {
		t.Errorf("bullet %+v: hit player = %+v, want sheriff %+v", bullets[0], got, sheriff)
	}
}

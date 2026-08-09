package gameloop_test

import (
	"testing"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
	"github.com/hornflakes/unpocologo"
)

var logger = unpocologo.New("1 vs 2")

func startGame(t *testing.T) (*gameloop.Game, [2]*testSocket, *gameloop.Queue) {
	t.Helper()
	game, sockets := newGameAndSockets()
	gameloop.GameStart(game)
	t.Cleanup(func() { gameloop.GameStop(game) })
	return game, sockets, gameloop.GameQueue(game)
}

func TestGameStart(t *testing.T) {
	_, _, queue := startGame(t)

	if got := queue; got == nil {
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
	game, sockets, queue := startGame(t)

	shoot := server.CreateSocketMessage(server.Shoot)

	sockets[0].in <- shoot
	gameloop.QueueWaitForAck(queue)

	sockets[1].in <- shoot
	gameloop.QueueWaitForAck(queue)

	gameloop.GameUpdateStateFromMessageQueue(game, logger)

	if got, want := len(gameloop.GameBullets(game)), 2; got != want {
		t.Errorf("got %d bullets, want %d", got, want)
	}
}

func TestGameUpdateStateFromMessageQueueRateLimit(t *testing.T) {
	game, sockets, queue := startGame(t)

	shoot := server.CreateSocketMessage(server.Shoot)

	sockets[0].in <- shoot
	gameloop.QueueWaitForAck(queue)

	gameloop.GameUpdateStateFromMessageQueue(game, logger)

	sockets[0].in <- shoot
	gameloop.QueueWaitForAck(queue)

	gameloop.GameUpdateStateFromMessageQueue(game, logger)

	if got, want := len(gameloop.GameBullets(game)), 1; got != want {
		t.Errorf("got %d bullets, want %d", got, want)
	}
}

func TestGameUpdateBulletsPositions(t *testing.T) {
	game, sockets, queue := startGame(t)

	shoot := server.CreateSocketMessage(server.Shoot)

	sockets[0].in <- shoot
	gameloop.QueueWaitForAck(queue)

	sockets[1].in <- shoot
	gameloop.QueueWaitForAck(queue)

	gameloop.GameUpdateStateFromMessageQueue(game, logger)

	if got, want := len(gameloop.GameBullets(game)), 2; got != want {
		t.Fatalf("got %d bullets, want %d", got, want)
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
	game, sockets, queue := startGame(t)

	shoot := server.CreateSocketMessage(server.Shoot)

	sockets[0].in <- shoot
	gameloop.QueueWaitForAck(queue)

	sockets[1].in <- shoot
	gameloop.QueueWaitForAck(queue)

	gameloop.GameUpdateStateFromMessageQueue(game, logger)

	if got, want := len(gameloop.GameBullets(game)), 2; got != want {
		t.Fatalf("got %d bullets, want %d", got, want)
	}

	gameloop.GameUpdateBulletsPositions(game, 983_000)

	if got, want := len(gameloop.GameBullets(game)), 2; got != want {
		t.Fatalf("after move: got %d bullets, want %d", got, want)
	}

	gameloop.GameCheckBulletBulletCollisions(game)

	if got, want := len(gameloop.GameBullets(game)), 0; got != want {
		t.Errorf("got %d bullets, want %d", got, want)
	}
}

func TestGameCheckBulletPlayerCollisions(t *testing.T) {
	game, sockets, queue := startGame(t)

	shoot := server.CreateSocketMessage(server.Shoot)
	sheriff := gameloop.GamePlayers(game)[1]

	sockets[0].in <- shoot
	gameloop.QueueWaitForAck(queue)

	gameloop.GameUpdateStateFromMessageQueue(game, logger)

	if got, want := len(gameloop.GameBullets(game)), 1; got != want {
		t.Fatalf("got %d bullets, want %d", got, want)
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

	if got, want := player, sheriff; got != want {
		t.Errorf("bullet %+v: hit player = %+v, want sheriff %+v", bullets[0], got, sheriff)
	}
}

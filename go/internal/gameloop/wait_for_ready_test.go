package gameloop_test

import (
	"context"
	"testing"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func TestWaitForReady(t *testing.T) {
	s0 := newTestSocket(1)
	s1 := newTestSocket(2)

	done := gameloop.WaitForReady(context.Background(), s0, s1)

	msg0 := <-s0.out
	msg1 := <-s1.out

	if got, want := msg0.Message.Type, server.Ready; got != want {
		t.Errorf("s0 out Type = %d, want %d", got, want)
	}
	if got, want := msg1.Message.Type, server.Ready; got != want {
		t.Errorf("s1 out Type = %d, want %d", got, want)
	}

	if got, want := msg0.Message.Msg, "enemyId=2"; got != want {
		t.Errorf("s0 Msg = %q, want %q", got, want)
	}
	if got, want := msg1.Message.Msg, "enemyId=1"; got != want {
		t.Errorf("s1 Msg = %q, want %q", got, want)
	}

	ready := server.CreateSocketMessage(server.Ready)
	s0.in <- ready
	s1.in <- ready

	select {
	case ok := <-done:
		if got, want := ok, true; got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("not ready within timeout (possible deadlock)")
	}
}

package gameloop_test

import (
	"testing"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func TestWaitForReady(t *testing.T) {
	s0 := newTestSocket()
	s1 := newTestSocket()

	done := gameloop.WaitForReady(s0, s1)

	msg0 := <-s0.out
	msg1 := <-s1.out

	want := server.Ready
	if got := msg0.Message.Type; got != want {
		t.Errorf("s0 out Type = %d, want %d", got, want)
	}
	if got := msg1.Message.Type; got != want {
		t.Errorf("s1 out Type = %d, want %d", got, want)
	}

	ready := server.CreateSocketMessage(server.Ready)
	s0.in <- ready
	s1.in <- ready

	select {
	case ok := <-done:
		if got := ok; got != true {
			t.Errorf("got %t, want true", got)
		}
	case <-time.After(time.Second):
		t.Fatal("not ready within timeout (possible deadlock)")
	}
}

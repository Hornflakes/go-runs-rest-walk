package gameloop_test

import (
	"testing"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func testMessages(t *testing.T, queue *gameloop.Queue, froms []uint) {
	t.Helper()

	messages := queue.Flush()

	if got, want := len(messages), len(froms); got != want {
		t.Fatalf("got %d messages, want %d messages", got, want)
	}

	for i, message := range messages {
		if got, want := message.From, froms[i]; got != want {
			t.Errorf("messages[%d].From = %d, want %d", i, got, want)
		}
	}

	messages = queue.Flush()

	if got, want := len(messages), 0; got != want {
		t.Errorf("got %d messages, want %d", got, want)
	}
}

func TestQueue(t *testing.T) {
	queue := gameloop.NewQueue()
	s0 := newTestSocket(1)
	s1 := newTestSocket(2)

	go queue.Start(s0, s1)
	defer queue.Stop()

	messages := queue.Flush()

	if got, want := len(messages), 0; got != want {
		t.Errorf("got %d messages, want %d when flushed", got, want)
	}

	shoot := server.CreateSocketMessage(server.Shoot)

	s0.in <- shoot
	gameloop.QueueWaitForAck(queue)

	testMessages(t, queue, []uint{1})

	s1.in <- shoot
	gameloop.QueueWaitForAck(queue)

	testMessages(t, queue, []uint{2})

	s1.in <- shoot
	gameloop.QueueWaitForAck(queue)

	s0.in <- shoot
	gameloop.QueueWaitForAck(queue)

	testMessages(t, queue, []uint{2, 1})
}

func TestQueueStop(t *testing.T) {
	queue := gameloop.NewQueue()
	s0 := newTestSocket(1)
	s1 := newTestSocket(2)

	done := make(chan struct{})

	go func() {
		queue.Start(s0, s1)
		close(done)
	}()

	queue.Stop()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("queue did not stop within timeout (possible deadlock)")
	}
}

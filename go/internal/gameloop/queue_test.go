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

	if len(messages) != len(froms) {
		t.Fatalf("got %d messages, want %d messages", len(messages), len(froms))
	}

	for i, message := range messages {
		if message.From != froms[i] {
			t.Errorf("messages[%d].From = %d, want %d", i, message.From, froms[i])
		}
	}

	messages = queue.Flush()
	if messages != nil {
		t.Errorf("got %d messages, want 0", len(messages))
	}
}

func TestQueue(t *testing.T) {
	queue := gameloop.NewQueue()
	s1 := newTestSocket()
	s2 := newTestSocket()

	go queue.Start(s1, s2)
	defer queue.Stop()

	messages := queue.Flush()
	if messages != nil {
		t.Errorf("got %d messages, want 0 when flushed", len(messages))
	}

	// TODO: deterministic testing with ack channels
	s1.in <- server.CreateSocketMessage(server.Shoot)
	time.Sleep(time.Millisecond)
	testMessages(t, queue, []uint{1})

	s2.in <- server.CreateSocketMessage(server.Shoot)
	time.Sleep(time.Millisecond)
	testMessages(t, queue, []uint{2})

	s2.in <- server.CreateSocketMessage(server.Shoot)
	s1.in <- server.CreateSocketMessage(server.Shoot)
	time.Sleep(time.Millisecond)
	testMessages(t, queue, []uint{2, 1})
}

func TestQueueStop(t *testing.T) {
	queue := gameloop.NewQueue()
	s1 := newTestSocket()
	s2 := newTestSocket()

	done := make(chan struct{})

	go func() {
		queue.Start(s1, s2)
		close(done)
	}()

	queue.Stop()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("queue did not stop within timeout (possible deadlock)")
	}
}

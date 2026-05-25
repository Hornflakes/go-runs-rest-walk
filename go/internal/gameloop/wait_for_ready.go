package gameloop

import (
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func sendAndWait(s1, s2 *server.Socket) <-chan bool {
	ready := make(chan bool)

	go func() {
		s1.Out() <- server.CreateSocketMessage(server.Ready)
		s2.Out() <- server.CreateSocketMessage(server.Ready)

		in1 := s1.In()
		in2 := s2.In()
		count := 0
		success := true

		for {
			select {
			case msg, ok := <-in1:
				if !ok {
					success = false
					break
				}

				if msg.Message.Type == server.Ready {
					count += 1
					in1 = nil
				}

			case msg, ok := <-in2:
				if !ok {
					success = false
					break
				}

				if msg.Message.Type == server.Ready {
					count += 1
					in2 = nil
				}
			}

			if count == 2 {
				break
			}
		}

		ready <- success
		close(ready)
	}()

	return ready
}

func WaitForReady(s1, s2 *server.Socket) <-chan bool {
	ready := make(chan bool)

	go func() {
		defer close(ready)

		select {
		case ok := <-sendAndWait(s1, s2):
			ready <- ok
		case <-time.After(30 * time.Second):
			ready <- false
		}
	}()

	return ready
}

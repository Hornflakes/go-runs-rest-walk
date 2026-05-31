package gameloop

import (
	"context"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func sendAndWait(ctx context.Context, s0, s1 server.Socket) <-chan bool {
	ready := make(chan bool)

	go func() {
		s0.Out() <- server.CreateReadyMessage(s1.PlayerId())
		s1.Out() <- server.CreateReadyMessage(s0.PlayerId())

		in1 := s0.In()
		in2 := s1.In()
		count := 0
		success := true

		for {
			select {
			case <-ctx.Done():
				success = false
				return

			case msg, ok := <-in1:
				if !ok {
					success = false
					return
				}

				if msg.Message.Type == server.Ready {
					count += 1
					in1 = nil
				}

			case msg, ok := <-in2:
				if !ok {
					success = false
					return
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

func WaitForReady(ctx context.Context, s0, s1 server.Socket) <-chan bool {
	ready := make(chan bool)

	go func() {
		defer close(ready)
		ctx, cancel := context.WithTimeout(ctx, ReadyTimeoutS)
		defer cancel()

		select {
		case ok := <-sendAndWait(ctx, s0, s1):
			ready <- ok
		case <-ctx.Done():
			ready <- false
		}
	}()

	return ready
}

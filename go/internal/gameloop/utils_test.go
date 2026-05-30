package gameloop_test

import (
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

type testSocket struct {
	playerId   uint64
	remoteAddr string
	in         chan server.SocketMessage
	out        chan server.SocketMessage
	closed     bool
}

func (s *testSocket) PlayerId() uint64                 { return s.playerId }
func (s *testSocket) RemoteAddr() string               { return s.remoteAddr }
func (s *testSocket) In() <-chan server.SocketMessage  { return s.in }
func (s *testSocket) Out() chan<- server.SocketMessage { return s.out }
func (s *testSocket) Closed() bool                     { return s.closed }
func (s *testSocket) Close() error {
	s.closed = true
	return nil
}

func newTestSocket(id uint64) *testSocket {
	return &testSocket{
		playerId:   id,
		remoteAddr: "foo",
		in:         make(chan server.SocketMessage, 1),
		out:        make(chan server.SocketMessage, 1),
		closed:     false,
	}
}

func newGameAndSockets() (*gameloop.Game, [2]*testSocket) {
	clock := &gameloop.SyntheticClock{}
	clock.SetNow(time.Unix(1, 0))
	sockets := [2]*testSocket{
		newTestSocket(1),
		newTestSocket(2),
	}
	game := gameloop.NewGameWithClock(sockets[0], sockets[1], clock)
	return game, sockets
}

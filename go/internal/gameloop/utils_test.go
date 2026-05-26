package gameloop_test

import "github.com/hornflakes/go-runs-rest-walk/internal/server"

type testSocket struct {
	remoteAddr string
	in         chan server.SocketMessage
	out        chan server.SocketMessage
	closed     bool
}

func (s *testSocket) RemoteAddr() string               { return s.remoteAddr }
func (s *testSocket) In() <-chan server.SocketMessage  { return s.in }
func (s *testSocket) Out() chan<- server.SocketMessage { return s.out }
func (s *testSocket) Closed() bool                     { return s.closed }

func newTestSocket() *testSocket {
	return &testSocket{
		remoteAddr: "foo",
		in:         make(chan server.SocketMessage, 1),
		out:        make(chan server.SocketMessage, 1),
		closed:     false,
	}
}

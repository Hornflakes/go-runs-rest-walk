package server

import "testing"

type testSocket struct {
	playerId   uint64
	remoteAddr string
	closed     bool
}

func (s *testSocket) PlayerId() uint64          { return s.playerId }
func (s *testSocket) setPlayerId(id uint64)     { s.playerId = id }
func (s *testSocket) RemoteAddr() string        { return s.remoteAddr }
func (s *testSocket) In() <-chan SocketMessage  { return nil }
func (f *testSocket) Out() chan<- SocketMessage { return make(chan SocketMessage, 1) }
func (s *testSocket) Closed() bool              { return s.closed }
func (s *testSocket) Close() error {
	s.closed = true
	return nil
}

func TestRegisterSocket(t *testing.T) {
	srv := NewServer()
	s0 := &testSocket{remoteAddr: "foo"}
	s1 := &testSocket{remoteAddr: "bar"}

	srv.registerSocket(s0)

	if got, want := s0.PlayerId(), uint64(1); got != want {
		t.Errorf("s0.PlayerId() = %d, want %d", got, want)
	}

	select {
	case <-srv.out:
		t.Fatal("got unexpected lonely pair")
	default:
	}

	srv.registerSocket(s1)

	if got, want := s1.PlayerId(), uint64(2); got != want {
		t.Errorf("s1.PlayerId() = %d, want %d", got, want)
	}

	pair := <-srv.out
	if pair[0] != s0 || pair[1] != s1 {
		t.Errorf("got %s and %s, want foo then bar", pair[0].RemoteAddr(), pair[1].RemoteAddr())
	}
}

func TestRegisterSocketSkipsClosedWaiter(t *testing.T) {
	srv := NewServer()
	s0 := &testSocket{remoteAddr: "foo", closed: true}
	s1 := &testSocket{remoteAddr: "bar"}
	s2 := &testSocket{remoteAddr: "baz"}

	srv.registerSocket(s0)
	srv.registerSocket(s1)

	select {
	case <-srv.out:
		t.Fatal("should not pair with closed waiter")
	default:
	}

	srv.registerSocket(s2)

	pair := <-srv.out
	if pair[0] != s1 || pair[1] != s2 {
		t.Errorf("got %s and %s, want bar then baz", pair[0].RemoteAddr(), pair[1].RemoteAddr())
	}
}

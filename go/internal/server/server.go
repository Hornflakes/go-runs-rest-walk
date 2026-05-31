package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/logger"
)

type Server struct {
	Out           <-chan [2]Socket
	out           chan [2]Socket
	waitingSocket Socket
	mutex         sync.Mutex
	nextPlayerId  atomic.Uint64
}

func NewServer() *Server {
	out := make(chan [2]Socket, 4)
	server := Server{
		Out:           out,
		out:           out,
		waitingSocket: nil,
	}
	return &server
}

func socketAlive(s Socket) bool {
	return s != nil && !s.Closed() && !s.Disconnected()
}

func (s *Server) registerSocket(socket Socket) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if pids, ok := socket.(playerIdSetter); ok {
		pids.setPlayerId(s.nextPlayerId.Add(1))
	}

	if socketAlive(s.waitingSocket) {
		s.out <- [2]Socket{s.waitingSocket, socket}
		s.waitingSocket = nil
		return
	}

	if s.waitingSocket != nil {
		s.waitingSocket.Close()
		s.waitingSocket = nil
	}

	s.waitingSocket = socket
}

func (s *Server) HandleNewConnection(w http.ResponseWriter, r *http.Request) {
	socket, err := NewSocket(w, r)
	if err != nil {
		logger.HardError(
			"websocket upgrade failed",
			fmt.Sprintf("addr=%s err=%v", r.RemoteAddr, err),
		)
		return
	}
	s.registerSocket(socket)

	socket.Out() <- CreateHelloMessage(socket.PlayerId())

	logger.Info(
		"websocket connected",
		logger.PlayerWithAddr(socket.PlayerId(), socket.RemoteAddr()),
	)
}

func WatchPairDisconnect(ctx context.Context, cancel context.CancelFunc, s0, s1 Socket) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s0.Disconnected() || s0.Closed() || s1.Disconnected() || s1.Closed() {
				cancel()
				return
			}
		}
	}
}

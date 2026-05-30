package server

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

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

func (s *Server) registerSocket(socket Socket) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if pids, ok := socket.(playerIdSetter); ok {
		pids.setPlayerId(s.nextPlayerId.Add(1))
	}

	if s.waitingSocket != nil && !s.waitingSocket.Closed() {
		s.out <- [2]Socket{s.waitingSocket, socket}
		s.waitingSocket = nil
	} else {
		s.waitingSocket = socket
	}
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

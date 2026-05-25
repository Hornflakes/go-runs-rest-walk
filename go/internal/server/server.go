package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/fatih/color"
)

type Server struct {
	Out <-chan [2]*Socket

	out           chan [2]*Socket
	waitingSocket *Socket
	mutex         sync.Mutex
}

func NewServer() *Server {
	out := make(chan [2]*Socket, 4)
	server := Server{
		Out:           out,
		out:           out,
		waitingSocket: nil,
	}
	return &server
}

func echoSocket(s *Socket) {
	for msg := range s.in {
		s.out <- msg
	}
}

// TODO: prevent pairing with a closed connection
func (s *Server) HandleNewConnection(w http.ResponseWriter, r *http.Request) {
	socket, err := NewSocket(w, r)
	if err != nil {
		log.Printf("%s | %v", color.RedString("websocket upgrade failed"), err)
		return
	}

	go echoSocket(socket)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.waitingSocket != nil {
		s.out <- [2]*Socket{s.waitingSocket, socket}
		s.waitingSocket = nil
	} else {
		s.waitingSocket = socket
	}
}

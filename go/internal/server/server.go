package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type Server struct {
	Out <-chan []*websocket.Conn

	out           chan []*websocket.Conn
	waitingSocket *websocket.Conn
	mutex         sync.Mutex
}

var upgrader = websocket.Upgrader{}

func NewServer() *Server {
	out := make(chan []*websocket.Conn, 4)
	server := Server{
		Out:           out,
		out:           out,
		waitingSocket: nil,
	}
	return &server
}

func serveEcho(c *websocket.Conn) {
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("%s | %v", color.YellowString("websocket read ended"), err)
			break
		}

		if mt != websocket.TextMessage {
			continue
		}

		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Printf("%s | %v", color.RedString("websocket write failed"), err)
			break
		}
	}
}

// TODO: prevent pairing with a closed connection
func (s *Server) HandleNewConnection(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s | %v", color.RedString("websocket upgrade failed"), err)
		return
	}

	go serveEcho(socket)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.waitingSocket != nil {
		s.out <- []*websocket.Conn{s.waitingSocket, socket}
		s.waitingSocket = nil
	} else {
		s.waitingSocket = socket
	}
}

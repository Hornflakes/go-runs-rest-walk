package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Socket interface {
	RemoteAddr() string
	In() <-chan SocketMessage
	Out() chan<- SocketMessage
	Closed() bool
}

type socketImpl struct {
	conn   *websocket.Conn
	in     <-chan SocketMessage
	out    chan<- SocketMessage
	closed bool
}

func (s *socketImpl) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *socketImpl) In() <-chan SocketMessage {
	return s.in
}

func (s *socketImpl) Out() chan<- SocketMessage {
	return s.out
}

func (s *socketImpl) Closed() bool {
	return s.closed
}

func NewSocket(w http.ResponseWriter, r *http.Request) (Socket, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	in := make(chan SocketMessage, 1)
	out := make(chan SocketMessage, 1)

	socket := socketImpl{
		conn:   conn,
		in:     in,
		out:    out,
		closed: false,
	}

	go func() {
		defer func() {
			close(in)
			close(out)
			conn.Close()
			socket.closed = true
		}()

		for {
			mt, payload, err := conn.ReadMessage()
			if err != nil {
				log.Printf("%s | %v", color.YellowString("websocket read ended"), err)
				break
			}

			if mt != websocket.TextMessage {
				continue
			}

			msg, err := fromSocket(payload)
			if err != nil {
				log.Printf("%s | %v", color.MagentaString("websocket message unmarshal failed"), err)
				continue
			}

			in <- msg
		}
	}()

	go func() {
		for msg := range out {
			bytes, err := json.Marshal(msg.Message)
			if err != nil {
				log.Printf("%s | %v", color.MagentaString("websocket message marshal failed"), err)
				continue
			}

			err = conn.WriteMessage(msg.Type, bytes)
			if err != nil {
				log.Printf("%s | %v", color.RedString("websocket message write failed"), err)
				break
			}
		}
	}()

	return &socket, nil
}

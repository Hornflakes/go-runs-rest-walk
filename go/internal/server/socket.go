package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Socket interface {
	RemoteAddr() string
	In() <-chan SocketMessage
	Out() chan<- SocketMessage
	Closed() bool
	Close() error
}

type socketImpl struct {
	conn      *websocket.Conn
	in        chan SocketMessage
	out       chan SocketMessage
	done      chan struct{}
	closeOnce sync.Once
	closed    bool
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

func (s *socketImpl) Close() error {
	var err error

	s.closeOnce.Do(func() {
		close(s.out)
		<-s.done
		err = s.conn.Close()
		s.closed = true
	})

	return err
}

func NewSocket(w http.ResponseWriter, r *http.Request) (Socket, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	socket := socketImpl{
		conn: conn,
		in:   make(chan SocketMessage, 1),
		out:  make(chan SocketMessage, 1),
		done: make(chan struct{}),
	}

	go func() {
		for {
			mt, payload, err := socket.conn.ReadMessage()
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

			socket.in <- msg
		}

		socket.Close()
	}()

	go func() {
		defer close(socket.done)

		for msg := range socket.out {
			bytes, err := json.Marshal(msg.Message)
			if err != nil {
				log.Printf("%s | %v", color.MagentaString("websocket message marshal failed"), err)
				continue
			}

			err = socket.conn.WriteMessage(msg.Type, bytes)
			if err != nil {
				log.Printf("%s | %v", color.RedString("websocket message write failed"), err)
				break
			}
		}
	}()

	return &socket, nil
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/logger"
)

var upgrader = websocket.Upgrader{}

type Socket interface {
	RemoteAddr() string
	PlayerId() uint64
	In() <-chan SocketMessage
	Out() chan<- SocketMessage
	Disconnected() bool
	Closed() bool
	Close() error
}

type playerIdSetter interface {
	setPlayerId(uint64)
}

type socketImpl struct {
	conn         *websocket.Conn
	playerId     uint64
	in           chan SocketMessage
	out          chan SocketMessage
	done         chan struct{}
	closeOnce    sync.Once
	disconnected atomic.Bool
	closed       bool
}

func (s *socketImpl) PlayerId() uint64      { return s.playerId }
func (s *socketImpl) setPlayerId(id uint64) { s.playerId = id }

func (s *socketImpl) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

func (s *socketImpl) In() <-chan SocketMessage {
	return s.in
}

func (s *socketImpl) Out() chan<- SocketMessage {
	return s.out
}

func (s *socketImpl) Disconnected() bool {
	return s.disconnected.Load()
}

func (s *socketImpl) markDisconnected() {
	if s.disconnected.Swap(true) {
		return
	}
	s.conn.Close()
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

func (s *socketImpl) logDetail(err error) string {
	return fmt.Sprintf("%s err=%v", logger.PlayerWithAddr(s.playerId, s.RemoteAddr()), err)
}

func normalReadEnd(err error) bool {
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection")
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
				if !normalReadEnd(err) {
					logger.Warn("websocket read ended", socket.logDetail(err))
				}
				break
			}

			if mt != websocket.TextMessage {
				continue
			}

			msg, err := fromSocket(payload)
			if err != nil {
				logger.SoftError("websocket message unmarshal failed", socket.logDetail(err))
				continue
			}

			socket.in <- msg
		}

		socket.markDisconnected()
	}()

	go func() {
		defer close(socket.done)

		for msg := range socket.out {
			bytes, err := json.Marshal(msg.Message)
			if err != nil {
				logger.SoftError("websocket message marshal failed", socket.logDetail(err))
				continue
			}

			err = socket.conn.WriteMessage(msg.Type, bytes)
			if err != nil {
				if socket.Disconnected() {
					continue
				}

				logger.HardError("websocket message write failed", socket.logDetail(err))
				break
			}
		}
	}()

	return &socket, nil
}

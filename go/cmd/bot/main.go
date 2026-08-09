package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
	"github.com/hornflakes/unpocologo"
)

var (
	playing  atomic.Bool
	playerId uint64
	enemyId  uint64
	pairLog  *unpocologo.Logger
)

func main() {
	url := "ws://127.0.0.1:37373/"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			unpocologo.Warnf("websocket read ended", "addr=%s err=%v", conn.LocalAddr(), err)
			return
		}

		msg, err := server.UnmarshalMessage(data)
		if err != nil {
			unpocologo.SoftErrorf("websocket message unmarshal failed", "addr=%s err=%v", conn.LocalAddr(), err)
			continue
		}

		switch msg.Type {
		case server.Hello:
			playerId, err = server.ParseHelloMessage(msg.Msg)
			if err != nil {
				unpocologo.HardErrorf("websocket hello message parse failed", "addr=%s err=%v", conn.LocalAddr(), err)
				return
			}

			unpocologo.Infof("websocket connected", "player=%d addr=%s", playerId, conn.LocalAddr())

		case server.Ready:
			if pairLog == nil {
				enemyId, err = server.ParseReadyMessage(msg.Msg)
				if err != nil {
					unpocologo.HardErrorf("websocket ready message parse failed", "addr=%s err=%v", conn.LocalAddr(), err)
					return
				}

				id0, id1 := playerId, enemyId
				if id0 > id1 {
					id0, id1 = id1, id0
				}
				pairLog = unpocologo.New(fmt.Sprintf("%d vs %d", id0, id1))
			}

			reply, err := json.Marshal(server.CreateMessage(server.Ready))
			if err != nil {
				pairLog.SoftErrorf("websocket message marshal failed", "err=%v", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, reply); err != nil {
				pairLog.HardErrorf("websocket message write failed", "err=%v", err)
				return
			}

			pairLog.Milestonef("websocket wrote ready", "player=%d enemy=%d", playerId, enemyId)

		case server.GameOn:
			pairLog.Info("game on", "")

			playing.Store(true)

			go func() {
				ticker := time.NewTicker(200 * time.Millisecond)
				defer ticker.Stop()
				for range ticker.C {
					if !playing.Load() {
						return
					}

					reply, err := json.Marshal(server.CreateMessage(server.Shoot))
					if err != nil {
						pairLog.SoftErrorf("websocket message marshal failed", "err=%v", err)
						continue
					}

					if err := conn.WriteMessage(websocket.TextMessage, reply); err != nil {
						pairLog.HardErrorf("websocket message write failed", "err=%v", err)
						return
					}
				}
			}()

		case server.GameOver:
			playing.Store(false)

			pairLog.Milestone("game over", msg.Msg)

			closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			deadline := time.Now().Add(time.Second)
			_ = conn.WriteControl(websocket.CloseMessage, closeMsg, deadline)

			return
		}
	}
}

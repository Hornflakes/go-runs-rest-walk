package main

import (
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

var playing atomic.Bool

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
			log.Printf("%s | %v", color.YellowString("websocket read ended"), err)
			return
		}

		msg, err := server.UnmarshalMessage(data)
		if err != nil {
			log.Printf("%s | %v", color.MagentaString("websocket message unmarshal failed"), err)
			continue
		}

		switch msg.Type {
		case server.Ready:
			reply, err := json.Marshal(server.CreateMessage(server.Ready))
			if err != nil {
				log.Printf("%s | %v", color.MagentaString("websocket message marshal failed"), err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, reply); err != nil {
				log.Printf("%s | %v", color.RedString("websocket message write failed"), err)
				return
			}
			log.Printf("%s", color.GreenString("websocket wrote ready"))

		case server.GameOn:
			log.Printf("%s", color.CyanString("game on"))
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
						log.Printf("%s | %v", color.MagentaString("websocket message marshal failed"), err)
						continue
					}

					if err := conn.WriteMessage(websocket.TextMessage, reply); err != nil {
						log.Printf("%s | %v", color.RedString("websocket message write failed"), err)
						return
					}
				}
			}()

		case server.GameOver:
			playing.Store(false)
			log.Printf("%s | %s", color.GreenString("game over"), msg.Msg)

			closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			deadline := time.Now().Add(time.Second)
			if err := conn.WriteControl(websocket.CloseMessage, closeMsg, deadline); err != nil {
				log.Printf("%s | %v", color.RedString("websocket close failed"), err)
			}

			return
		}
	}
}

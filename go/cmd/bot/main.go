package main

import (
	"encoding/json"
	"log"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
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
			log.Printf("%s | %v", color.YellowString("websocket read ended"), err)
			return
		}

		msg, err := server.UnmarshalMessage(data)
		if err != nil {
			log.Printf("%s | %v", color.MagentaString("websocket message marshal failed"), err)
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
			log.Println("TODO: game on")
		case server.Shoot:
			log.Println("TODO: shoot")
		case server.GameOver:
			log.Println("TODO: game over")
			return
		}
	}
}

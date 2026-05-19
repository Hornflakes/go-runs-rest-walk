package server

import (
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func HandleEcho(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s | %v", color.RedString("websocket upgrade failed"), err)
		return
	}

	go func() {
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
	}()
}

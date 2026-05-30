package main

import (
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func main() {
	srv := server.NewServer()

	go func() {
		for pair := range srv.Out {
			log.Printf("%s | %s <-> %s", color.GreenString("websocket connections paired"), pair[0].RemoteAddr(), pair[1].RemoteAddr())

			go func(p [2]server.Socket) {
				ok := <-gameloop.WaitForReady(p[0], p[1])
				if ok {
					log.Printf("%s", color.GreenString("websocket pair ready handshake ok"))
					go gameloop.NewGame(p[0], p[1]).Run()
				} else {
					log.Printf("%s", color.YellowString("websocket pair ready handshake failed"))
				}
			}(pair)
		}
	}()

	http.HandleFunc("/", srv.HandleNewConnection)
	log.Fatal(http.ListenAndServe(":37373", nil))
}

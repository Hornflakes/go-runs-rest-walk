package main

import (
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func main() {
	srv := server.NewServer()

	go func() {
		for pair := range srv.Out {
			log.Printf("%s | %v <-> %v", color.GreenString("websocket connections paired"), pair[0].RemoteAddr(), pair[1].RemoteAddr())
		}
	}()

	http.HandleFunc("/", srv.HandleNewConnection)
	log.Fatal(http.ListenAndServe(":37373", nil))
}

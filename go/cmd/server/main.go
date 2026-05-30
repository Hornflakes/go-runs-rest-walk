package main

import (
	"log"
	"net/http"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/logger"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func main() {
	srv := server.NewServer()

	go func() {
		for pair := range srv.Out {
			log := logger.ForPair(pair[0].PlayerId(), pair[1].PlayerId())
			log.Milestone("websockets paired", "")

			go func(p [2]server.Socket, l *logger.Logger) {
				ok := <-gameloop.WaitForReady(p[0], p[1])
				if ok {
					l.Milestone("websockets ready handshake ok", "")

					go gameloop.NewGame(p[0], p[1]).Run()
				} else {
					l.Warn("websockets ready handshake failed", "")

					p[0].Close()
					p[1].Close()
				}
			}(pair, log)
		}
	}()

	http.HandleFunc("/", srv.HandleNewConnection)

	logger.Info("server listening", "addr=:37373")

	log.Fatal(http.ListenAndServe(":37373", nil))
}

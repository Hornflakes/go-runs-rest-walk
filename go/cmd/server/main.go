package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/hornflakes/go-runs-rest-walk/internal/gameloop"
	"github.com/hornflakes/go-runs-rest-walk/internal/logger"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func main() {
	verbose := flag.Bool("verbose", false, "log each registered player shot")
	flag.Parse()

	srv := server.NewServer()

	go func() {
		for pair := range srv.Out {
			log := logger.ForPair(pair[0].PlayerId(), pair[1].PlayerId())
			log.Milestone("websockets paired", "")

			go func(p [2]server.Socket, l *logger.Logger) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				go server.WatchPairDisconnect(ctx, cancel, p[0], p[1])

				ok := <-gameloop.WaitForReady(ctx, p[0], p[1])
				cancel()

				if !ok {
					l.Warn("websockets ready handshake failed", "")

					_ = p[0].Close()
					_ = p[1].Close()
					return
				}

				l.Milestone("websockets ready handshake ok", "")

				gameloop.NewGame(p[0], p[1], *verbose).Run()
			}(pair, log)
		}
	}()

	http.HandleFunc("/", srv.HandleNewConnection)

	logger.Info("server listening", "addr=:37373")

	log.Fatal(http.ListenAndServe(":37373", nil))
}

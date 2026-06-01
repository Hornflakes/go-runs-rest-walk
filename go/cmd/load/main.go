package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/logger"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

var (
	gamesOver      atomic.Uint64
	gamesFailed    atomic.Uint64
	clientsStarted atomic.Uint64
	clientsDone    atomic.Uint64
)

func closeConnGracefully(conn *websocket.Conn, playing *atomic.Bool, shootDone chan struct{}) {
	playing.Store(false)

	if shootDone != nil {
		select {
		case <-shootDone:
		case <-time.After(200 * time.Millisecond):
		}
	}

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	deadline := time.Now().Add(time.Second)
	_ = conn.WriteControl(websocket.CloseMessage, closeMsg, deadline)
}

func playOneGame(ctx context.Context, url string, fireIntervalTime time.Duration) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return err
	}

	var playing atomic.Bool
	var shootDone chan struct{}

	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			closeConnGracefully(conn, &playing, shootDone)
			_ = conn.Close()
		})
	}
	defer shutdown()

	go func() {
		<-ctx.Done()
		shutdown()
	}()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		msg, err := server.UnmarshalMessage(data)
		if err != nil {
			continue
		}

		switch msg.Type {
		case server.Hello:
			server.ParseHelloMessage(msg.Msg)

		case server.Ready:
			server.ParseReadyMessage(msg.Msg)

			reply, err := json.Marshal(server.CreateMessage(server.Ready))
			if err != nil {
				return err
			}

			if err := conn.WriteMessage(websocket.TextMessage, reply); err != nil {
				return err
			}

		case server.GameOn:
			playing.Store(true)
			done := make(chan struct{})
			shootDone = done

			go func() {
				defer close(done)
				ticker := time.NewTicker(fireIntervalTime)
				defer ticker.Stop()

				for range ticker.C {
					if !playing.Load() {
						return
					}

					body, err := json.Marshal(server.CreateMessage(server.Shoot))
					if err != nil {
						return
					}

					if err := conn.WriteMessage(websocket.TextMessage, body); err != nil {
						return
					}
				}
			}()

		case server.GameOver:
			if strings.HasPrefix(msg.Msg, "winner=") && strings.Contains(msg.Msg, "histogram=") {
				fmt.Println(msg.Msg)
				gamesOver.Add(1)
			}
			return nil
		}
	}
}

func runClient(ctx context.Context, url string, fireIntervalTime time.Duration, games int) {
	clientsStarted.Add(1)
	defer clientsDone.Add(1)

	played := 0

	for {
		if ctx.Err() != nil {
			return
		}
		if games > 0 && played >= games {
			return
		}

		err := playOneGame(ctx, url, fireIntervalTime)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			gamesFailed.Add(1)

			time.Sleep(50 * time.Millisecond)
			continue
		}
		played++
	}
}

func main() {
	connections := flag.Int("connections", 10, "number of websockets")
	games := flag.Int("games", 0, "number of games per connection, 0 = run until Ctrl+C")
	host := flag.String("host", "127.0.0.1", "host to connect to")
	port := flag.Int("port", 37373, "port to connect to")
	path := flag.String("path", "/", "path to connect to")
	stagger := flag.Int("stagger", 50, "stagger between connections in ms")
	fire := flag.Int("fire", 200, "fire rate in ms")
	flag.Parse()

	if *connections < 1 {
		log.Fatal("connections must be at least 1")
	}
	if *path == "" || (*path)[0] != '/' {
		log.Fatal("path must start with /")
	}

	url := fmt.Sprintf("ws://%s:%d%s", *host, *port, *path)
	staggerIntervalTime := time.Duration(*stagger) * time.Millisecond
	fireIntervalTime := time.Duration(*fire) * time.Millisecond

	logger.Milestone("load", fmt.Sprintf("url=%s connections=%d games=%d stagger=%s fire=%s",
		url, *connections, *games, staggerIntervalTime, fireIntervalTime))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	for id := 0; id < *connections; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			select {
			case <-time.After(time.Duration(id) * staggerIntervalTime):
			case <-ctx.Done():
				return
			}
			runClient(ctx, url, fireIntervalTime, *games)
		}(id)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.Info("load heartbeat",
					fmt.Sprintf("games_over=%d games_failed=%d clients_started=%d/%d clients_done=%d/%d",
						gamesOver.Load(), gamesFailed.Load(),
						clientsStarted.Load(), *connections,
						clientsDone.Load(), *connections))
			}
		}
	}()

	wg.Wait()

	logger.Milestone("load finished",
		fmt.Sprintf("games_over=%d games_failed=%d clients_started=%d/%d clients_done=%d/%d",
			gamesOver.Load(), gamesFailed.Load(),
			clientsStarted.Load(), *connections,
			clientsDone.Load(), *connections))
}

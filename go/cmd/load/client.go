package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func closeConnGracefully(writer *wsWriter, shootDone chan struct{}) {
	if shootDone != nil {
		select {
		case <-shootDone:
		case <-time.After(200 * time.Millisecond):
		}
	}

	_ = writer.writeClose(websocket.CloseNormalClosure, "")
}

func startTickerFire(
	playing *atomic.Bool,
	w *wsWriter,
	shootMsg []byte,
	fireInterval time.Duration,
	done chan struct{},
) {
	go func() {
		defer close(done)

		ticker := time.NewTicker(fireInterval)
		defer ticker.Stop()

		for range ticker.C {
			if !playing.Load() {
				return
			}
			_ = w.writeText(shootMsg)
		}
	}()
}

func playOneGame(ctx context.Context, url string, clientId int, pool *burstPool, fireIntervalTime time.Duration, shootMsg []byte) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return err
	}

	writer := &wsWriter{conn: conn}
	var playing atomic.Bool
	var shootDone chan struct{}

	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			playing.Store(false)

			if pool != nil {
				pool.unregister(clientId)
			}

			closeConnGracefully(writer, shootDone)
			_ = writer.conn.Close()
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

			if err := writer.writeText(reply); err != nil {
				return err
			}

		case server.GameOn:
			playing.Store(true)

			if pool != nil {
				pool.register(clientId, writer)
			} else {
				done := make(chan struct{})
				shootDone = done
				startTickerFire(&playing, writer, shootMsg, fireIntervalTime, done)
			}

		case server.GameOver:
			if strings.HasPrefix(msg.Msg, "winner=") && strings.Contains(msg.Msg, "histogram=") {
				fmt.Println(msg.Msg)
				gamesOver.Add(1)
			}
			return nil
		}
	}
}

func runClient(ctx context.Context, url string, clientId int, pool *burstPool, fireIntervalTime time.Duration, shootMsg []byte, games int) {
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

		err := playOneGame(ctx, url, clientId, pool, fireIntervalTime, shootMsg)
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

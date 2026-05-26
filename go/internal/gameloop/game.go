package gameloop

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

type Game struct {
	sockets [2]server.Socket
	queue   *GameQueue
}

func NewGame(s1, s2 server.Socket) *Game {
	return &Game{sockets: [2]server.Socket{s1, s2}}
}

func (g *Game) Run() {
	g.sockets[0].Out() <- server.CreateSocketMessage(server.GameOn)
	g.sockets[1].Out() <- server.CreateSocketMessage(server.GameOn)

	g.queue = NewGameQueue()
	go g.queue.Start(g.sockets[0], g.sockets[1])
	defer g.queue.Stop()

	ticks := 0
	startTime := time.Now()

	for {
		g.queue.Flush()

		tickStart := time.Now().UnixMicro()
		ticks++

		now := time.Now().UnixMicro()
		sleepUs := 16_000 - (now - tickStart)
		if sleepUs > 0 {
			time.Sleep(time.Duration(sleepUs) * time.Microsecond)
		}

		if ticks >= 180 || time.Since(startTime) >= 3*time.Second {
			break
		}
	}

	log.Printf("%s | ticks=%d elapsed=%v", color.GreenString("game loop finished"), ticks, time.Since(startTime))
}

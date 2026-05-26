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
	players [2]*Player
	bullets []*Bullet
}

func NewGame(s1, s2 server.Socket) *Game {
	return &Game{sockets: [2]server.Socket{s1, s2}, players: [2]*Player{
		NewPlayer(Vector2D{1024, 0}, Vector2D{-1, 0}, 128),
		NewPlayer(Vector2D{-1024, 0}, Vector2D{1, 0}, 256),
	}}
}

func (g *Game) updateStateFromMessageQueue() {
	messages := g.queue.Flush()
	for _, message := range messages {
		if message.Message.Type == server.Shoot {
			player := g.players[message.From-1]
			fired := player.Fire()

			if fired {
				bullet := CreateBulletFromPlayer(player, 32.0)
				g.bullets = append(g.bullets, &bullet)

				log.Printf("%s | player=%d bullet=%d", color.CyanString("player shot"), message.From, len(g.bullets))
			}
		}
	}
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
		g.updateStateFromMessageQueue()

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

package gameloop

import (
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

type Game struct {
	sockets [2]server.Socket
	queue   *Queue
	players [2]*Player
	bullets []*Bullet
}

func NewGame(s1, s2 server.Socket) *Game {
	return &Game{sockets: [2]server.Socket{s1, s2}, players: [2]*Player{
		NewPlayer(Vector2D{1024, 0}, Vector2D{-1, 0}, 128),
		NewPlayer(Vector2D{-1024, 0}, Vector2D{1, 0}, 256),
	}}
}

func (g *Game) start() {
	g.queue = NewQueue()
	go g.queue.Start(g.sockets[0], g.sockets[1])
}

func (g *Game) stop() {
	if g.queue != nil {
		g.queue.Stop()
	}
}

func (g *Game) updateStateFromMessageQueue() {
	messages := g.queue.Flush()
	for _, message := range messages {
		if message.Message.Type == server.Shoot {
			player := g.players[message.From-1]
			fired := player.Fire()

			if fired {
				bullet := CreateBulletFromPlayer(player, 1.0)
				g.bullets = append(g.bullets, &bullet)

				log.Printf("%s | player=%d bullet=%d", color.CyanString("player shot"), message.From, len(g.bullets))
			}
		}
	}
}

func (g *Game) updateBulletsPositions(delta int64) {
	deltaMs := float64(delta) / 1000
	for _, bullet := range g.bullets {
		bullet.Rect.X += deltaMs * bullet.Velocity[0]
		bullet.Rect.Y += deltaMs * bullet.Velocity[1]
	}
}

func (g *Game) checkBulletBulletCollisions() {
outerLoop:
	for i := 0; i < len(g.bullets); i++ {
		for j := i + 1; j < len(g.bullets); j++ {
			if g.bullets[i].Rect.Collides(&g.bullets[j].Rect) {
				g.bullets = append(g.bullets[:j], g.bullets[j+1:]...)
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				break outerLoop
			}
		}
	}
}

func (g *Game) checkBulletPlayerCollisions() *Player {
	for _, player := range g.players {
		for _, bullet := range g.bullets {
			if bullet.Rect.Collides(&player.Rect) {
				return player
			}
		}
	}
	return nil
}

func (g *Game) getPlayerSocket(player *Player) server.Socket {
	if player == g.players[0] {
		return g.sockets[0]
	}
	return g.sockets[1]
}

func (g *Game) getOtherPlayer(player *Player) *Player {
	if player == g.players[0] {
		return g.players[1]
	}
	return g.players[0]
}

func (g *Game) Run() {
	g.sockets[0].Out() <- server.CreateSocketMessage(server.GameOn)
	g.sockets[1].Out() <- server.CreateSocketMessage(server.GameOn)

	g.start()
	defer g.stop()

	ticks := 0
	tickStartTime := time.Now()
	lastLoopTime := time.Now().UnixMicro()

	for {
		ticks++
		startTime := time.Now().UnixMicro()
		deltaTime := startTime - lastLoopTime

		g.updateStateFromMessageQueue()
		g.updateBulletsPositions(deltaTime)
		g.checkBulletBulletCollisions()

		loser := g.checkBulletPlayerCollisions()
		if loser != nil {
			winner := g.getOtherPlayer(loser)
			winnerSocket := g.getPlayerSocket(winner)
			loserSocket := g.getPlayerSocket(loser)

			winnerSocket.Out() <- server.CreateSocketMessage(server.GameOver)
			loserSocket.Out() <- server.CreateSocketMessage(server.GameOver)
			break
		}

		nowTime := time.Now().UnixMicro()
		sleepUs := 16_000 - (nowTime - startTime)
		if sleepUs > 0 {
			time.Sleep(time.Duration(sleepUs) * time.Microsecond)
		}
		lastLoopTime = startTime
	}

	log.Printf("%s | ticks=%d elapsed=%v bullets=%d", color.GreenString("game over"), ticks, time.Since(tickStartTime), len(g.bullets))
}

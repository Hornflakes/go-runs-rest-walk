package gameloop

import (
	"fmt"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
	"github.com/hornflakes/go-runs-rest-walk/internal/stats"
	"github.com/hornflakes/unpocologo"
)

type Game struct {
	sockets [2]server.Socket
	queue   *Queue
	players [2]*Player
	bullets []*Bullet
	clock   Clock
	stats   *stats.GameFrameStats

	verbose bool
}

func newGameWithClock(s0, s1 server.Socket, clock Clock) *Game {
	return &Game{sockets: [2]server.Socket{s0, s1}, players: [2]*Player{
		newPlayerWithClock(Player0Spawn, Player0Dir, Player0FireRateMs, clock),
		newPlayerWithClock(Player1Spawn, Player1Dir, Player1FireRateMs, clock),
	},
		clock: clock,
		stats: stats.NewGameFrameStats()}
}

func NewGame(s0, s1 server.Socket, verbose bool) *Game {
	game := newGameWithClock(s0, s1, &RealClock{})
	game.verbose = verbose
	return game
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

func (g *Game) updateStateFromMessageQueue(logger *unpocologo.Logger) {
	messages := g.queue.Flush()
	for _, message := range messages {
		if message.Message.Type == server.Shoot {
			player := g.players[message.From-1]
			fired := player.Fire()

			if fired {
				bullet := CreateBulletFromPlayer(player, BulletSpeedMs)
				g.bullets = append(g.bullets, &bullet)

				if g.verbose {
					logger.Infof("player shot", "player=%d bullet=%d",
						g.sockets[message.From-1].PlayerId(), len(g.bullets))
				}
			}
		}
	}
}

func (g *Game) updateBulletsPositions(deltaTime int64) {
	deltaMs := float64(deltaTime) / MicrosPerMs
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
	g.start()
	defer g.stop()

	stats.AddActiveGame()
	defer stats.RemoveActiveGame()

	logger := unpocologo.New(fmt.Sprintf("%d vs %d", g.sockets[0].PlayerId(), g.sockets[1].PlayerId()))

	g.sockets[0].Out() <- server.CreateSocketMessage(server.GameOn)
	g.sockets[1].Out() <- server.CreateSocketMessage(server.GameOn)

	logger.Info("game on", "")

	var winnerId uint64
	ticks := 0
	tickStartTime := g.clock.Now()

	lastLoopTime := g.clock.Now().UnixMicro()

	for {
		ticks++

		startTime := g.clock.Now().UnixMicro()
		deltaTime := startTime - lastLoopTime

		if ticks > 1 {
			g.stats.AddDeltaTime(deltaTime)
		}

		g.updateStateFromMessageQueue(logger)
		g.updateBulletsPositions(deltaTime)
		g.checkBulletBulletCollisions()

		loser := g.checkBulletPlayerCollisions()
		if loser != nil {
			winner := g.getOtherPlayer(loser)
			winnerSocket := g.getPlayerSocket(winner)
			loserSocket := g.getPlayerSocket(loser)
			winnerId = winnerSocket.PlayerId()

			winnerSocket.Out() <- server.CreateWinnerMessage(winnerId, g.stats)
			loserSocket.Out() <- server.CreateLoserMessage()

			_ = winnerSocket.Close()
			_ = loserSocket.Close()

			break
		}

		nowTime := g.clock.Now().UnixMicro()
		sleepUs := TickTargetMicros - (nowTime - startTime)
		if sleepUs > 0 {
			time.Sleep(time.Duration(sleepUs) * time.Microsecond)
		}
		lastLoopTime = startTime
	}

	logger.Milestonef("game over",
		"winner=%d histogram=%s active_games=%d ticks=%d elapsed=%s bullets=%d",
		winnerId,
		g.stats,
		stats.ActiveGames,
		ticks,
		g.clock.Now().Sub(tickStartTime),
		len(g.bullets),
	)
}

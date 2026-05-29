package gameloop

func GameQueue(game *Game) *Queue { return game.queue }

func GameBullets(game *Game) []*Bullet {
	return game.bullets
}

func GamePlayers(game *Game) [2]*Player {
	return game.players
}

var NewGameWithClock = newGameWithClock
var NewPlayerWithClock = newPlayerWithClock

var GameStart = (*Game).start
var GameStop = (*Game).stop
var GameUpdateStateFromMessageQueue = (*Game).updateStateFromMessageQueue
var GameUpdateBulletsPositions = (*Game).updateBulletsPositions
var GameCheckBulletBulletCollisions = (*Game).checkBulletBulletCollisions
var GameCheckBulletPlayerCollisions = (*Game).checkBulletPlayerCollisions

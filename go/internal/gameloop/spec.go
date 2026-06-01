package gameloop

import "time"

const (
	PlayerWidth  = 100
	PlayerHeight = 100
	BulletWidth  = 35
	BulletHeight = 3
)

const (
	Player0SpawnX     = 2500
	Player1SpawnX     = -2500
	Player0FireRateMs = 180
	Player1FireRateMs = 300
)

var (
	Player0Spawn = Vector2D{Player0SpawnX, 0}
	Player1Spawn = Vector2D{Player1SpawnX, 0}
	Player0Dir   = Vector2D{-1, 0}
	Player1Dir   = Vector2D{1, 0}
)

const (
	BulletSpeedMs    = 1.0
	TickTargetMicros = 16_000
	MicrosPerMs      = 1000
)

const ReadyTimeoutS = 30 * time.Second

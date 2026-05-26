package gameloop

import "time"

type Player struct {
	Rect         Rect
	Dir          Vector2D
	FireRate     int64
	lastFireTime int64
}

const (
	PlayerWidth  = 64
	PlayerHeight = 64
	BulletWidth  = 8
	BulletHeight = 2
)

func NewPlayer(pos, dir Vector2D, fireRate int64) *Player {
	return &Player{
		Rect: Rect{
			X:      pos[0],
			Y:      pos[1],
			Width:  PlayerWidth,
			Height: PlayerHeight,
		},
		Dir:          dir,
		FireRate:     fireRate,
		lastFireTime: 0,
	}
}

func (p *Player) Fire() bool {
	now := time.Now().UnixMilli()

	if p.FireRate > now-p.lastFireTime {
		return false
	}

	p.lastFireTime = now
	return true
}

type Bullet struct {
	Rect     Rect
	Velocity Vector2D
}

func newBullet() Bullet {
	return Bullet{
		Rect:     Rect{Width: BulletWidth, Height: BulletHeight},
		Velocity: Vector2D{0, 0},
	}
}

func CreateBulletFromPlayer(player *Player, speed float64) Bullet {
	bullet := newBullet()

	if player.Dir[0] == 1 {
		bullet.Rect.SetPosition(player.Rect.X+player.Rect.Width+1, 0)
	} else {
		bullet.Rect.SetPosition(player.Rect.X-BulletWidth-1, 0)
	}

	bullet.Velocity[0] = player.Dir[0] * speed
	bullet.Velocity[1] = player.Dir[1] * speed

	return bullet
}

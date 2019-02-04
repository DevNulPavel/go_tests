package gameserver

import (
	"math"
	"sync/atomic"
)

const BULLE_MAGIC_NUMBER uint8 = 4

// Variables
var BULLET_ID uint32 = 0

type Bullet struct {
	ID uint32  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	SX float64 `json:"sx"`
	SY float64 `json:"sy"`
}

func NewBullet(x, y, radius int16, angle float32) *Bullet {
	const BULLET_INITIAL_SPEED float64 = 120.0

	// Увеличиваем id
	curId := atomic.AddUint32(&BULLET_ID, 1)
	angleRad := float64(angle)/180.0*math.Pi + math.Pi/2.0

	bullet := &Bullet{
		ID: curId,
		X:  float64(x) + math.Cos(angleRad)*float64(radius),
		Y:  float64(y) + math.Sin(angleRad)*float64(radius),
		SX: math.Cos(angleRad) * BULLET_INITIAL_SPEED,
		SY: math.Sin(angleRad) * BULLET_INITIAL_SPEED,
	}
	return bullet
}

func (bullet *Bullet) WorldTick(delta float64) {
	bullet.X += bullet.SX * delta
	bullet.Y += bullet.SY * delta
}

package gameserver

import (
    "bytes"
    "encoding/binary"
    "math"
)

const BULLE_MAGIC_NUMBER uint8 = 4

type Bullet struct {
    X float64
    Y float64
    SX float64
    SY float64
}

func NewBullet(x,y int16, angle float32) Bullet {
    const BULLET_INITIAL_SPEED float64 = 120.0

    bullet := Bullet{
        X: float64(x),
        Y: float64(y),
        SX: math.Sin(float64(angle)) * BULLET_INITIAL_SPEED,
        SY: math.Cos(float64(angle)) * BULLET_INITIAL_SPEED,
    }
    return bullet
}

func (bullet *Bullet) WorldTick(delta float64) {
    bullet.X += bullet.SX * delta
    bullet.Y += bullet.SY * delta
}

func (bullet *Bullet) ConvertToBytes() ([]byte, error) {
    buffer := new(bytes.Buffer)
    // MagicNumber
    err := binary.Write(buffer, binary.BigEndian, BULLE_MAGIC_NUMBER)
    if err != nil {
        return []byte{}, err
    }
    // X
    err = binary.Write(buffer, binary.BigEndian, int16(bullet.X))
    if err != nil {
        return []byte{}, err
    }
    // Y
    err = binary.Write(buffer, binary.BigEndian, int16(bullet.Y))
    if err != nil {
        return []byte{}, err
    }
    return buffer.Bytes(), nil
}
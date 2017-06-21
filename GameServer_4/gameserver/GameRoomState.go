package gameserver

import (
	"bytes"
	"encoding/binary"
)

const (
	ROOM_STATUS_NOT_IN_GAME = 0
	ROOM_STATUS_IN_GAME     = 1
	ROOM_STATUS_FAIL        = 2
)

type GameRoomState struct {
	ID            int32
	Status        int8
	Width         int16
	Height        int16
	BallPosX      int16
	BallPosY      int16
	BallPosSpeedX float64
	BallPosSpeedY float64
}

func (state *GameRoomState) ConvertToBytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	// ID
	err := binary.Write(buffer, binary.BigEndian, state.ID)
	if err != nil {
		return []byte{}, err
	}
	// Status
	err = binary.Write(buffer, binary.BigEndian, state.Status)
	if err != nil {
		return []byte{}, err
	}
	// Width
	err = binary.Write(buffer, binary.BigEndian, state.Width)
	if err != nil {
		return []byte{}, err
	}
	// Height
	err = binary.Write(buffer, binary.BigEndian, state.Height)
	if err != nil {
		return []byte{}, err
	}
	// BallPosX
	err = binary.Write(buffer, binary.BigEndian, state.BallPosX)
	if err != nil {
		return []byte{}, err
	}
	// BallPosY
	err = binary.Write(buffer, binary.BigEndian, state.BallPosY)
	if err != nil {
		return []byte{}, err
	}

	return buffer.Bytes(), nil
}

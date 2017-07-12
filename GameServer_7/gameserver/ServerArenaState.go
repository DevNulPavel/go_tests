package gameserver

import (
	"bytes"
	"encoding/binary"
)

const GAME_ROOM_STATE_MAGIC_NUMBER uint8 = 2

const (
	GAME_ROOM_STATUS_ACTIVE    = 0
	GAME_ROOM_STATUS_COMPLETED = 1
)

type GameArenaState struct {
	ID     uint32
	Status int8
}

func (state *GameArenaState) ConvertToBytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	// MagicNumber
	err := binary.Write(buffer, binary.BigEndian, GAME_ROOM_STATE_MAGIC_NUMBER)
	if err != nil {
		return []byte{}, err
	}
	// ID
	err = binary.Write(buffer, binary.BigEndian, state.ID)
	if err != nil {
		return []byte{}, err
	}
	// Status
	err = binary.Write(buffer, binary.BigEndian, state.Status)
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}

func (state *GameArenaState) WorldTick(delta float64) {
	if state.Status != GAME_ROOM_STATUS_ACTIVE {
		return
	}
}

func (gameRoomState *GameArenaState) Reset() {
	gameRoomState.Status = GAME_ROOM_STATUS_ACTIVE
}

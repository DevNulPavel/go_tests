package gameserver

import (
	"bytes"
	"encoding/binary"
	"container/list"
)

const CLIENT_STATE_MAGIC_NUMBER uint8 = 1

const (
	CLIENT_STATUS_IN_GAME = 0
	CLIENT_STATUS_FAIL    = 1
)

// Client state structure
type ClientState struct {
	ID      uint32
	X       int16
	Y       int16
	Angle   float32
	Frags   uint16
	Status  int8
	Bullets *list.List
}

func NewState(id uint32, x, y int16) ClientState {
	state := ClientState{
		ID:      id,
		X:       x,
		Y:       y,
		Angle:   0.0,
		Frags:   0,
		Status:  CLIENT_STATUS_IN_GAME,
		Bullets: list.New(),
	}
	return state
}

func (state *ClientState) ConvertToBytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	// MagicNumber
	err := binary.Write(buffer, binary.BigEndian, CLIENT_STATE_MAGIC_NUMBER)
	if err != nil {
		return []byte{}, err
	}
	// ID
	err = binary.Write(buffer, binary.BigEndian, state.ID)
	if err != nil {
		return []byte{}, err
	}
	// X
	err = binary.Write(buffer, binary.BigEndian, state.X)
	if err != nil {
		return []byte{}, err
	}
	// Y
	err = binary.Write(buffer, binary.BigEndian, state.Y)
	if err != nil {
		return []byte{}, err
	}
	// Angle
	err = binary.Write(buffer, binary.BigEndian, state.Angle)
	if err != nil {
		return []byte{}, err
	}
	// Frags
	err = binary.Write(buffer, binary.BigEndian, state.Frags)
	if err != nil {
		return []byte{}, err
	}
	// Status
	err = binary.Write(buffer, binary.BigEndian, state.Status)
	if err != nil {
		return []byte{}, err
	}
	// Bullets count
	bulletsCount := uint8(state.Bullets.Len())
	err = binary.Write(buffer, binary.BigEndian, bulletsCount)
	if err != nil {
		return []byte{}, err
	}
	// Bullets
	bulletsBytes := make([]byte, 0, state.Bullets.Len()*8)
    it := state.Bullets.Front()
	for i := 0; i < state.Bullets.Len(); i++ {
		bulletData, err := it.Value.(Bullet).ConvertToBytes()
		if err == nil {
			bulletsBytes = append(bulletsBytes, bulletData...)
		}
	}
	buffer.Write(bulletsBytes)

	return buffer.Bytes(), nil
}

// конвертация в строку
//func (mes *ClienState) String() string {
//	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
//}

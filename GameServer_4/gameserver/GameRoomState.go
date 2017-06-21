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

type GameRoomState struct {
	ID               int32
	Status           int8
	Width            int16
	Height           int16
	BallPosX         int16
	BallPosY         int16
	BallSpeedX       float64
	BallSpeedY       float64
	clientLeftState  ClientState
	clientRightState ClientState
}

func (state *GameRoomState) ConvertToBytes() ([]byte, error) {
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
	// Client1
	client1Data, err := state.clientLeftState.ConvertToBytes()
	if err != nil {
		return []byte{}, err
	}
	_, err = buffer.Write(client1Data)
	if err != nil {
		return []byte{}, err
	}
	// Client2
	client2Data, err := state.clientRightState.ConvertToBytes()
	if err != nil {
		return []byte{}, err
	}
	_, err = buffer.Write(client2Data)
	if err != nil {
		return []byte{}, err
	}

	return buffer.Bytes(), nil
}

func (state *GameRoomState) WorldTick(delta float64) {
	if state.Status != GAME_ROOM_STATUS_ACTIVE {
		return
	}

	// Рассчет новой позиции
	nextPosX := int16(float64(state.BallPosX) * state.BallSpeedX)
	nextPosY := int16(float64(state.BallPosY) * state.BallSpeedY)

	// Проверка по Y
	if nextPosY < 0 {
		state.BallSpeedY = -state.BallSpeedY
		state.BallPosY = int16(float64(state.BallPosY) * state.BallSpeedY)
	}
	if nextPosY > state.Height {
		state.BallSpeedY = -state.BallSpeedY
		state.BallPosY = int16(float64(state.BallPosY) * state.BallSpeedY)
	}

	// Проверка по X
	const panelWidth int16 = 20
	leftBorder := panelWidth
	rightborder := state.Width - panelWidth
	// Слева
	if nextPosX < leftBorder {
		minY := state.clientLeftState.Y - state.clientLeftState.Height/2
		maxY := state.clientLeftState.Y - state.clientLeftState.Height/2

		if (nextPosY > minY) && (nextPosY < maxY) {
			state.BallSpeedX = -state.BallSpeedX
			state.BallPosX = int16(float64(state.BallPosX) * state.BallSpeedX)
		} else {
			state.Status = GAME_ROOM_STATUS_COMPLETED
			state.clientLeftState.Status = CLIENT_STATUS_FAIL
			state.clientRightState.Status = CLIENT_STATUS_WIN
			state.BallPosX = nextPosX
			state.BallPosX = nextPosY
			state.BallSpeedX = 0.0
			state.BallSpeedY = 0.0
		}
	}
	// Справа
	if nextPosX > rightborder {
		minY := state.clientRightState.Y - state.clientRightState.Height/2
		maxY := state.clientRightState.Y - state.clientRightState.Height/2

		if (nextPosY > minY) && (nextPosY < maxY) {
			state.BallSpeedX = -state.BallSpeedX
			state.BallPosX = int16(float64(state.BallPosX) * state.BallSpeedX)
		} else {
			state.Status = GAME_ROOM_STATUS_COMPLETED
			state.clientLeftState.Status = CLIENT_STATUS_WIN
			state.clientRightState.Status = CLIENT_STATUS_FAIL
			state.BallPosX = nextPosX
			state.BallPosX = nextPosY
			state.BallSpeedX = 0.0
			state.BallSpeedY = 0.0
		}
	}
}

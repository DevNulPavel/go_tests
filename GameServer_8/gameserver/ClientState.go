package gameserver

import (
	"container/list"
)

const (
	CLIENT_STATUS_IN_GAME = 0
	CLIENT_STATUS_FAIL    = 1
)

// Client state structure
type ClientState struct {
	ID      uint32     `json:"id"`
	Size    uint8      `json:"size"`
	X       int16      `json:"x"`
	Y       int16      `json:"y"`
	Angle   float32    `json:"angle"`
	Frags   uint16     `json:"frags"`
	Status  int8       `json:"status"`
	Bullets *list.List `json:"bullets"`
}

func NewState(id uint32, x, y int16) *ClientState {
	const clientSize = 32

	state := &ClientState{
		ID:      id,
		Size:    clientSize,
		X:       x,
		Y:       y,
		Angle:   0.0,
		Frags:   0,
		Status:  CLIENT_STATUS_IN_GAME,
		Bullets: list.New(),
	}
	return state
}

// конвертация в строку
//func (mes *ClienState) String() string {
//	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
//}

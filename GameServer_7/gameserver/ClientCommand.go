package gameserver

import (
	"encoding/json"
)

const (
	CLIENT_COMMAND_TYPE_MOVE  uint8 = 0
	CLIENT_COMMAND_TYPE_SHOOT uint8 = 1
)

type ClientCommand struct {
	ID          uint32  `json:"id"`
	Type        uint8   `json:"type"`
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	VX          float32 `json:"vx"`
	VY          float32 `json:"vy"`
	Duration    float32 `json:"duration"`
	VisualState uint8   `json:"visualState"`
	AnimName    string  `json:"animName"`
}

func NewClientCommand(data []byte) (*ClientCommand, error) {
	command := &ClientCommand{}
	err := json.Unmarshal(data, command)
	return command, err
}

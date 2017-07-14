package gameserver

import (
	"encoding/json"
)

const (
	CLIENT_COMMAND_TYPE_MOVE  uint8 = 0
	CLIENT_COMMAND_TYPE_SHOOT uint8 = 1
)

type ClientCommand struct {
	ID          uint32 `json:"id"`
	Type        uint8  `json:"type"`
	X           int16  `json:"x"`
	Y           int16  `json:"y"`
	VisualState uint8   `json:"visualState"`
}

func NewClientCommand(data []byte) (*ClientCommand, error) {
	command := &ClientCommand{}
	err := json.Unmarshal(data, command)
	return command, err
}

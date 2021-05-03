package gameserver

import (
	"encoding/json"
)

const (
	CLIENT_STATUS_IN_GAME = 0
	CLIENT_STATUS_FAIL    = 1
	CLIENT_STATUS_WIN     = 2
)

// ServerClient state structure
type ServerClientState struct {
	Type           string  `json:"type"`
	ID             uint32  `json:"id"`
	RotationX      float64 `json:"rx"`
	RotationY      float64 `json:"ry"`
	RotationZ      float64 `json:"rz"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	VX             float64 `json:"vx"`
	VY             float64 `json:"vy"`
	Duration       float64 `json:"duration"`
	Status         int8    `json:"status"`
	VisualState    uint8   `json:"visualState"`
	AnimName       string  `json:"animName"`
	StartSkillName string  `json:"startSkillName"`
	TotalDamage    uint32  `json:"totalDamage"`
}

func NewServerClientState(id uint32) ServerClientState {
	state := ServerClientState{
		Type: "ClientState",
		ID:   id,
	}
	return state
}

func (client *ServerClientState) ToBytes() ([]byte, error) {
	return json.Marshal(client)
}

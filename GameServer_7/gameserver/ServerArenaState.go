package gameserver

import (
	"encoding/json"
)

const (
	GAME_ROOM_STATUS_ACTIVE    = 0
	GAME_ROOM_STATUS_COMPLETED = 1
)

type GameArenaState struct {
	ID      uint32              `json:"id"`
	Type    string              `json:"type"`
	Status  int8                `json:"status"`
	Clients []ServerClientState `json:"clients"`
}

func NewServerArenaState(id uint32) GameArenaState {
	state := GameArenaState{
		ID:      id,
		Type:    "ArenaState",
		Clients: []ServerClientState{},
	}
	return state
}

func (state *GameArenaState) ToBytes() ([]byte, error) {
	return json.Marshal(state)
}

func (state *GameArenaState) WorldTick(delta float64) {
	if state.Status != GAME_ROOM_STATUS_ACTIVE {
		return
	}
}

func (state *GameArenaState) Reset() {
	state.Status = GAME_ROOM_STATUS_ACTIVE
}

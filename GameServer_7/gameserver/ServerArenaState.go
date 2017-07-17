package gameserver

import (
	"encoding/json"
)

const (
	GAME_ROOM_STATUS_ACTIVE    = 0
	GAME_ROOM_STATUS_COMPLETED = 1
)

type GameArenaState struct {
	Type    string              `json:"type"`
	ID      uint32              `json:"id"`
	Status  int8                `json:"status"`
	Clients []ServerClientState `json:"clients"`
}

func NewServerArenaState(id uint32) GameArenaState {
	state := GameArenaState{
        Type:    "ArenaState",
		ID:      id,
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

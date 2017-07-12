package gameserver

import (
	"encoding/json"
)

const (
	GAME_ROOM_STATUS_ACTIVE    = 0
	GAME_ROOM_STATUS_COMPLETED = 1
)

type GameArenaState struct {
	Type   string `json:"type"`
	ID     uint32 `json:"id"`
	Status int8   `json:"status"`
}

func NewServerArenaState() GameArenaState {
	state := GameArenaState{
		Type: "ArenaState",
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

func (gameRoomState *GameArenaState) Reset() {
	gameRoomState.Status = GAME_ROOM_STATUS_ACTIVE
}

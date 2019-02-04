package gameserver

const (
	GAME_STATE_MESSAGE_INIT_PLAYER uint8 = 0
	GAME_STATE_MESSAGE_WORLD_STATE uint8 = 1
)

type GameStateMessage struct {
	Type        uint8         `json:"type"`
	WorldData   WorldInfo     `json:"world"`
	ClienStates []ClientState `json:"clients"`
}

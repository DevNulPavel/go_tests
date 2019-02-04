package gameserver

const (
	CLIENT_COMMAND_TYPE_MOVE  uint8 = 0
	CLIENT_COMMAND_TYPE_SHOOT uint8 = 1
)

type ClientCommand struct {
	ID    uint32  `json:"id"`
	X     int16   `json:"x"`
	Y     int16   `json:"y"`
	Angle float32 `json:"angle"`
	Type  uint8   `json:"type"`
}

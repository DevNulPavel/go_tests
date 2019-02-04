package gameserver

const (
	CLIENT_STATUS_IN_GAME = 0
	CLIENT_STATUS_FAIL    = 1
	CLIENT_STATUS_WIN     = 2
)

const (
	CLIENT_TYPE_LEFT  = 0
	CLIENT_TYPE_RIGHT = 1
)

type ClientState struct {
	ID     uint32 `json:"id"`
	Type   uint8  `json:"t"`
	Y      int16  `json:"y"`
	Height int16  `json:"h"`
	Status int8   `json:"st"`
}

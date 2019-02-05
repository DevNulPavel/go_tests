package gameserver

type FromPlayerMessage struct {
	ID uint32 `json:"id"`
	Y  int16  `json:"y"`
}

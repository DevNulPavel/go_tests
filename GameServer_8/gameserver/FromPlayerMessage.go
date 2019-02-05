package gameserver

type FromPlayerMessage struct {
	ID uint32  `json:"id"`
	Y  float32 `json:"y"`
}

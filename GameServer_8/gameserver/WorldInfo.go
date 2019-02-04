package gameserver

type WorldInfo struct {
	SizeX uint16 `json:"sizeX"`
	SizeY uint16 `json:"sizeY"`
}

func NewWorldInfo() *WorldInfo {
	const worldSizeX = 400
	const worldSizeY = 400

	info := &WorldInfo{
		SizeX: worldSizeX,
		SizeY: worldSizeY,
	}
	return info
}

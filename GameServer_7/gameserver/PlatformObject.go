package gameserver

const (
	PLATFORM_OBJECT_TYPE_FLOOR       = 0
	PLATFORM_OBJECT_TYPE_WALL        = 1
	PLATFORM_OBJECT_TYPE_CORNER      = 2
	PLATFORM_OBJECT_TYPE_ARCHE       = 3
	PLATFORM_OBJECT_TYPE_PILLAR      = 4
	PLATFORM_OBJECT_TYPE_COFFIN      = 5
	PLATFORM_OBJECT_TYPE_ENVIRONMENT = 6
	PLATFORM_OBJECT_TYPE_DECOR       = 7
)

type PlatformObject struct {
	Id          string  `json:"id"`          // имя
	Type        uint8   `json:"type"`        // тип
	Probability float64 `json:"probability"` // вероятность появления
	Rotation    int8    `json:"rotation"`    // поворот
	X           int16   `json:"positionX"`
	Y           int16   `json:"positionY"`
	SizeX       int16   `json:"sizeX"`
	SizeY       int16   `json:"sizeY"`
	Cells       []int8 `json:"cells"`
}

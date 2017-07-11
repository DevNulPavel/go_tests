package gameserver

type PlatformObjectType uint8

const (
	PLATFORM_OBJ_TYPE_FLOOR       PlatformObjectType = 0
	PLATFORM_OBJ_TYPE_WALL        PlatformObjectType = 1
	PLATFORM_OBJ_TYPE_CORNER      PlatformObjectType = 2
	PLATFORM_OBJ_TYPE_ARCHE       PlatformObjectType = 3
	PLATFORM_OBJ_TYPE_PILLAR      PlatformObjectType = 4
	PLATFORM_OBJ_TYPE_COFFIN      PlatformObjectType = 5
	PLATFORM_OBJ_TYPE_ENVIRONMENT PlatformObjectType = 6
	PLATFORM_OBJ_TYPE_DECOR       PlatformObjectType = 7
)

type PlatformObjectInfo struct {
	Id          string             `json:"id"`          // имя
	Type        PlatformObjectType `json:"type"`        // тип
	Probability float64            `json:"probability"` // вероятность появления
	Rotation    int8               `json:"rotation"`    // поворот
	X           int16              `json:"positionX"`
	Y           int16              `json:"positionY"`
	Width       int16              `json:"sizeX"`
	Height      int16              `json:"sizeY"`
	Cells       []int8             `json:"cells"`
}

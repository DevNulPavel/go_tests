package gameserver

const (
	MONSTER_STATE_STATUS_ALIVE = 0
	MONSTER_STATE_STATUS_DEAD  = 1
)

type ServerMonsterState struct {
	Type          string  `json:"type"`
	ID            uint32  `json:"id"`
	Name          string  `json:"name"`
	RotX          float64 `json:"rx"`
	RotY          float64 `json:"ry"`
	RotZ          float64 `json:"rz"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Status        uint8   `json:"status"`
	Health        int16   `json:"health"`
	VisualState   int16   `json:"visualState"`
	AnimationName string  `json:"animName"`
}

func NewServerMonsterState(id uint32) ServerMonsterState {
	state := ServerMonsterState{
		Type: "MonsterState",
		ID:   id,
	}
	return state
}

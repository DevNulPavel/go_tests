package gameserver

const (
	GAME_ROOM_STATUS_ACTIVE    = 0
	GAME_ROOM_STATUS_COMPLETED = 1
)

type GameRoomState struct {
	ID         uint32  `json:"id"`
	Status     int8    `json:"status"`
	Width      int16   `json:"width"`
	Height     int16   `json:"height"`
	BallPosX   float64 `json:"ballPosX"`
	BallPosY   float64 `json:"ballPosY"`
	BallSpeedX float64 `json:"ballSpeedX"`
	BallSpeedY float64 `json:"ballSpeedY"`
}

func (gameRoomState *GameRoomState) Reset(speedX, speedY float64) {
	gameRoomState.Status = GAME_ROOM_STATUS_ACTIVE
	gameRoomState.BallSpeedY = speedX
	gameRoomState.BallSpeedX = speedY
	gameRoomState.BallPosX = float64(gameRoomState.Width / 2)
	gameRoomState.BallPosY = float64(gameRoomState.Height / 2)
}

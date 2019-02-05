package gameserver

func WorldTick(delta float64, state *GameRoomState, leftClientState *ClientState, rightClientState *ClientState) {
	if state.Status != GAME_ROOM_STATUS_ACTIVE {
		return
	}

	// Рассчет новой позиции
	nextPosX := state.BallPosX + delta*state.BallSpeedX
	nextPosY := state.BallPosY + delta*state.BallSpeedY

	// Проверка по Y
	if nextPosY < float64(0.0) {
		state.BallSpeedY = -state.BallSpeedY
		nextPosY = state.BallPosY + delta*state.BallSpeedY
	}
	if nextPosY > float64(state.Height) {
		state.BallSpeedY = -state.BallSpeedY
		nextPosY = state.BallPosY + delta*state.BallSpeedY
	}

	// Проверка по X
	const panelWidth float64 = 34.0
	leftBorder := panelWidth
	rightborder := float64(state.Width) - panelWidth
	// Слева
	if nextPosX < leftBorder {
		minY := float64(leftClientState.Y) - float64(leftClientState.Height)/2.0
		maxY := float64(leftClientState.Y) + float64(leftClientState.Height)/2.0

		if (nextPosY > minY) && (nextPosY < maxY) {
			state.BallSpeedX = -state.BallSpeedX
			nextPosX = state.BallPosX + delta*state.BallSpeedX
		} else {
			state.Status = GAME_ROOM_STATUS_COMPLETED
			state.BallSpeedX = 0.0
			state.BallSpeedY = 0.0
			leftClientState.Status = CLIENT_STATUS_FAIL
			rightClientState.Status = CLIENT_STATUS_WIN
		}
	}
	// Справа
	if nextPosX > rightborder {
		minY := float64(rightClientState.Y) - float64(rightClientState.Height)/2.0
		maxY := float64(rightClientState.Y) + float64(rightClientState.Height)/2.0

		if (nextPosY > minY) && (nextPosY < maxY) {
			state.BallSpeedX = -state.BallSpeedX
			nextPosX = state.BallPosX + delta*state.BallSpeedX
		} else {
			state.Status = GAME_ROOM_STATUS_COMPLETED
			state.BallSpeedX = 0.0
			state.BallSpeedY = 0.0
			leftClientState.Status = CLIENT_STATUS_WIN
			rightClientState.Status = CLIENT_STATUS_FAIL
		}
	}

	state.BallPosX = nextPosX
	state.BallPosY = nextPosY

	//log.Printf("delta=%f, x=%f, y=%f, sy=%f, sx=%f\n", delta, state.BallPosX, state.BallPosY, state.BallSpeedX, state.BallSpeedY)
}

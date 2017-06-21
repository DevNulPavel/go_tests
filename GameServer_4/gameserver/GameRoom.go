package gameserver

import (
	"sync/atomic"
	"time"
)

var LAST_ID int32 = 0

type GameRoom struct {
	roomId               int32
	server               *Server
	client1              *Client
	client2              *Client
	gameRoomState        GameRoomState
	isFullCh             chan (chan bool)
	addClientCh          chan *Client
	deleteClientCh       chan *Client
	clientStateUpdatedCh chan bool
	exitLoopCh           chan bool
}

func NewGameRoom(server *Server) *GameRoom {
	newRoomId := atomic.AddInt32(&LAST_ID, 1)

	const width = 600
	const height = 400
	roomState := GameRoomState{
		ID:            newRoomId,
		Status:        ROOM_STATUS_NOT_IN_GAME,
		Width:         width,
		Height:        height,
		BallPosX:      width / 2,
		BallPosY:      height / 2,
		BallPosSpeedX: 4.0,
		BallPosSpeedY: 4.0,
	}

	room := GameRoom{
		roomId:               newRoomId,
		server:               server,
		client1:              nil,
		client2:              nil,
		gameRoomState:        roomState,
		isFullCh:             make(chan (chan bool)),
		addClientCh:          make(chan *Client),
		deleteClientCh:       make(chan *Client),
		clientStateUpdatedCh: make(chan bool),
		exitLoopCh:           make(chan bool),
	}
	return &room
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (room *GameRoom) Exit() {
	room.exitLoopCh <- true
}

func (room *GameRoom) AddClient(client *Client) {
	room.addClientCh <- client
}

func (room *GameRoom) DeleteClient(client *Client) {
	room.deleteClientCh <- client
}

func (room *GameRoom) ClientStateUpdated(client *Client) {
	room.clientStateUpdatedCh <- true
}

func (room *GameRoom) GetIsFull() bool {
	testCh := make(chan bool)
	room.isFullCh <- testCh
	return <-testCh
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (room *GameRoom) sendAllNewState() {
}

func (room *GameRoom) worldTick(delta float64) {
	room.sendAllNewState()
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (room *GameRoom) StartLoop() {
	const updatePeriodMS = 20

	worldUpdateTime := time.Millisecond * updatePeriodMS
	timer := time.NewTimer(worldUpdateTime)
	timer.Stop()
	timerActive := false

	lastTickTime := time.Now()

	for {
		select {
		// Канал добавления нового юзера
		case client := <-room.addClientCh:
			clientAdded := false
			if room.client1 == nil {
				room.client1 = client
				clientAdded = true
			} else if room.client2 == nil {
				room.client2 = client
				clientAdded = true
			}

			canStartGame := (room.client1 != nil) && (room.client2 != nil)
			if clientAdded && canStartGame && !timerActive {
				// Запуск таймера
				timerActive = true
				lastTickTime = time.Now()
				timer.Reset(worldUpdateTime)
			}

		// Канал удаления нового юзера
		case client := <-room.deleteClientCh:
			deleted := false
			if room.client1 == client {
				room.client1.Close()
				room.client1 = nil
				deleted = true
			} else if room.client2 == client {
				room.client2.Close()
				room.client2 = nil
				deleted = true
			}

			if timerActive && deleted {
				timer.Stop()
				timerActive = false
			}

		// Канал таймера
		case <-timer.C:
			delta := time.Now().Sub(lastTickTime).Seconds()
			lastTickTime = time.Now()

			timer.Reset(worldUpdateTime)

			room.worldTick(delta)

		// Канал проверки заполнения
		case resultChannel := <-room.isFullCh:
			if (room.client1 == nil) || (room.client2 == nil) {
				resultChannel <- false
			} else {
				resultChannel <- true
			}

		// Выход из цикла обработки событий
		case <-room.exitLoopCh:
			// Timer
			if timerActive {
				timer.Stop()
			}
			// Clients
			if room.client1 != nil {
				room.client1.Close()
			}
			if room.client2 != nil {
				room.client2.Close()
			}
			// Server
			room.server.DeleteRoom(room)
			return
		}
	}
}

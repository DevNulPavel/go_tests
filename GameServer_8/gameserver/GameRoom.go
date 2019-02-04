package gameserver

import (
	"golang.org/x/net/websocket"
	"sync/atomic"
	"time"
)

const (
	BALL_SPEED = 50.0
)

var LAST_ID uint32 = 0

type GameRoom struct {
	roomId               uint32
	server               *Server
	clientLeft           *Client
	clientRight          *Client
	gameRoomState        GameRoomState
	isFullCh             chan (chan bool)
	addClientByConnCh    chan *websocket.Conn
	deleteClientCh       chan *Client
	clientStateUpdatedCh chan bool
	exitLoopCh           chan bool
}

func NewGameRoom(server *Server) *GameRoom {
	newRoomId := atomic.AddUint32(&LAST_ID, 1)

	const width = 800
	const height = 600
	roomState := GameRoomState{
		ID:         newRoomId,
		Status:     GAME_ROOM_STATUS_ACTIVE,
		Width:      width,
		Height:     height,
		BallPosX:   width / 2,
		BallPosY:   height / 2,
		BallSpeedX: BALL_SPEED,
		BallSpeedY: BALL_SPEED,
	}

	room := GameRoom{
		roomId:               newRoomId,
		server:               server,
		clientLeft:           nil,
		clientRight:          nil,
		gameRoomState:        roomState,
		isFullCh:             make(chan (chan bool)),
		addClientByConnCh:    make(chan *websocket.Conn),
		deleteClientCh:       make(chan *Client),
		clientStateUpdatedCh: make(chan bool),
		exitLoopCh:           make(chan bool),
	}
	return &room
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (room *GameRoom) StartLoop() {
	go room.mainLoop()
}

func (room *GameRoom) Exit() {
	room.exitLoopCh <- true
}

func (room *GameRoom) AddClientForConnection(connection *websocket.Conn) {
	room.addClientByConnCh <- connection
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
	// Создание сообщения
	var message PlayerMessage
	message.Type = PLAYER_MESSAGE_TYPE_WORLD_STATE
	message.RoomState = room.gameRoomState // TODO: Sync?
	if room.clientLeft != nil {
		message.LeftClientState = room.clientLeft.GetCurrentState()
	}
	if room.clientRight != nil {
		message.RightClientState = room.clientRight.GetCurrentState()
	}

	// Отправка сообщения
	if room.clientLeft != nil {
		room.clientLeft.QueueSendGameState(message)
	}
	if room.clientRight != nil {
		room.clientRight.QueueSendGameState(message)
	}
}

func (room *GameRoom) worldTick(delta float64) {
	if (room.clientLeft == nil) || (room.clientRight == nil) {
		return
	}

	WorldTick(delta, &room.gameRoomState, &room.clientLeft.state, &room.clientRight.state)

	room.sendAllNewState()
}

func (room *GameRoom) mainLoop() {
	const updatePeriodMS = 20

	worldUpdateTime := time.Millisecond * updatePeriodMS
	timer := time.NewTimer(worldUpdateTime)
	timer.Stop()
	timerActive := false

	lastTickTime := time.Now()

	for {
		select {
		// Канал добавления нового юзера
		case connection := <-room.addClientByConnCh:
			var client *Client = nil
			clientAdded := false
			if room.clientLeft == nil {
				client = NewClient(connection, CLIENT_TYPE_LEFT, room)
				room.clientLeft = client
				clientAdded = true
			} else if room.clientRight == nil {
				client = NewClient(connection, CLIENT_TYPE_RIGHT, room)
				room.clientRight = client
				clientAdded = true
			}

			if client != nil {
				client.StartLoop()
				client.QueueSendCurrentClientState()
			}

			canStartGame := (room.clientLeft != nil) && (room.clientRight != nil)

			if clientAdded && canStartGame {
				room.gameRoomState.Reset(BALL_SPEED, -BALL_SPEED)
			}

			if clientAdded && canStartGame && !timerActive {
				// Запуск таймера
				timerActive = true
				lastTickTime = time.Now()
				timer.Reset(worldUpdateTime)
			}

		// Канал обновления состояния юзера
		case <-room.clientStateUpdatedCh:
			if timerActive == false {
				room.sendAllNewState()
			}

		// Канал удаления нового юзера
		case client := <-room.deleteClientCh:
			deleted := false
			if room.clientLeft == client {
				room.clientLeft.Close()
				room.clientLeft = nil
				deleted = true
			} else if room.clientRight == client {
				room.clientRight.Close()
				room.clientRight = nil
				deleted = true
			}

			if timerActive && deleted {
				timer.Stop()
				timerActive = false

				room.sendAllNewState()
			}

		// Канал таймера
		case <-timer.C:
			delta := time.Now().Sub(lastTickTime).Seconds()
			lastTickTime = time.Now()

			timer.Reset(worldUpdateTime)

			room.worldTick(delta)

		// Канал проверки заполнения
		case resultChannel := <-room.isFullCh:
			if (room.clientLeft == nil) || (room.clientRight == nil) {
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
			if room.clientLeft != nil {
				room.clientLeft.Close()
			}
			if room.clientRight != nil {
				room.clientRight.Close()
			}
			// Server
			room.server.DeleteRoom(room)
			return
		}
	}
}

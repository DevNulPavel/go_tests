package gameserver

import (
	"log"
	"sync/atomic"
	"time"
)

const (
	BALL_SPEED           = 60.0
	MESSAGES_BUFFER_SIZE = 100
)

var LAST_ID uint32 = 0

type GameRoom struct {
	roomId               uint32
	server               *Server
	clientLeft           *Client
	clientRight          *Client
	gameRoomState        GameRoomState
	isFullAtomic         uint32
	messagesCh           chan ServerMessage
	deleteClientCh       chan *Client
	clientStateUpdatedCh chan bool
	exitLoopCh           chan bool
}

func NewGameRoom(server *Server) *GameRoom {
	newRoomId := atomic.AddUint32(&LAST_ID, 1)

	const width = 600
	const height = 400
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
		isFullAtomic:         0,
		messagesCh:           make(chan ServerMessage, MESSAGES_BUFFER_SIZE),
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

func (room *GameRoom) HandleMessage(message ServerMessage) {
	if len(room.messagesCh) < (MESSAGES_BUFFER_SIZE - 1) {
		room.messagesCh <- message
	} else {
		log.Printf("Messages buffer full for room: %d\n", room.roomId)
	}
}

func (room *GameRoom) DeleteClient(client *Client) {
	room.deleteClientCh <- client
}

func (room *GameRoom) ClientStateUpdated(client *Client) {
	room.clientStateUpdatedCh <- true
}

func (room *GameRoom) GetIsFull() bool {
	isFull := bool(atomic.LoadUint32(&room.isFullAtomic) > 0)
	return isFull
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

func (room *GameRoom) sendAllNewState() {
	gameStateBytes, err := room.gameRoomState.ConvertToBytes()
	if err != nil {
		return
	}

	if room.clientLeft != nil {
		room.clientLeft.QueueSendGameState(gameStateBytes)
	}
	if room.clientRight != nil {
		room.clientRight.QueueSendGameState(gameStateBytes)
	}
}

func (room *GameRoom) worldTick(delta float64) {
	if (room.clientLeft == nil) || (room.clientRight == nil) {
		return
	}

    if room.clientLeft.IsReady() && room.clientRight.IsReady() {
        room.gameRoomState.clientLeftState = room.clientLeft.GetCurrentState()
        room.gameRoomState.clientRightState = room.clientRight.GetCurrentState()

        room.gameRoomState.WorldTick(delta)

        room.sendAllNewState()
    }
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
		case message := <-room.messagesCh:
			// Определяем, для какого клиента это сообщение
			var foundClient *Client = nil
			clientFound := false
			if room.clientLeft.address == message.address { // TODO: ???
				clientFound = true
				foundClient = room.clientLeft
			} else if room.clientRight.address == message.address { // TODO: ???
				clientFound = true
				foundClient = room.clientRight
			}

			if clientFound {
                foundClient.HandleIncomingMessage(message.data)

			} else {
                // Создаем клиента
				var newClient *Client = nil
				clientAdded := false
				if room.clientLeft == nil {
					newClient = NewClient(message.address, CLIENT_TYPE_LEFT, room)
					room.clientLeft = newClient
					clientAdded = true
				} else if room.clientRight == nil {
					newClient = NewClient(message.address, CLIENT_TYPE_RIGHT, room)
					room.clientRight = newClient
					clientAdded = true
				}

				// Инициализация клиента
				if newClient != nil {
					newClient.StartLoop()
                    newClient.QueueSendCurrentClientState()
				}

				canStartGame := (room.clientLeft != nil) && (room.clientRight != nil)

				// Сброс игры
				if clientAdded && canStartGame {
					room.gameRoomState.Reset(BALL_SPEED, -BALL_SPEED)
				}

				// Запуск таймера
				if clientAdded && canStartGame && !timerActive {
					timerActive = true
					lastTickTime = time.Now()
					timer.Reset(worldUpdateTime)
				}
			}

			if (room.clientLeft != nil) && (room.clientRight != nil) {
				atomic.StoreUint32(&room.isFullAtomic, 1)
			} else {
				atomic.StoreUint32(&room.isFullAtomic, 0)
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
				room.clientLeft = nil
				deleted = true
			} else if room.clientRight == client {
				room.clientRight = nil
				deleted = true
			}

			if timerActive && deleted {
				timer.Stop()
				timerActive = false

				room.sendAllNewState()
			}

			// Изменяем статус заполненности клиента
			if (room.clientLeft != nil) && (room.clientRight != nil) {
				atomic.StoreUint32(&room.isFullAtomic, 1)
			} else {
				atomic.StoreUint32(&room.isFullAtomic, 0)
			}

		// Канал таймера
		case <-timer.C:
			delta := time.Now().Sub(lastTickTime).Seconds()
			lastTickTime = time.Now()
			timer.Reset(worldUpdateTime)

			room.worldTick(delta)

		// Выход из цикла обработки событий
		case <-room.exitLoopCh:
			// Timer
			if timerActive {
				timer.Stop()
			}
			// Clients
			if room.clientLeft != nil {
				room.server.DeleteRoomForAddress(room.clientLeft.address)
			}
			if room.clientRight != nil {
				room.server.DeleteRoomForAddress(room.clientRight.address)
			}
			return
		}
	}
}

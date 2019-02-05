package gameserver

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
	"sync"
	"sync/atomic"
)

const UPDATE_QUEUE_SIZE = 100

// Variables
var MAX_ID uint32 = 0

// Client ... Структура клиента
type Client struct {
	gameRoom     *GameRoom
	socket       *WebSocket
	id           uint32
	mutex        sync.RWMutex
	state        ClientState
	uploadDataCh chan ToPlayerMessage
	exitReadCh   chan bool
	exitWriteCh  chan bool
}

// NewClient ... Конструктор
func NewClient(connection *WebSocket, clientType uint8, gameRoom *GameRoom) *Client {
	if connection == nil {
		panic("No connection")
	}
	if gameRoom == nil {
		panic("No game server")
	}

	// Увеличиваем id
	curId := atomic.AddUint32(&MAX_ID, 1)

	// Конструируем клиента и его каналы
	clientState := ClientState{
		ID:     curId,
		Type:   clientType,
		Y:      100,
		Height: 100,
		Status: CLIENT_STATUS_IN_GAME,
	}
	uploadDataCh := make(chan ToPlayerMessage, UPDATE_QUEUE_SIZE) // В канале апдейтов может накапливаться максимум 1000 апдейтов
	exitReadCh := make(chan bool, 1)
	exitWriteCh := make(chan bool, 1)

	return &Client{
		gameRoom:     gameRoom,
		socket:       connection,
		id:           curId,
		mutex:        sync.RWMutex{},
		state:        clientState,
		uploadDataCh: uploadDataCh,
		exitReadCh:   exitReadCh,
		exitWriteCh:  exitWriteCh,
	}
}

func (client *Client) Close() {
	client.socket.Close()
	log.Printf("Connection closed for client %d", client.id)
}

func (client *Client) GetCurrentState() ClientState {
	client.mutex.Lock()
	stateCopy := client.state
	client.mutex.Unlock()
	return stateCopy
}

// QueueSendAllStates ... Пишем сообщение клиенту
func (client *Client) QueueSendGameState(gameState ToPlayerMessage) {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadDataCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for client %d", client.id)
		// TODO: Ждем таймаут??
		//client.server.DeleteClient(client)
		//client.exitWriteChannel <- true
		//client.exitReadChannel <- true
		return
	} else {
		client.uploadDataCh <- gameState
	}
}

// QueueSendCurrentClientState ... Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadDataCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for client %d", client.id)
		// TODO: Ждем таймаут??
		//client.server.DeleteClient(client)
		//client.exitWriteChannel <- true
		//client.exitReadChannel <- true
		return
	} else {
		state := client.GetCurrentState()

		var message ToPlayerMessage
		message.Type = PLAYER_MESSAGE_TYPE_PLAYER_INIT
		if state.Type == CLIENT_TYPE_LEFT {
			message.LeftClientState = state
		} else if state.Type == CLIENT_TYPE_RIGHT {
			message.RightClientState = state
		}

		client.uploadDataCh <- message
	}
}

// Запускаем ожидания записи и чтения
func (client *Client) StartLoop() {
	go client.loopWrite() // в отдельной горутине
	go client.loopRead()
}

func (client *Client) StopLoop() {
	client.exitWriteCh <- true
	client.exitReadCh <- true
	client.Close()
}

// Ожидание записи
func (client *Client) loopWrite() {
	//log.Println("StartSyncListenLoop write to client:", client.id)
	for {
		select {
		// Отправка записи клиенту
		case message := <-client.uploadDataCh:
			// С помощью библиотеки websocket производим кодирование сообщения и отправку на сокет
			err := websocket.JSON.Send(client.socket.connection, message) // Функция синхронная
			if err != nil {
				client.Close()
				client.gameRoom.DeleteClient(client)
				client.exitReadCh <- true // Выход из loopRead
				log.Printf("LoopWrite exit by ERROR (%s), clientId = %d\n", err, client.id)
				return
			}
		// Получение флага выхода из функции
		case <-client.exitWriteCh:
			log.Println("LoopWrite exit, clientId =", client.id)
			return
		}
	}
}

// Ожидание чтения
func (client *Client) loopRead() {
	//log.Println("Listening read from client")
	for {
		select {
		// Получение флага выхода
		case <-client.exitReadCh:
			log.Println("LoopRead exit, clientId =", client.id)
			return

		// Чтение данных из сокета
		default:
			// Выполняем получение данных из вебсокета и декодирование из Json в структуру
			var message FromPlayerMessage
			err := websocket.JSON.Receive(client.socket.connection, &message) // Функция синхронная

			if err == io.EOF {
				// Отправляем в очередь сообщение выхода для loopWrite
				client.gameRoom.DeleteClient(client)
				client.Close()
				client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него
				log.Println("loopRead->exit")
				return
			} else if err != nil {
				// Ошибка
				client.gameRoom.DeleteClient(client)
				client.Close()
				client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него
				log.Printf("loopRead->exit by ERROR (%s), clientId = %d\n", err, client.id)
				return
			} else {
				updated := false

				client.mutex.Lock()
				// Сбновляем состояние данного клиента
				if message.ID == client.state.ID {
					client.state.Y = (int16)(message.Y)
				}
				client.mutex.Unlock()

				if updated == true {
					// Отправляем обновление состояния всем
					client.gameRoom.ClientStateUpdated(client)
				}
			}
		}
	}
}

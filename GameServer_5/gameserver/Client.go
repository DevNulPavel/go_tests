package gameserver

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const UPDATE_QUEUE_SIZE = 100

// Variables
var MAX_ID uint32 = 0

// Client ... Структура клиента
type Client struct {
	gameRoom      *GameRoom
	address       *net.UDPAddr
	id            uint32
	mutex         sync.RWMutex
	isReadyAtomic uint32
	state         ClientState
	inDataCh      chan []byte
	uploadDataCh  chan []byte
	exitReadCh    chan bool
	exitWriteCh   chan bool
}

// NewClient ... Конструктор
func NewClient(address *net.UDPAddr, clientType uint8, gameRoom *GameRoom) *Client {
	if address == nil {
		panic("No address")
	}
	if gameRoom == nil {
		panic("No game room")
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
	inDataCh := make(chan []byte, UPDATE_QUEUE_SIZE)
	uploadDataCh := make(chan []byte, UPDATE_QUEUE_SIZE) // В канале апдейтов может накапливаться максимум 1000 апдейтов
	exitReadCh := make(chan bool, 1)
	exitWriteCh := make(chan bool, 1)

	return &Client{
		gameRoom:      gameRoom,
		address:       address,
		id:            curId,
		mutex:         sync.RWMutex{},
		isReadyAtomic: 0,
		state:         clientState,
		inDataCh:      inDataCh,
		uploadDataCh:  uploadDataCh,
		exitReadCh:    exitReadCh,
		exitWriteCh:   exitWriteCh,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (client *Client) GetCurrentState() ClientState {
	client.mutex.Lock()
	stateCopy := client.state
	client.mutex.Unlock()

	return stateCopy
}

// Обрабатываем входящее соединение
func (client *Client) HandleIncomingMessage(data []byte) {
	if len(client.inDataCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Incoming queue full for client %d", client.id)
		return
	} else {
		client.inDataCh <- data
	}
}

// Пишем сообщение клиенту c игровым состоянием
func (client *Client) QueueSendGameState(stateData []byte) {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadDataCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Upload queue full for client %d", client.id)
		return
	} else {
		client.uploadDataCh <- stateData
	}
}

// Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadDataCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Upload queue for client %d", client.id)
		return
	} else {
		client.mutex.RLock()
		currentUserStateCopy := client.state
		client.mutex.RUnlock()

		data, err := currentUserStateCopy.ConvertToBytes()
		if err != nil {
			log.Printf("State upload error for client %d: %s\n", client.id, err)
		}

		client.uploadDataCh <- data
	}
}

func (client *Client) IsReady() bool {
	isReady := bool(atomic.LoadUint32(&client.isReadyAtomic) > 0)
	return isReady
}

// Запускаем ожидания записи и чтения (блокирующая функция)
func (client *Client) StartLoop() {
	go client.loopWrite() // в отдельной горутине
	go client.loopRead()
}

func (client *Client) StopLoop() {
	client.exitWriteCh <- true
	client.exitReadCh <- true
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Ожидание записи
func (client *Client) loopWrite() {
	//log.Println("StartSyncListenLoop write to client:", client.id)
	for {
		select {
		// Отправка записи клиенту
		case payloadData := <-client.uploadDataCh:
			// Отсылаем
			message := ServerMessage{address: client.address, data: payloadData}
			client.gameRoom.server.SendMessage(message)

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

	// Специальный таймер, который отслеживает долгое отсутствие входящих данных,
	// если данные долго не приходят - считаем клиента отвалившимся
	const checkPeriodMS = 4000
	checkTime := time.Millisecond * checkPeriodMS
	timer := time.NewTimer(checkTime)
	timer.Stop()

	for {
		select {
		// Читаем
		case data := <-client.inDataCh:
			// Декодирование
			state, err := NewClientState(data)

			if (err == nil) && (state.ID > 0) {
				// Сбновляем состояние данного клиента
				client.mutex.Lock()
				client.state.Y = state.Y
				client.mutex.Unlock()

				// Отправляем обновление состояния всем
				client.gameRoom.ClientStateUpdated(client)

				// Выставляем флаг готовности
				atomic.StoreUint32(&client.isReadyAtomic, 1)
			}

			// Сброс ожидания
			timer.Reset(checkTime)

		// Слишком долго ждали ответа - выходим
		case <-timer.C:
			timer.Stop()
			client.exitWriteCh <- true
			log.Println("LoopRead exit by timeout, clientId =", client.id)
			return

		// Получение флага выхода
		case <-client.exitReadCh:
			timer.Stop()
			log.Println("LoopRead exit, clientId =", client.id)
			return
		}
	}
}

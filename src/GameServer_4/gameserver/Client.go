package gameserver

import (
	"encoding/binary"
	"io"
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
	gameRoom     *GameRoom
	connection   *net.TCPConn
	id           uint32
	state        ClientState
	mutex        sync.RWMutex
	uploadDataCh chan []byte
	exitReadCh   chan bool
	exitWriteCh  chan bool
}

// NewClient ... Конструктор
func NewClient(connection *net.TCPConn, clientType uint8, gameRoom *GameRoom) *Client {
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
	uploadDataCh := make(chan []byte, UPDATE_QUEUE_SIZE) // В канале апдейтов может накапливаться максимум 1000 апдейтов
	exitReadCh := make(chan bool, 1)
	exitWriteCh := make(chan bool, 1)

	return &Client{
		gameRoom:     gameRoom,
		connection:   connection,
		id:           curId,
		state:        clientState,
		mutex:        sync.RWMutex{},
		uploadDataCh: uploadDataCh,
		exitReadCh:   exitReadCh,
		exitWriteCh:  exitWriteCh,
	}
}

func (client *Client) Close() {
	client.connection.Close()
	log.Printf("Connection closed for client %d", client.id)
}

func (client *Client) GetCurrentState() ClientState {
	client.mutex.Lock()
	stateCopy := client.state
	client.mutex.Unlock()
	return stateCopy
}

// QueueSendAllStates ... Пишем сообщение клиенту
func (client *Client) QueueSendGameState(gameState []byte) {
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

// Запускаем ожидания записи и чтения (блокирующая функция)
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
		case payloadData := <-client.uploadDataCh:
			// Размер данных
			dataBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(dataBytes, uint32(len(payloadData)))

			// Данные для отправки
			sendData := append(dataBytes, payloadData...)

			// Таймаут
			timeout := time.Now().Add(30 * time.Second)
			(*client.connection).SetWriteDeadline(timeout)

			// Отсылаем
			writenCount, err := (*client.connection).Write(sendData)
			if (err != nil) || (writenCount < len(sendData)) {
				client.Close()
				client.gameRoom.DeleteClient(client)
				client.exitReadCh <- true // Выход из loopRead
				if err != nil {
					log.Printf("LoopWrite exit by ERROR (%s), clientId = %d\n", err, client.id)
				} else if writenCount < len(sendData) {
					log.Printf("LoopWrite exit by less bytes - %d from %d, clientId = %d\n", writenCount, len(sendData), client.id)
				}
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
			// Ожидается, что за 10 минут что-то придет, иначе - это отвал
			timeout := time.Now().Add(10 * time.Minute)
			(*client.connection).SetReadDeadline(timeout)

			// Размер данных
			dataSizeBytes := make([]byte, 4)
			readCount, err := (*client.connection).Read(dataSizeBytes)

			// Ошибка чтения данных
			if (err != nil) || (readCount < 4) {
				client.gameRoom.DeleteClient(client)
				client.Close()
				client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него

				if err == io.EOF {
					log.Printf("LoopRead exit by disconnect, clientId = %d\n", client.id)
				} else if err != nil {
					log.Printf("LoopRead exit by ERROR (%s), clientId = %d\n", err, client.id)
				} else if readCount < 4 {
					log.Printf("LoopRead exit - read less 8 bytes (%d bytes), clientId = %d\n", readCount, client.id)
				}
				return
			}
			dataSize := binary.BigEndian.Uint32(dataSizeBytes)

			// Ожидается, что будут данные в течении 20 секунд - иначе отвал
			timeout = time.Now().Add(20 * time.Second)
			(*client.connection).SetReadDeadline(timeout)

			// Данные
			data := make([]byte, dataSize)
			readCount, err = (*client.connection).Read(data)

			// Ошибка чтения данных
			if (err != nil) || (uint32(readCount) < dataSize) {
				client.gameRoom.DeleteClient(client)
				client.Close()
				client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него

				if err == io.EOF {
					log.Printf("LoopRead exit by disconnect, clientId = %d\n", client.id)
				} else if err != nil {
					log.Printf("LoopRead exit by ERROR (%s), clientId = %d\n", err, client.id)
				} else if uint32(readCount) < dataSize {
					log.Printf("LoopRead exit - read less %d bytes (%d bytes), clientId = %d\n", dataSize, readCount, client.id)
				}
				return
			}

			if readCount > 0 {
				// Декодирование
				state, err := NewClientState(data)

				if (err == nil) && (state.ID > 0) {
					// Сбновляем состояние данного клиента
					client.mutex.Lock()
					client.state.Y = state.Y
					client.mutex.Unlock()

					// Отправляем обновление состояния всем
					client.gameRoom.ClientStateUpdated(client)
				}
			}
		}
	}
}

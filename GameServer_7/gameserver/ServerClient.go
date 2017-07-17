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

// Структура клиента
type ServerClient struct {
	serverArena  *ServerArena
	connection   *net.TCPConn
	id           uint32
	stateValid   bool
	state        ServerClientState
	mutex        sync.RWMutex
	uploadDataCh chan []byte
	exitReadCh   chan bool
	exitWriteCh  chan bool
}

// Конструктор
func NewClient(connection *net.TCPConn, serverArena *ServerArena) *ServerClient {
	if connection == nil {
		panic("No connection")
	}
	if serverArena == nil {
		panic("No game server")
	}

	// Увеличиваем id
	curId := atomic.AddUint32(&MAX_ID, 1)

	// Конструируем клиента и его каналы
	clientState := NewServerClientState(curId)
	clientState.Status = CLIENT_STATUS_IN_GAME

	return &ServerClient{
		serverArena:  serverArena,
		connection:   connection,
		id:           curId,
		stateValid:   false,
		state:        clientState,
		mutex:        sync.RWMutex{},
		uploadDataCh: make(chan []byte, UPDATE_QUEUE_SIZE), // В канале апдейтов может накапливаться максимум 1000 апдейтов
		exitReadCh:   make(chan bool, 1),
		exitWriteCh:  make(chan bool, 1),
	}
}

func (client *ServerClient) Close() {
	client.connection.Close()
	log.Printf("Connection closed for client %d", client.id)
}

func (client *ServerClient) IsValidState() bool {
	client.mutex.RLock()
	validCopy := client.stateValid
	client.mutex.RUnlock()
	return validCopy
}

func (client *ServerClient) GetCurrentState(withReset bool) ServerClientState {
	client.mutex.RLock()
	stateCopy := client.state
	if withReset {
		client.state.Duration = 0.0
	}
	client.mutex.RUnlock()
	return stateCopy
}

func (client *ServerClient) GetCurrentStateData(withReset bool) []byte {
	client.mutex.RLock()
	stateData, err := client.state.ToBytes()
	if withReset {
		client.state.Duration = 0.0
	}
	client.mutex.RUnlock()

	if err != nil {
		log.Printf("State data make error for client %d: %s\n", client.id, err)
	}

	return stateData
}

// Пишем сообщение клиенту
func (client *ServerClient) QueueSendData(data []byte) {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadDataCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for client %d", client.id)
		return
	} else {
		client.uploadDataCh <- data
	}
}

// Пишем сообщение клиенту только с его состоянием
func (client *ServerClient) QueueSendCurrentClientState() {
	data := client.GetCurrentStateData(false)
	client.QueueSendData(data)
}

// Запускаем ожидания записи и чтения (блокирующая функция)
func (client *ServerClient) StartLoop() {
	go client.loopWrite() // в отдельной горутине
	go client.loopRead()
}

func (client *ServerClient) StopLoop() {
	client.exitWriteCh <- true
	client.exitReadCh <- true
	client.Close()
}

// Ожидание записи
func (client *ServerClient) loopWrite() {
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
				client.serverArena.DeleteClient(client)
				client.Close()
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
func (client *ServerClient) loopRead() {
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
				client.serverArena.DeleteClient(client)
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

			// Ожидается, что будут данные в течении 30 секунд - иначе отвал
			timeout = time.Now().Add(30 * time.Second)
			(*client.connection).SetReadDeadline(timeout)

			// Данные
			data := make([]byte, dataSize)
			readCount, err = (*client.connection).Read(data)

			// Ошибка чтения данных
			if (err != nil) || (uint32(readCount) < dataSize) {
				client.serverArena.DeleteClient(client)
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
				command, err := NewClientCommand(data)
				if err != nil {
					client.serverArena.DeleteClient(client)
					client.Close()
					client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него

					log.Printf("Error read command, clientId = %d, command = %s\n", client.id, string(data))
					return
				}

				client.mutex.Lock()
				client.stateValid = true
				client.state.RotationX = command.RotationX
                client.state.RotationY = command.RotationY
                client.state.RotationZ = command.RotationZ
				client.state.X = command.X
				client.state.Y = command.Y
				client.state.VX = command.VX
				client.state.VY = command.VY
				client.state.Duration = command.Duration
				client.state.VisualState = command.VisualState
				client.state.AnimName = command.AnimName
				client.state.StartSkillName = command.StartSkillName
				client.mutex.Unlock()

				// ставим в очередь обновление
				client.serverArena.ClientStateUpdated(client, false)
			}
		}
	}
}

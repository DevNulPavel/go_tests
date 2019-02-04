package gameserver

import (
	"container/list"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
)

const UPDATE_QUEUE_SIZE = 100

// Variables
var MAX_ID uint32 = 0

type ClientShootUpdateResult struct {
	clientID uint32  `json:"id"`
	bullet   *Bullet `json:"bullets"`
	client   *Client `json:"clients"`
}

type ClientPositionInfo struct {
	clientID uint32  `json:"id"`
	x        int16   `json:"x"`
	y        int16   `json:"y"`
	size     uint8   `json:"size"`
	client   *Client `json:"client"`
}

// Структура клиента
type Client struct {
	server        *Server
	connection    *websocket.Conn
	id            uint32
	mutex         sync.RWMutex
	state         *ClientState
	uploadStateCh chan GameStateMessage
	exitReadCh    chan bool
	exitWriteCh   chan bool
}

// NewClient ... Конструктор
func NewClient(connection *websocket.Conn, server *Server) *Client {
	if connection == nil {
		panic("No connection")
	}
	if server == nil {
		panic("No game server")
	}

	// Увеличиваем id
	curId := atomic.AddUint32(&MAX_ID, 1)

	// Состояние для отгрузки клиенту
	clientState := NewState(curId, int16(rand.Int()%200+100), int16(rand.Int()%200+100))

	// Конструируем клиента и его каналы
	uploadDataCh := make(chan GameStateMessage, UPDATE_QUEUE_SIZE) // В канале апдейтов может накапливаться максимум 1000 апдейтов
	exitReadCh := make(chan bool, 1)
	exitWriteCh := make(chan bool, 1)

	return &Client{
		server:        server,
		connection:    connection,
		id:            curId,
		mutex:         sync.RWMutex{},
		state:         clientState,
		uploadStateCh: uploadDataCh,
		exitReadCh:    exitReadCh,
		exitWriteCh:   exitWriteCh,
	}
}

func (client *Client) GetCurrentState() ClientState {
	client.mutex.RLock()
	stateData := *client.state
	client.mutex.RUnlock()
	return stateData
}

func (client *Client) UpdateCurrentState(delta float64, worldSizeX, worldSizeY uint16) (bool, []ClientShootUpdateResult, ClientPositionInfo) {
	maxX := float64(worldSizeX)
	maxY := float64(worldSizeY)

	hasNews := false
	bullets := []ClientShootUpdateResult{}
	deleteBullets := []*list.Element{}

	client.mutex.Lock()

	// Position info
	positionInfo := ClientPositionInfo{
		clientID: client.id,
		x:        client.state.X,
		y:        client.state.Y,
		size:     client.state.Size,
		client:   client,
	}

	// Bullets
	if client.state.Status != CLIENT_STATUS_FAIL {
		// обновление позиций пуль с удалением старых
		it := client.state.Bullets.Front()
		for i := 0; i < client.state.Bullets.Len(); i++ {

			bul := it.Value.(*Bullet)
			bul.WorldTick(delta)

			// Проверяем пулю на выход из карты
			if (bul.X > 0) && (bul.X < maxX) && (bul.Y > 0) && (bul.Y < maxY) {
				clientBulletPair := ClientShootUpdateResult{
					clientID: client.id,
					client:   client,
					bullet:   bul,
				}
				bullets = append(bullets, clientBulletPair)
				hasNews = true
			} else {
				deleteBullets = append(deleteBullets, it)
				hasNews = true
			}

			it = it.Next()
		}
		// Удаление старых
		for _, it := range deleteBullets {
			client.state.Bullets.Remove(it)
		}
	}
	client.mutex.Unlock()
	return hasNews, bullets, positionInfo
}

func (client *Client) IncreaseFrag(bullet *Bullet) {
	client.mutex.Lock()
	{
		// Frag increase
		client.state.Frags++
		// Delete bullet
		it := client.state.Bullets.Front()
		for i := 0; i < client.state.Bullets.Len(); i++ {
			bul := it.Value.(*Bullet)
			if bul.ID == bullet.ID {
				client.state.Bullets.Remove(it)
				break
			}
			it = it.Next()
		}
	}
	client.mutex.Unlock()
}

func (client *Client) SetFailStatus() {
	client.mutex.Lock()
	client.state.Status = CLIENT_STATUS_FAIL
	client.mutex.Unlock()
}

// Пишем сообщение клиенту
func (client *Client) QueueSendGameState(gameStateMessage GameStateMessage) {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadStateCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for state %d", client.id)
		return
	} else {
		client.uploadStateCh <- gameStateMessage
	}
}

// Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
	// Если очередь превышена - считаем, что юзер отвалился
	if len(client.uploadStateCh)+1 > UPDATE_QUEUE_SIZE {
		log.Printf("Queue full for state %d", client.id)
		return
	} else {
		var message GameStateMessage

		// Type
		message.Type = GAME_STATE_MESSAGE_INIT_PLAYER

		// World info
		message.WorldData = *client.server.worldInfo

		// Clients data
		message.ClienStates = append(message.ClienStates, client.GetCurrentState())

		client.uploadStateCh <- message
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
	client.connection.Close()
}

// Ожидание записи
func (client *Client) loopWrite() {
	//log.Println("StartSyncListenLoop write to state:", state.id)
	for {
		select {
		// Отправка записи клиенту
		case worldStateInfo := <-client.uploadStateCh:
			// С помощью библиотеки websocket производим кодирование сообщения и отправку на сокет
			err := websocket.JSON.Send(client.connection, worldStateInfo) // Функция синхронная
			if err != nil {
				log.Println("Error:", err.Error())
				client.server.DeleteClient(client)
				client.exitReadCh <- true // для метода loopRead, чтобы выйти из него
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
	//log.Println("Listening read from state")
	for {
		select {
		// Получение флага выхода
		case <-client.exitReadCh:
			log.Println("LoopRead exit, clientId =", client.id)
			return

		// Чтение данных из сокета
		default:
			// Выполняем получение данных из вебсокета и декодирование из Json в структуру
			var command ClientCommand
			err := websocket.JSON.Receive(client.connection, &command) // Функция синхронная

			if err == io.EOF {
				// Клиент отключился
				client.server.DeleteClient(client)
				client.connection.Close()
				client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него
				log.Printf("LoopRead exit by disconnect, clientId = %d\n", client.id)
				return
			} else if err != nil {
				// Произошла ошибка
				client.server.DeleteClient(client)
				client.connection.Close()
				client.exitWriteCh <- true // для метода loopWrite, чтобы выйти из него
				log.Printf("LoopRead exit by ERROR (%s), clientId = %d\n", err, client.id)
				return
			} else {
				// Обновление состояния
				client.mutex.Lock()
				{
					client.state.X = command.X
					client.state.Y = command.Y
					client.state.Angle = command.Angle
					// Дополнительные действия
					switch command.Type {
					case CLIENT_COMMAND_TYPE_MOVE:
						break
					case CLIENT_COMMAND_TYPE_SHOOT:
						// Выстрел, создаем новую пулю
						bullet := NewBullet(client.state.X, client.state.Y, int16(client.state.Size)/2, client.state.Angle)
						client.state.Bullets.PushBack(bullet)
						break
					}
				}
				client.mutex.Unlock()

				// Запрашиваем отправку обновления состояния всем
				client.server.QueueSendAllNewState()
			}
		}
	}
}

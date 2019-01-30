package game_server

import "fmt"
import "io"
import "log"
import "sync"
import "sync/atomic"
import "math/rand"
import "golang.org/x/net/websocket"

// Constants
const CHANNEL_BUF_SIZE = 100

// Variables
var maxId uint32 = 1

// Структура клиента
type Client struct {
	id                uint32
	wSocket           *websocket.Conn
	server            *Server
	stateMutex        sync.Mutex
	state             ClienState
	usersStateChannel chan []ClienState
	exitChannel       chan bool
}

// Конструктор
func NewClient(ws *websocket.Conn, server *Server) *Client {
	if ws == nil {
		panic("No socket")
	}
	if server == nil {
		panic("No game_server")
	}

	newID := atomic.AddUint32(&maxId, 1)

	// Конструируем клиента и его каналы
	clientState := ClienState{newID, float64(rand.Int() % 600), float64(rand.Int() % 600)}
	usersStateChannel := make(chan []ClienState, CHANNEL_BUF_SIZE)
	successChannel := make(chan bool)

	return &Client{newID, ws, server, sync.Mutex{}, clientState, usersStateChannel, successChannel}
}

// Пишем сообщение клиенту
func (client *Client) GetState() ClienState {
	client.stateMutex.Lock()
	stateCopy := client.state
	client.stateMutex.Unlock()
	return stateCopy
}

// Пишем сообщение клиенту
func (client *Client) SetState(state ClienState) {
	client.stateMutex.Lock()
	client.state = state
	client.stateMutex.Unlock()
}

// Пишем сообщение клиенту
func (client *Client) QueueSendAllStates(states []ClienState) {
	select {
	// Пишем сообщение в канал
	case client.usersStateChannel <- states:
		{
			//log.Println("Client wrote:", message)
		}
	default:
		{
			// Удаляем клиента если уже нет канала
			client.server.QueueDeleteClient(client)
			err := fmt.Errorf("client %d disconnected", client.id)
			client.server.QueueSendErr(err)
			client.QueueSendExit() // Вызываем выход из горутины loopWrite
			return
		}
	}
}

// Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
	currentUserStateArray := []ClienState{client.state}
	select {
	// Пишем сообщение в канал
	case client.usersStateChannel <- currentUserStateArray:
		{
			//log.Println("Client wrote:", message)
		}
	default:
		{
			// Удаляем клиента если уже нет канала
			client.server.QueueDeleteClient(client)
			err := fmt.Errorf("client %d disconnected", client.id)
			client.server.QueueSendErr(err)
			client.QueueSendExit() // Вызываем выход из горутины loopWrite
			return
		}
	}
}

// Отправляем успешный результат
func (client *Client) QueueSendExit() {
	client.exitChannel <- true
}

// Запускаем ожидания записи и чтения (блокирующая функция)
func (client *Client) SyncListen() {
	go client.loopWrite() // в отдельной горутине
	client.loopRead()
}

// Ожидание записи
func (client *Client) loopWrite() {
	//log.Println("SyncListen write to client")
	for {
		select {
		// Отправка записи клиенту
		case message := <-client.usersStateChannel:
			//log.Println("Send:", message)

			// С помощью библиотеки websocket производим кодирование сообщения и отправку на сокет
			err := websocket.JSON.Send(client.wSocket, message) // Функция синхронная
			if err != nil {
				log.Println("loopWrite->exit")
				client.QueueSendExit() // для метода loopRead, чтобы выйти из него
				return
			}
		// Получение флага выхода из функции
		case <-client.exitChannel:
			client.server.QueueDeleteClient(client)
			log.Println("loopWrite->exit")
			client.QueueSendExit() // для метода loopRead, чтобы выйти из него
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
		case <-client.exitChannel:
			client.server.QueueDeleteClient(client)
			log.Println("loopRead->exit")
			client.QueueSendExit() // для метода loopWrite, чтобы выйти из него
			return

		// Чтение данных из webSocket
		default:
			// Выполняем получение данных из вебсокета и декодирование из Json в структуру
			var state ClienState
			err := websocket.JSON.Receive(client.wSocket, &state) // Функция синхронная

			if err == io.EOF {
				// Отправляем в очередь сообщение выхода для loopWrite
				client.QueueSendExit()
				log.Println("loopRead->exit")
				return
			} else if err != nil {
				// Ошибка
				client.server.QueueSendErr(err)
			} else {
				if state.Id > 0 {
					// Сбновляем состояние данного клиента
					client.SetState(state)
				}

				// Отправляем обновление состояния всем
				//log.Println("Send all:", msg)
				client.server.QueueSendAll()
			}
		}
	}
}

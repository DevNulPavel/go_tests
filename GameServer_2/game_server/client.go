package game_server

import "fmt"
import "io"
import "log"
import (
    "golang.org/x/net/websocket"
    "math/rand"
)

// Constants
const CHANNEL_BUF_SIZE = 100

// Variables
var maxId int = 1

// Структура клиента
type Client struct {
	id                int
	wSocket           *websocket.Conn
	server            *Server
    state             ClienState
	usersStateChannel chan []ClienState
	exitChannel       chan bool
}

// Конструктор
func NewClient(ws *websocket.Conn, server *Server) *Client{
	if ws == nil {
		panic("No socket")
	}
	if server == nil {
		panic("No game_server")
	}

    maxId++

	// Конструируем клиента и его каналы
    clientState := ClienState{maxId, rand.Int() % 600, rand.Int() % 600}
    usersStateChannel := make(chan []ClienState, CHANNEL_BUF_SIZE)
	successChannel := make(chan bool)

	return &Client{maxId, ws, server, clientState, usersStateChannel, successChannel}
}

// Пишем сообщение клиенту
func (client *Client) QueueSendAllStates(states []ClienState) {
	select{
		// Пишем сообщение в канал
        case client.usersStateChannel <- states:{
            //log.Println("Client wrote:", message)
        }
        default: {
        	// Удаляем клиента
        	client.server.QueueDeleteClient(client)
        	err := fmt.Errorf("Client %d disconnected", client.id)
        	client.server.QueueSendErr(err)
            client.QueueSendExit() // Вызываем выход из горутины loopWrite
            return
        }
    }
}

// Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
    currentUserStateArray := []ClienState{client.state}
    select{
    // Пишем сообщение в канал
    case client.usersStateChannel <- currentUserStateArray:{
        //log.Println("Client wrote:", message)
    }
    default: {
        // Удаляем клиента
        client.server.QueueDeleteClient(client)
        err := fmt.Errorf("Client %d disconnected", client.id)
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
				websocket.JSON.Send(client.wSocket, message) // Функция синхронная

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
			case <- client.exitChannel:
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
					return
				} else if err != nil {
					// Ошибка
					client.server.QueueSendErr(err)
				} else {
                    if state.Id > 0 {
                        // Сбновляем состояние данного клиента
                        client.state = state
                    }

					// Отправляем обновление состояния всем
                    //log.Println("Send all:", msg)
					client.server.QueueSendAll()
				}
			}
	}
}





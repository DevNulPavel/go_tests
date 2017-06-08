package game_server

import (
	"net"
	"fmt"
	"io"
	"log"
    "math/rand"
    "encoding/json"
)

// Constants
const CHANNEL_BUF_SIZE = 10

// Variables
var maxId int = 1

// Структура клиента
type Client struct {
	server            *Server
    connection        *net.Conn
    id                int
    state             ClienState
    encoder           *json.Encoder
    decoder           *json.Decoder
	usersStateChannel chan []ClienState
	exitChannel       chan bool
}

// Конструктор
func NewClient(connection *net.Conn, server *Server) *Client{
	if server == nil {
		panic("No game_server")
	}

    // Увеличиваем id
    maxId++

	// Конструируем клиента и его каналы
    clientState := ClienState{maxId, float64(rand.Int() % 600), float64(rand.Int() % 600)}
    encoder := json.NewEncoder(*connection)
    decoder := json.NewDecoder(*connection)
    usersStateChannel := make(chan []ClienState, CHANNEL_BUF_SIZE)
	successChannel := make(chan bool)

	return &Client{
        server,
        connection,
        maxId,
        clientState,
        encoder,
        decoder,
        usersStateChannel,
        successChannel,
    }
}

// Пишем сообщение клиенту
func (client *Client) QueueSendAllStates(states []ClienState) {
	select{
		// Пишем сообщение в канал
        case client.usersStateChannel <- states:
            //log.Println("Client wrote:", message)

        // Удаляем клиента раз у нас произошла ошибка какая-то
        default:
        	client.server.QueueDeleteClient(client)
        	err := fmt.Errorf("Client %d disconnected", client.id)
        	client.server.QueueSendErr(err)
            client.QueueSendExit() // Вызываем выход из горутины loopWrite
            return
    }
}

// Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
    currentUserStateArray := []ClienState{client.state}
    select{
    // Пишем сообщение в канал
    case client.usersStateChannel <- currentUserStateArray:
        //log.Println("Client wrote:", message)

    // Удаляем клиента если нельзя отправлять
    default:
        client.server.QueueDeleteClient(client)
        err := fmt.Errorf("Client %d disconnected", client.id)
        client.server.QueueSendErr(err)
        client.QueueSendExit() // Вызываем выход из горутины loopWrite
        return
    }
}

// Отправляем успешный результат
func (client *Client) QueueSendExit() {
	client.exitChannel <- true
}

// Запускаем ожидания записи и чтения (блокирующая функция)
func (client *Client) StartSyncListenLoop() {
	go client.loopWrite() // в отдельной горутине
	client.loopRead()
}

// Ожидание записи
func (client *Client) loopWrite() {
	//log.Println("StartSyncListenLoop write to client")
	for {
		select {
			// Отправка записи клиенту
			case message := <-client.usersStateChannel:
				//log.Println("Send:", message)
                // Синхронная функция отправки данных в сокет
                err := client.encoder.Encode(message)
                if err != nil { // Ошибка
                    client.server.QueueSendErr(err)
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
			case <- client.exitChannel:
				client.server.QueueDeleteClient(client)
                log.Println("loopRead->exit")
                client.QueueSendExit() // для метода loopWrite, чтобы выйти из него
				return

			// Чтение данных из webSocket
			default:
                // Выполняем получение данных из вебсокета и декодирование из Json в структуру
				var state ClienState
				err := client.decoder.Decode(state)

				if err == io.EOF {
                    // Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
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





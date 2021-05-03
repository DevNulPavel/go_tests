package chat

import "fmt"
import "io"
import "log"
import "golang.org/x/net/websocket"

// Constants
const CHANNEL_BUF_SIZE = 100

// Variables
var maxId int = 0

// Структура клиента
type Client struct {
	id          int
	wSocket     *websocket.Conn
	server      *Server
	msgChannel  chan *Message
	exitChannel chan bool
}

// Constructor
func NewClient(ws *websocket.Conn, server *Server) *Client{
	if ws == nil {
		panic("No socket")
	}
	if server == nil {
		panic("No server")
	}

    maxId++

	// Конструируем клиента и его каналы
	messageChannel := make(chan *Message, CHANNEL_BUF_SIZE)
	successChannel := make(chan bool)

	return &Client{maxId, ws, server, messageChannel, successChannel}
}

// Пишем сообщение клиенту
func (client *Client) QueueWrite(message *Message) {
	select{
		// Пишем сообщение в канал
        case client.msgChannel <- message:{
            //log.Println("Client wrote:", message)
        }
        default: {
        	// Удаляем клиента
        	client.server.QueueDeleteClient(client)
        	err := fmt.Errorf("Client %d disconnected", client.id)
        	client.server.QueueSendErr(err)
            client.QueueSendExit() // Вызываем выход из горутины listenWrite
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
	go client.listenWrite() // в отдельной горутине
	client.listenRead()
}

// Ожидание записи
func (client *Client) listenWrite() {
	//log.Println("SyncListen write to client")
	for {
		select {
			// Отправка записи клиенту
			case message := <-client.msgChannel:{
				//log.Println("Send:", message)

                // С помощью библиотеки websocket производим кодирование сообщения и отправку на сокет
				websocket.JSON.Send(client.wSocket, message) // Функция синхронная
            }

            // Получение флага выхода из функции
            case <-client.exitChannel:{
                client.server.QueueDeleteClient(client)
                log.Println("listenWrite->exit")
                client.QueueSendExit() // для метода listenRead, чтобы выйти из него
                return
			}
		}
	}
}

// Ожидание чтения
func (client *Client) listenRead() {
	//log.Println("Listening read from client")
	for {
		select {
			// Получение успешности
			case <- client.exitChannel:
				client.server.QueueDeleteClient(client)
                log.Println("listenRead->exit")
                client.QueueSendExit() // для метода listenWrite, чтобы выйти из него
				return

			// Чтение данных из webSocket
			default:
                // Выполняем получение данных из вебсокета и декодирование из Json в структуру
				var msg Message
				err := websocket.JSON.Receive(client.wSocket, &msg) // Функция синхронная

				if err == io.EOF {
                    // Отправляем в очередь сообщение выхода для listenWrite
					client.QueueSendExit()
					return
				} else if err != nil {
					// Ошибка
					client.server.QueueSendErr(err)
				} else {
					// Отправляем сообщение всем
                    //log.Println("Send all:", msg)
					client.server.QueueSendAll(&msg)
				}
			}
	}
}





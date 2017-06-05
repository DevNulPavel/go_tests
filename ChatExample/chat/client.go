package chat

import "fmt"
import "io"
import "log"
import "golang.org/x/net/websocket"

// Constants
const CHANNEL_BUF_SIZE = 100

// Variables
var maxId int = 0

// Client code
type Client struct {
	id int
	wSocket *websocket.Conn
	server *Server
	msgChannel chan *Message
	successChannel chan bool
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

	// Конструируем клиента
	messageChannel := make(chan *Message, CHANNEL_BUF_SIZE)
	successChannel := make(chan bool)

	return &Client{maxId, ws, server, messageChannel, successChannel}
}

func (this *Client) GetConn() *websocket.Conn  {
	return this.wSocket
}

// Пишем сообщение клиенту
func (this *Client) Write(message *Message) {
	select{
		// Пишем сообщение в канал
        case this.msgChannel <- message:{
            log.Println("Client write:", message)
        }
        default: {
        	// Удаляем клиента
        	this.server.Del(this)
        	error := fmt.Errorf("Client %d disconnected", this.id)
        	this.server.SendErr(error)
        }
    }
}

// Отправляем успешный результат
func (this *Client) SendSuccess() {
	this.successChannel <- true
}

// Запускаем ожидания записи и чтения
func (this *Client) Listen() {
	go this.listenWrite() // в отдельной горутине
	this.listenRead()
}

// Ожидание записи
func (this *Client) listenWrite() {
	log.Println("Listen write to client")

	for {
		select {
			// Отправка записи клиенту
			case message := <-this.msgChannel:{
				log.Println("Send:", message)
				websocket.JSON.Send(this.wSocket, message)
            }

            // Получение успешности
            case <-this.successChannel:{
                this.server.Del(this)
                this.successChannel <- true // для метода listenRead
                return
			}
		}
	}
}

// Ожидание чтения
func (this *Client) listenRead() {
	log.Println("Listening read from client")
	for {
		select {
			// Получение успешности
			case <- this.successChannel:
				this.server.Del(this)
				this.successChannel <- true // для метода listenWrite
				return

			// Чтение данных из webSocket
			default:
				var msg Message
				err := websocket.JSON.Receive(this.wSocket, &msg)

				if err == io.EOF {
					this.successChannel <- true
				} else if err != nil {
					// Ошибка
                    log.Println("Send error")
					this.server.SendErr(err)
				} else {
					// Отправляем сообщение всем
                    log.Println("Send all:", msg)
					this.server.SendAll(&msg)
				}
			}
	}
}





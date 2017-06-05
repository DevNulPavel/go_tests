package chat

import "log"
import "net/http"
import "golang.org/x/net/websocket"

type Server struct {
    pattern string
    messages []*Message
    clients map[int]*Client
    addChannel chan *Client
    deleteChannel chan *Client
    sendAllChannel chan *Message
    successChannel chan bool
    errorChannel chan error
}


// Создание нового сервера
func NewServer(pattern string) *Server {
    messages := []*Message{}
    clients := make(map[int]*Client)
    addChannel := make(chan *Client)
    deleteChannel := make(chan *Client)
    sendAllChannel := make(chan *Message)
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    return &Server{
        pattern,
        messages,
        clients,
        addChannel,
        deleteChannel,
        sendAllChannel,
        successChannel,
        errorChannel,
    }
}

// Добавление клиента
func (this *Server) Add(c *Client) {
    this.addChannel <- c
}

func (this *Server) Del(c *Client) {
    this.deleteChannel <- c
}

func (this *Server) SendAll(msg *Message) {
    this.sendAllChannel <- msg
}

func (this *Server) SendDone() {
    this.successChannel <- true
}

func (this *Server) SendErr(err error) {
    this.errorChannel <- err
}

// Отправка всех последних сообщений
func (this *Server) sendPastMessages(c *Client) {
    for _, msg := range this.messages {
        c.Write(msg)
    }
}

// Отправить всем сообщение
func (this *Server) sendAll(msg *Message) {
    for _, c := range this.clients {
        c.Write(msg)
    }
}

// Listen and serve.
// It serves client connection and broadcast request.
func (this *Server) Listen() {

    log.Println("Listening server...")

    // Обработчик подключения WebSocket
    onConnected := func(ws *websocket.Conn) {
        // Функция автоматического закрытия
        defer func() {
            err := ws.Close()
            if err != nil {
                this.errorChannel <- err
            }
        }()

        // Создание нового клиента
        client := NewClient(ws, this)
        this.Add(client)
        client.Listen()
    }
    http.Handle(this.pattern, websocket.Handler(onConnected))
    log.Println("Created handler")

    // Обработка каналов
    for {
        select {
            // Добавление нового юзера
            case c := <-this.addChannel:
                this.clients[c.id] = c
                log.Println("Added new client: now", len(this.clients), "clients connected.")
                this.sendPastMessages(c)

            // Удаление клиента
            case c := <-this.deleteChannel:
                log.Println("Delete client")
                delete(this.clients, c.id)

            // Отправка сообщения всем клиентам
            case msg := <-this.sendAllChannel:
                log.Println("Send all:", msg)
                this.messages = append(this.messages, msg)
                this.sendAll(msg)

            // Была какая-то ошибка
            case err := <-this.errorChannel:
                log.Println("Error:", err.Error())

            // Завершение работы
            case <-this.successChannel:
                return
        }
    }
}
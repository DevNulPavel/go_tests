package chat

import "log"
import "net/http"
import "golang.org/x/net/websocket"

type Server struct {
    messages       []*Message
    clients        map[int]*Client
    addChannel     chan *Client
    deleteChannel  chan *Client
    sendAllChannel chan *Message
    exitChannel    chan bool
    errorChannel   chan error
}


// Создание нового сервера
func NewServer() *Server {
    messages := []*Message{}
    clients := make(map[int]*Client)
    addChannel := make(chan *Client)
    deleteChannel := make(chan *Client)
    sendAllChannel := make(chan *Message)
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    return &Server{
        messages,
        clients,
        addChannel,
        deleteChannel,
        sendAllChannel,
        successChannel,
        errorChannel,
    }
}

// Добавление клиента через очередь
func (server *Server) QueueAddNewClient(c *Client) {
    server.addChannel <- c
}

// Удаление клиента через очередь
func (server *Server) QueueDeleteClient(c *Client) {
    server.deleteChannel <- c
}

// Отправить всем сообщения через очередь
func (server *Server) QueueSendAll(msg *Message) {
    server.sendAllChannel <- msg
}

func (server *Server) QueueExitServer() {
    server.exitChannel <- true
}

func (server *Server) QueueSendErr(err error) {
    server.errorChannel <- err
}

func (server *Server) StartAsyncListen()  {
    go server.mainListenFunction()
}

// Отправка всех последних сообщений
func (server *Server) sendPastMessages(c *Client) {
    for _, msg := range server.messages {
        c.QueueWrite(msg)
    }
}

// Отправить всем сообщение
func (server *Server) sendAll(msg *Message) {
    for _, c := range server.clients {
        c.QueueWrite(msg)
    }
}

func (server *Server) deleteClientFromMap(client *Client)  {
    // Даже если нету клиента в мапе - ничего страшного
    delete(server.clients, client.id)
}

// Основная функция прослушивания
func (server *Server) mainListenFunction() {

    log.Println("Listening server...")

    // Обработчик подключения WebSocket
    onConnectedHandler := func(ws *websocket.Conn) {
        // Функция автоматического закрытия
        defer func() {
            err := ws.Close()
            if err != nil {
                server.errorChannel <- err
            }
        }()

        // Создание нового клиента
        client := NewClient(ws, server)
        server.QueueAddNewClient(client) // выставляем клиента в очередь на добавление
        client.SyncListen()              // блокируется выполнение на данной функции, пока не выйдет клиент
        log.Println("WebSocket connect handler out")
    }
    http.Handle("/websocket", websocket.Handler(onConnectedHandler))
    log.Println("Web socket handler created")

    // Обработка каналов в главной горутине
    for {
        select {
            // Добавление нового юзера
            case c := <-server.addChannel:
                server.clients[c.id] = c
                log.Println("Added new client: now", len(server.clients), "clients connected.")
                server.sendPastMessages(c)

            // Удаление клиента
            case c := <-server.deleteChannel:
                log.Println("Delete client")
                server.deleteClientFromMap(c)

            // Отправка сообщения всем клиентам
            case msg := <-server.sendAllChannel:
                log.Println("Send all:", msg)
                // Пополняем список сообщений на сервере
                server.messages = append(server.messages, msg)
                // Вызываем отправку сообщений всем клиентам
                server.sendAll(msg)

            // Была какая-то ошибка
            case err := <-server.errorChannel:
                log.Println("Error:", err.Error())

            // Завершение работы
            case <-server.exitChannel:
                return
        }
    }
}
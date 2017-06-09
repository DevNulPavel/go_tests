package game_server

import (
    "net"
    "log"
)

type Server struct {
    listener       *net.Listener
    clients        map[int]*Client
    addChannel     chan *Client
    deleteChannel  chan *Client
    sendAllChannel chan bool
    exitChannel    chan bool
    errorChannel   chan error
}


// Создание нового сервера
func NewServer() *Server {
    clients := make(map[int]*Client)
    addChannel := make(chan *Client)
    deleteChannel := make(chan *Client)
    sendAllChannel := make(chan bool)
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    return &Server{
        nil,
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
func (server *Server) QueueSendAll() {
    server.sendAllChannel <- true
}

func (server *Server) QueueExitServer() {
    server.exitChannel <- true
}

func (server *Server) QueueSendErr(err error) {
    server.errorChannel <- err
}

func (server *Server) StartSyncListen()  {
    go server.startAsyncSocketAcceptListener()
    server.mainQueueHandleFunction() // Блокируется выполнение на этом методе
}

// Отправка состояния конкретному клиенту
func (server *Server) sendStateToClient(c *Client) {
    // Создать состояние текущее
    clientStates := []ClienState{}
    for _, client := range server.clients {
        clientStates = append(clientStates, client.state)
    }

    // Отослать юзеру
    c.QueueSendAllStates(clientStates)
}

// Отправить всем сообщение
func (server *Server) sendAllNewState() {
    // Создать состояние текущее
    clientStates := make([]ClienState, 0, len(server.clients))
    for _, client := range server.clients {
        clientStates = append(clientStates, client.state)
    }

    // Отослать всем
    for _, c := range server.clients {
        c.QueueSendAllStates(clientStates)
    }
}

func (server *Server) addClientToMap(client *Client)  {
    server.clients[client.id] = client
}

func (server *Server) deleteClientFromMap(client *Client)  {
    // Даже если нету клиента в мапе - ничего страшного
    delete(server.clients, client.id)
}

// Работа с новыми соединением идет в отдельной горутине
func (server *Server) newAsyncServerConnectionHandler(c *net.Conn) {
    // Создание нового клиента
    client := NewClient(c, server)
    server.QueueAddNewClient(client) // Выставляем клиента в очередь на добавление
    client.StartSyncListenLoop()     // Блокируется выполнение на данной функции, пока не выйдет клиент

    (*c).Close()

    /*
    // Собственные экземпляры декодеров
    decoder := json.NewDecoder(c)
    encoder := json.NewEncoder(c)

    for {
        // Получаем сообщение
        state :=

        err := decoder.Decode(&message)
        if err != nil {
            fmt.Println(err)
            fmt.Println("Client out (error)")
            return
        } else {
            //fmt.Println("Received:", message)

            // Working
            // 1
            param1, ok1 := message.Params["param1"].(string)
            if ok1 {
                param1 = "Text handled for " + param1
                message.Params["param1"] = param1
            }
            // 2
            param2, ok2 := message.Params["param2"].(float64)
            if ok2 {
                param2 *= 10
                message.Params["param2"] = int(param2)
            }
            // 3
            param3, ok3 := message.Params["param3"].(float64)
            if ok3 {
                param3 *= float64(10.0)
                message.Params["param3"] = float64(param3)
            }
            // 4
            param4, ok4 := message.Params["param4"].([]interface{})
            if ok4 {
                for i, value := range param4 {
                    param4[i] = value.(float64) * 10
                }
                message.Params["param4"] = param4
            }
        }

        // Отправляем ответ
        err = encoder.Encode(message)
        if err != nil {
            fmt.Println(err)
            return
        }
    }*/
}

// Обработка входящих подключений
func (server *Server) startAsyncSocketAcceptListener()  {
    // Создание листенера
    createdListener, err := net.Listen("tcp", ":9999")
    if err != nil {
        log.Println("Server listener start error")
        server.QueueExitServer()
        return
    }
    defer createdListener.Close() // TODO: Может быть не нужно? уже есть в exitAsyncSocketListener

    // Сохраняем листенер для возможности закрытия
    server.listener = &createdListener

    for {
        // Для возможности выхода из цикла
        if server.listener == nil {
            return
        }

        // Ожидаем новое подключение
        c, err := (*server.listener).Accept()
        if err != nil {
            server.QueueSendErr(err)
            continue
        }

        // Раз появилось новое соединение - запускаем его в работу с отдельной горутине
        go server.newAsyncServerConnectionHandler(&c)
    }
}

// Выход из обработчика событий
func (server *Server) exitAsyncSocketListener()  {
    if server.listener != nil {
        (*server.listener).Close()
        server.listener = nil
    }
}

// Основная функция прослушивания
func (server *Server) mainQueueHandleFunction() {
    // Обработка каналов в главной горутине
    for {
        select {
            // Добавление нового юзера
            case c := <-server.addChannel:
                server.addClientToMap(c)
                c.QueueSendCurrentClientState() // После добавления на сервере - отправляем клиенту состояние
                server.sendAllNewState()

            // Удаление клиента
            case c := <-server.deleteChannel:
                //log.Println("Delete client")
                server.deleteClientFromMap(c)
                server.sendAllNewState()

            // Отправка сообщения всем клиентам
            case <-server.sendAllChannel:
                // Вызываем отправку сообщений всем клиентам
                server.sendAllNewState()

            // Была какая-то ошибка
            case err := <-server.errorChannel:
                log.Println("Error:", err.Error())

            // Завершение работы
            case <-server.exitChannel:
                server.exitAsyncSocketListener()
                return
        }
    }
}
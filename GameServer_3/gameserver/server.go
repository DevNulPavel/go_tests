package gameserver

import (
	"log"
	"net"
	"time"
)

type Server struct {
	listener        *net.TCPListener
	clients         map[int]*Client
	needSendAllFlag bool
	addChannel      chan *Client
	deleteChannel   chan *Client
	sendAllChannel  chan bool
	exitChannel     chan bool
	listenerExitCh  chan bool
}

// Создание нового сервера
func NewServer() *Server {
	clients := make(map[int]*Client)
	addChannel := make(chan *Client)
	deleteChannel := make(chan *Client)
	sendAllChannel := make(chan bool)
	exitChannel := make(chan bool)
	listenerExitCh := make(chan bool, 1)

	return &Server{
		nil,
		clients,
		false,
		addChannel,
		deleteChannel,
		sendAllChannel,
		exitChannel,
		listenerExitCh,
	}
}

// Добавление клиента через очередь
func (server *Server) AddNewClient(c *Client) {
	server.addChannel <- c
}

// Удаление клиента через очередь
func (server *Server) DeleteClient(c *Client) {
	server.deleteChannel <- c
}

// Отправить всем сообщения через очередь
func (server *Server) SendAll() {
	server.sendAllChannel <- true
}

func (server *Server) ExitServer() {
	server.exitChannel <- true
}

func (server *Server) StartSyncListen() {
	go server.startAsyncSocketAcceptListener()
	server.mainQueueHandleFunction() // Блокируется выполнение на этом методе
}

// Отправка состояния конкретному клиенту
func (server *Server) sendStateToClient(c *Client) {
	// Создать состояние текущее
	clientStates := []ClienState{}
	for _, client := range server.clients {
		state := client.GetCurrentStateWithTimeReset()
		clientStates = append(clientStates, state)
	}

	// Отослать юзеру
	c.QueueSendAllStates(clientStates)
}

// Отправить всем сообщение
func (server *Server) sendAllNewState() {
	// Создать состояние текущее
	clientStates := make([]ClienState, 0, len(server.clients))
	for _, client := range server.clients {
		state := client.GetCurrentStateWithTimeReset()
		clientStates = append(clientStates, state)
	}

	// Отослать всем
	for _, c := range server.clients {
		c.QueueSendAllStates(clientStates)
	}
}

func (server *Server) addClientToMap(client *Client) {
	server.clients[client.id] = client
}

func (server *Server) deleteClientFromMap(client *Client) bool {
	// Даже если нету клиента в мапе - ничего страшного
	if _, exists := server.clients[client.id]; exists {
		delete(server.clients, client.id)
		return true
	}
	return false
}

// Работа с новыми соединением идет в отдельной горутине
func (server *Server) newAsyncServerConnectionHandler(c *net.TCPConn) {
	// Создание нового клиента
	client := NewClient(c, server)
	server.AddNewClient(client)  // Выставляем клиента в очередь на добавление
	client.StartSyncListenLoop() // Блокируется выполнение на данной функции, пока не выйдет клиент
	client.Close()
	log.Printf("Server goroutine closed for client %d\n", client.id)
}

// Обработка входящих подключений
func (server *Server) startAsyncSocketAcceptListener() {
	address, err := net.ResolveTCPAddr("tcp", ":9999")
	if err != nil {
		log.Println("Server address resolve error")
		server.ExitServer()
		return
	}

	// Создание листенера
	createdListener, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Println("Server listener start error")
		server.ExitServer()
		return
	}
	defer createdListener.Close() // TODO: Может быть не нужно? уже есть в exitAsyncSocketListener

	// Сохраняем листенер для возможности закрытия
	server.listener = createdListener

	for {
		select {
		case <-server.listenerExitCh:
			log.Print("Socket listener exit") // Либо наш лиснер закрылся и надо будет выйти из цикла
			return

		default:
			// Ожидаем новое подключение
			c, err := (*server.listener).AcceptTCP()
			if err != nil {
				log.Printf("Accept error: %s\n", err) // Либо наш лиснер закрылся и надо будет выйти из цикла
				continue
			}
			c.SetKeepAlive(true)
			c.SetNoDelay(true)

			// Раз появилось новое соединение - запускаем его в работу с отдельной горутине
			go server.newAsyncServerConnectionHandler(c)
		}
	}
}

// Выход из обработчика событий
func (server *Server) exitAsyncSocketListener() {
	server.listenerExitCh <- true
	(*server.listener).Close()
}

// Основная функция прослушивания
func (server *Server) mainQueueHandleFunction() {
	const updatePeriodMS = 50 // 20 FPS
	worldUpdateTime := time.Millisecond * updatePeriodMS
	ticker := time.NewTicker(worldUpdateTime)
	log.Printf("Server world update period = %dms\n", updatePeriodMS)

	// Обработка каналов в главной горутине
	for {
		select {
		// Добавление нового юзера
		case c := <-server.addChannel:
			log.Printf("Client %d added\n", c.id)
			server.addClientToMap(c)
			c.QueueSendCurrentClientState() // После добавления на сервере - отправляем клиенту состояние
			server.sendAllNewState()

		// Удаление клиента
		case c := <-server.deleteChannel:
			deleted := server.deleteClientFromMap(c)
			if deleted {
				log.Printf("Client %d deleted\n", c.id)
				server.sendAllNewState()
			}

		// Отправка сообщения всем клиентам
		case <-server.sendAllChannel:
			server.needSendAllFlag = true

			// Проверяем необходимость разослать всем новый статус
		case <-ticker.C:
			if server.needSendAllFlag {
				server.sendAllNewState()
				server.needSendAllFlag = false
			}

		// Завершение работы
		case <-server.exitChannel:
			ticker.Stop()
			server.exitAsyncSocketListener()
			return
		}
	}
}

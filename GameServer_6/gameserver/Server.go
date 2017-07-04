package gameserver

import (
	"log"
	"net"
	"time"
    "sync/atomic"
)

type Server struct {
	listener       *net.TCPListener
	listenerExitCh chan bool
	loopExitCh     chan bool
	clients        map[uint32]*Client
	makeClientCh   chan *net.TCPConn
	removeClientCh chan *Client
	needSendAll    uint32
}

// Создание нового сервера
func NewServer() *Server {
	server := Server{
		listener:       nil,
		listenerExitCh: make(chan bool),
		loopExitCh:     make(chan bool),
		clients:        make(map[uint32]*Client),
		makeClientCh:   make(chan *net.TCPConn),
		removeClientCh: make(chan *Client),
		needSendAll:    0,
	}
	return &server
}

func (server *Server) StartServer() {
	server.asyncSocketAcceptListener()
	server.mainLoop()
}

func (server *Server) ExitServer() {
	server.exitAsyncSocketListener()
	server.exitMainLoop()
}

func (server *Server) DeleteClient(client *Client) {
    server.removeClientCh <- client
}

func (server *Server) QueueSendAllNewState() {
    atomic.StoreUint32(&server.needSendAll, 1)
}

////////////////////////////////////////////////////////////////////////////////////////////////

func (server *Server) сreateClient(connection *net.TCPConn) {
	server.makeClientCh <- connection
}

// Обработка входящих подключений
func (server *Server) asyncSocketAcceptListener() {
	address, err := net.ResolveTCPAddr("tcp", ":9999")
	if err != nil {
		log.Println("Server address resolve error")
		server.ExitServer()
		return
	}

	// Создание листенера
	createdListener, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Printf("Server listener start error: %s\n", err)
		server.ExitServer()
		return
	}

	// Сохраняем листенер для возможности закрытия
	server.listener = createdListener

	// Функция-цикл обработки входящих подключений
	loopFunction := func() {
		for {
			select {
			case <-server.listenerExitCh:
				server.listener.Close()
				log.Print("Socket listener exit") // Наш лиснер закрылся и надо будет выйти из цикла
				return

			default:
				// Ожидаем новое подключение
				connection, err := server.listener.AcceptTCP()
				if err != nil {
					log.Printf("Accept error: %s\n", err) // Наш лиснер закрылся и надо будет выйти из цикла
					continue
				}

				log.Printf("Connection accepted\n")

				err = connection.SetKeepAlive(true)
				if err != nil {
					log.Printf("Kep alive set error: %s\n", err)
				}

				err = connection.SetNoDelay(true)
				if err != nil {
					log.Printf("Kep alive set error: %s\n", err)
				}

				// Раз появилось новое соединение - запускаем его в работу с отдельной горутине
				server.сreateClient(connection)
			}
		}
	}

	go loopFunction()
}

// Выход из листенера
func (server *Server) exitAsyncSocketListener() {
	if server.listener != nil {
		server.listener.Close()
	}
	server.listenerExitCh <- true
}

func (server *Server) sendAllGameState() {
}

func (server *Server) worldTick(delta float64) {

}

// Основная функция прослушивания
func (server *Server) mainLoop() {
	loopFunction := func() {
		const updatePeriodMS = time.Millisecond * 20
		timer := time.NewTimer(updatePeriodMS)
		lastTickTime := time.Now()

		for {
			select {
			// Обрабатываем новое подключение
			case connection := <-server.makeClientCh:
				log.Printf("Make state call\n")

				newClient := NewClient(connection, server)
				server.clients[newClient.id] = newClient

				newClient.StartLoop()
				newClient.QueueSendCurrentClientState()

				server.sendAllGameState()

			// Обработка удаления клиентов
			case client := <-server.removeClientCh:
				delete(server.clients, client.id)
				server.sendAllGameState()

			// Основной серверный таймер, который обновляет серверный мир
			case <-timer.C:
				delta := time.Now().Sub(lastTickTime).Seconds()
				lastTickTime = time.Now()

				server.worldTick(delta)

				if atomic.LoadUint32(&server.needSendAll) > 0 {
					atomic.StoreUint32(&server.needSendAll, 0)
					server.sendAllGameState()
				}

			// Завершение работы
			case <-server.loopExitCh:
				timer.Stop()
				log.Print("Main loop exit") // Наш лиснер закрылся и надо будет выйти из цикла
				return
			}
		}
	}
	go loopFunction()
}

func (server *Server) exitMainLoop() {
	server.loopExitCh <- true
}

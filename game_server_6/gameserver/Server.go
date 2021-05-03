package gameserver

import (
	"log"
	"net"
	"sync/atomic"
	"time"
)

type Server struct {
	listener       *net.TCPListener
	listenerExitCh chan bool
	loopExitCh     chan bool
	worldInfo      *WorldInfo
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
		worldInfo:      NewWorldInfo(),
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
    // Clients data
    clientsData := make([]byte, 0)
    for _, client := range server.clients {
		stateCopyBytes, err := client.GetCurrentStateData()
        if err != nil {
            log.Printf("Client state marshal error\n")
            continue
        }
        clientsData = append(clientsData, stateCopyBytes...)
    }

    // World Info
	worldInfoBytes, err := server.worldInfo.ConvertToBytes(clientsData)
	if err != nil {
		log.Printf("World info marshal error\n")
		return
	}

	// Send all
	for _, client := range server.clients {
		client.QueueSendGameState(worldInfoBytes)
	}
}

func (server *Server) worldTick(delta float64) {
	needSendUpdate := false

	updateResults := []ClientShootUpdateResult{}
    clientsPositions := make([]ClientPositionInfo, 0, len(server.clients))
	for _, client := range server.clients {
        // Calling update
		hasNews, bulletsResult, positionInfo := client.UpdateCurrentState(delta, server.worldInfo.SizeX, server.worldInfo.SizeY)

        // Clients positions
        clientsPositions = append(clientsPositions, positionInfo)

        // Bullets results
        if hasNews {
            updateResults = append(updateResults, bulletsResult...)
            needSendUpdate = true
        }
	}

	for i := 0; i < len(updateResults); i++ {
		bulletInfo := updateResults[i]

		for j := 0; j < len(clientsPositions); j++ {
			receiver := clientsPositions[j]
			if bulletInfo.clientID == receiver.clientID {
				continue
			}

			bul := bulletInfo.bullet

			halfSize := int16(receiver.size / 2)
			minX := float64(receiver.x - halfSize)
			maxX := float64(receiver.x + halfSize)
			minY := float64(receiver.y - halfSize)
			maxY := float64(receiver.y + halfSize)

			if (bul.X > minX) && (bul.X < maxX) && (bul.Y > minY) && (bul.Y < maxY) {
                log.Printf("Kill client %d\n", receiver.clientID)
                bulletInfo.client.IncreaseFrag(bulletInfo.bullet)
				receiver.client.SetFailStatus()
				needSendUpdate = true
			}
		}
	}

	if needSendUpdate {
		atomic.StoreUint32(&server.needSendAll, 1)
	}
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
				server.worldInfo.ClientsCount = uint16(len(server.clients))

				newClient.StartLoop()
				newClient.QueueSendCurrentClientState()

				server.sendAllGameState()

			// Обработка удаления клиентов
			case client := <-server.removeClientCh:
				log.Printf("Delete client call\n")

				delete(server.clients, client.id)
				server.worldInfo.ClientsCount = uint16(len(server.clients))

				server.sendAllGameState()

			// Основной серверный таймер, который обновляет серверный мир
			case <-timer.C:
                timer.Reset(updatePeriodMS)
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

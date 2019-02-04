package gameserver

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type Server struct {
	loopExitCh     chan bool
	worldInfo      *WorldInfo
	clients        map[uint32]*Client
	makeClientCh   chan *websocket.Conn
	removeClientCh chan *Client
	needSendAll    uint32
}

// Создание нового сервера
func NewServer() *Server {
	server := Server{
		loopExitCh:     make(chan bool),
		worldInfo:      NewWorldInfo(),
		clients:        make(map[uint32]*Client),
		makeClientCh:   make(chan *websocket.Conn),
		removeClientCh: make(chan *Client),
		needSendAll:    0,
	}
	return &server
}

func (server *Server) StartServer() {
	server.setupWebSocketListener()
	server.startMainLoop()
}

func (server *Server) ExitServer() {
	server.exitMainLoop()
}

func (server *Server) DeleteClient(client *Client) {
	server.removeClientCh <- client
}

func (server *Server) QueueSendAllNewState() {
	atomic.StoreUint32(&server.needSendAll, 1)
}

////////////////////////////////////////////////////////////////////////////////////////////////

func (server *Server) сreateClientForConnection(ws *websocket.Conn) {
	server.makeClientCh <- ws
}

func (server *Server) setupWebSocketListener() {
	onConnectedHandler := func(ws *websocket.Conn) {
		// Создание нового клиента
		server.сreateClientForConnection(ws) // выставляем клиента в очередь на добавление (синхронно)
		log.Println("WebSocket connect handler out")
	}
	http.Handle("/websocket", websocket.Handler(onConnectedHandler))
	log.Println("Web socket handler created")
}

func (server *Server) sendAllGameState() {
	var message GameStateMessage

	// Type
	message.Type = GAME_STATE_MESSAGE_WORLD_STATE

	// World info
	message.WorldData = *server.worldInfo

	// Clients data
	for _, client := range server.clients {
		message.ClienStates = append(message.ClienStates, client.GetCurrentState())
	}

	// Send all
	for _, client := range server.clients {
		client.QueueSendGameState(message)
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
func (server *Server) startMainLoop() {
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
				log.Printf("Delete client call\n")

				delete(server.clients, client.id)

				server.sendAllGameState()

			// Основной серверный таймер, который обновляет серверный мир
			case <-timer.C:
				// Сбрасываем таймер
				timer.Reset(updatePeriodMS)
				delta := time.Now().Sub(lastTickTime).Seconds()
				lastTickTime = time.Now()

				// обновляем состояние мира
				server.worldTick(delta)

				// Проверка, что нужно отсылать данные на сервер
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

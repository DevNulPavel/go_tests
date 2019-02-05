package gameserver

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

type Server struct {
	loopExitCh   chan bool
	gameRooms    map[uint32]*GameRoom
	removeRoomCh chan *GameRoom
	makeClientCh chan *WebSocket
}

// Создание нового сервера
func NewServer() *Server {
	server := Server{
		loopExitCh:   make(chan bool),
		gameRooms:    make(map[uint32]*GameRoom),
		removeRoomCh: make(chan *GameRoom),
		makeClientCh: make(chan *WebSocket),
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

func (server *Server) DeleteRoom(room *GameRoom) {
	server.removeRoomCh <- room
}

func (server *Server) setupWebSocketListener() {
	onConnectedHandler := func(ws *websocket.Conn) {
		log.Println("WebSocket connect handler in")
		connection := WebSocket{ws, make(chan bool)}
		server.makeClientCh <- &connection // Раз появилось новое соединение - запускаем его в работу
		<-connection.closeChannel          // Блокируем, иначе при выходе из функции произойдет выход
		log.Println("WebSocket connect handler out")
	}
	http.Handle("/websocket", websocket.Handler(onConnectedHandler))
	log.Println("Web socket handler created")
}

// Основная функция прослушивания
func (server *Server) startMainLoop() {
	loopFunction := func() {
		log.Println("Start main loop")
		for {
			select {
			// Обрабатываем новое подключение
			case connection := <-server.makeClientCh:
				log.Printf("Make client call\n")

				roomFound := false
				for _, gameRoom := range server.gameRooms {
					if gameRoom.GetIsFull() == false {
						gameRoom.AddClientForConnection(connection)
						roomFound = true
						break
					}
				}
				// Не нашли подходящей свободной комнаты
				if roomFound == false {
					newGameRoom := NewGameRoom(server)
					server.gameRooms[newGameRoom.roomId] = newGameRoom

					newGameRoom.StartLoop()

					newGameRoom.AddClientForConnection(connection)
				}

			// Обработка удаления комнаты
			case room := <-server.removeRoomCh:
				delete(server.gameRooms, room.roomId)

			// Завершение работы
			case <-server.loopExitCh:
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

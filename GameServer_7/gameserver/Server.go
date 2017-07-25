package gameserver

import (
	"errors"
	"log"
	"net"
)

type Server struct {
	isActive       bool
	listener       *net.TCPListener
	listenerExitCh chan bool
	loopExitCh     chan bool
	gameRooms      map[uint32]*ServerArena
	removeRoomCh   chan *ServerArena
	makeClientCh   chan *net.TCPConn
}

// Создание нового сервера
func NewServer() *Server {
	server := Server{
		isActive:       false,
		listener:       nil,
		listenerExitCh: make(chan bool),
		loopExitCh:     make(chan bool),
		gameRooms:      make(map[uint32]*ServerArena),
		removeRoomCh:   make(chan *ServerArena),
		makeClientCh:   make(chan *net.TCPConn),
	}
	return &server
}

func (server *Server) ExitServer() error {
	// TODO: Atomic???
	if server.isActive == true {
		server.exitAsyncSocketListener()
		server.exitMainLoop()
		server.isActive = false
		return nil
	}
	return errors.New("Server already stopped")
}

func (server *Server) StartListen() error {
	// TODO: Atomic???
	if server.isActive == false {
		// Listener
		err := server.asyncSocketAcceptListener()
		if err != nil {
			return err
		}
		// Loop
		server.mainLoop()
		// Flag
		server.isActive = true
		return nil
	}
	return errors.New("Server already active")
}

func (server *Server) DeleteRoom(room *ServerArena) {
	server.removeRoomCh <- room
}

// Обработка входящих подключений
func (server *Server) asyncSocketAcceptListener() error {
	address, err := net.ResolveTCPAddr("tcp", ":9999")
	if err != nil {
		log.Println("Server address resolve error")
		server.ExitServer()
		return err
	}

	// Создание листенера
	createdListener, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Printf("Server listener start error: %s\n", err)
		server.ExitServer()
		return err
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
				c, err := server.listener.AcceptTCP()
				if err != nil {
					log.Printf("Accept error: %s\n", err) // Наш лиснер закрылся и надо будет выйти из цикла
					return
				}

				log.Printf("Connection accepted\n")

				err = c.SetKeepAlive(true)
				if err != nil {
					log.Printf("Set keep alive error: %s\n", err) // Наш лиснер закрылся и надо будет выйти из цикла
					return
				}

				c.SetNoDelay(true)
				if err != nil {
					log.Printf("Set no delay error: %s\n", err) // Наш лиснер закрылся и надо будет выйти из цикла
					return
				}

				// Раз появилось новое соединение - запускаем его в работу с отдельной горутине
				server.makeClientCh <- c
			}
		}
	}

	go loopFunction()
	return nil
}

// Выход из листенера
func (server *Server) exitAsyncSocketListener() {
	if server.listener != nil {
		server.listener.Close()
	}
	server.listenerExitCh <- true
}

// Основная функция прослушивания
func (server *Server) mainLoop() {
	loopFunction := func() {
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
					newGameRoom, err := NewServerArena(server)
					if err != nil {
						log.Printf("Failed server create: %s\n", err)
					} else {
						server.gameRooms[newGameRoom.arenaId] = newGameRoom
						newGameRoom.StartLoop()
						newGameRoom.AddClientForConnection(connection)
					}
				}

			// Обработка удаления комнаты
			case room := <-server.removeRoomCh:
				delete(server.gameRooms, room.arenaId)

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

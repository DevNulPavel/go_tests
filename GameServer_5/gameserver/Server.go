package gameserver

import (
	"fmt"
	"log"
	"net"
)

type ServerMessage struct {
	address *net.UDPAddr
	data    []byte
}

type Server struct {
	// Connection
	conn                *net.UDPConn
	connReadLoopExitCh  chan bool
	connWriteLoopExitCh chan bool
	connReadDataCh      chan ServerMessage
	connWriteDataCh     chan ServerMessage
	// Main loop
	mainLoopExitCh chan bool
	// Logic
	gameRooms    map[*net.UDPAddr]*GameRoom
	removeRoomCh chan *net.UDPAddr
}

// Создание нового сервера
func NewServer() *Server {
	const writeBufferSize = 100
	const readBufferSize = 100

	server := Server{
		// Connection
		conn:                nil,
		connReadLoopExitCh:  make(chan bool),
		connWriteLoopExitCh: make(chan bool),
		connReadDataCh:      make(chan ServerMessage, readBufferSize),
		connWriteDataCh:     make(chan ServerMessage, writeBufferSize),
		// Main loop
		mainLoopExitCh: make(chan bool),
		// Logic
		gameRooms:    make(map[*net.UDPAddr]*GameRoom),
		removeRoomCh: make(chan *net.UDPAddr),
	}
	return &server
}

func (server *Server) ExitServer() {
	server.exitAsyncSocketListener()
	server.exitMainLoop()
}

func (server *Server) StartServer() {
	connectionStarted := server.asyncConnectionhandler()
	if connectionStarted {
		server.mainLoop()
        log.Println("Server started");
	}else {
        log.Println("Server NOT started");
    }
}

func (server *Server) SendMessage(message ServerMessage) {
    server.connWriteDataCh <- message
}

func (server *Server) DeleteRoomForAddress(address *net.UDPAddr) {
	server.removeRoomCh <- address
}

// Обработка входящих подключений
func (server *Server) asyncConnectionhandler() bool {
	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		fmt.Printf("Server address resolve error: %s\n", err)
		return false
	}

	// Прослушивание сервера
	connection, err := net.ListenUDP("udp", address)
	if err != nil {
		fmt.Printf("Listen UDP start error: %s\n", err)
		return false
	}

	// Сохраняем листенер для возможности закрытия
	server.conn = connection

	// Функции-циклы обработки
	readLoopFunction := func() {
		for {
			select {
			case <-server.connReadLoopExitCh:
				server.conn.Close()
				log.Print("Connection read exit") // Наш лиснер закрылся и надо будет выйти из цикла
				return

			default:
				dataBuffer := make([]byte, 128)
				readCount, address, err := server.conn.ReadFromUDP(dataBuffer)
				if err != nil {
					fmt.Printf("UDP read error: %s\n", err)
					continue // TODO: ???
				} else if readCount == 0 {
					fmt.Printf("UDP read 0 bytes\n")
					continue // TODO: ???
				}

                log.Printf("Message received\n")

				message := ServerMessage{address: address, data: dataBuffer[0:readCount]}
				server.connReadDataCh <- message
			}
		}
	}
	writeLoopFunction := func() {
		for {
			select {
			case <-server.connWriteLoopExitCh:
				server.conn.Close()
				log.Print("Connection write exit") // Наш лиснер закрылся и надо будет выйти из цикла
				return

			case writeMessage := <-server.connWriteDataCh:
				writtenCount, err := server.conn.WriteToUDP(writeMessage.data, writeMessage.address)
				if err != nil {
					fmt.Printf("UDP write error: %s\n", err) // TODO: ???
				} else if writtenCount < len(writeMessage.data) {
					fmt.Printf("UDP writeen less bytes: %d from %d\n", writtenCount, len(writeMessage.data)) // TODO: ???
				}

                log.Printf("Message sent\n")
			}
		}
	}

	go readLoopFunction()
	go writeLoopFunction()

	return true
}

// Выход из листенера
func (server *Server) exitAsyncSocketListener() {
	if server.conn != nil {
		server.conn.Close()
	}
	server.connReadLoopExitCh <- true
	server.connWriteLoopExitCh <- true
}

// Основная функция прослушивания
func (server *Server) mainLoop() {
	loopFunction := func() {
		for {
			select {
			// Обрабатываем входищие сообщения
			case message := <-server.connReadDataCh:

				room, roomFound := server.gameRooms[message.address]
				if roomFound {
					// Обрабатываем сообщение
					room.HandleMessage(message)
				} else {
					freeRoomFound := false
					for _, room := range server.gameRooms {
						if room.GetIsFull() == false {
							// Задаем соответствие комнаты и адреса
							server.gameRooms[message.address] = room

							// Обрабатываем сообщение
							room.HandleMessage(message)

							// Нашли комнату
							freeRoomFound = true
							break
						}
					}

					if freeRoomFound == false {
						newGameRoom := NewGameRoom(server)
						server.gameRooms[message.address] = newGameRoom
						newGameRoom.StartLoop()

						// Обрабатываем сообщение
                        newGameRoom.HandleMessage(message)
					}
				}

			// Обработка удаления комнаты
			case address := <-server.removeRoomCh:
				delete(server.gameRooms, address)

			// Завершение работы
			case <-server.mainLoopExitCh:
				log.Print("Main loop exit") // Наш лиснер закрылся и надо будет выйти из цикла
				return
			}
		}
	}
	go loopFunction()
}

func (server *Server) exitMainLoop() {
	server.mainLoopExitCh <- true
}

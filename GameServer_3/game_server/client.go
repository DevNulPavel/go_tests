package gameserver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"log"
    "bufio"
    "time"
)

// Variables
var maxID int = 1

// Client ... Структура клиента
type Client struct {
	server            *Server
	connection        *net.Conn
	id                int
	state             ClienState
    writer            *bufio.Writer
    reader            *bufio.Reader
	usersStateChannel chan []ClienState
	exitChannel       chan bool
}

// NewClient ... Конструктор
func NewClient(connection *net.Conn, server *Server) *Client {
	if connection == nil {
		panic("No connection")
	}
	if server == nil {
		panic("No game server")
	}

	// Увеличиваем id
	maxID++

	// Конструируем клиента и его каналы
	clientState := ClienState{maxID, float64(rand.Int() % 100), float64(rand.Int() % 100), 0}
    writer := bufio.NewWriter(*connection)
    reader := bufio.NewReader(*connection)
	usersStateChannel := make(chan []ClienState, 1000) // В канале апдейтов может накапливаться максимум 10 апдейтов
	successChannel := make(chan bool)

	return &Client{
		server,
		connection,
		maxID,
		clientState,
        writer,
        reader,
		usersStateChannel,
		successChannel,
	}
}

// QueueSendAllStates ... Пишем сообщение клиенту
func (client *Client) QueueSendAllStates(states []ClienState) {
	select {
	// Пишем сообщение в канал
	case client.usersStateChannel <- states:
		//log.Println("Client wrote:", message)

	// Удаляем клиента раз у нас произошла ошибка какая-то
	default:
        err := fmt.Errorf("Client %d disconnected", client.id)
		client.server.DeleteClient(client)
		client.server.SendErr(err)
		client.QueueSendExit() // Вызываем выход из горутины loopWrite
		return
	}
}

// QueueSendCurrentClientState ... Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
	currentUserStateArray := []ClienState{client.state}
	select {
	// Пишем сообщение в канал
	case client.usersStateChannel <- currentUserStateArray:
		//log.Println("Client wrote:", message)

	// Удаляем клиента если нельзя отправлять
	default:
		client.server.DeleteClient(client)
		err := fmt.Errorf("Client %d disconnected", client.id)
		client.server.SendErr(err)
		client.QueueSendExit() // Вызываем выход из горутины loopWrite
		return
	}
}

// Отправляем успешный результат
func (client *Client) QueueSendExit() {
	client.exitChannel <- true
}

// Запускаем ожидания записи и чтения (блокирующая функция)
func (client *Client) StartSyncListenLoop() {
	go client.loopWrite() // в отдельной горутине
	client.loopRead()
}

// Ожидание записи
func (client *Client) loopWrite() {
	log.Println("StartSyncListenLoop write to client:", client.id)
	for {
		select {
		// Отправка записи клиенту
		case message := <-client.usersStateChannel:
			// Данные
			jsonDataBytes, err := json.Marshal(message)
			if err != nil {
				continue
			}

			// Размер
			dataBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(dataBytes, uint64(len(jsonDataBytes)))

            log.Printf("Send to client %d: %s\n", client.id, string(jsonDataBytes))

            timeout := time.Now().Add(5 * time.Second)
            (*client.connection).SetWriteDeadline(timeout)

			// Отсылаем
            client.writer.Write(dataBytes)
            client.writer.Write(jsonDataBytes)
            err = client.writer.Flush()
            if err != nil {
                client.server.DeleteClient(client)
                // TODO: client.QueueSendExit() надо ли??
                log.Println("LoopWrite exit by ERROR, clientId =", client.id)
                return
            }

			//(*client.connection).Write(dataBytes)
			//(*client.connection).Write(jsonDataBytes)

			//tempBytes := make([]byte, len(dataBytes))
			//(*client.connection).Read(tempBytes)

		// Получение флага выхода из функции
		case <-client.exitChannel:
			client.server.DeleteClient(client)
            log.Println("LoopWrite exit, clientId =", client.id)
			return
		}
	}
}

// Ожидание чтения
func (client *Client) loopRead() {
	//log.Println("Listening read from client")
	for {
		select {
		// Получение флага выхода
		case <-client.exitChannel:
			client.server.DeleteClient(client)
			client.QueueSendExit() // для метода loopWrite, чтобы выйти из него
            log.Println("LoopRead exit, clientId =", client.id)
			return

		// Чтение данных из webSocket
		default:
			// Размер данных
			dataSizeBytes := make([]byte, 8)
			readCount, err := client.reader.Read(dataSizeBytes)
			if (err != nil) || (readCount == 0) {
				client.server.DeleteClient(client)
				client.QueueSendExit() // для метода loopWrite, чтобы выйти из него
                log.Println("LoopRead exit, clientId =", client.id)
				return
			}
			dataSize := binary.LittleEndian.Uint64(dataSizeBytes)

			// Данные
			data := make([]byte, dataSize)
			readCount, err = client.reader.Read(data)

            if err == io.EOF {
				// Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
				client.QueueSendExit()
				return
			} else if err != nil {
				// Ошибка
				client.server.SendErr(err)
				// TODO: ???
				// Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
				client.QueueSendExit()
				return
			} else {
				if readCount > 0 {
					// Декодирование из Json в структуру
					var state ClienState
					err := json.Unmarshal(data, &state)

					if (err == nil) && (state.ID > 0) {
                        log.Printf("Client %d received: %s \n", client.id, string(data))

						// Сбновляем состояние данного клиента
						client.state = state

						// Отправляем обновление состояния всем
						//log.Println("Send all:", msg)
						client.server.SendAll()
					}
				} else {
					// Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
					client.QueueSendExit()
                    log.Println("LoopRead exit, clientId =", client.id)
					return
				}
			}
		}
	}
}

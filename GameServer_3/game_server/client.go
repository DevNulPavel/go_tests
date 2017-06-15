package gameserver

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"log"
    "bufio"
    "time"
)

const UPDATE_QUEUE_SIZE = 100

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
	exitReadChannel   chan bool
    exitWriteChannel  chan bool
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
	usersStateChannel := make(chan []ClienState, UPDATE_QUEUE_SIZE) // В канале апдейтов может накапливаться максимум 1000 апдейтов
    exitReadChannel := make(chan bool)
    exitWriteChannel := make(chan bool)

	return &Client{
		server,
		connection,
		maxID,
		clientState,
        writer,
        reader,
		usersStateChannel,
        exitReadChannel,
        exitWriteChannel,
	}
}

// QueueSendAllStates ... Пишем сообщение клиенту
func (client *Client) QueueSendAllStates(states []ClienState) {
	// Если очередь превышена - считаем, что юзер отвалился
    if len(client.usersStateChannel)+1 > UPDATE_QUEUE_SIZE {
        log.Printf("Queue full for client %d", client.id)
        // TODO: Ждем таймаут??
        //client.server.DeleteClient(client)
        //client.exitWriteChannel <- true
        //client.exitReadChannel <- true
        return
    }else{
		client.usersStateChannel <- states
	}
}

// QueueSendCurrentClientState ... Пишем сообщение клиенту только с его состоянием
func (client *Client) QueueSendCurrentClientState() {
    // Если очередь превышена - считаем, что юзер отвалился
    if len(client.usersStateChannel)+1 > UPDATE_QUEUE_SIZE {
        log.Printf("Queue full for client %d", client.id)
        // TODO: Ждем таймаут??
        //client.server.DeleteClient(client)
        //client.exitWriteChannel <- true
        //client.exitReadChannel <- true
        return
    }else{
        currentUserStateArray := []ClienState{client.state}
        client.usersStateChannel <- currentUserStateArray
    }
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

            //log.Printf("Send to client %d: %s\n", client.id, string(jsonDataBytes))

            timeout := time.Now().Add(30 * time.Second)
            (*client.connection).SetWriteDeadline(timeout)

			// Отсылаем
            client.writer.Write(dataBytes)
            client.writer.Write(jsonDataBytes)
            err = client.writer.Flush()
            if err != nil {
                client.server.DeleteClient(client)
                // TODO: client.QueueSendExit() надо ли??
                client.exitReadChannel <- true // Выход из loopRead
                log.Println("LoopWrite exit by ERROR, clientId =", client.id)
                return
            }

			//(*client.connection).Write(dataBytes)
			//(*client.connection).Write(jsonDataBytes)

			//tempBytes := make([]byte, len(dataBytes))
			//(*client.connection).Read(tempBytes)

		// Получение флага выхода из функции
		case <-client.exitWriteChannel:
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
		case <-client.exitReadChannel:
			client.server.DeleteClient(client)
            log.Println("LoopRead exit, clientId =", client.id)
			return

		// Чтение данных из webSocket
		default:
            // Ожидается, что за 10 минут что-то придет, иначе - это отвал
            timeout := time.Now().Add(10 * time.Minute)
            (*client.connection).SetReadDeadline(timeout)

			// Размер данных
			dataSizeBytes := make([]byte, 8)
			readCount, err := client.reader.Read(dataSizeBytes)
			if (err != nil) || (readCount == 0) {
				client.server.DeleteClient(client)
                client.exitWriteChannel <- true // для метода loopWrite, чтобы выйти из него
                log.Println("LoopRead exit, clientId =", client.id)
				return
			}
			dataSize := binary.LittleEndian.Uint64(dataSizeBytes)

            // Ожидается, что будут данные в течении 20 секунд - иначе отвал
            timeout = time.Now().Add(20 * time.Second)
            (*client.connection).SetReadDeadline(timeout)

            // Данные
			data := make([]byte, dataSize)
			readCount, err = client.reader.Read(data)

            if err == io.EOF {
				// Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
                client.exitWriteChannel <- true
                log.Println("LoopRead exit, clientId =", client.id)
				return
			} else if err != nil {
                // TODO: ???
				// Ошибка
                log.Printf("Client %d error: ", client.id, err)
				// Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
                client.exitWriteChannel <- true
                log.Println("LoopRead exit, clientId =", client.id)
				return
			} else {
				if readCount > 0 {
					// Декодирование из Json в структуру
					var state ClienState
					err := json.Unmarshal(data, &state)

					if (err == nil) && (state.ID > 0) {
                        //log.Printf("Client %d received: %s \n", client.id, string(data))

						// Сбновляем состояние данного клиента
						client.state = state

						// Отправляем обновление состояния всем
						client.server.SendAll()
					}
				} else {
                    // Разрыв соединения - отправляем в очередь сообщение выхода для loopWrite
                    client.exitWriteChannel <- true
                    log.Println("LoopRead exit, clientId =", client.id)
					return
				}
			}
		}
	}
}

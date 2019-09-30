package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ivpusic/grpool"
)

var users map[string]*User
var getUsersChan chan UserSearchRequest
var workersPool *grpool.Pool

//////////////////////////////////////////////////////////////////////

// UserSearchRequest user search request with result channel
type UserSearchRequest struct {
	addr       *net.UDPAddr
	resultChan chan *User
}

// User ???
type User struct {
	clientAddress    *net.UDPAddr
	lastActiveTime   time.Time
	dataMutex        sync.Mutex
	lastPacketNumber uint64
}

func (user *User) handleData(data []byte, c *net.UDPConn) {
	const counterOffset = 200

	// Делаем валидацию
	if user == nil {
		log.Printf("User do not exists")
	}

	if len(data) < counterOffset+8 {
		log.Printf("Invalid UDP packages size: %d", len(data))
		return
	}

	// Парсим прочитанные данные
	receivedPacketNumber := binary.BigEndian.Uint64(data[counterOffset : counterOffset+8])
	//log.Printf("Received package number: %d", receivedPacketNumber)

	// Может быть уже получали пакет или это более старый?
	user.dataMutex.Lock()
	if receivedPacketNumber <= user.lastPacketNumber {
		log.Printf("Invalid UDP packages seq: %d received, %d last", receivedPacketNumber, user.lastPacketNumber)
	} else {
		// Это следующий пакет, все ок
		user.lastPacketNumber = receivedPacketNumber
	}
	user.dataMutex.Unlock()

	// timer := time.NewTimer(10 * time.Millisecond)
	// <-timer.C

	// Теперь очередь ответной записи
	writtenCount, err := c.WriteToUDP(data, user.clientAddress)
	if err != nil {
		log.Println(err)
		return
	} else if writtenCount < len(data) {
		log.Printf("Written less bytes - %d from \n", writtenCount, len(data))
		return
	}
}

// GetUserForAddr request user for address
func GetUserForAddr(addr *net.UDPAddr) *User {
	req := UserSearchRequest{addr, make(chan *User)}
	getUsersChan <- req
	return <-req.resultChan
}

func usersLoop() {
	// Список юзерв
	users = make(map[string]*User)

	// Канал получения нового юзера
	getUsersChan = make(chan UserSearchRequest)

	// Периодичность проверки отвалившихся юзеров
	const clearTime = 20 * time.Second
	timer := time.NewTimer(clearTime)
	for {
		select {
		case req := <-getUsersChan:
			// Ищем пользователя, если нету - создаем нового
			user, ok := users[req.addr.String()]
			if ok {
				user.lastActiveTime = time.Now()
				req.resultChan <- user
			} else {
				// Создание нового
				newUser := &User{
					clientAddress:    req.addr,
					lastActiveTime:   time.Now(),
					lastPacketNumber: 0,
				}
				users[req.addr.String()] = newUser
				req.resultChan <- newUser
			}

		// Удаляем неактивных юзеров
		case <-timer.C:
			timer.Reset(clearTime)

			// Если юзер неактивен в течение 60 секунд - считаем его невалидным
			const userInactiveDuration = 60 * time.Second

			// Ищем юзеров, которые которых надо удалить
			deleteKeys := make([]string, 0)
			for key, val := range users {
				testTime := val.lastActiveTime.Add(userInactiveDuration)
				if testTime.Before(time.Now()) {
					deleteKeys = append(deleteKeys, key)
				}
			}

			// Удаляем юзеров
			if len(deleteKeys) > 0 {
				for _, key := range deleteKeys {
					delete(users, key)
					log.Printf("User for address %s deleted", key)
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////

func handleServerConnectionRaw(c *net.UDPConn, serverAddress *net.UDPAddr, wg *sync.WaitGroup, threadIndex int) {
	defer wg.Done()

	// Гарантированный размер датаграммы
	const dataSize = 508
	udpBuffer := make([]byte, dataSize)

	for {
		// Читаем данные в буффер
		readCount, clientAddress, err := c.ReadFromUDP(udpBuffer)
		if (err == io.EOF) || (readCount == 0) { // Нужно ли проверять readCount
			//log.Println("Disconnected")
			continue
		} else if err != nil {
			log.Println(err)
			continue
		}

		//log.Printf("Handle message in thread: %d", threadIndex)

		// Создаем копию с данными
		receivedData := make([]byte, readCount)
		copy(receivedData, udpBuffer[0:readCount])

		// Закидываем в пулл задачу по обработке
		workersPool.JobQueue <- func() {
			// Получаем юзера для текущего адреса
			user := GetUserForAddr(clientAddress)

			// Обрабатываем полученные данные
			user.handleData(receivedData, c)
		}
	}
}

func main() {
	// Пулл обработчиков
	workersPool = grpool.NewPool(128, 128)
	defer workersPool.Release()

	// Запускаем корутину работы с юзерами
	go usersLoop()

	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		log.Println(err)
		return
	}

	// Создание приемника новых подключений
	connection, err := net.ListenUDP("udp", address)
	if err == nil {
		log.Print("New connection created\n")

		// Выставляем размер буффера
		//connection.SetReadBuffer(1024)
		//connection.SetWriteBuffer(1024)

		// Создаем горутины, которые просто обрабатывают I/O
		const goroutinesCount = 64
		var wg sync.WaitGroup
		for i := 0; i < goroutinesCount; i = i + 1 {
			wg.Add(1)
			go handleServerConnectionRaw(connection, address, &wg, i)
		}
		wg.Wait()

		// Закрытиие соединения
		connection.Close()
	} else {
		log.Println("Error in accept: %s", err.Error())
	}
}

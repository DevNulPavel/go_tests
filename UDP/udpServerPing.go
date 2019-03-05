package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var usersMutex sync.Mutex
var users map[string]*User

//////////////////////////////////////////////////////////////////////

// User ???
type User struct {
	clientAddress    *net.UDPAddr
	lastActiveTime   time.Time
	dataMutex        sync.Mutex
	lastPacketNumber uint64
}

func (user *User) handleData(data []byte) {
	const counterOffset = 200

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
	if receivedPacketNumber <= user.lastPacketNumber {
		log.Printf("Invalid UDP packages seq: %d received, %d last", receivedPacketNumber, user.lastPacketNumber)
		return
	}

	// Это следующий пакет, все ок
	user.dataMutex.Lock()
	user.lastPacketNumber = receivedPacketNumber
	user.dataMutex.Unlock()
}

//////////////////////////////////////////////////////////////////////

func getUserForAddr(addr *net.UDPAddr) *User {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	// Ищем пользователя, если нету - создаем нового
	user, ok := users[addr.String()]
	if ok {
		user.lastActiveTime = time.Now()
		return user
	}

	// Создание нового
	newUser := &User{
		clientAddress:    addr,
		lastActiveTime:   time.Now(),
		lastPacketNumber: 0,
	}
	users[addr.String()] = newUser
	return newUser
}

// Функция чистит юзеров от которых давно не было ответа
func clearOldUsers() {
	const clearTime = 20 * time.Second
	const oldUserDuration = 15 * time.Second

	timer := time.NewTimer(clearTime)
	for {
		<-timer.C
		timer.Reset(clearTime)

		//log.Printf("Old user clear tick")

		usersMutex.Lock()
		deleteKeys := make([]string, 0)
		for key, val := range users {
			testTime := val.lastActiveTime.Add(oldUserDuration)
			if testTime.Before(time.Now()) {
				deleteKeys = append(deleteKeys, key)
			}
		}
		for _, key := range deleteKeys {
			delete(users, key)
			log.Printf("User for address %s deleted", key)
		}
		usersMutex.Unlock()
	}
}

func handleServerConnectionRaw(c *net.UDPConn, serverAddress *net.UDPAddr, wg *sync.WaitGroup, threadIndex int) {
	defer wg.Done()

	const dataSize = 1024
	udpBuffer := make([]byte, dataSize)

	for {
		// Читаем данные в буффер
		readCount, clientAddress, err := c.ReadFromUDP(udpBuffer)
		if (err == io.EOF) || (readCount == 0) { // Нужно ли проверять readCount
			log.Println("Disconnected")
			continue
		} else if err != nil {
			log.Println(err)
			continue
		}

		//log.Printf("Handle message in thread: %d", threadIndex)

		// Получаем юзера для текущего адреса
		user := getUserForAddr(clientAddress)
		user.handleData(udpBuffer[0:readCount])

		// Теперь очередь ответной записи
		writtenCount, err := c.WriteToUDP(udpBuffer[0:readCount], clientAddress)
		if err != nil {
			log.Println(err)
			continue
		} else if writtenCount < readCount {
			log.Printf("Written less bytes - %d from \n", writtenCount, readCount)
			continue
		}
	}
}

func main() {
	users = make(map[string]*User)

	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		log.Println(err)
		return
	}

	// Горутина удаления старых пользователей
	go clearOldUsers()

	// Создание соединения
	connection, err := net.ListenUDP("udp", address)
	if err == nil {
		log.Print("New connection created\n")

		// Выставляем размер буффера
		//connection.SetReadBuffer(1024)
		//connection.SetWriteBuffer(1024)

		// Создаем горутины, которые обрабатывают данные
		const goroutinesCount = 10
		var wg sync.WaitGroup
		for i := 0; i < goroutinesCount; i = i + 1 {
			wg.Add(1)
			go handleServerConnectionRaw(connection, address, &wg, i)
		}
		wg.Wait()
	} else {
		log.Println("Error in accept: %s", err.Error())
	}

	// Закрытиие соединения
	connection.Close()
}

package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ivpusic/grpool"
	"github.com/pkg/profile"
)

type Stopper interface {
	Stop()
}

var withProfiling bool = true
var usersMutex sync.Mutex
var users map[string]*User
var workersPool *grpool.Pool
var profileStopper Stopper

//////////////////////////////////////////////////////////////////////

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

//////////////////////////////////////////////////////////////////////

func getUserForAddr(addr *net.UDPAddr) *User {
	// Ищем пользователя, если нету - создаем нового
	usersMutex.Lock()
	defer usersMutex.Unlock()

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

		if len(deleteKeys) > 0 {
			for _, key := range deleteKeys {
				delete(users, key)
				log.Printf("User for address %s deleted", key)
			}
		}
		usersMutex.Unlock()
	}
}

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
			user := getUserForAddr(clientAddress)

			// Обрабатываем полученные данные
			user.handleData(receivedData, c)
		}
	}
}

func main() {
	if withProfiling {
		//profileStopper = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		profileStopper = profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	}

	users = make(map[string]*User)

	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		log.Println(err)
		return
	}

	// Пулл обработчиков
	workersPool = grpool.NewPool(128, 128)
	defer workersPool.Release()

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

		if withProfiling {
			log.Print("Press any key to continue profiling\n")
			var input string
			fmt.Scanln(&input)
		} else {
			wg.Wait()
		}

		// Закрытиие соединения
		connection.Close()
	} else {
		log.Println("Error in accept: %s", err.Error())
	}

	if withProfiling {
		profileStopper.Stop()
	}
}

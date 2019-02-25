package main

import (
	"log"
	"net"
	"net/http"
	"sync"
	"syscall"

	_ "net/http/pprof"

	"github.com/DevNulPavel/easygo/netpoll"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/ivpusic/grpool"
)

var poller netpoll.Poller
var workersPool *grpool.Pool

// User - тип для обработки соединения
type User struct {
	mutex   sync.RWMutex
	conn    net.Conn
	desc    *netpoll.Desc
	reading bool
	closed  bool
}

// SetReading выставляет статус чтения
func (user *User) SetReading(status bool) {
	user.mutex.Lock()
	user.reading = status
	user.mutex.Unlock()
}

// IsReading возвращает статус чтения
func (user *User) IsReading() bool {
	user.mutex.RLock()
	readingCopy := user.reading
	user.mutex.RUnlock()
	return readingCopy
}

// Close закрывает все дескрипторы
func (user *User) Close() {
	user.mutex.Lock()

	poller.Stop(user.desc)
	user.conn.Close()
	user.desc.Close()
	user.closed = true

	user.mutex.Unlock()
}

// IsClosed проверяет, не закрыто ли соединение
func (user *User) IsClosed() bool {
	user.mutex.RLock()
	closedCopy := user.closed
	user.mutex.RLock()
	return closedCopy
}

////////////////////////////////////////////////////////////////

// Вариант с использованием библиотеки WS
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем WS соединение из http
	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	// Ручное создание дескриптора
	/*desc, err := netpoll.Handle(conn, netpoll.EventRead | netpoll.EventEdgeTriggered)
	if err != nil {
		// handle error
	}*/

	// Создаем дескриптор для poll на чтение c постоянным вызовом коллбека
	//descriptor := netpoll.Must(netpoll.HandleRead(connection))

	// Создаем дескриптор для poll на чтение с разовым вызовом коллбека
	descriptor := netpoll.Must(netpoll.HandleReadOnce(connection))

	// Создаем временный объект
	user := &User{
		sync.RWMutex{},
		connection,
		descriptor,
		false,
		false,
	}

	// Устанавливаем обработчик для данного соединения
	handleCallback := func(ev netpoll.Event) {
		//log.Printf("Callback with event: %s", ev.String())

		// Вырубили соединение
		if ((ev & netpoll.EventReadHup) != 0) || ((ev & netpoll.EventWriteHup) != 0) {
			user.Close()
			return
		}

		// Можем читать данные
		if (ev & netpoll.EventRead) != 0 {
			// Вычитывать данные надо все-таки в том же самом потоке и коллбеке (если у нас режим не OneShot)
			//data, code, err := wsutil.ReadClientData(user.conn)
			/*_, _, err := wsutil.ReadClientData(user.conn)
			if err != nil {
				user.Close()
				return
			} */

			// Закидываем в пулл задачу по обработке
			workersPool.JobQueue <- func() {
				//log.Printf("Thread function begin")

				if user.IsClosed() {
					log.Printf("Thread function exit by closed status")
					return
				}

				// В режиме OneShot можно вычитывать в отдельной горутине
				//data, code, err := wsutil.ReadClientData(user.conn)
				_, _, err := wsutil.ReadClientData(user.conn)
				if err != nil {
					user.Close()
					return
				}
				//log.Printf("msg: %s, code: %d", string(data), code)

				responseData := []byte("Response")
				err = wsutil.WriteServerText(user.conn, responseData)
				if err != nil {
					user.Close()
					return
				}

				// Снова запускаем отслеживание событий на данном дескрипторе (только для OneShot)
				poller.Resume(user.desc)
			}
		}
	}
	poller.Start(user.desc, handleCallback)
}

func main() {
	// Увеличиваем системные лимиты
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	log.Printf("Set total max connections limit: %d", rLimit.Cur)

	// Подключаем профайлинг
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	// Пулл обработчиков
	workersPool = grpool.NewPool(128, 128)
	defer workersPool.Release()

	// Создаем пулер
	var err error
	poller, err = netpoll.New(nil)
	if err != nil {
		panic(err)
	}

	// Запуск веб сервера
	http.HandleFunc("/", wsHandler)
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal(err)
	}
}

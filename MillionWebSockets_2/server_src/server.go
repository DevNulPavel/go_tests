package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
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
	reading int32
	closed  bool
}

// PushReading добавляет к счетчику количество чтений
func (user *User) PushReading() {
	atomic.AddInt32(&user.reading, 1)
}

// PopReading извлекает одно чтение
func (user *User) PopReading() {
	atomic.AddInt32(&user.reading, -1)
}

// IsReading возвращает активно ли чтение
func (user *User) IsReading() bool {
	status := atomic.LoadInt32(&user.reading) > 0
	return status
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

// TryToStartReading стартует чтение если еще не активно
func (user *User) TryToStartReading() {
	// Закидываем в пулл задачу по обработке
	workersPool.JobQueue <- func() {
		//log.Printf("Thread function begin")

		// На всякий случай проверяем, что у нас не закрыто соединение
		if user.IsClosed() {
			log.Printf("Thread function exit by closed status (1)")
			return
		}

		// В режиме OneShot можно вычитывать в отдельной горутине
		//data, code, err := wsutil.ReadClientData(user.conn)
		_, _, err := wsutil.ReadClientData(user.conn)
		if err != nil {
			log.Printf("Thread function return with close with read error: %v", err)
			if err == io.EOF {
				user.Close()
			}
			return
		}

		// Возобновляем оповещения о появлении новых данных
		poller.Resume(user.desc)

		// Выводим сообщение полученное
		//log.Printf("msg: %s, code: %d", string(data), code)

		// На всякий случай проверяем, что у нас не закрыто соединение
		if user.IsClosed() {
			log.Printf("Thread function exit by closed status (2)")
			return
		}

		// Пишем ответ
		responseData := []byte("Response")
		err = wsutil.WriteServerText(user.conn, responseData)
		if err != nil {
			log.Printf("Thread function return by close with write")
			if err == io.EOF {
				user.Close()
			}
			return
		}

	}
}

////////////////////////////////////////////////////////////////

// Вариант с использованием библиотеки WS
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем WS соединение из http
	connection, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	// Создаем дескриптор для poll на чтение c постоянным вызовом коллбека
	//descriptor := netpoll.Must(netpoll.HandleRead(connection))

	// Создаем дескриптор для poll на чтение с разовым вызовом коллбека
	descriptor := netpoll.Must(netpoll.HandleReadOnce(connection))

	// Создаем временный объект
	user := &User{
		sync.RWMutex{},
		connection,
		descriptor,
		0,
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
			// Для режима OneShot
			user.TryToStartReading()
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

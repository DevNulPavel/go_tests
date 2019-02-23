package main

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"syscall"

	"github.com/DevNulPavel/easygo/netpoll"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var poller netpoll.Poller = nil

type User struct {
	conn         net.Conn
	desc         *netpoll.Desc
	needResponse bool
}

func (user *User) Close() {
	poller.Stop(user.desc)
	user.conn.Close()
}

// Вариант с использованием библиотеки WS
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем WS соединение из http
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}

	// Ручное создание дескриптора
	/*desc, err := netpoll.Handle(conn, netpoll.EventRead | netpoll.EventEdgeTriggered)
	if err != nil {
		// handle error
	}*/

	// Создаем дескриптор для poll на чтение/запись
	desc := netpoll.Must(netpoll.HandleReadWrite(conn))

	// Создаем временный объект
	user := &User{
		conn,
		desc,
		false,
	}

	// Устанавливаем обработчик для данного соединения
	handleCallback := func(ev netpoll.Event) {
		log.Printf("Callback with event: %s", ev.String())

		// Вырубили соединение
		if ((ev & netpoll.EventReadHup) != 0) || ((ev & netpoll.EventWriteHup) != 0) {
			user.Close()
			return
		}

		// Можем читать данные
		if (ev & netpoll.EventRead) != 0 {
			data, code, err := wsutil.ReadClientData(conn)
			if err != nil {
				user.Close()
				return
			}
			log.Printf("msg: %s, code: %d", string(data), code)
			user.needResponse = true
		}

		// Можем писать данные ответа
		if ((ev & netpoll.EventWrite) != 0) && (user.needResponse == true) {
			responseData := []byte("Response")
			err = wsutil.WriteServerText(conn, responseData)
			if err != nil {
				user.Close()
				return
			}
			//log.Printf("Response write success")
			user.needResponse = false
		}
	}
	poller.Start(desc, handleCallback)
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

	// Создаем пулер
	var err error = nil
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

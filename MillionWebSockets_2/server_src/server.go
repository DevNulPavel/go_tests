package main

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mailru/easygo/netpoll"
	"log"
	"net/http"
	_ "net/http/pprof"
	"syscall"
)

var poller netpoll.Poller = nil

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

	// Создаем дескриптор для poll на чтение
	desc := netpoll.Must(netpoll.HandleRead(conn))

	// Устанавливаем обработчик для данного соединения
	handleCallback := func(ev netpoll.Event) {
		// Вырубили соединение
		if (ev & netpoll.EventReadHup) != 0 {
			poller.Stop(desc)
			conn.Close()
			return
		}
		// Можем читать данные
		if (ev & netpoll.EventRead) != 0 {
			// _, err := ioutil.ReadAll(conn)
			data, code, err := wsutil.ReadClientData(conn)
			if err != nil {
				// TODO: Close? handle error?
				poller.Stop(desc)
				conn.Close()
				return
			}
			log.Printf("msg: %s, code: %d", string(data), code)
		}
		// Можем писать данные
		if (ev & netpoll.EventWrite) != 0 {
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

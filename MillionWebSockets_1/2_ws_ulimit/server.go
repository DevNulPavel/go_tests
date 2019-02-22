package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"syscall"
)

var count int64 = 0

func ws(w http.ResponseWriter, r *http.Request) {
	// Создаем объект, который получает данные
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Атомарно увеличиваем счетчик обрабатываемых соединений
	new := atomic.AddInt64(&count, 1)

	// Выводим количество соединений
	if (new % 100) == 0 {
		log.Printf("Total number of connections: %v", new)
	}

	// Отложенное уменьшение количества соединений на выход
	defer func() {
		new := atomic.AddInt64(&count, -1)
		if (new % 100) == 0 {
			log.Printf("Total number of connections: %v", new)
		}
	}()

	// Читаем данные из сокета
	for {
		//_, msg, err := conn.ReadMessage()
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		//log.Printf("msg: %s", string(msg))
	}
}

func main() {
	// Увеличиваем системные лимиты до максимума
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	log.Printf("Set total max connections limit: %d", rLimit.Cur)

	// Включаем pprof
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("Pprof failed: %v", err)
		}
	}()

	// Запускаем веб-сервер
	http.HandleFunc("/", ws)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

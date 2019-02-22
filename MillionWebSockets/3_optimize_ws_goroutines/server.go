package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	_ "net/http/pprof"
	"syscall"
)

var epoller *epoll = nil

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем соединение
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	// Добавляем соединение в обработчик epool
	if err := epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection")
		conn.Close()
	}
}

// Обработчик epool
func Start() {
	// В бесконечном цикле получаем соединения в которых обновились данные
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			// Читаем данные из соединения
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("Failed to remove %v", err)
				}
			} else {
				//log.Printf("msg: %s", string(msg))
			}
		}
	}
}

func main() {
	// Увеличиваем системные лимиты на максимальное количество соединений
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	log.Printf("Set total max connections limit: %d", rLimit.Cur)

	// Включаем профилирование
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	// Запускаем epool
	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}
	go Start()

	// Запускаем http сервер
	http.HandleFunc("/", wsHandler)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

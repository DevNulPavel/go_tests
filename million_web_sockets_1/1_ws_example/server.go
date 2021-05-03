package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// Обработчик веб-сокет подключения
func ws(w http.ResponseWriter, r *http.Request) {
	// Создаем объект, который представляет из себя обработчик сокета
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	// Читаем данные из веб-сокета
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}
		log.Printf("msg: %s", string(msg))
	}
}

func main() {
	// Запускаем веб сервер
	http.HandleFunc("/", ws)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

package main

import "log"
import "net/http"
import "./game_server"

func main() {
	log.SetFlags(log.Lshortfile)

	// Запуск сервера
	server := game_server.NewServer()
	server.StartAsyncListen()

	// HTTP сервер
	http.Handle("/", http.FileServer(http.Dir("web")))
	err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
}

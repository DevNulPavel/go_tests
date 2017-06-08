package main

import "log"
import "./game_server"

func main() {
	log.SetFlags(log.Lshortfile)

	// Запуск сервера
	server := game_server.NewServer()
	server.StartSyncListen()
}

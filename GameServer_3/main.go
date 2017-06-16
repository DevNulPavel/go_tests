package main

import "log"
import "./game_server"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	// Запуск сервера
	server := gameserver.NewServer()
	server.StartSyncListen()
}

package main

import (
	"log"
	"./gameserver"
)

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	// Запуск сервера
	server := gameserver.NewServer()
	server.StartSyncListen()

	//<-time.After(time.Second * 30)
	//server.ExitServer()
}

package main

import (
	"./gameserver"
	"log"
	"fmt"
)

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	// Запуск сервера
	server := gameserver.NewServer()
	server.StartListen()

	for {
		var input string
    	fmt.Scanln(&input)

		if input == "exit"{
			server.ExitServer()
			break
		}
	}

	//<-time.After(time.Second * 30)
	//server.ExitServer()
}

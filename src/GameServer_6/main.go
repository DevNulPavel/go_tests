package main

import (
	"./gameserver"
	"fmt"
	//"github.com/pkg/profile"
	"log"
	//"os"
	//"runtime/trace"
)

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	/*f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()*/


	// Запуск сервера
	server := gameserver.NewServer()
	server.StartServer()

	log.Print("Server started")

	for {
		var input string
		fmt.Scanln(&input)

		if input == "exit" {
			server.ExitServer()
			break
		}
	}

	//<-time.After(time.Second * 30)
	//server.ExitServer()
}

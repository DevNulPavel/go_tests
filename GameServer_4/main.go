package main

import (
	"./gameserver"
	"log"
	"fmt"
	"runtime/trace"
	"os"
    "github.com/pkg/profile"
)

func main() {
	defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

    err = trace.Start(f)
    if err != nil {
        panic(err)
    }

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

	defer trace.Stop()

	//<-time.After(time.Second * 30)
	//server.ExitServer()
}

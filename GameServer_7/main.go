package main

import (
	"./gameserver"
    //"fmt"
    "os"
	//"log"
	//"runtime/trace"
    //"github.com/pkg/profile"
    "log"
)

func testLoadPlatform() {
    f, err := os.Open("data/platforms.json")
    if err != nil {
        log.Println(err)
        return
    }
    defer f.Close()

    platforms, err := gameserver.NewPlatformsFromReader(f)
    if err != nil {
        log.Println(err)
        return
    }

    for key, value := range platforms{
        log.Printf("%s: %s\n", key, value.SymbolName)
    }
}

func main() {
	/*defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

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
    defer trace.Stop()
    */

	/*// Запуск сервера
	server := gameserver.NewServer()
	server.StartListen()

	for {
		var input string
    	fmt.Scanln(&input)

		if input == "exit"{
			server.ExitServer()
			break
		}
	}*/

	//<-time.After(time.Second * 30)
	//server.ExitServer()

    testLoadPlatform()
}

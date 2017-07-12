package main

import (
	"./gameserver"
	//"fmt"
	//"log"
	//"runtime/trace"
	//"github.com/pkg/profile"
	"log"
)

func testLoadPlatform() {
	// Load platforms
	platforms, err := gameserver.NewPlatformsFromFile("data/platforms.json")
	if err != nil {
		log.Println(err)
		return
	}

	// Load levels
	levels, err := gameserver.NewLevelsFromFile("data/level_graphics.json")
	if err != nil {
		log.Println(err)
		return
	}

	// Make arena
	item, exists := levels["egypt"]
	if exists {
		platformsForArena := make([]*gameserver.PlatformInfo, 0)
		for _, key := range item.Platforms {
			value, ok := platforms[key]
			if ok {
				platformsForArena = append(platformsForArena, value)
			}
		}

		if len(platformsForArena) > 0 {
			arena := gameserver.NewArena(platformsForArena)
			log.Printf("Arena: %v\n", arena)
		}
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

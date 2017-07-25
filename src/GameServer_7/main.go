package main

import (
	"gameserver"
	//"fmt"
	//"log"
	//"runtime/trace"
	//"github.com/pkg/profile"
	"fmt"
	"log"
	"github.com/pquerna/ffjson/ffjson"
)

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

	err := gameserver.MakeApp()
	if err != nil {
		log.Printf("App not created: %s\n", err)
		return
	}

	err = gameserver.GetApp().RunServer()
	if err != nil {
		log.Printf("Server not started: %s\n", err)
		return
	}

	for {
		var input string
		fmt.Scanln(&input)

		if input == "exit" {
			gameserver.GetApp().ExitServer()
			break
		}
	}
}

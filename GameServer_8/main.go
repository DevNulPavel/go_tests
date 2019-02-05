package main

import (
	"./gameserver"
	//"fmt"
	"log"
	"net/http"
)

func main() {
	// Запуск сервера
	server := gameserver.NewServer()
	server.StartServer()

	// HTTP сервер
	http.Handle("/", http.FileServer(http.Dir("web")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

	/*for {
		var input string
		fmt.Scanln(&input)

		if input == "exit" {
			server.ExitServer()
			break
		}
	}*/

	//<-time.After(time.Second * 30)
	//server.ExitServer()
}

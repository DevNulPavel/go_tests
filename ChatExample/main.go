package main

import "log"
import "net/http"
import "./chat"

func main() {
	log.SetFlags(log.Lshortfile)

	// websocket server
	server := chat.NewServer("/websocket")
	go server.Listen()

	// HTTP Server
	// static files
	http.Handle("/", http.FileServer(http.Dir("web")))
	// Listen
	error := http.ListenAndServe(":8080", nil)
	log.Fatal(error)
}

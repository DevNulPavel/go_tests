package main

import "log"
import "net/http"
import "./chat"

func main() {
    log.SetFlags(log.Lshortfile)

    // websocket server
    server := chat.NewServer("/entry")
    go server.Listen()

    // HTTP Server
    // static files
    http.Handle("/", http.FileServer(http.Dir("webroot")))
    error := http.ListenAndServe(":8080", nil)
    log.Fatal(error)
}
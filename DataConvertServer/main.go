package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Resources TEMP DIR:", os.TempDir())

	// Tools pathes
	initializeToolsPathes()

	// Direct tcp server
	go tcpServer()

	// HTTP server
	startHttpServer()
}

package main

import (
	"log"
	"os"
	"flag"
)

func main() {
    // Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Resources TEMP DIR:", os.TempDir())

	// Parse flags
    tcpPort := flag.Int("tcpPort", TCP_SERVER_PORT, "TCP port value")
    httpPort := flag.Int("httpPort", HTTP_SERVER_PORT, "HTTP port value")
    contentFolder := flag.String("contentPath", "", "Content path")
    flag.Parse()

	// Tools pathes
	initializeToolsPathes()

	// Direct tcp server
	go tcpServer(*tcpPort)

	// HTTP server
	startHttpServer(*httpPort, *contentFolder)
}

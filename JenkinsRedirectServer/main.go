package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Resources TEMP DIR:", os.TempDir())

	// Parse flags
	httpPort := flag.Int("httpPort", HTTP_SERVER_PORT, "HTTP port value")
	contentFolder := flag.String("contentPath", "", "Content path")
	flag.Parse()

	// HTTP server
	startHttpServer(*httpPort, *contentFolder)
}

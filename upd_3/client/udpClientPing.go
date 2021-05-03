package main

import (
	"fmt"
	"log"
	"net"
)

const serverAddressString = "127.0.0.1:9999" // "127.0.0.1:9999" "devnulpavel.ddns.net:9999"

func rawClient() {
	// Определяем адрес
	serverAddress, err := net.ResolveUDPAddr("udp", serverAddressString)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Resolved ip address: %s\n", serverAddress)

	// Подключение к серверу
	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
}

func main() {
}

package main

import (
	"fmt"
	"net"
	"time"
)

func HandleServerConnectionRaw(c *net.UDPConn, address *net.UDPAddr) {
	defer c.Close()

	const dataSize = 1024*64
	udpBuffer := make([]byte, dataSize)

	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

		readCount, receiveAddress, err := c.ReadFromUDP(udpBuffer)
		if err != nil {
			fmt.Println(err)
			return
		} else if readCount == 0 {
			fmt.Println("Disconnected")
			return
		}

        time.Sleep(1000 * time.Millisecond)

		// Теперь очередь ответной записи
		writtenCount, err := c.WriteToUDP(udpBuffer[0:readCount], receiveAddress)
		if err != nil {
			fmt.Println(err)
			return
		} else if writtenCount < readCount {
			fmt.Printf("Written less bytes - %d from \n", writtenCount, readCount)
			return
		}
	}
}

func server() {
	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Прослушивание сервера
	connection, err := net.ListenUDP("udp", address)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Server started\n")
	HandleServerConnectionRaw(connection, address)
}

func main() {
	go server()

	var input string
	fmt.Scanln(&input)
}

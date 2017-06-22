package main

import (
	"fmt"
	"net"
	"time"
    "encoding/binary"
)

func HandleServerConnectionRaw(c *net.UDPConn, address *net.UDPAddr) {
	defer c.Close()

	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

        udpBuffer := make([]byte, 1024)

        readCount, receiveAddress, err := c.ReadFromUDP(udpBuffer)
        if err != nil {
            fmt.Println(err)
            return
        }
        if readCount == 0{
            return
        }

        dataSize := binary.BigEndian.Uint32(udpBuffer[0:4])

        fmt.Printf("Data size: %d\n", dataSize)

        receiveData := udpBuffer[3:dataSize+4]

        fmt.Println("Received:", string(receiveData))

        // Теперь очередь ответной записи??
        writeBytes := []byte("ok")
        _, err = c.WriteToUDP(writeBytes, receiveAddress)
        if err != nil {
            fmt.Println(err)
            return
        }
	}
}

func server() {
	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp4", "localhost:9002")
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

    HandleServerConnectionRaw(connection, address)
}

func main() {
	go server()

	var input string
	fmt.Scanln(&input)
}

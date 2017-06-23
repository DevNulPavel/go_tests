package main

import (
	"fmt"
	"net"
	"time"
)

func handleServerConnectionRaw(c net.Conn) {
	defer c.Close()

    const dataSize = 8
    dataBytes := make([]byte, dataSize)

	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

		readCount, err := c.Read(dataBytes)
		if err != nil {
			fmt.Println(err)
			return
		} else if readCount == 0 {
			fmt.Println("Disconnected")
			return
		} else if readCount < dataSize {
			fmt.Println("Read size error")
			return
		}

		// Теперь очередь ответной записи??
		writeCount, err := c.Write(dataBytes)
		if err != nil {
			fmt.Println(err)
			return
		} else if writeCount < dataSize {
			fmt.Println("Write size error")
			return
		}
	}
}

func pingServer() {
    address, err := net.ResolveTCPAddr("tcp", ":9999")
    if err != nil {
        fmt.Println(err)
        return
    }

	// Прослушивание сервера
	ln, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		// Принятие соединения
		c, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
        err = c.SetNoDelay(true)
        if err != nil {
            fmt.Println(err)
            return
        }

		// Запуск горутины
		go handleServerConnectionRaw(c)
	}
}

func main() {
	go pingServer()

	var input string
	fmt.Scanln(&input)
}

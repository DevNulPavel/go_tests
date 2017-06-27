package main

import (
	"fmt"
	"net"
	"time"
)

func handleServerConnectionRaw(c net.Conn) {
	defer c.Close()

	const dataSize = 1024 * 64
	dataBytes := make([]byte, dataSize)

	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

		readCount, err := c.Read(dataBytes)
		if err != nil {
            fmt.Printf("Ping server read error: %s\n", err)
			return
		} else if readCount == 0 {
			fmt.Println("Disconnected")
			return
		}

        //fmt.Printf("Read size: %d\n", readCount)
        time.Sleep(1000 * time.Millisecond)

		// Теперь очередь ответной записи??
		writeCount, err := c.Write(dataBytes[0:readCount])
		if err != nil {
			fmt.Printf("Ping server write error: %s\n", err)
			return
		} else if writeCount < readCount {
			fmt.Printf("Write size error: %d from %d\n", writeCount, readCount)
			return
		}

        //fmt.Printf("Write size: %d\n", writeCount)
	}
}

func pingServer() {
	fmt.Print("Ping server started\n")

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

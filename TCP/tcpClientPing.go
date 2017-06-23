package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

func rawClientPing() {
    address, err := net.ResolveTCPAddr("tcp", "devnulpavel.ddns.net:9999")
    if err != nil {
        fmt.Println(err)
        return
    }

	// Подключение к серверу
	c, err := net.DialTCP("tcp", nil, address) // devnulpavel.ddns.net
	if err != nil {
		fmt.Println(err)
		return
	}
    err = c.SetNoDelay(true)
    if err != nil {
        fmt.Println(err)
        return
    }

	defer c.Close() // Отложеннное закрытие при выходе

	// TODO: в одном блоке данных могут быть получены сразу 2 сообщения

	const testDataSize = 8
	testData := make([]byte, testDataSize)

	// Бесконечный цикл записи
	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

		nowTime := uint64(time.Now().UnixNano())

		binary.BigEndian.PutUint64(testData, nowTime)

		writeSuccess := false
		writtenBytes := 0
		for {
			currentWritten, err := c.Write(testData[writtenBytes:])
			if err == nil {
				writtenBytes += currentWritten
				if writtenBytes == testDataSize {
					writeSuccess = true
					break
				} else {
					writtenBytes--
				}
			} else {
				log.Println(err)
				break
			}
		}

		if writeSuccess {
			// Теперь очередь чтения
			readSize, err := c.Read(testData)
			if err != nil {
				fmt.Println(err)
				return
			} else if readSize == 0 {
				fmt.Println("Disconnected")
				return
			}

			sendTimeUint := binary.BigEndian.Uint64(testData)
			sendTime := time.Unix(0, int64(sendTimeUint))

			pingValue := float64(time.Now().Sub(sendTime).Nanoseconds()) / 1000 / 1000
			fmt.Printf("Ping value = %fmsec\n", pingValue)

		} else {
			fmt.Println("Write failed")
			break
		}

        time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	rawClientPing()

	var input string
	fmt.Scanln(&input)
}

package main

import (
	"fmt"
	"log"
	"net"
	"time"
    "bytes"
    "encoding/binary"
)

func rawClient() {
	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", "127.0.0.1:9002")
	if err != nil {
        fmt.Println(err)
        return
	}

	// Подключение к серверу
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	// TODO: в одном блоке данных могут быть получены сразу 2 сообщения

	// Бесконечный цикл записи
	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

		testData := []byte("Test message")
		testDataSize := uint32(len(testData))

        buffer := new(bytes.Buffer)
        binary.Write(buffer, binary.BigEndian, testDataSize)
        buffer.Write(testData)

        uploadData := buffer.Bytes()


        // Пытаемся записать данные
		writeSuccess := false
		writtenBytes := 0
		for {
			currentWritten, err := c.Write(uploadData[writtenBytes:])
			if err == nil {
				writtenBytes += currentWritten
				if writtenBytes == len(uploadData){
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
			fmt.Println("Write success")

			// Теперь очередь чтения
			getData := make([]byte, 2)
			receivedCount, _, err := c.ReadFromUDP(getData)
			if err != nil {
                fmt.Println(err)
				return
			}
            fmt.Printf("Received data size = %d\n", receivedCount)

            // Проверяем результат
            if bytes.Equal(getData, []byte("ok")) {
                fmt.Println("Response OK")
            }else{
                fmt.Println("Response FAIL")
            }
		} else {
			fmt.Println("Write failed")
			break
		}
	}
}

func main() {
	rawClient()

	var input string
	fmt.Scanln(&input)
}

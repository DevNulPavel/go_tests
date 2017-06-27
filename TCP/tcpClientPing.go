package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func rawClientPing() {
	address, err := net.ResolveTCPAddr("tcp", "devnulpavel.ddns.net:9999") // devnulpavel.ddns.net   192.168.1.3
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

	// TODO:
    // в одном блоке данных могут быть получены сразу 2 сообщения
    // или сообщение может придти не полностью

	const timeBlockBegin = 128
	const testDataSize = 1024*1
	testData := make([]byte, testDataSize)

    lastReadTime := time.Now()

	// Бесконечный цикл записи
	for {
		timeVal := time.Now().Add(5 * time.Minute)
		c.SetDeadline(timeVal)

		nowTime := uint64(time.Now().UnixNano())

		binary.BigEndian.PutUint64(testData[timeBlockBegin : timeBlockBegin+8], nowTime)

        // Пишем
        writeTryCount := 0
        totalWrittenSize := 0
        for totalWrittenSize < testDataSize {
            writtenSize, err := c.Write(testData[totalWrittenSize:])
            if err != nil {
                fmt.Printf("Write error: %s\n", err)
                return
            }
            totalWrittenSize += writtenSize
            writeTryCount++
        }
        if totalWrittenSize < testDataSize {
            fmt.Printf("Written less than needed: %d from %d\n", totalWrittenSize, testDataSize)
            return
        }

        // Теперь очередь чтения
        readTryCount := 0
        totalReadSize := 0
        for totalReadSize < testDataSize {
            readSize, err := c.Read(testData[totalReadSize:])
            if err != nil {
                fmt.Printf("Reading error: %s\n", err)
                return
            } else if readSize == 0 {
                fmt.Println("Disconnected")
                return
            }
            totalReadSize += readSize
            readTryCount++
        }
        if totalReadSize < testDataSize {
            fmt.Printf("Read less than written: %d from %d\n", totalReadSize, testDataSize)
            return
        }

        now := time.Now()
        delayBetweenReadings := float64(now.Sub(lastReadTime).Nanoseconds()) / 1000.0 / 1000.0
        lastReadTime = now

        sendTimeUint := binary.BigEndian.Uint64(testData[timeBlockBegin : timeBlockBegin+8])
        sendTime := time.Unix(0, int64(sendTimeUint))

        netPingValue := float64(time.Now().Sub(sendTime).Nanoseconds()) / 1000.0 / 1000.0
        fmt.Printf("Readings delay = %fmsec, NetPing = %fmsec, writeTryCount = %d, readTryCount = %d\n", delayBetweenReadings, netPingValue, writeTryCount, readTryCount)
	}
}

func main() {
	rawClientPing()

	var input string
	fmt.Scanln(&input)
}

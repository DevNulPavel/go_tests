package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func rawClient() {
	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", "devnulpavel.ddns.net:9999") // devnulpavel.ddns.net
	if err != nil {
		fmt.Println(err)
		return
	}
    fmt.Printf("Resolved ip address: %s\n", address)

	// Подключение к серверу
	c, err := net.DialUDP("udp", nil, address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

    readErrorCounter := 0

	const dataSize = 16
	data := make([]byte, dataSize)
	var counter uint64 = 0

	// Бесконечный цикл записи
	for {
		sendTime := uint64(time.Now().UnixNano())
		binary.BigEndian.PutUint64(data[0:8], sendTime)
		binary.BigEndian.PutUint64(data[8:16], counter)

		// Пытаемся записать данные
        c.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
        currentWritten, err := c.Write(data)
		if err != nil {
            if err, ok := err.(net.Error); ok && err.Timeout() {
                fmt.Printf("Is WRITE timeout error: %s\n", err)
                continue
            }else{
                fmt.Println(err)
                return
            }
			return
		} else if currentWritten < dataSize {
			fmt.Printf("Written less bytes - %d\n", currentWritten)
			return
		}

		// теперь читаем
        c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		receivedCount, senderAddress, err := c.ReadFromUDP(data)
		if err != nil {
            if err, ok := err.(net.Error); ok && err.Timeout() {
                fmt.Printf("Is READ timeout error: %s\n", err)
                readErrorCounter++
                if readErrorCounter > 5 {
                    fmt.Println("Disconnected by many read timeouts")
                    return
                }else {
                    continue
                }
            }else{
                fmt.Printf("Read error: %s\n", err)
                return
            }
		} else if receivedCount == 0 {
            fmt.Println("Disconnected")
            return
        } else if receivedCount < dataSize {
            fmt.Printf("Received less data size - %d\n", receivedCount)
        }
        // Reset read counter error
        readErrorCounter = 0

		receivedSendTimeUint64 := binary.BigEndian.Uint64(data[0:8])
        receivedCounterUint64 := binary.BigEndian.Uint64(data[8:16])

        if receivedCounterUint64 != counter {
            fmt.Println("Receive counter error")
            continue
        }

        counter++

        // Ping
		receivedSendTime := time.Unix(0, int64(receivedSendTimeUint64))
		ping := float64(time.Now().Sub(receivedSendTime).Nanoseconds()) / 1000.0 / 1000.0
		fmt.Printf("Ping = %fms, from adress: %s\n", ping, senderAddress)

        //time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	rawClient()

	var input string
	fmt.Scanln(&input)
}

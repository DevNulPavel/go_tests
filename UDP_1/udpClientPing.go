package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
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

	// Гарантированный размер датаграммы
	const dataSize = 508
	const timeOffset = 0
	const counterOffset = 200
	const requestsCount = 100000

	var counter uint64 = 1 // По-умолчанию 0
	data := make([]byte, dataSize)
	readErrorCounter := 0
	startTime := time.Now()

	for i := 0; i < requestsCount; i++ {
		// Формируем данные
		sendTime := uint64(time.Now().UnixNano())
		binary.BigEndian.PutUint64(data[timeOffset:timeOffset+8], sendTime)
		binary.BigEndian.PutUint64(data[counterOffset:counterOffset+8], counter)

		// Пытаемся записать данные
		conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
		currentWritten, err := conn.Write(data)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				log.Printf("Is WRITE timeout error: %s\n", err)
				continue
			} else {
				log.Println(err)
				return
			}
			return
		} else if currentWritten < dataSize {
			log.Printf("Written less bytes - %d\n", currentWritten)
			return
		}

		// Теперь читаем
		conn.SetReadDeadline(time.Now().Add(5000 * time.Millisecond))
		receivedCount, receivedServerAddress, err := conn.ReadFromUDP(data)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				log.Printf("Is READ timeout error: %s\n", err)
				readErrorCounter++
				if readErrorCounter > 5 {
					log.Println("Disconnected by many read timeouts")
					return
				} else {
					continue
				}
			} else {
				log.Printf("Read error: %s\n", err)
				return
			}
		} else if receivedCount == 0 {
			log.Println("Disconnected")
			return
		} else if receivedServerAddress.IP.Equal(serverAddress.IP) == false {
			log.Printf("Invalid received server address: %s, must be: %s", receivedServerAddress, serverAddress)
			continue
		} else if receivedCount < dataSize {
			log.Printf("Received less data size - %d\n", receivedCount)
		}

		// Reset read counter error
		readErrorCounter = 0

		// Парсим прочитанные данные
		//receivedSendTimeUint64 := binary.BigEndian.Uint64(data[timeOffset : timeOffset+8])
		receivedCounterUint64 := binary.BigEndian.Uint64(data[counterOffset : counterOffset+8])

		// Проверяем валидность данных
		if receivedCounterUint64 != counter {
			log.Println("Receive counter error")
			continue
		}

		counter++

		// Ping
		//receivedSendTime := time.Unix(0, int64(receivedSendTimeUint64))
		//ping := float64(time.Now().Sub(receivedSendTime).Nanoseconds()) / 1000.0 / 1000.0
		//fmt.Printf("Ping = %fms, from adress: %s\n", ping, serverAddress)

		//time.Sleep(1000 * time.Millisecond)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()

	requestsPerSec := requestsCount / duration
	fmt.Printf("Requests per sec value: %f", requestsPerSec)
}

func main() {
	rawClient()

	//var input string
	//fmt.Scanln(&input)
}

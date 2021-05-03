package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type UDPMessage struct {
	conn *net.UDPConn
	addr *net.UDPAddr
	data []byte
}

type UDPClient struct {
	conn       *net.UDPConn
	addr       *net.UDPAddr
	lastActive time.Time
}

var clients = make(map[string]UDPClient)
var clientsReceivedCh = make(chan UDPMessage, 128) // Делаем буфферизацию на 128 сообщений
var clientsSendCh = make(chan []byte, 128)         // Делаем буфферизацию на 128 сообщений

var bufferIndex uint16 = 0
var totalBufferSize uint32 = 1024 * 1024
var fullBuffer = make([]byte, totalBufferSize)
var chunkSize uint16 = 256
var totalChunksCount uint32 = totalBufferSize / uint32(chunkSize)
var curChunkBegin uint32 = 0

func clientsLoop() {
	checkTimer := time.NewTimer(time.Second)
	for {
		select {
		case mesg := <-clientsReceivedCh:
			// TODO: строка как ключ - не очень быстро, но надежно
			clients[mesg.addr.String()] = UDPClient{
				mesg.conn,
				mesg.addr,
				time.Now(),
			}

		case data := <-clientsSendCh:
			for _, cl := range clients {
				_, err := cl.conn.WriteToUDP(data, cl.addr)
				if err != nil {
					fmt.Println("Write error: %s", err)
				}
			}
		case <-checkTimer.C:
			removeList := make([]string, 0)
			for key, val := range clients {
				if val.lastActive.Add(time.Second * 60).Before(time.Now()) {
					removeList = append(removeList, key)
				}
			}
			for _, key := range removeList {
				delete(clients, key)
			}
		}
	}
}

func sendDataLoop() {
	for i := 1; i <= 20; i++ {
		dataForSend := fullBuffer[curChunkBegin : uint32(curChunkBegin)+uint32(chunkSize)]
		chunkIndex := curChunkBegin
		curChunkBegin += uint32(chunkSize)

		metaInfoSize := 18
		msg := make([]byte, uint32(metaInfoSize)+uint32(chunkSize))

		var byteOffset = 0
		binary.BigEndian.PutUint16(msg[byteOffset:], bufferIndex)
		byteOffset += 2
		binary.BigEndian.PutUint32(msg[byteOffset:], totalBufferSize)
		byteOffset += 4
		binary.BigEndian.PutUint32(msg[byteOffset:], totalChunksCount)
		byteOffset += 4
		binary.BigEndian.PutUint32(msg[byteOffset:], chunkIndex)
		byteOffset += 4
		binary.BigEndian.PutUint32(msg[byteOffset:], curChunkBegin)
		byteOffset += 2
		binary.BigEndian.PutUint16(msg[byteOffset:], chunkSize)
		byteOffset += 2

		copy(msg[byteOffset:], dataForSend)

		clientsSendCh <- msg

		// Старт заново
		if uint32(curChunkBegin)+uint32(chunkSize) > totalBufferSize {
			curChunkBegin = 0
			bufferIndex++
			fmt.Println("Start again")
		}
	}
}

func main() {
	// Определяем адрес
	address, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		log.Println(err)
		return
	}

	// Создание приемника новых подключений
	connection, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Println("Error in accept: %s", err)
		return
	}
	defer connection.Close()

	log.Println("Listening started")

	readBuffer := make([]byte, 512)

	for {
		readCount, addr, err := connection.ReadFromUDP(readBuffer)
		if err != nil {
			log.Println("Read error: %s", err)
			return
		}

		resultBuffer := make([]byte, readCount)
		copy(readBuffer[0:readCount], resultBuffer)
		clientsReceivedCh <- UDPMessage{
			connection,
			addr,
			resultBuffer,
		}
	}
}

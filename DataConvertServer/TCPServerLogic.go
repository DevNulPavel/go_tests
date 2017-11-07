package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
    "crypto/md5"
    "bytes"
)

const TCP_SERVER_PORT = 10000
var CONFIG_DATA_SALT []byte = []byte{ 0xBA, 0xBA, 0xEB, 0x53, 0x78, 0x88, 0x32, 0x91 }



// TODO: Можно заменить на io.ReadAll()???
func readToFixedSizeBuffer(c net.Conn, dataBuffer []byte) int {
	dataBufferLen := len(dataBuffer)
	totalReadCount := 0
	for {
		readCount, err := c.Read(dataBuffer[totalReadCount:])
		if err == io.EOF {
			break
		}
		if readCount == 0 {
			break
		}
		if checkErr(err) {
			break
		}

		totalReadCount += readCount

		if totalReadCount == dataBufferLen {
			break
		}
	}
	return totalReadCount
}

func convertDataForConnection(c net.Conn, convertType, srcFileExtLen, resultFileExtLen, paramsStrSize byte, dataSize uint32) {
    if (srcFileExtLen == 0) || (resultFileExtLen == 0) {
        return
    }

    // Md5 calc
    hash := md5.New()

    // Input file extention
    srcFileExt := make([]byte, srcFileExtLen)
    totalReadCount := readToFixedSizeBuffer(c, srcFileExt)
    if totalReadCount < int(srcFileExtLen) {
        return
    }
    hash.Write(srcFileExt)

    // Result file extention
    resultFileExt := make([]byte, resultFileExtLen)
    totalReadCount = readToFixedSizeBuffer(c, resultFileExt)
    if totalReadCount < int(resultFileExtLen) {
        return
    }
    hash.Write(resultFileExt)

    // Convert parameters
    convertParams := make([]byte, paramsStrSize)
    totalReadCount = readToFixedSizeBuffer(c, convertParams)
    if totalReadCount < int(paramsStrSize) {
        return
    }
    hash.Write(convertParams)

    // Data salt
    hash.Write(CONFIG_DATA_SALT)

    // Hash receive
    dataHash := make([]byte, 16)
    totalReadCount = readToFixedSizeBuffer(c, dataHash)
    if totalReadCount < 16 {
        return
    }

    // Hash comparison
    receivedDataHash := hash.Sum(nil)
    if bytes.Equal(dataHash, receivedDataHash) == false {
        log.Printf("Invalid data hashes!!!")
        return
    }

    // File data
	dataBytes := make([]byte, dataSize)
	totalReadCount = readToFixedSizeBuffer(c, dataBytes)
	if uint32(totalReadCount) < dataSize {
		return
	}

	// Temp file udid
	uuid, err := newUUID()
	if checkErr(err) {
		return
	}

	// Save file
	filePath := os.TempDir() + uuid + string(srcFileExt)
	err = ioutil.WriteFile(filePath, dataBytes, 0644)
	if checkErr(err) {
		return
	}

	// Result file path
	resultFile := os.TempDir() + uuid + string(resultFileExt)

	// Defer remove files
	defer os.Remove(filePath)
	defer os.Remove(resultFile)

	// File convert
	err = convertFile(filePath, resultFile, uuid, convertType, string(convertParams))
	if checkErr(err) {
		return
	}

	// Open result file
	file, err := os.Open(resultFile)
	if checkErr(err) {
		return
	}
	defer file.Close()

	// Send file size
	stat, err := file.Stat()
	if checkErr(err) {
		return
	}
	statBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(statBytes, uint32(stat.Size()))
	writtenCount, writeErr := c.Write(statBytes)
	if writtenCount < 4 {
		return
	}
	if checkErr(writeErr) {
		return
	}

	// Send file
	var currentByte int64 = 0
	fileSendBuffer := make([]byte, 1024)
	for {
		fileReadCount, fileErr := file.ReadAt(fileSendBuffer, currentByte)
		if fileReadCount == 0 {
			break
		}

		writtenCount, writeErr := c.Write(fileSendBuffer[:fileReadCount])
		if checkErr(writeErr) {
			return
		}

		currentByte += int64(writtenCount)

		if (fileErr == io.EOF) && (fileReadCount == writtenCount) {
			break
		}
	}
}

func handleServerConnectionRaw(c net.Conn) {
	defer c.Close()

	timeVal := time.Now().Add(5 * time.Minute)
	c.SetDeadline(timeVal)
	c.SetWriteDeadline(timeVal)
	c.SetReadDeadline(timeVal)

	// Read convertDataForConnection type
	const metaSize = 8
	metaData := make([]byte, metaSize)
	totalReadCount := readToFixedSizeBuffer(c, metaData)
	if totalReadCount < metaSize {
		return
	}

	// Parse bytes
	convertType := metaData[0]
    srcFileExtLen := metaData[1]
    resultFileExtLen := metaData[2]
    paramsStrSize := metaData[3]
	dataSize := binary.BigEndian.Uint32(metaData[4:8])

	// Converting
	convertDataForConnection(c, convertType, srcFileExtLen, resultFileExtLen, paramsStrSize, dataSize)
}

func tcpServer() {
	// Прослушивание сервера
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", TCP_SERVER_PORT))
	if err != nil {
		log.Println(err)
		return
	}
	for {
		// Принятие соединения
		c, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		// Запуск горутины
		go handleServerConnectionRaw(c)
	}
}

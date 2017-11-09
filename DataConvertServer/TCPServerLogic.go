package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"time"
)

const TCP_SERVER_PORT = 10000

// Request type
const REQUEST_TYPE_PROC_COUNT = 1
const REQUEST_TYPE_CONVERT = 2

var CONFIG_DATA_SALT = []byte{0xBA, 0xBA, 0xEB, 0x53, 0x78, 0x88, 0x32, 0x91}


func convertDataForConnection(c net.Conn, convertType, srcFileExtLen, resultFileExtLen, paramsStrSize byte, dataSize uint32) {
	if (srcFileExtLen == 0) || (resultFileExtLen == 0) {
		return
	}

	// Md5 calc
	hash := md5.New()

	// Input file extention
	srcFileExt := make([]byte, srcFileExtLen)
	_, err := io.ReadFull(c, srcFileExt)
	if err != nil {
		return
	}
	hash.Write(srcFileExt)

	// Result file extention
	resultFileExt := make([]byte, resultFileExtLen)
	_, err = io.ReadFull(c, resultFileExt)
	if err != nil {
		return
	}
	hash.Write(resultFileExt)

	// Convert parameters
	convertParams := make([]byte, paramsStrSize)
	_, err = io.ReadFull(c, convertParams)
	if err != nil {
		return
	}
	hash.Write(convertParams)

	// Data salt
	hash.Write(CONFIG_DATA_SALT)

	// Hash receive
	dataHash := make([]byte, 16)
	_, err = io.ReadFull(c, dataHash)
	if err != nil {
		return
	}

	// Hash comparison
	calculatedDataHash := hash.Sum(nil)
	if bytes.Equal(dataHash, calculatedDataHash) == false {
		log.Printf("Invalid data hashes!!!")
		return
	}

	// File data
	dataBytes := make([]byte, dataSize)
	_, err = io.ReadFull(c, dataBytes)
	if err != nil {
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

	// File send
	io.Copy(c, file)
	file.Close()
	os.Remove(resultFile)
}

func handleServerConnectionRaw(c net.Conn) {
	defer c.Close()

	timeVal := time.Now().Add(5 * time.Minute)
	c.SetDeadline(timeVal)
	c.SetWriteDeadline(timeVal)
	c.SetReadDeadline(timeVal)

	requestTypeArr := make([]byte, 1)
	readCount, err := c.Read(requestTypeArr)
	if (readCount == 0) || (err != nil) {
		return
	}

	requestType := requestTypeArr[0]
	switch requestType {
	case REQUEST_TYPE_PROC_COUNT:
		numCount := byte(runtime.NumCPU())
		c.Write([]byte{numCount})
	case REQUEST_TYPE_CONVERT:
		// Read convertDataForConnection type
		const metaSize = 8
		metaData := make([]byte, metaSize)
		_, err := io.ReadFull(c, metaData)
		if err != nil {
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
}

func tcpServer(customPort int) {
    port := TCP_SERVER_PORT
    if customPort != 0 {
        port = customPort
    }

	// Прослушивание сервера
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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

package main

import (
    "encoding/binary"
    "net"
    "os"
    "io/ioutil"
    "io"
    "time"
    "strings"
    "fmt"
    "log"
)

const (
    TCP_SERVER_PORT = 10000
)


func convertDataForConnection(c net.Conn, convertType byte, dataSize int, srcFileExt, dstFileExt string) {
    dataBytes := make([]byte, dataSize)
    totalReadCount := readToFixedSizeBuffer(c, dataBytes)
    if totalReadCount < dataSize {
        return
    }

    uuid, err := newUUID()
    if checkErr(err) {
        return
    }

    // Save file
    filePath := os.TempDir() + uuid + srcFileExt
    err = ioutil.WriteFile(filePath, dataBytes, 0644)
    if checkErr(err) {
        return
    }

    // Result file path
    resultFile := os.TempDir() + uuid + dstFileExt

    // Defer remove files
    defer os.Remove(filePath)
    defer os.Remove(resultFile)

    // File convert
    err = convertFile(filePath, resultFile, uuid, convertType)
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

    timeVal := time.Now().Add(2 * time.Minute)
    c.SetDeadline(timeVal)
    c.SetWriteDeadline(timeVal)
    c.SetReadDeadline(timeVal)

    // Read convertDataForConnection type
    const metaSize = 21
    metaData := make([]byte, metaSize)
    totalReadCount := readToFixedSizeBuffer(c, metaData)
    if totalReadCount < metaSize {
        return
    }

    // Parse bytes
    convertType := metaData[0]
    dataSize := int(binary.BigEndian.Uint32(metaData[1:5]))
    srcFileExt := strings.Replace(string(metaData[5:13]), " ", "", -1)
    dstFileExt := strings.Replace(string(metaData[13:21]), " ", "", -1)

    //log.Println(convertTypeStr, dataSize)

    // Converting
    convertDataForConnection(c, convertType, dataSize, srcFileExt, dstFileExt)
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

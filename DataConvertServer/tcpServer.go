package main

import (
    "net"
    "time"
    "fmt"
    "strings"
    "encoding/binary"
    "crypto/rand"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
)

const (
    CONVERT_TYPE_IMAGE_TO_PVR   = 1
    CONVERT_TYPE_IMAGE_TO_PVRGZ = 2
    CONVERT_TYPE_SOUND          = 3
)

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
    uuid := make([]byte, 16)
    n, err := io.ReadFull(rand.Reader, uuid)
    if n != len(uuid) || err != nil {
        return "", err
    }
    // variant bits; see section 4.1.1
    uuid[8] = uuid[8]&^0xc0 | 0x80
    // version 4 (pseudo-random); see section 4.1.3
    uuid[6] = uuid[6]&^0xf0 | 0x40
    return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func checkErr(e error) bool {
    // TODO: Print stack
    if e != nil {
        fmt.Println(e)
        return true
    }
    return false
}

func convert(c net.Conn, convertType byte, dataSize int, srcFileExt, dstFileExt string) {
    dataBytes := make([]byte, dataSize)
    totalReadCount := 0
    for {
        bytesRef := dataBytes[totalReadCount:]
        curReadCount, readErr := c.Read(bytesRef)
        if readErr == io.EOF {
            break
        }
        if curReadCount == 0 {
            break
        }
        if checkErr(readErr) {
            break
        }
        totalReadCount += curReadCount
    }
    if totalReadCount < dataSize {
        return
    }

    uuid, err := newUUID()
    if checkErr(err) {
        return
    }

    // Save file
    filePath := "/tmp/" + uuid + srcFileExt
    err = ioutil.WriteFile(filePath, dataBytes, 0644)
    if checkErr(err) {
        return
    }

    // Result file path
    resultFile := "/tmp/" + uuid + dstFileExt

    // Defer remove files
    defer os.Remove(filePath)
    defer os.Remove(resultFile)

    // Convert file
    switch convertType {
    case CONVERT_TYPE_IMAGE_TO_PVR:
        pvrToolPath := "/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI"
        commandText := fmt.Sprintf("%s -f PVRTC2_4 -dither -q pvrtcbest -i %s -o %s", pvrToolPath, filePath, resultFile)
        command := exec.Command("bash", "-c", commandText)
        err = command.Run()
        if checkErr(err) {
            return
        }
    case CONVERT_TYPE_IMAGE_TO_PVRGZ:
        tempFileName := "/tmp/" + uuid + ".pvr"
        pvrToolPath := "/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI"
        convertCommandText := fmt.Sprintf("%s -f r8g8b8a8 -dither -q pvrtcbest -i %s -o %s; gzip -f --suffix gz -9 %s", pvrToolPath, filePath, tempFileName, tempFileName)
        command := exec.Command("bash", "-c", convertCommandText)
        err = command.Run()
        if checkErr(err) {
            return
        }
    case CONVERT_TYPE_SOUND:
        commandText := fmt.Sprintf("ffmpeg -i %s %s", filePath, resultFile)
        command := exec.Command("bash", "-c", commandText)
        err = command.Run()
        if checkErr(err) {
            return
        }
    }

    // Open result file
    file, err := os.Open(resultFile)
    if checkErr(err) {
        return
    }

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

        if (fileErr == io.EOF) && (fileReadCount == writtenCount)  {
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

    // Read convert type
    const metaSize = 21
    metaData := make([]byte, metaSize)
    totalReadCount := 0
    for {
        readCount, err := c.Read(metaData[totalReadCount:])
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
    }
    if totalReadCount < metaSize {
        return
    }

    // Parse bytes
    convertType := metaData[0]
    dataSize := int(binary.BigEndian.Uint32(metaData[1:5]))
    srcFileExt := strings.Replace(string(metaData[5:13]), " ", "", -1)
    dstFileExt := strings.Replace(string(metaData[13:21]), " ", "", -1)

    //fmt.Println(convertTypeStr, dataSize)

    // Converting
    convert(c, convertType, dataSize, srcFileExt, dstFileExt)
}

func server() {
    // Прослушивание сервера
    ln, err := net.Listen("tcp", ":10000")
    if err != nil {
        fmt.Println(err)
        return
    }
    for {
        // Принятие соединения
        c, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        // Запуск горутины
        go handleServerConnectionRaw(c)
    }
}

func main() {
    go server()

    var input string
    fmt.Scanln(&input)
}
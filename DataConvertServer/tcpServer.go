package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
    "path/filepath"
	"html/template"
)

const (
	SERVER_PORT = 10000
)

const (
	PVR_TOOL_PATH = "/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI"
	FFMPEG_PATH   = "ffmpeg"
)

const (
	CONVERT_TYPE_IMAGE_TO_PVR     = 1
    CONVERT_TYPE_IMAGE_TO_PVRGZ16 = 2
	CONVERT_TYPE_IMAGE_TO_PVRGZ32 = 3
	CONVERT_TYPE_SOUND            = 4
)

// Templates
var mainPageTemplate *template.Template = nil
var resultFilePageTemplate *template.Template = nil


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
		log.Println(e)
		return true
	}
	return false
}

func readToFizedSizeBuffer(c net.Conn, dataBuffer []byte) int {
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

func convertFile(srcFilePath, resultFile, uuid string, convertType byte) error {
	// Convert file
	switch convertType {
	case CONVERT_TYPE_IMAGE_TO_PVR:
		commandText := fmt.Sprintf("%s -f PVRTC2_4 -dither -q pvrtcbest -i %s -o %s", PVR_TOOL_PATH, srcFilePath, resultFile)
		command := exec.Command("bash", "-c", commandText)
		err := command.Run()
		return err
	case CONVERT_TYPE_IMAGE_TO_PVRGZ32:
		tempFileName := os.TempDir() + uuid + ".pvr"
		convertCommandText := fmt.Sprintf("%s -f r8g8b8a8 -dither -q pvrtcbest -i %s -o %s; gzip -f --suffix gz -9 %s", PVR_TOOL_PATH, srcFilePath, tempFileName, tempFileName)
		command := exec.Command("bash", "-c", convertCommandText)
		err := command.Run()
		return err
    case CONVERT_TYPE_IMAGE_TO_PVRGZ16:
        tempFileName := os.TempDir() + uuid + ".pvr"
        convertCommandText := fmt.Sprintf("%s -f r4g4b4a4 -dither -q pvrtcbest -i %s -o %s; gzip -f --suffix gz -9 %s", PVR_TOOL_PATH, srcFilePath, tempFileName, tempFileName)
        command := exec.Command("bash", "-c", convertCommandText)
        err := command.Run()
        return err
	case CONVERT_TYPE_SOUND:
		commandText := fmt.Sprintf("%s -y -i %s %s", FFMPEG_PATH, srcFilePath, resultFile)
		command := exec.Command("bash", "-c", commandText)
		err := command.Run()
		return err
	}
	return errors.New("Invalid convert type")
}

func convertDataForConnection(c net.Conn, convertType byte, dataSize int, srcFileExt, dstFileExt string) {
	dataBytes := make([]byte, dataSize)
	totalReadCount := readToFizedSizeBuffer(c, dataBytes)
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
	totalReadCount := readToFizedSizeBuffer(c, metaData)
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
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", SERVER_PORT))
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

func loadHtmlTemplates()  {
    var err error = nil
    // MainPage
    mainPageTemplate, err = template.ParseFiles("templates/fileConvertTemplate.html")
    if checkErr(err) {
        return
    }
}

func httpRootFunc(writer http.ResponseWriter, req *http.Request)  {
    // TODO: For debug!!!
    loadHtmlTemplates()

    if req.Method == "POST" {
        // Receive file data
        receivedFileData, fileHeader, err := req.FormFile("transferFile")
        if err != nil {
            http.Error(writer, err.Error(), 500)
            return
        }
        defer receivedFileData.Close()

        // Input file ext
        inputFileExt := filepath.Ext(fileHeader.Filename)
        if len(inputFileExt) == 0{
            http.Error(writer, err.Error(), 500)
            return
        }

        // Temp udid
        uuid, err := newUUID()
        if checkErr(err) {
            return
        }

        // Save file
        sourceFilePath := os.TempDir() + uuid + inputFileExt
        sourceFileFile, err := os.Create(sourceFilePath)
        if err != nil {
            http.Error(writer, err.Error(), 500)
            return
        }
        io.Copy(sourceFileFile, receivedFileData)

        // Manual closing after copy
        sourceFileFile.Close()
        receivedFileData.Close()

        // Remove defer source file
        defer os.Remove(sourceFilePath)

        // Convert type
        resultFilePath := ""
        convertType := req.FormValue("convertType")
        switch convertType {
        case "pvr":
            resultFilePath = os.TempDir() + uuid + ".pvr"
            err = convertFile(sourceFilePath, resultFilePath, uuid, CONVERT_TYPE_IMAGE_TO_PVR)
        case "pvrgz16":
            resultFilePath = os.TempDir() + uuid + ".pvrgz"
            err = convertFile(sourceFilePath, resultFilePath, uuid, CONVERT_TYPE_IMAGE_TO_PVRGZ16)
        case "pvrgz32":
            resultFilePath = os.TempDir() + uuid + ".pvrgz"
            err = convertFile(sourceFilePath, resultFilePath, uuid, CONVERT_TYPE_IMAGE_TO_PVRGZ32)
        case "m4a":
            resultFilePath = os.TempDir() + uuid + ".m4a"
            err = convertFile(sourceFilePath, resultFilePath, uuid, CONVERT_TYPE_SOUND)
        case "ogg":
            resultFilePath = os.TempDir() + uuid + ".ogg"
            err = convertFile(sourceFilePath, resultFilePath, uuid, CONVERT_TYPE_SOUND)
        default:
            http.Error(writer, err.Error(), 500)
            return
        }
        if err != nil {
            http.Error(writer, err.Error(), 500)
            return
        }

        // Source file remove + derer remove result file
        os.Remove(sourceFilePath)
        defer os.Remove(resultFilePath)

        // Out file
        uploadFile, err := os.Open(resultFilePath)
        if err != nil {
            http.Error(writer, err.Error(), 500)
            return
        }
        io.Copy(writer, uploadFile)

        uploadFile.Close()
        os.Remove(resultFilePath)
        return
    }

    mainPageTemplate.Execute(writer, nil)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

    log.Println("Resources TEMP DIR:", os.TempDir())

	// Direct tcp server
	go tcpServer()

	//fmt.Println("Press enter to exit")
	//var input string
	//fmt.Scanln(&input)

    // Http server
    loadHtmlTemplates()
	http.HandleFunc("/", httpRootFunc)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":8080", nil)
}

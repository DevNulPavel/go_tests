package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	HTTP_SERVER_PORT = 8080
)

// Templates
var mainPageTemplate *template.Template = nil

type HttpReceivedFileInfo struct {
	inputFileName string
	inputFileExt  string
	filePath      string
	fileUUID      string
}

type HttpSendFileInfo struct {
	filePath   string
	uploadName string
}

func loadHtmlTemplates() {
	var err error = nil
	// MainPage
	mainPageTemplate, err = template.ParseFiles("templates/fileConvertTemplate.html")
	if checkErr(err) {
		return
	}
}

func httpConvertWorkerFunction(inputChannel <-chan HttpReceivedFileInfo, resultChannel chan<- *HttpSendFileInfo, convertType, convertParams string) {

	for fileInfo := range inputChannel {
		// Convert type
		var err error = nil
		resultFilePath := ""
		uploadFileName := ""
		switch convertType {
		case "pvr":
			const extention = ".pvr"
			resultFilePath = os.TempDir() + fileInfo.fileUUID + extention
			uploadFileName = strings.Replace(fileInfo.inputFileName, fileInfo.inputFileExt, extention, -1)
            if len(convertParams) == 0 {
                convertParams = "-f PVRTC1_4 -pot + -dither -q pvrtcbest"
            }
			err = convertFile(fileInfo.filePath, resultFilePath, fileInfo.fileUUID, CONVERT_TYPE_IMAGE_PVR, convertParams)
		case "pvrgz16":
			const extention = ".pvrgz"
			resultFilePath = os.TempDir() + fileInfo.fileUUID + extention
			uploadFileName = strings.Replace(fileInfo.inputFileName, fileInfo.inputFileExt, extention, -1)
            if len(convertParams) == 0 {
                convertParams = "-f r4g4b4a4 -dither"
            }
			err = convertFile(fileInfo.filePath, resultFilePath, fileInfo.fileUUID, CONVERT_TYPE_IMAGE_PVRGZ, convertParams)
		case "pvrgz32":
			const extention = ".pvrgz"
			resultFilePath = os.TempDir() + fileInfo.fileUUID + extention
			uploadFileName = strings.Replace(fileInfo.inputFileName, fileInfo.inputFileExt, extention, -1)
            if len(convertParams) == 0 {
                convertParams = "-f r8g8b8a8 -dither"
            }
			err = convertFile(fileInfo.filePath, resultFilePath, fileInfo.fileUUID, CONVERT_TYPE_IMAGE_PVRGZ, convertParams)
		case "webp":
			const extention = ".webp"
			resultFilePath = os.TempDir() + fileInfo.fileUUID + extention
			uploadFileName = strings.Replace(fileInfo.inputFileName, fileInfo.inputFileExt, extention, -1)
            if len(convertParams) == 0 {
                convertParams = "-q 96"
            }
			err = convertFile(fileInfo.filePath, resultFilePath, fileInfo.fileUUID, CONVERT_TYPE_IMAGE_WEBP, convertParams)
		case "m4a":
			const extention = ".m4a"
			resultFilePath = os.TempDir() + fileInfo.fileUUID + extention
			uploadFileName = strings.Replace(fileInfo.inputFileName, fileInfo.inputFileExt, extention, -1)
            /*if len(convertParams) == 0 {
                convertParams = ""
            }*/
			err = convertFile(fileInfo.filePath, resultFilePath, fileInfo.fileUUID, CONVERT_TYPE_SOUND_FFMPEG, convertParams)
		case "ogg":
			const extention = ".ogg"
			resultFilePath = os.TempDir() + fileInfo.fileUUID + extention
			uploadFileName = strings.Replace(fileInfo.inputFileName, fileInfo.inputFileExt, extention, -1)
            /*if len(convertParams) == 0 {
                convertParams = ""
            }*/
			err = convertFile(fileInfo.filePath, resultFilePath, fileInfo.fileUUID, CONVERT_TYPE_SOUND_FFMPEG, convertParams)
		default:
			os.Remove(fileInfo.filePath)
			resultChannel <- nil
			return
		}
		if err != nil {
			os.Remove(fileInfo.filePath)
			resultChannel <- nil
			return
		}

		// Source file remove + derer remove result file
		os.Remove(fileInfo.filePath)

		// Add to list of files
		info := HttpSendFileInfo{
			filePath:   resultFilePath,
			uploadName: uploadFileName,
		}
		resultChannel <- &info
	}
}

func httpRootFunc(writer http.ResponseWriter, req *http.Request) {
    // TODO: Debug
    loadHtmlTemplates()

	if req.Method == "POST" {
		receivedFiles := make([]HttpReceivedFileInfo, 0)

		// Read input files count
		req.ParseMultipartForm(2 << 16)
		headers := req.MultipartForm.File["transferFile"]
		if len(headers) == 0 {
			http.Error(writer, errors.New("No files").Error(), 500)
			return
		}

		// Save files to temp folders
		for _, fileHeader := range headers {
			// Receive file data
			receivedFileData, err := fileHeader.Open()
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}

			// Input file ext
			inputFileName := fileHeader.Filename
			inputFileExt := filepath.Ext(inputFileName)
			if len(inputFileExt) == 0 {
				receivedFileData.Close()
				continue
			}

			// Temp udid
			uuid, err := newUUID()
			if checkErr(err) {
				receivedFileData.Close()
				return
			}

			// Save file
			sourceFilePath := os.TempDir() + uuid + inputFileExt
			sourceFile, err := os.Create(sourceFilePath)
			if err != nil {
				receivedFileData.Close()
				http.Error(writer, err.Error(), 500)
				return
			}
			io.Copy(sourceFile, receivedFileData)

			// Manual closing after copy
			sourceFile.Close()
			receivedFileData.Close()

			fileInfo := HttpReceivedFileInfo{
				inputFileName: inputFileName,
				inputFileExt:  inputFileExt,
				filePath:      sourceFilePath,
				fileUUID:      uuid,
			}
			receivedFiles = append(receivedFiles, fileInfo)
		}

		if len(receivedFiles) == 0 {
			http.Error(writer, errors.New("No files").Error(), 500)
			return
		}

		// Send files list
		sendFiles := make([]HttpSendFileInfo, 0)

		// Convert Type
		convertType := req.FormValue("convertType")

		// Custom parameters
        customParameters := req.FormValue("customParams")

		// Workers pool
		convertChannel := make(chan HttpReceivedFileInfo)
		resultChannel := make(chan *HttpSendFileInfo)
		for i := 0; (i < len(receivedFiles)) && (i < runtime.NumCPU()); i++ {
			go httpConvertWorkerFunction(convertChannel, resultChannel, convertType, customParameters)
		}

		// Perform convert
		for _, fileInfo := range receivedFiles {
			convertChannel <- fileInfo
		}

		// Get results
		for i := 0; i < len(receivedFiles); i++ {
			convertResult := <-resultChannel
			if convertResult != nil {
				sendFiles = append(sendFiles, *convertResult)
			}
		}

		// Close channels for exit from workers pool
		close(convertChannel)
		close(resultChannel)

		// Check files count
		if len(sendFiles) == 0 {
			http.Error(writer, errors.New("No files").Error(), 500)
			return
		}

		// Compress if needed
		var sendInfo HttpSendFileInfo
		if len(sendFiles) == 1 {
			// Only one file - send as it
			sendInfo = sendFiles[0]
		} else {
			// Compress many files into one file
			zipFileUDID, err := newUUID()
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}

			// Compress files into zip
			zipFilePath := os.TempDir() + zipFileUDID + ".zip"
			filesPathes := make([]string, 0, len(sendFiles))
			filesAliases := make([]string, 0, len(sendFiles))
			for _, sendFileInfo := range sendFiles {
				filesPathes = append(filesPathes, sendFileInfo.filePath)
				filesAliases = append(filesAliases, sendFileInfo.uploadName)
			}
			// Compress into zip
			err = zipFiles(zipFilePath, filesPathes, filesAliases)
			if err != nil {
				// Files remove
				for _, sendFileInfo := range sendFiles {
					os.Remove(sendFileInfo.filePath)
				}
				// Send error
				http.Error(writer, err.Error(), 500)
				return
			}

			// Manual remove files
			for _, sendFileInfo := range sendFiles {
				os.Remove(sendFileInfo.filePath)
			}

			// Update send info
			sendInfo.filePath = zipFilePath
			sendInfo.uploadName = "files_archive.zip"
		}

		// Defer result remove
		defer os.Remove(sendInfo.filePath)

		// Out file
		uploadFile, err := os.Open(sendInfo.filePath)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
		uploadFileStat, err := uploadFile.Stat()
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}

		// Header for file upload
		writer.Header().Set("ContentType", "application/octet-stream")
		writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=\"%s\"", sendInfo.uploadName))
		writer.Header().Set("Content-Length:", fmt.Sprintf("%d", uploadFileStat.Size()))

		// Write file to stream
		io.Copy(writer, uploadFile)

		// Close and remove files
		uploadFile.Close()
		os.Remove(sendInfo.filePath)

		return
	}

	mainPageTemplate.Execute(writer, nil)
}

func startHttpServer(customPort int) {
	port := HTTP_SERVER_PORT
	if customPort != 0 {
		port = customPort
	}

	// Http server
	loadHtmlTemplates()
	http.HandleFunc("/", httpRootFunc)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

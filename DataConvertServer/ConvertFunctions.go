package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"log"
)

const (
	CONVERT_TYPE_IMAGE_PVR    = 1
	CONVERT_TYPE_IMAGE_PVRGZ  = 2
	CONVERT_TYPE_SOUND_FFMPEG = 3
)

const (
	PVR_TOOL_PATH = "/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI"
	FFMPEG_PATH   = "ffmpeg"
)

func convertFile(srcFilePath, resultFile, uuid string, convertType byte, paramsStr string) error {
	// Convert file
	switch convertType {
	case CONVERT_TYPE_IMAGE_PVR:
		// Params examples
		// -f PVRTC2_4 -dither -q pvrtcbest
		commandText := fmt.Sprintf("%s %s -i %s -o %s", PVR_TOOL_PATH, paramsStr, srcFilePath, resultFile)
		command := exec.Command("bash", "-c", commandText)
		err := command.Run()
		if err != nil {
			output, _ := command.CombinedOutput()
			log.Println(string(output))
		}
		return err
	case CONVERT_TYPE_IMAGE_PVRGZ:
		// Params examples
		// -f r8g8b8a8 -dither -q pvrtcbest
		// -f r4g4b4a4 -dither -q pvrtcbest
		tempFileName := os.TempDir() + uuid + ".pvr"
		convertCommandText := fmt.Sprintf("%s %s -i %s -o %s; gzip -f --suffix gz -9 %s", PVR_TOOL_PATH, paramsStr, srcFilePath, tempFileName, tempFileName)
		command := exec.Command("bash", "-c", convertCommandText)
		err := command.Run()
		if err != nil {
			output, _ := command.CombinedOutput()
			log.Println(string(output))
		}
		return err
	case CONVERT_TYPE_SOUND_FFMPEG:
		commandText := fmt.Sprintf("%s -y %s -i %s %s", FFMPEG_PATH, paramsStr, srcFilePath, resultFile)
		command := exec.Command("bash", "-c", commandText)
		err := command.Run()
		if err != nil {
			output, _ := command.CombinedOutput()
			log.Println(string(output))
		}
		return err
	}
	return errors.New("Invalid convert type")
}

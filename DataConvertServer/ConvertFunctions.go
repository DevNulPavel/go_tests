package main

import (
    "fmt"
    "os/exec"
    "os"
    "errors"
)

const (
    CONVERT_TYPE_IMAGE_TO_PVR     = 1
    CONVERT_TYPE_IMAGE_TO_PVRGZ16 = 2
    CONVERT_TYPE_IMAGE_TO_PVRGZ32 = 3
    CONVERT_TYPE_SOUND            = 4
)

const (
    PVR_TOOL_PATH = "/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI"
    FFMPEG_PATH   = "ffmpeg"
)

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

#! /usr/bin/env bash

export PVR_TOOL_PATH '/Applications/Imagination/PowerVR_Graphics/PowerVR_Tools/PVRTexTool/CLI/OSX_x86/PVRTexToolCLI'
export FFMPEG_TOOL_PATH '/usr/local/bin/ffmpeg'
export WEBP_TOOL_PATH '/usr/local/bin/cwebp'
# launchctl setenv PATH $PATH
./DataConvertServer -tcpPort 10000 -httpPort 18000
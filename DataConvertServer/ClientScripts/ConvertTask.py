#! /usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import os.path
import multiprocessing
import subprocess
import GlobalVariables as GV
from ConvertStatus import *
from RemoteAccess import *


####################################################################################################################################

class ConvertTask:
    def __init__(self, fileFullPath, resultFolder, resultFileExt, convertType, bits, customParams, processingFolder, platform, remoteServerConvert):
        self.fileFullPath = fileFullPath
        self.resultFolder = resultFolder
        self.convertType = convertType
        self.processingFolder = processingFolder
        self.platform = platform
        self.remoteServerConvert = remoteServerConvert
        self.resultFileExt = resultFileExt
        self.bits = bits
        self.customParams = customParams

    def makeResultFullDirPath(self, platform):
        dirName = GV.PLATFORMS_INFO[platform]["folder"]
        resultFullDirPath = os.path.join(self.resultFolder, dirName, self.processingFolder)
        resultFullDirPath = os.path.abspath(resultFullDirPath)
        return resultFullDirPath

    def perform(self):
        # Get result full path
        resultFullDirPath = self.makeResultFullDirPath(self.platform)

        # Make result folder
        os.makedirs(resultFullDirPath, exist_ok=True)

        # New filename
        fileName = os.path.split(self.fileFullPath)[1]
        fileNameExt = os.path.splitext(fileName)[1]
        newFileName = str.replace(fileName, fileNameExt, self.resultFileExt)

        # Result file path
        resultFilePath = os.path.join(resultFullDirPath, newFileName)

        # Convert params
        if self.convertType == GV.CONVERT_TYPE_IMAGE_PVR:  # PVR
            checkKey = "PVR"
            defaultParamsVal = GV.PVR_DEFAULT_PARAMS
        elif self.convertType == GV.CONVERT_TYPE_IMAGE_PVRGZ:  # PVRGZ
            if self.bits == 32:
                checkKey = "PVRGZ32"
                defaultParamsVal = GV.PVRGZ32_DEFAULT_PARAMS
            elif self.bits == 16:
                checkKey = "PVRGZ16"
                defaultParamsVal = GV.PVRGZ16_DEFAULT_PARAMS
            else:
                return None
        elif self.convertType == GV.CONVERT_TYPE_IMAGE_WEBP:  # WEBP
            checkKey = "WEBP"
            defaultParamsVal = GV.WEBP_DEFAULT_PARAMS
        elif self.convertType == GV.CONVERT_TYPE_FFMPEG:  # FFMPEG
            checkKey = "FFMPEG"
            defaultParamsVal = GV.CUSTOM_PARAMS_CONFIG
        else:
            return None

        # Convert params
        if checkKey in self.customParams:
            convertParams = self.customParams[checkKey]
        elif checkKey in GV.CUSTOM_PARAMS_CONFIG:
            convertParams = GV.CUSTOM_PARAMS_CONFIG[checkKey]
        else:
            convertParams = defaultParamsVal

        # Platfrom convert params
        if type(convertParams) is dict:
            key = GV.PLATFORMS_INFO[self.platform]["paramName"]
            if key in convertParams:
                convertParams = convertParams[key]
            else:
                convertParams = defaultParamsVal

        # Converting
        if self.remoteServerConvert:
            process = multiprocessing.Process(target=serverConvertFunctionMultiproc,
                                              args=(str(self.fileFullPath),
                                                    fileNameExt.encode('utf-8'),
                                                    str(resultFilePath),
                                                    self.resultFileExt.encode('utf-8'),
                                                    self.convertType,
                                                    convertParams.encode('utf-8')))
            process.start()
            return MultiprocessStatus(process)
        else:
            if self.convertType == GV.CONVERT_TYPE_IMAGE_PVR:       # PVR
                command = "{0} {1} -i '{2}' -o '{3}'".format(GV.PVR_TOOL_PATH, convertParams, self.fileFullPath, resultFilePath)
                process = subprocess.Popen(command, shell=True)
                return SubprocessStatus(process)

            elif self.convertType == GV.CONVERT_TYPE_IMAGE_PVRGZ:   # PVRGZ
                tempFilePath = os.path.split(self.fileFullPath)[1]
                tempFilePathExt = os.path.splitext(tempFilePath)[1]
                tempFilePath = str.replace(tempFilePath, tempFilePathExt, ".pvr")
                tempFilePath = os.path.join(resultFullDirPath, tempFilePath)

                command = "{0} {1} -i '{2}' -o '{3}'; gzip -f --suffix gz -9 '{3}'".format(
                    GV.PVR_TOOL_PATH,
                    convertParams,
                    self.fileFullPath,
                    tempFilePath)
                process = subprocess.Popen(command, shell=True)
                return SubprocessStatus(process)

            elif self.convertType == GV.CONVERT_TYPE_IMAGE_WEBP:    # WEBP
                command = "{0} {1} '{2}' -o '{3}'".format(GV.WEBP_TOOL_PATH, convertParams, self.fileFullPath,
                                                          resultFilePath)
                process = subprocess.Popen(command, shell=True)
                return SubprocessStatus(process)

            elif self.convertType == GV.CONVERT_TYPE_FFMPEG:  # FFMPEG
                command = "{0} -i '{1}' -y {2} '{3}'".format(GV.FFMPEG_TOOL_PATH, self.fileFullPath, convertParams, resultFilePath)
                process = subprocess.Popen(command, shell=True)
                return SubprocessStatus(process)
            else:
                return None

    def canBeRestarted(self):
        if self.remoteServerConvert:
            return True
        return False

    def restart(self):
        self.remoteServerConvert = False
        return self.perform()


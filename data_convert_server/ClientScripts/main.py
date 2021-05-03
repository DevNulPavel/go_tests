#! /usr/bin/env python3
# -*- coding: utf-8 -*-

import os.path
import json
import getopt
import hashlib
import GlobalVariables as GV
from ConvertTask import *
from ConvertTasksQueue import *


def md5ForFile(filePath):
    hash_md5 = hashlib.md5()
    with open(filePath, "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            hash_md5.update(chunk)
    return hash_md5.hexdigest()


def safeMakeDirs(dirPath):
    if not os.path.exists(dirPath):
        try:
            os.makedirs(dirPath)
        except:
            pass


def handleFile(folderConfigData, fileFullPath, processingFolder, targetsFolder, customParams, filesHashesDict, filesNamesDict):
    if os.path.exists(fileFullPath) == False:
        return

    # Calc file hash
    currentHash = md5ForFile(fileFullPath)

    # Check hash exists
    if fileFullPath in filesHashesDict:
        savedHash = filesHashesDict[fileFullPath]
        # Compare hashes
        if savedHash == currentHash:
            return

    needSaveHash = False

    sourceFilePath = os.path.join(processingFolder, os.path.split(fileFullPath)[1])

    # Compare file ext
    fileExt = os.path.splitext(fileFullPath)[1]
    fileExt = fileExt.lower()
    if fileExt in GV.IMAGES_FILES_EXTENTIONS:       # Images
        needSaveHash = True

        # Formats
        imageFormatIOS = folderConfigData.get("imageFormatIOS", "pvrgz")
        imageFormatAndroid = folderConfigData.get("imageFormatAndroid", "pvrgz")
        pvrgzBitSize = folderConfigData.get("imagePVRGZBits", 32)

        # Convert list
        convertList = [(imageFormatIOS, GV.PLATFORM_IOS),
                       (imageFormatAndroid, GV.PLATFORM_ANDROID)]

        # Converting
        for imageFormat, platform in convertList:
            # Make platform dict
            if platform not in filesNamesDict:
                filesNamesDict[platform] = {}

            if imageFormat == "pvrgz":
                convertTask = ConvertTask(fileFullPath, targetsFolder, ".pvrgz", GV.CONVERT_TYPE_IMAGE_PVRGZ, pvrgzBitSize,
                                          customParams, processingFolder, platform, GV.REMOTE_SERVER_ACCESSIBLE)
                GV.CONVERT_QUEUE.append(convertTask)
                filesNamesDict[platform][sourceFilePath] = sourceFilePath.replace(fileExt, ".pvrgz")
            elif imageFormat == "pvr":
                convertTask = ConvertTask(fileFullPath, targetsFolder, ".pvr", GV.CONVERT_TYPE_IMAGE_PVR, 0,
                                          customParams, processingFolder, platform, GV.REMOTE_SERVER_ACCESSIBLE)
                GV.CONVERT_QUEUE.append(convertTask)
                filesNamesDict[platform][sourceFilePath] = sourceFilePath.replace(fileExt, ".pvr")
            elif imageFormat == "webp":
                convertTask = ConvertTask(fileFullPath, targetsFolder, ".webp", GV.CONVERT_TYPE_IMAGE_WEBP, 0,
                                          customParams, processingFolder, platform, GV.REMOTE_SERVER_ACCESSIBLE)
                GV.CONVERT_QUEUE.append(convertTask)
                filesNamesDict[platform][sourceFilePath] = sourceFilePath.replace(fileExt, ".webp")

    elif fileExt in GV.SOUNDS_FILES_EXTENTIONS:     # Images
        needSaveHash = True

        soundFormatIOS = folderConfigData.get("soundFormatIOS", "m4a")
        soundFormatAndroid = folderConfigData.get("soundFormatAndroid", "ogg")

        # Convert list
        convertList = [(soundFormatIOS, GV.PLATFORM_IOS),
                       (soundFormatAndroid, GV.PLATFORM_ANDROID)]

        for soundFormat, platform in convertList:
            # Make platform dict
            if platform not in filesNamesDict:
                filesNamesDict[platform] = {}

            outFileExt = "." + soundFormat
            convertTask = ConvertTask(fileFullPath, targetsFolder, outFileExt, GV.CONVERT_TYPE_FFMPEG, 0,
                                      customParams, processingFolder, platform, GV.REMOTE_SERVER_ACCESSIBLE)
            GV.CONVERT_QUEUE.append(convertTask)
            filesNamesDict[platform][sourceFilePath] = sourceFilePath.replace(fileExt, outFileExt)

    elif fileExt in GV.VIDEO_FILES_EXTENTIONS:      # Video
        needSaveHash = True

        videoFormatIOS = folderConfigData.get("videoFormatIOS", "webm")
        videoFormatAndroid = folderConfigData.get("videoFormatAndroid", "webm")

        # Convert list
        convertList = [(videoFormatIOS, GV.PLATFORM_IOS),
                       (videoFormatAndroid, GV.PLATFORM_ANDROID)]

        for videoFormat, platform in convertList:
            # Make platform dict
            if platform not in filesNamesDict:
                filesNamesDict[platform] = {}

            outFileExt = "." + videoFormat
            convertTask = ConvertTask(fileFullPath, targetsFolder, outFileExt, GV.CONVERT_TYPE_FFMPEG, 0,
                                      customParams, processingFolder, platform, GV.REMOTE_SERVER_ACCESSIBLE)
            GV.CONVERT_QUEUE.append(convertTask)
            filesNamesDict[platform][sourceFilePath] = sourceFilePath.replace(fileExt, outFileExt)

    # Save hash
    if needSaveHash:
        filesHashesDict[fileFullPath] = currentHash

    # Check convert queue
    checkConvertQueue()


def processHashesFile(filesHashesDict, hashesFilePath, resourcesFolder, targetsFolder):
    # Calculate deleted files
    deletedFiles = []
    for sourceFilePath in filesHashesDict:
        sourceFileFullPath = os.path.join(resourcesFolder, sourceFilePath)
        sourceFileFullPath = os.path.abspath(sourceFileFullPath)
        if os.path.exists(sourceFileFullPath) == False:
            deletedFiles.append(sourceFilePath)

    # Delete from hash dict
    for key in deletedFiles:
        del filesHashesDict[key]

    # Save hashes
    safeMakeDirs(os.path.dirname(hashesFilePath))
    with open(hashesFilePath, "w") as hashesFile:
        json.dump(filesHashesDict, hashesFile, indent=4, sort_keys=True)

    # Delete old result files
    if len(deletedFiles) > 0:
        for platformInfoKey in GV.PLATFORMS_INFO.keys():
            platformResFolder = os.path.abspath(
                os.path.join(targetsFolder, GV.PLATFORMS_INFO[platformInfoKey]["folder"]))
            platformFileFullPath = os.path.join(platformResFolder, GV.NAMES_FILE_NAME)
            if os.path.exists(platformFileFullPath):
                # Load json
                with open(platformFileFullPath, "r") as file:
                    filesInfoJson = json.load(file)
                # Delete files + update json
                for path in deletedFiles:
                    path = os.path.relpath(path, resourcesFolder)
                    if path in filesInfoJson:
                        resultFilePath = os.path.join(platformResFolder, filesInfoJson[path])
                        if os.path.exists(resultFilePath):
                            os.remove(resultFilePath)
                        del filesInfoJson[path]
                # Save updated json
                with open(platformFileFullPath, "w") as file:
                    json.dump(filesInfoJson, file, indent=4, sort_keys=True)


def processNewNamesFiles(filesNamesDict, targetsFolder, resourcesFolder):
    # Save names + handle removed files
    for platformKey in filesNamesDict:
        resultDict = filesNamesDict[platformKey]
        platformFileFullPath = os.path.join(targetsFolder, GV.PLATFORMS_INFO[platformKey]["folder"], GV.NAMES_FILE_NAME)

        if os.path.exists(platformFileFullPath):
            # Update old file
            with open(platformFileFullPath, "r") as namesFile:
                namesOldJson = json.load(namesFile)
            # Update old json
            namesOldJson.update(resultDict)
            # Check file exists
            deleteKeys = []
            for key in namesOldJson:
                # Source file full path
                sourceFullPath = os.path.abspath(os.path.join(resourcesFolder, key))
                if os.path.exists(sourceFullPath) == False:
                    deleteKeys.append(key)
                    value = namesOldJson[key]
                    # Old target file delete
                    resultFullPath = os.path.join(targetsFolder, GV.PLATFORMS_INFO[platformKey]["folder"], value)
                    resultFullPath = os.path.abspath(resultFullPath)
                    # Delete old file from target folder
                    if os.path.exists(resultFullPath):
                        os.remove(resultFullPath)

            # Remove not existing files
            for key in deleteKeys:
                del namesOldJson[key]

            # Save json
            with open(platformFileFullPath, "w") as namesFile:
                json.dump(namesOldJson, namesFile, indent=4, sort_keys=True)
        else:
            # New file
            with open(platformFileFullPath, "w") as namesFile:
                json.dump(resultDict, namesFile, indent=4, sort_keys=True)


def processResources(configPath, hashesFilePath, resourcesFolder, targetsFolder):
    # Config
    with open(configPath, "r") as configFile:
        configDict = json.load(configFile)
    if configDict == None:
        return

    # Hashes
    filesHashesDict = {}
    if os.path.exists(hashesFilePath):
        with open(hashesFilePath, "r") as hashesFile:
            filesHashesDict = json.load(hashesFile)

    # Files names
    filesNamesDict = {}

    # Custom server params
    if "connectionParams" in configDict:
        params = configDict["connectionParams"]
        if "server" in params:
            GV.SERVER_ADDRESS = params["server"]
        if "port" in params:
            GV.SERVER_PORT = params["port"]

    # Update server meta info after addresses update
    updateServerMeta()

    # Custom convert params
    if "customConvertParams" in configDict:
        GV.CUSTOM_PARAMS_CONFIG = configDict["customConvertParams"]

    # Process folders
    if "folders" in configDict:
        folders = configDict["folders"]
        for folderConfigData in folders:
            if "folder" not in folderConfigData:
                continue

            folder = folderConfigData["folder"]

            # Folder custom params
            if "foldersCustomParams" in folderConfigData:
                customParams = folderConfigData["foldersCustomParams"]
            else:
                customParams = {}

            # Ignore folders
            if "ignoreFolders" in folderConfigData:
                ignoreFolders = folderConfigData["ignoreFolders"]
            else:
                ignoreFolders = []

            # Recursive or not?
            if "recursive" in folderConfigData:
                recursive = bool(folderConfigData["recursive"])
            else:
                recursive = True

            folderFullPath = os.path.abspath(os.path.join(resourcesFolder, folder))
            for root, dirs, files in os.walk(folderFullPath, topdown=True):
                relativePath = os.path.relpath(root, folderFullPath)

                # Ignore folders
                continueWalk = False
                for folderName in ignoreFolders:
                    if relativePath.startswith(folderName):
                        continueWalk = True
                        break
                if continueWalk:
                    continue

                # Cur folder params (Path begining check maybe???)
                if relativePath != ".":
                    relativePath = "./"+relativePath
                foundParams = None
                maxCommonPrefix = 0
                for paramKey in customParams:
                    if paramKey != ".":
                        testPath = "./"+paramKey
                    else:
                        testPath = paramKey
                    commonPrefix = os.path.commonprefix([relativePath, testPath])
                    if (commonPrefix != None) and (len(commonPrefix) > maxCommonPrefix):
                        maxCommonPrefix = len(commonPrefix)
                        foundParams = customParams[paramKey]

                # Get custom params value
                if foundParams:
                    curFolderParams = foundParams
                else:
                    curFolderParams = {}

                # Cur processing folder
                processingFolder = os.path.relpath(root, resourcesFolder)

                # Handle files
                for file in files:
                    if file.startswith("."):
                        continue
                    if not os.path.splitext(file)[1]:
                        continue

                    fileFullPath = os.path.abspath(os.path.join(resourcesFolder, root, file))
                    handleFile(folderConfigData, fileFullPath, processingFolder, targetsFolder, curFolderParams, filesHashesDict, filesNamesDict)

                if recursive == False:
                    break

    # Process files
    if "files" in configDict:
        files = configDict["files"]
        for fileConfigData in files:
            if "file" not in fileConfigData:
                continue

            filePath = fileConfigData["file"]

            # Path
            fileFolder = os.path.split(filePath)[0]
            fileFullPath = os.path.abspath(os.path.join(resourcesFolder, filePath))

            # Custom params
            if "customParams" in fileConfigData:
                customParams = fileConfigData["customParams"]
            else:
                customParams = {}

            handleFile(fileConfigData, fileFullPath, fileFolder, targetsFolder, customParams, filesHashesDict, filesNamesDict)

    # Process hashes + handle removed files
    processHashesFile(filesHashesDict, hashesFilePath, resourcesFolder, targetsFolder)

    # Save names + handle removed files
    processNewNamesFiles(filesNamesDict, targetsFolder, resourcesFolder)


if __name__ == '__main__':
    # Script dir
    GV.SCRIPT_ROOT_DIR = os.path.dirname(os.path.abspath(__file__))

    # Params
    exampleString = "main.py -c <jsonConfigFile> -h <jsonHashesFile> -i <resourcesFolder> -o <targetsResFolder>"
    try:
        opts, args = getopt.getopt(sys.argv[1:], "c:h:i:o:", ["config=", "hashes=", "ifolder=", "ofolder="])
    except getopt.GetoptError:
        print(exampleString)
        sys.exit(2)

    # Check parameters length
    if len(opts) < 2:
        print(exampleString)
        sys.exit(2)

    # Parse parameters
    configFilePath = ''
    hashesFilePath = ''
    inputFolder = ''
    outputFolder = ''
    for opt, arg in opts:
        if opt in ("-c", "--config"):
            configFilePath = arg
        elif opt in ("-h", "--hashes"):
            hashesFilePath = arg
        elif opt in ("-i", "--ifolder"):
            inputFolder = arg
        elif opt in ("-o", "--ofolder"):
            outputFolder = arg

    # Update tools pathes
    pvrPath = os.getenv("PVR_TOOL_PATH", "")
    if len(pvrPath) > 0:
        GV.PVR_TOOL_PATH = pvrPath
    ffmpegPath = os.getenv("FFMPEG_TOOL_PATH", "")
    if len(ffmpegPath) > 0:
        GV.FFMPEG_TOOL_PATH = ffmpegPath
    webpPath = os.getenv("WEBP_TOOL_PATH", "")
    if len(webpPath) > 0:
        GV.WEBP_TOOL_PATH = webpPath

    # Process config
    if configFilePath and hashesFilePath and inputFolder and outputFolder:
        # Resources processing
        processResources(configFilePath, hashesFilePath, inputFolder, outputFolder)
        # Wait complete all convert
        waitConvertComplete()

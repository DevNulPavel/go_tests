#! /usr/bin/env python3
# -*- coding: utf-8 -*-

import socket
import struct
import hashlib
import os
import GlobalVariables as GV

def updateServerMeta():
    # Resolve DNS Address
    GV.SERVER_IP_ADDRESS = socket.gethostbyname(GV.SERVER_ADDRESS)
    if GV.SERVER_IP_ADDRESS:
        # Check server accessible
        GV.REMOTE_SERVER_ACCESSIBLE = checkServerAvailable(GV.SERVER_IP_ADDRESS, GV.SERVER_PORT)
        if GV.REMOTE_SERVER_ACCESSIBLE:
            # NUM proc count
            GV.MAX_PROCESSES_COUNT = int(getNumServerProcCount(GV.SERVER_IP_ADDRESS, GV.SERVER_PORT))
    else:
        GV.REMOTE_SERVER_ACCESSIBLE = False


def checkServerAvailable(address, port):
    # Create a TCP socket
    s = socket.socket()
    s.settimeout(3)
    print("Attempting to connect to %s:%s" % (address, port))
    try:
        s.connect((address, port))
        s.close()
        print("Connected to %s:%s" % (address, port))
        return True
    except:
        print("Connection to %s:%s failed" % (address, port))
        return False


def getNumServerProcCount(address, port):
    # Create a TCP socket
    s = socket.socket()
    try:
        s.connect((address, port))
        s.send(struct.pack(">B", GV.REQUEST_TYPE_PROC_COUNT))
        buffer = s.recv(1)
        s.close()
        if len(buffer) == 1:
            print("Process count on %s:%s = %d" % (address, port, buffer[0]))
            return buffer[0]
        return True
    except:
        print("Process count check failed on %s:%s" % (address, port))
        return 4


def serverConvertFunctionMultiproc(srcFilePath, srcFileExt, resultFilePath, resultFileExt, convertType, convertParams):
    if len(srcFileExt) == 0 or len(resultFileExt) == 0:
        exit(1)
        return

    s = socket.socket()
    try:
        # File size info
        statinfo = os.stat(srcFilePath)
        sizeInBytes = statinfo.st_size

        # Connect
        s.connect((GV.SERVER_IP_ADDRESS, GV.SERVER_PORT))

        # Send convert type
        s.send(struct.pack(">B", GV.REQUEST_TYPE_CONVERT))

        # Big endian
        typeBytes = struct.pack(">B", convertType)  # 1 byte
        srcFileExtBytes = struct.pack(">B", len(srcFileExt))  # 1 byte
        resultFileExtBytes = struct.pack(">B", len(resultFileExt))  # 1 byte
        paramsStrSize = struct.pack(">B", len(convertParams))  # 1 byte
        sizeBytes = struct.pack(">I", sizeInBytes)  # 4 bytes

        # Send meta data (8 bytes)
        metaDataBuffer = bytearray()
        metaDataBuffer += typeBytes
        metaDataBuffer += srcFileExtBytes
        metaDataBuffer += resultFileExtBytes
        metaDataBuffer += paramsStrSize
        metaDataBuffer += sizeBytes
        s.send(metaDataBuffer)

        # Hash for params
        hashCalc = hashlib.md5()
        hashCalc.update(srcFileExt)
        hashCalc.update(resultFileExt)
        hashCalc.update(convertParams)
        hashCalc.update(GV.CONFIG_DATA_SALT)
        hashForConfig = hashCalc.digest()

        # Send info
        s.send(srcFileExt)
        s.send(resultFileExt)
        s.send(convertParams)
        s.send(hashForConfig)

        # Send file
        sendOffset = 0
        with open(srcFilePath, "rb") as file:
            file.seek(sendOffset, 0)
            fileBuffer = file.read(1024)
            while fileBuffer:
                sendCount = s.send(fileBuffer)
                if sendCount == 0:
                    s.close()
                    exit(1)
                    return
                sendOffset += sendCount
                file.seek(sendOffset, 0)
                fileBuffer = file.read(1024)

        # Receive file
        inputFileSizeBytes = s.recv(4)
        if inputFileSizeBytes == None:
            print("Connection lost")
            s.close()
            exit(1)
            return
        if len(inputFileSizeBytes) == 0:
            print("Connection lost")
            s.close()
            exit(1)
            return
        if len(inputFileSizeBytes) < 4:
            s.close()
            exit(1)
            return

        receiveBytesSize = struct.unpack(">I", inputFileSizeBytes)[0]
        dataBuffer = bytearray()
        while True:
            receivedData = s.recv(1024)
            if not receivedData:
                break
            dataBuffer += receivedData

        if len(dataBuffer) < receiveBytesSize:
            s.close()
            exit(1)
            return

        # Save file
        with open(resultFilePath, "wb") as file:
            file.write(dataBuffer)

        s.close()
        exit(0)  # Ok
        return

    except Exception as e:
        print(e)
        s.close()
        exit(1)
        return
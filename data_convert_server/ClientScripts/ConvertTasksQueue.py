#! /usr/bin/env python3
# -*- coding: utf-8 -*-

import sys
import time
import GlobalVariables as GV
from ConvertTask import *


def waitAnyProcessComplete(targetCount):
    # TODO: Active waiting :-(
    while len(GV.CONVERT_PROCESSES_LIST) > targetCount:
        time.sleep(0.05)
        processesCopy = GV.CONVERT_PROCESSES_LIST.copy()
        for processInfo in processesCopy:
            # Check status
            complete = processInfo[0].isComplete()
            if complete:
                # Remove process from list
                GV.CONVERT_PROCESSES_LIST.remove(processInfo)
                # Try again execute if error status
                isSuccess = processInfo[0].isSuccess()
                if isSuccess == False:
                    # Resen num proc count
                    GV.MAX_PROCESSES_COUNT = 4

                    convertTask = processInfo[1]
                    # Раз удаленная конвертация не удалась, пробуем локальную конвертацию
                    if convertTask.canBeRestarted():
                        processStatus = convertTask.restart()
                        if processStatus:
                            GV.CONVERT_PROCESSES_LIST.append((processStatus, convertTask))
                    else:
                        print("Error convert file: %s", convertTask.fileFullPath)
                        sys.exit(2)


def checkConvertQueue():
    while len(GV.CONVERT_QUEUE) > 0:
        # Pop task from queue and execute
        convertTask = GV.CONVERT_QUEUE.pop(0)
        if isinstance(convertTask, ConvertTask):
            processStatus = convertTask.perform()
            if processStatus:
                GV.CONVERT_PROCESSES_LIST.append((processStatus, convertTask))

        # Limit converting processes size
        waitAnyProcessComplete(GV.MAX_PROCESSES_COUNT - 1)


def waitConvertComplete():
    # Limit converting processes size
    waitAnyProcessComplete(0)
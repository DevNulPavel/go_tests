#! /usr/bin/env python3
# -*- coding: utf-8 -*-


####################################################################################################################################

class ProcessStatus:
    def isComplete(self):
        return True

    def isSuccess(self):
        return True

####################################################################################################################################

class SubprocessStatus:
    def __init__(self, process):
        self.process = process

    def isComplete(self):
        processStatus = self.process.poll()
        if processStatus != None:
            return True
        return False

    def isSuccess(self):
        processStatus = self.process.poll()
        if (processStatus != None) and (processStatus == 0):
            return True
        return True

####################################################################################################################################

class MultiprocessStatus:
    def __init__(self, process):
        self.process = process

    def isComplete(self):
        processStatus = self.process.exitcode
        if processStatus != None:
            return True
        return False

    def isSuccess(self):
        processStatus = self.process.exitcode
        if (processStatus != None) and (processStatus == 0):
            return True
        return False
#! /usr/bin/env bash

PLIST_NAME=com.gameinsight.DoxygenServer.plist
PLIST_FOLDER=/Library/LaunchDaemons
PLIST_PATH=$PLIST_FOLDER/$PLIST_NAME

launchctl unload $PLIST_PATH
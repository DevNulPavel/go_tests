#! /usr/bin/env bash

PLIST_NAME=com.gameinsight.FilesConvert.plist
PLIST_FOLDER=/Library/LaunchDaemons
PLIST_PATH=$PLIST_FOLDER/$PLIST_NAME

sudo launchctl unload $PLIST_PATH
sudo rm -rf $PLIST_PATH
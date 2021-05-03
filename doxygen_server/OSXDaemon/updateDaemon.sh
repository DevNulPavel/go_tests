#! /usr/bin/env bash

PLIST_NAME=com.gameinsight.DoxygenServer.plist
PLIST_FOLDER=/Library/LaunchDaemons
PLIST_PATH=$PLIST_FOLDER/$PLIST_NAME

launchctl unload $PLIST_PATH
cp $PLIST_NAME $PLIST_PATH
chmod 644 $PLIST_PATH
chown root:wheel $PLIST_PATH
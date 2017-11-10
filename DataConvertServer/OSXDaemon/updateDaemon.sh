#! /usr/bin/env bash

PLIST_FOLDER=/Library/LaunchDaemons

sudo launchctl unload $PLIST_FOLDER/com.gameinsight.FilesConvert.plist
sudo cp com.gameinsight.FilesConvert.plist $PLIST_FOLDER/com.gameinsight.FilesConvert.plist
sudo chmod 755 $PLIST_FOLDER/com.gameinsight.FilesConvert.plist
sudo chown root:wheel $PLIST_FOLDER/com.gameinsight.FilesConvert.plist
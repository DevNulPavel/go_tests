#! /usr/bin/env bash

sudo launchctl unload /Library/LaunchDaemons/com.gameinsight.FilesConvert.plist
sudo cp com.gameinsight.FilesConvert.plist /Library/LaunchDaemons/com.gameinsight.FilesConvert.plist
sudo chmod 755 /Library/LaunchDaemons/com.gameinsight.FilesConvert.plist
sudo chown root:wheel /Library/LaunchDaemons/com.gameinsight.FilesConvert.plist
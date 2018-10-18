#! /usr/bin/env bash

PLIST_FOLDER=/Library/LaunchDaemons

sudo launchctl unload $PLIST_FOLDER/com.gameinsight.JenkinsRedirectServer.plist
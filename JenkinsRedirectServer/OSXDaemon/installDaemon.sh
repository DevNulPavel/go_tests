#! /usr/bin/env bash

PLIST_NAME=com.gameinsight.JenkinsRedirectServer.plist
PLIST_FOLDER=/Library/LaunchDaemons
PLIST_PATH=$PLIST_FOLDER/$PLIST_NAME

cp $PLIST_NAME $PLIST_PATH
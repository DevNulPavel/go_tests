#! /usr/bin/env bash

ZIP_FILE_NAME=bundle.zip

go build

zip -u $ZIP_FILE_NAME JenkinsRedirectServer
zip -ur $ZIP_FILE_NAME templates/
zip -ur $ZIP_FILE_NAME static/
zip -ur $ZIP_FILE_NAME OSXDaemon/
# zip -urj $ZIP_FILE_NAME OSXDaemon/filesConvertDaemon.sh

rm JenkinsRedirectServer
#! /usr/bin/env bash

ZIP_FILE_NAME=bundle.zip

go build

zip -u $ZIP_FILE_NAME DataConvertServer
zip -ur $ZIP_FILE_NAME templates/
zip -ur $ZIP_FILE_NAME static/
zip -urj $ZIP_FILE_NAME OSXDaemon/filesConvertDaemon.sh

rm DataConvertServer
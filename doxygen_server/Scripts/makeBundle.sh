#! /usr/bin/env bash

ZIP_FILE_NAME=bundle.zip

go build

zip -u $ZIP_FILE_NAME DoxygenServer
zip -ur $ZIP_FILE_NAME templates/
zip -ur $ZIP_FILE_NAME doxygen_html/
zip -ur $ZIP_FILE_NAME OSXDaemon/
# zip -urj $ZIP_FILE_NAME OSXDaemon/filesConvertDaemon.sh

rm DoxygenServer
#! /usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

HOST="pi@devnulpavel.ddns.net"
TARGET_DIR="/home/pi/Projects/GoTests/TCP"
SRC_DIR="$SCRIPT_DIR/.."

ssh pi@devnulpavel.ddns.net mkdir -p $TARGET_DIR
rsync -h -v -r --delete -e ssh $SRC_DIR $HOST:$TARGET_DIR
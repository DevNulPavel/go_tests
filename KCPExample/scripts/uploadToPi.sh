#! /usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

HOST="pi@devnulpavel.ddns.net"
TARGET_DIR="/home/pi/Projects/GoTests/KCPExample"
SRC_DIR="$SCRIPT_DIR/.."

ssh $HOST mkdir -p $TARGET_DIR
rsync -h -v -r --delete -e ssh $SRC_DIR $HOST:$TARGET_DIR
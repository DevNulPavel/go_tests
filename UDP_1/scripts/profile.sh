#! /usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

go build -o $SCRIPT_DIR/../profileServer $SCRIPT_DIR/../udpServerPing.go

$SCRIPT_DIR/../profileServer

#go tool pprof -web $SCRIPT_DIR/../profileServer $SCRIPT_DIR/../cpu.pprof
go tool pprof -web $SCRIPT_DIR/../profileServer $SCRIPT_DIR/../mem.pprof

rm $SCRIPT_DIR/../profileServer
# rm $SCRIPT_DIR/../cpu.pprof
rm $SCRIPT_DIR/../mem.pprof

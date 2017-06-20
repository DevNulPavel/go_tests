#! /usr/bin/env bash

protoc --go_out=. *.proto
protoc --cpp_out=. *.proto
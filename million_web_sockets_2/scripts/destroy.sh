#!/bin/bash

docker rm -vf $(docker ps -q --filter label=million_web_sockets_2)

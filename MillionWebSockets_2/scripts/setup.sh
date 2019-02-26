#!/bin/bash

## This script helps regenerating multiple client instances in different network namespaces using Docker
## This helps to overcome the ephemeral source port limitation
## Usage: ./connect <connections> <number of clients> <server ip>
## Server IP is usually the Docker gateway IP address, which is 172.17.0.1 by default
## Number of clients helps to speed up connections establishment at large scale, in order to make the demo faster

# Example: ./scripts/setup.sh 1000 10 10.51.48.106


CONNECTIONS=$1
REPLICAS=$2
IP=$3

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Просто собрка
# go build -o client $DIR/../client_src/client.go

# Создание образа
# --rm
# docker run -t -i --name ClientTest -v $DIR/../client_src:/go/src/client_src golang bash -c "cd /go/src/client_src; go get; go build;" #  cd /go; bash
# docker commit ClientTest devnul/ClientTest
# docker push ClientTest

# Запуск вручную
# docker run --rm -t -i -l million_web_sockets_2 devnul/client_test /go/bin/client_src -conn=${CONNECTIONS} -ip=${IP}
# docker run --rm -t -i -l million_web_sockets_2 devnul/client_test /go/bin/client_src -conn=1000 -ip=10.51.48.106

# Запуск инстансов образа
for (( c=0; c<${REPLICAS}; c++ ))
do
    #docker run -l million_web_sockets_2 -v $DIR/../client:/client -d alpine /client -conn=${CONNECTIONS} -ip=${IP}
    docker run --rm -d -l million_web_sockets_2 devnul/client_test /go/bin/client_src -conn=${CONNECTIONS} -ip=${IP}
done

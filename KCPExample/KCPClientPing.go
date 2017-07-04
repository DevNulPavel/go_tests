package main

import (
    "github.com/xtaci/kcp-go"
    "log"
    "time"
    "encoding/binary"
)

func main() {
    kcpconn, err := kcp.DialWithOptions("devnulpavel.ddns.net:9999", nil, 10, 3) // devnulpavel.ddns.net
    if err != nil{
        log.Printf("Dial error %s\n", err)
        return
    }

    //kcpconn.SetACKNoDelay(true)
    kcpconn.SetNoDelay(1, 10, 2, 1)
    kcpconn.SetStreamMode(true)

    const dataSize = 8
    data := make([]byte, dataSize)

    const requestsCount = 100
    startTime := time.Now()
    for i := 0; i < requestsCount; i++ {
        // Write to server
        nowTimeInt64 := uint64(time.Now().UnixNano())
        binary.BigEndian.PutUint64(data, nowTimeInt64)

        writtenCount, err := kcpconn.Write(data)
        if err != nil {
            log.Printf("Write error %s\n", err)
            return
        } else if writtenCount < dataSize {
            log.Printf("Writen less bytes: %d from %d\n", writtenCount, dataSize)
            return
        }

        // Read from server
        readCount, err := kcpconn.Read(data)
        if err != nil {
            log.Printf("Read error %s\n", err)
            return
        } else if readCount == 0 {
            log.Printf("Server disconnected\n")
            return
        } else if readCount < dataSize {
            log.Printf("Read less bytes: %d from %d\n", readCount, dataSize)
            return
        }

        sendTimeInt64 := int64(binary.BigEndian.Uint64(data))
        sendTime := time.Unix(0, sendTimeInt64)
        pingValueMS := time.Now().Sub(sendTime).Nanoseconds() / 1000 / 1000

        log.Printf("Ping value: %dms", pingValueMS)
    }

    endTime := time.Now()
    duration := endTime.Sub(startTime).Seconds()

    requestsPerSec := requestsCount/duration
    log.Printf("Requests per sec value: %f", requestsPerSec)
}

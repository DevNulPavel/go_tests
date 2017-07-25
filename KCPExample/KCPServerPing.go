package main

import (
    "github.com/xtaci/kcp-go"
    "log"
)

func sessionFunction(session *kcp.UDPSession) {
    defer session.Close()

    //session.SetStreamMode(true)
    //session.SetNoDelay(1, 10, 2, 1)

    const dataSize = 8
    data := make([]byte, dataSize)

    for {
        // Reading
        readCount, err := session.Read(data)
        if err != nil {
            log.Printf("Read error: %s\n", err)
            return
        }else if readCount == 0 {
            log.Printf("Client disconnect\n", readCount, dataSize)
            return
        }else if readCount < dataSize {
            log.Printf("Read less bytes than needed: %d from %d\n", readCount, dataSize)
            return
        }

        // Writing
        writeCount, err := session.Write(data)
        if err != nil {
            log.Printf("Write error: %s\n", err)
            return
        }else if writeCount < dataSize {
            log.Printf("Write less bytes than needed: %d from %d\n", readCount, dataSize)
            return
        }
    }
}

func main() {
    listener, err := kcp.ListenWithOptions(":9999", nil, 10, 3)
    if err != nil {
        log.Printf("Listener create error: %s\n", err)
        return
    }

    for {
        session, err := listener.AcceptKCP()
        if err != nil {
            log.Printf("Session accept error: %s\n", err)
            return
        }

        go sessionFunction(session)
    }
}

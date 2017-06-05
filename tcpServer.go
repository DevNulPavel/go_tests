package main

import "fmt"
import "net"
import "encoding/gob"


func handleServerConnection(c net.Conn) {
    defer c.Close()

    decoder := gob.NewDecoder(c) 
    encoder := gob.NewEncoder(c)

    for {
        // Получаем сообщение
        var msg string
        err := decoder.Decode(&msg)
        if err != nil {
            fmt.Println(err)
            fmt.Println("Client out (error)")
            return
        } else {
            fmt.Println("Received:", msg)
        }

        // Отправляем ответ
        err = encoder.Encode("ok")
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func server() {
    // Прослушивание сервера
    ln, err := net.Listen("tcp", ":9999")
    if err != nil {
        fmt.Println(err)
        return
    }
    for {
        // Принятие соединения
        c, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        // Запуск горутины
        go handleServerConnection(c)
    }
}

func main() {
    go server()

    var input string
    fmt.Scanln(&input)
}
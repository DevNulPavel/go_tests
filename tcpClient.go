package main

import "fmt"
import "net"
import "time"
import "encoding/gob"

func client() {
    // Подключение к серверу
    c, err := net.Dial("tcp", "127.0.0.1:9999")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer c.Close()

    encoder := gob.NewEncoder(c)
    decoder := gob.NewDecoder(c) 

    for {
        // Отправка сообщения серверу
        msg := "Hello World"
        fmt.Println("Sending:", msg)

        err = encoder.Encode(msg)
        if err != nil {
            fmt.Println(err)
            return
        }

        time.Sleep(time.Millisecond * 20)

        // Получаем сообщение
        err := decoder.Decode(&msg)
        if err != nil {
            fmt.Println(err)
            return
        } else if msg == "ok" {
            fmt.Println("Sent")
        } else{
            fmt.Println("Sending error")
            return
        }
    }
}

func main() {
    client()

    var input string
    fmt.Scanln(&input)
}
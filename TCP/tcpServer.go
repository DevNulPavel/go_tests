package main

import "fmt"
import "net"
import "encoding/gob"
import "encoding/json"
import "./request"


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

func handleServerConnectionJson(c net.Conn) {
    defer c.Close()

    decoder := json.NewDecoder(c)
    encoder := json.NewEncoder(c)

    for {
        // Получаем сообщение
        message := request.Request{}

        err := decoder.Decode(&message)
        if err != nil {
            fmt.Println(err)
            fmt.Println("Client out (error)")
            return
        } else {
            //fmt.Println("Received:", message)

            // Working
            // 1
            param1, ok1 := message.Params["param1"].(string)
            if ok1 {
                param1 = "Text handled for " + param1
                message.Params["param1"] = param1
            }
            // 2
            param2, ok2 := message.Params["param2"].(float64)
            if ok2 {
                param2 *= 10
                message.Params["param2"] = int(param2)
            }
            // 3
            param3, ok3 := message.Params["param3"].(float64)
            if ok3 {
                param3 *= float64(10.0)
                message.Params["param3"] = float64(param3)
            }
            // 4
            param4, ok4 := message.Params["param4"].([]interface{})
            if ok4 {
                for i, value := range param4 {
                    param4[i] = value.(float64) * 10
                }
                message.Params["param4"] = param4
            }
        }

        // Отправляем ответ
        err = encoder.Encode(message)
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func HandleServerConnectionRaw(c net.Conn) {
    defer c.Close()

    for {
        bytes := make([]byte, 0)

        dataReceived := false
        for {
            readCount, err := c.Read(bytes)
            if err != nil {
                break
            }
            fmt.Println("Read count:", readCount)
            bytesLen := len(bytes)
            if (bytesLen > 0) && (bytes[bytesLen-1] == byte(0)) {
                dataReceived = true
                break
            }
        }

        if dataReceived {
            fmt.Println("Received:", bytes)
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
        go handleServerConnectionJson(c)
    }
}

func main() {
    go server()

    var input string
    fmt.Scanln(&input)
}
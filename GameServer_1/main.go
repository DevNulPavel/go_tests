package main

import (
    "log"
    "net/http"
    "net/url"
    "html/template"
    "github.com/gorilla/websocket"
    "./game"
)

const (
    ADDRESS string = ":8080"
)

func homeHandler(c http.ResponseWriter, r *http.Request) {
    // Создание темплейта
    var homeTempl = template.Must(template.ParseFiles("templates/home.html"))
    data := struct {
        Host       string
        RoomsCount int
    }{r.Host, game.GetRoomsCount()}

    // Обрабатываем шаблон и отдаем его в соединение
    homeTempl.Execute(c, data)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    // ????
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    _, ok := err.(websocket.HandshakeError)
    if ok {
        http.Error(w, "Not a websocket handshake", 400)
        return
    } else if err != nil {
        return
    }

    // ????
    playerName := "Player"
    params, _ := url.ParseQuery(r.URL.RawQuery)
    if len(params["name"]) > 0 {
        playerName = params["name"][0]
    }

    // Получаем свободную комнату
    var room *game.Room = game.GetFreeOrNewRoom()

    // Создание игрока и соединения
    player := game.NewPlayer(playerName)
    pConn := game.NewPlayerConn(ws, player, room)
    pConn.RunAsyncReceiver()

    // Цепляем игрока к комнате
    room.QueueJoinToRoom(pConn)

    log.Printf("Player: %s has joined to Room: %s", pConn.Player.Name, *(room.GetName()))
}

func main() {
    // Утановка обработчиков
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/ws", wsHandler)

    // Обработка статики
    http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, r.URL.Path[1:])
    })

    // запуск сервера
    if err := http.ListenAndServe(ADDRESS, nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
package main

import (
    "log"
    "./game"
    "./utils"
)

var allRooms = make(map[string]*Room)
var freeRooms = make(map[string]*Room)
var roomsCount int = 0

type Room struct {
    name string
    playerConns map[*PlayerConnection]bool      // Зарегистрированные соединения
    updateAll chan bool                         // Канал обновления состояния соединений
    join chan *PlayerConnection                 // Канал для регистрации соединения в комнате
    leave chan *PlayerConnection                // Канал для выхода из комнаты
}

// Запуск главного цикла опроса комнаты
func (r *Room) runRoomMainLoop() {
    exitFromLoop := false
    for exitFromLoop == false {
        select {
            // Присоединение игроков
            case c := <-r.join:
                r.playerConns[c] = true
                // Отправка обновления состояния игрокам
                r.sendUpdateAllPlayers()

                // если комната зополнилась - удаляем ее из свободных
                if len(r.playerConns) == 2 {
                    // Удаляем из свободных комнат
                    delete(freeRooms, r.name)
                    // Связываем игроков
                    var p []*game.Player
                    for k, _ := range r.playerConns {
                        p = append(p, k.player)
                    }
                    game.PairPlayers(p[0], p[1])
                }

            // Выход игроков
            case c := <-r.leave:
                // Выход
                c.player.Leave()
                // Отправка обновления состояния игрокам
                r.sendUpdateAllPlayers()
                // Удаляем соединение
                delete(r.playerConns, c)
                if len(r.playerConns) == 0 {
                    exitFromLoop = true
                    break;
                }

            // Просто рассылка обновления всем юзерам
            case <-r.updateAll:
                r.sendUpdateAllPlayers()
        }
    }

    // Очистка комнат
    delete(allRooms, r.name)
    delete(freeRooms, r.name)
    roomsCount -= 1
    log.Print("Room closed:", r.name)
}

// Рассылка обновления всем юзерам
func (r *Room) sendUpdateAllPlayers() {
    for c := range r.playerConns {
        c.SendStateAsync() // Асинхронный вызов
    }
}

func NewRoom(name string) *Room {
    if name == "" {
        name = utils.RandString(16)
    }

    room := &Room{
        name:        name,
        playerConns: make(map[*PlayerConnection]bool),
        updateAll:   make(chan bool),
        join:        make(chan *PlayerConnection),
        leave:       make(chan *PlayerConnection),
    }

    allRooms[name] = room
    freeRooms[name] = room

    // run Room
    go room.run()

    roomsCount += 1

    return room
}

package game

import (
    "log"
    "../utils"
)

var allRooms = make(map[string]*Room)
var freeRooms = make(map[string]*Room)
var roomsCount int = 0

type Room struct {
    name             string
    playerConns      map[*PlayerConnection]bool // Зарегистрированные соединения
    updateAllChannel chan bool                  // Канал обновления состояния соединений
    joinChannel      chan *PlayerConnection     // Канал для регистрации соединения в комнате
    leaveChannel     chan *PlayerConnection     // Канал для выхода из комнаты
}

func (r *Room) QueueJoinToRoom(c *PlayerConnection) {
    r.joinChannel <- c
}

func (r *Room) GetName() *string {
    return &r.name
}

// Запуск главного цикла опроса комнаты
func (r *Room) runRoomMainLoop() {
    exitFromLoop := false
    for exitFromLoop == false {
        select {
            // Присоединение игроков
            case c := <-r.joinChannel:
                r.playerConns[c] = true
                // Отправка обновления состояния игрокам
                r.sendUpdateAllPlayers()

                // если комната зополнилась - удаляем ее из свободных
                if len(r.playerConns) == 2 {
                    // Удаляем из свободных комнат
                    delete(freeRooms, r.name)
                    // Связываем игроков
                    var p []*Player
                    for k, _ := range r.playerConns {
                        p = append(p, k.Player)
                    }
                    PairPlayers(p[0], p[1])
                }

            // Выход игроков
            case c := <-r.leaveChannel:
                // Выход
                c.Player.Leave()
                // Отправка обновления состояния игрокам
                r.sendUpdateAllPlayers()
                // Удаляем соединение
                delete(r.playerConns, c)
                if len(r.playerConns) == 0 {
                    exitFromLoop = true
                    break;
                }

            // Просто рассылка обновления всем юзерам
            case <-r.updateAllChannel:
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
        name:             name,
        playerConns:      make(map[*PlayerConnection]bool),
        updateAllChannel: make(chan bool),
        joinChannel:      make(chan *PlayerConnection),
        leaveChannel:     make(chan *PlayerConnection),
    }

    allRooms[name] = room
    freeRooms[name] = room

    roomsCount += 1

    return room
}

func GetRoomsCount() int  {
    return roomsCount
}

func GetFreeOrNewRoom() *Room  {
    var room *Room
    if len(freeRooms) > 0 {
        for _, r := range freeRooms {
            room = r
            break
        }
    } else {
        room = NewRoom("")
    }
    return room
}
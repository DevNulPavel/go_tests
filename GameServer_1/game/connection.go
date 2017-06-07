package game

import (
    "github.com/gorilla/websocket"
)

type PlayerConnection struct {
    ws     *websocket.Conn
    Player *Player
    room   *Room
}

func NewPlayerConn(ws *websocket.Conn, player *Player, room *Room) *PlayerConnection {
    pc := &PlayerConnection{ws, player, room}
    return pc
}

// Запуск асинхронного полечения данных
func (pc *PlayerConnection) RunAsyncReceiver()  {
    go pc.receiver()
}

// Получение данных из вебсокета
func (pc *PlayerConnection) receiver() {
    for {
        _, command, err := pc.ws.ReadMessage()
        if err != nil {
            break
        }
        // Выполняем комманду у игрока
        pc.Player.Command(string(command))
        // Заказываем полное обновление у игроков
        pc.room.updateAllChannel <- true
    }
    // Заказываем выход из комнаты
    pc.room.leaveChannel <- pc
    // Закрытие вебсокета
    pc.ws.Close()
}

// Метод отправки текущего состояния на сокет
func (pc *PlayerConnection) SendStateAsync() {
    go func() {
        msg := pc.Player.GetState()
        err := pc.ws.WriteMessage(websocket.TextMessage, []byte(msg))
        if err != nil {
            // Была ошибка, поэтому запрашиваем выход из комнаты
            pc.room.leaveChannel <- pc
            // Закрытие сокета
            pc.ws.Close()
        }
    }()
}

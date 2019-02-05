package gameserver

import (
	"golang.org/x/net/websocket"
)

type WebSocket struct {
	connection   *websocket.Conn
	closeChannel chan bool
}

func (socket *WebSocket) Close() {
	socket.connection.Close()
	socket.closeChannel <- true
}

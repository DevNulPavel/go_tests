package gameserver

import (
	"golang.org/x/net/websocket"
)

type WebSocket struct {
	connection   *websocket.Conn
	closeChannel chan bool
}

func MakeWebSocket(ws *websocket.Conn) *WebSocket {
	connection := WebSocket{ws, make(chan bool, 1)}
	return &connection
}

func (socket *WebSocket) Close() {
	socket.connection.Close()
	socket.closeChannel <- true
}

func (socket *WebSocket) WaitClose() {
	<-socket.closeChannel
}

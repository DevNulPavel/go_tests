package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"

	"WebIM/models"
)

// Отправка события через вебсокет пользователям
func broadcastWebSocket(event models.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		beego.Error("Fail to marshal event:", err)
		return
	}

	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
		// Получаем вебсокет
		ws := sub.Value.(Subscriber).Conn
		if ws != nil {
			// Пишем данные в сокет
			if ws.WriteMessage(websocket.TextMessage, data) != nil {
				// Если не можем писать в сокет, то отрубаем пользователя
				unsubscribe <- sub.Value.(Subscriber).Name
			}
		}
	}
}

// Контроллер веб-сокета
type WebSocketController struct {
	baseController
}

// Обработчик метода GET
func (this *WebSocketController) Get() {
	// Проверяем на всякий случай, что установлено имя пользователя
	uname := this.GetString("uname")
	if len(uname) == 0 {
		this.Redirect("/", 302)
		return
	}

	// Устанавливаем шаблон и данные для шаблона
	this.TplName = "websocket.html"
	this.Data["IsWebSocket"] = true
	this.Data["UserName"] = uname
}

// Обработка запросов вебсокета
func (this *WebSocketController) Join() {
	// Получаем имя и проверяем
	uname := this.GetString("uname")
	if len(uname) == 0 {
		this.Redirect("/", 302)
		return
	}

	// Обновляем данные вебсокета из HTTP запроса, размер буффера записи и чтения - 1024
	ws, err := websocket.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(this.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}

	// Подключаем к чату пользователя с вебсокетом
	Join(uname, ws)
	defer Leave(uname)

	// Запускаем цикл получения данных из сокета
	for {
		// Если у нас происходит ошибка чтения данных из вебсокета, то выходим из данного цикла и метода
		_, p, err := ws.ReadMessage()
		if err != nil {
			return
		}
		// Отправляем сообщение в чат от пользователя
		publish <- newEvent(models.EVENT_MESSAGE, uname, string(p))
	}
}

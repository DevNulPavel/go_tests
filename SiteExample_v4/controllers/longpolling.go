package controllers

import (
	"WebIM/models"
)

// LongPollingController контроллер
type LongPollingController struct {
	baseController
}

// Обрабатывает GET соединения для контроллера
func (this *LongPollingController) Join() {
	// Проверка имени
	uname := this.GetString("uname")
	if len(uname) == 0 {
		this.Redirect("/", 302)
		return
	}

	// Подключение к чат-комнате
	Join(uname, nil)

	// Назначаем шаблон пулинга
	this.TplName = "longpolling.html"

	// Данные для шаблона?
	this.Data["IsLongPolling"] = true
	this.Data["UserName"] = uname
}

// Обработка POST запроса
func (this *LongPollingController) Post() {
	// Назначаем шаблон пулинга
	this.TplName = "longpolling.html"

	// Получаем данные шаблона?
	uname := this.GetString("uname")
	content := this.GetString("content")

	if len(uname) == 0 || len(content) == 0 {
		return
	}

	// Отправляем новый ивент для всех
	publish <- newEvent(models.EVENT_MESSAGE, uname, content)
}

// Метод обрабатываем получени архивов контроллера
func (this *LongPollingController) Fetch() {
	lastReceived, err := this.GetInt("lastReceived")
	if err != nil {
		return
	}

	events := models.GetEvents(int(lastReceived))
	if len(events) > 0 {
		this.Data["json"] = events
		this.ServeJSON()
		return
	}

	// Ожидаем новые сообщения
	ch := make(chan bool)
	waitingList.PushBack(ch)
	<-ch

	// Отправляем Json??
	this.Data["json"] = models.GetEvents(int(lastReceived))
	this.ServeJSON()
}

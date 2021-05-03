package controllers

import (
	"container/list"
	"time"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"

	"WebIM/models"
)

var (
	// Канал для добавления новых пользователей
	subscribe = make(chan Subscriber, 10)
	// Канал для выхода пользователей
	unsubscribe = make(chan string, 10)
	// Канал для отправки событий
	publish = make(chan models.Event, 10)
	// Списки для Long pool
	waitingList = list.New()
	subscribers = list.New()
)

type Subscription struct {
	Archive []models.Event      // Все события из архива
	New     <-chan models.Event // Канал новых событий
}

type Subscriber struct {
	Name string
	Conn *websocket.Conn // Only for WebSocket users; otherwise nil.
}

// Функция вызывается при инициализации данного пакета
func init() {
	go chatroom()
}

// Создаем новое событие
func newEvent(ep models.EventType, user, msg string) models.Event {
	return models.Event{
		Type:      ep,
		User:      user,
		Timestamp: int(time.Now().Unix()),
		Content:   msg,
	}
}

// Подключаемся c помощью WebSocket
func Join(user string, ws *websocket.Conn) {
	subscribe <- Subscriber{Name: user, Conn: ws}
}

// Отключаем пользователя
func Leave(user string) {
	unsubscribe <- user
}

////////////////////////////////////////////////////////////////////////////////

func isUserExist(subscribers *list.List, user string) bool {
	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
		if sub.Value.(Subscriber).Name == user {
			return true
		}
	}
	return false
}

// Функция-горутина, которая обрабатывает сообщения чата
func chatroom() {
	for {
		select {
		// Новый пользователь
		case sub := <-subscribe:
			// Проверяем, есть ли уже пользователь
			if !isUserExist(subscribers, sub.Name) {
				subscribers.PushBack(sub) // Добавляем пользователя в конец

				// Создаем новое событие для отправки
				publish <- newEvent(models.EVENT_JOIN, sub.Name, "")
				beego.Info("New user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			} else {
				beego.Info("Old user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			}
		// Обработка новых событий
		case event := <-publish:
			// Оповещаем список ожидающий
			for ch := waitingList.Back(); ch != nil; ch = ch.Prev() {
				ch.Value.(chan bool) <- true
				waitingList.Remove(ch)
			}

			// Отправляем сообщение по WebSocket
			broadcastWebSocket(event)
			models.AddEventToArchive(event)

			if event.Type == models.EVENT_MESSAGE {
				beego.Info("Message from", event.User, "; Content:", event.Content)
			}
		// Отключение пользователей
		case unsub := <-unsubscribe:
			for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
				if sub.Value.(Subscriber).Name == unsub {
					subscribers.Remove(sub)
					// Закрываем WebSocket соединение
					ws := sub.Value.(Subscriber).Conn
					if ws != nil {
						ws.Close()
						beego.Error("WebSocket closed:", unsub)
					}
					publish <- newEvent(models.EVENT_LEAVE, unsub, "") // Отправляем всем сообщение об отключении
					break
				}
			}
		}
	}
}

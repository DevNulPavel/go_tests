package routers

import (
	"github.com/astaxie/beego"

	"WebIM/controllers"
)

func init() {
	// Регистрируем корневой контроллер
	beego.Router("/", &controllers.AppController{})
	// Регистрируем метод AppController.Join для обработки POST запросов (c помощью рефлексии)
	beego.Router("/join", &controllers.AppController{}, "post:Join")

	// Метод пулинга данных с сервера
	beego.Router("/lp", &controllers.LongPollingController{}, "get:Join")
	beego.Router("/lp/post", &controllers.LongPollingController{})
	beego.Router("/lp/fetch", &controllers.LongPollingController{}, "get:Fetch")

	// Работа с вебсокетом
	beego.Router("/ws", &controllers.WebSocketController{})
	beego.Router("/ws/join", &controllers.WebSocketController{}, "get:Join")

}

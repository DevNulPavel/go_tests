package routers

import (
	"github.com/astaxie/beego"
	"quickstart/controllers"
)

func init() {
	// Регистрируем контроллер для корневого пути
	beego.Router("/", &controllers.MainController{})
}

package controllers

import (
	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

// Будет вызываться на метод Get
func (c *MainController) Get() {
	c.Data["TestLink"] = "beego.me"
	c.Data["TestEmail"] = "astaxie@gmail.com"
	c.TplName = "index.tpl" // Назначаем шаблон, который будет обрабатываться и туда подставятся значения
}

// Будет вызываться на метод Post
func (c *MainController) Post() {
}

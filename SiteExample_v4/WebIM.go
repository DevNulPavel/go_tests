package main

import (
	"github.com/astaxie/beego"
	"github.com/beego/i18n"

	_ "WebIM/routers" // Нужно для инициализации пакета перед вызовом функции main
)

const (
	APP_VER = "0.1.1.0227"
)

func main() {
	beego.Info(beego.BConfig.AppName, APP_VER)

	// Register template functions.
	beego.AddFuncMap("i18n", i18n.Tr)

	beego.Run()
}

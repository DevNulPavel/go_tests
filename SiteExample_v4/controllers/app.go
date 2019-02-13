package controllers

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
)

var langTypes []string // Список поддерживаемых языков

// Функция, вызываемая при инициализации пакета
func init() {
	// Инициализируем список языков из конфига
	langTypes = strings.Split(beego.AppConfig.String("lang_types"), "|")

	// Загружаем файлы локали в соответствии с конфигом
	for _, lang := range langTypes {
		beego.Trace("Loading language: " + lang)
		if err := i18n.SetMessage(lang, "conf/"+"locale_"+lang+".ini"); err != nil {
			beego.Error("Fail to set message file:", err)
			return
		}
	}
}

// Данный контроллер является базовым роутером для других роутеров, содержит базовые структуры внутри и базовые методы
type baseController struct {
	beego.Controller // Интерфейс контроллера
	i18n.Locale      // Реализация работы с локалью
}

// Реализация метода подготовки данных базовым контроллером
// Используется для подготовки языка и настроек
func (this *baseController) Prepare() {
	// Сбрасываем язык
	this.Lang = "" // Поле из i18n.Locale.

	// Получаем информацию о языке из хедера запроса ('Accept-Language' поле)
	al := this.Ctx.Request.Header.Get("Accept-Language")
	if len(al) > 4 {
		al = al[:5] // Сравниваем первые 5 символов
		if i18n.IsExist(al) {
			this.Lang = al
		}
	}

	// Основной язык - английский, если не подобрали нужный
	if len(this.Lang) == 0 {
		this.Lang = "en-US"
	}

	// Выставляем для обработки шаблона данные о языке
	this.Data["Lang"] = this.Lang
}

// Контроллер приложения, который отображает экран приглашения с выбором пользователя и выбором метода работы с чатом
type AppController struct {
	baseController // Использует базовый контроллер
}

// Реализация метода GET данным контроллером
func (this *AppController) Get() {
	this.TplName = "welcome.html" // Выставляем шаблон данного контроллера
}

// Join method handles POST requests for AppController.
// Реализация метода Join для POST запросов, отвечает за переход по нужному адресу
func (this *AppController) Join() {
	// Получаем значения из формы
	uname := this.GetString("uname")
	tech := this.GetString("tech")

	// Проверяем валидный или нет
	if len(uname) == 0 {
		this.Redirect("/", 302) // Редиректим на корневую страницу
		return
	}

	// Настраиваем ссылку в зависимости от выбранной технологии
	switch tech {
	case "longpolling":
		this.Redirect("/lp?uname="+uname, 302)
	case "websocket":
		this.Redirect("/ws?uname="+uname, 302)
	default:
		this.Redirect("/", 302)
	}

	// Всегда нужно выполнять return после redirect?
	return
}

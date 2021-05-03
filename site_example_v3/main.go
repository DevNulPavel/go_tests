package main

import (
	"net/http"
	"text/template"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/session"
	_ "quickstart/routers"
)

var globalSessions *session.Manager = nil
var globalCache cache.Cache = nil
var logger *logs.BeeLogger = nil

func initLogs() {
	logger = logs.NewLogger(10000)
	logger.SetLogger("console", "")

	// Настраиваем вывод номера строки
	logger.EnableFuncCallDepth(true)
	// Настройка глубины пропускаемых слоев
	logger.SetLogFuncCallDepth(3)

	logger.Trace("trace %s %s", "param1", "param2")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warning")
	logger.Error("error")
	logger.Critical("critical")
}

func initCache() {
	globalCache, err := cache.NewCache("memory", `{"interval":60}`)
	if err != nil {
		return
	}
	globalCache.Put("astaxie", 1, 10*time.Second)
	globalCache.Get("astaxie")
	globalCache.IsExist("astaxie")
	globalCache.Delete("astaxie")
}

// Инициализация менеджера сессий для передачи кук клиенту
func initSessionManager() {
	//`{"cookieName":, "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "sessionIDHashFunc": "sha1", "sessionIDHashKey": "", "cookieLifeTime": 3600, "providerConfig": ""}`
	config := session.ManagerConfig{
		CookieName:      "gosessionid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		DisableHTTPOnly: false,
		Secure:          false,
		CookieLifeTime:  3600,
		ProviderConfig:  "",
		// Domain                  string `json:"domain"`
		// SessionIDLength         int64  `json:"sessionIDLength"`
		// EnableSidInHTTPHeader   bool   `json:"EnableSidInHTTPHeader"`
		// SessionNameInHTTPHeader string `json:"SessionNameInHTTPHeader"`
		// EnableSidInURLQuery     bool   `json:"EnableSidInURLQuery"`
		// SessionIDPrefix         string `json:"sessionIDPrefix"`
	}
	globalSessions, _ = session.NewManager("memory", &config)
	go globalSessions.GC()
}

func login(w http.ResponseWriter, r *http.Request) {
	// Создаем объект сессии
	sess, _ := globalSessions.SessionStart(w, r)
	// Отложенно освобождаем
	defer sess.SessionRelease(w)

	// Получаем имя пользователя
	//username := sess.Get("username")

	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, nil)
	} else {
		sess.Set("username", r.Form["username"])
	}
}

func main() {
	initLogs()
	initSessionManager()
	initCache()

	// Дополнителньая установка пути для статики
	beego.SetStaticPath("/testStatic", "static")

	// Запуск HTTP сервера BeeGo на нужном порте
	beego.Run()
}

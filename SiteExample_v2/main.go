package main

import "fmt"
import "net/http"
import "html/template"
import "github.com/codegangsta/martini"
import "github.com/gavruk/go-blog-example/models"
import "github.com/martini-contrib/render"
import "github.com/russross/blackfriday"
import "./model"
import "./helpers"

/////////////////////////////////////////////

var loadedTemplates *template.Template
var posts map[string]*model.Post

/////////////////////////////////////////////

func indexHandler(writer http.ResponseWriter, request *http.Request) {
    loadedTemplates.ExecuteTemplate(writer, "index", posts)
}

func writeHandler(writer http.ResponseWriter, request *http.Request) {
    loadedTemplates.ExecuteTemplate(writer, "write", nil)
}

func editHandler(writer http.ResponseWriter, request *http.Request) {
    // получаем id поста и проверяем наличие его на сервере
    postId := request.FormValue("id")
    post, found := posts[postId]
    if !found {
        http.NotFound(writer, request)
        return
    }

    loadedTemplates.ExecuteTemplate(writer, "write", post)
}

func savePostHandler(writer http.ResponseWriter, request *http.Request) {
    id := request.FormValue("id")
    title := request.FormValue("title")
    content := request.FormValue("content")

    // в зависимости от того, есть у нас id или нет - редактируем или создаем новый
    var post *model.Post
    if id != "" {
        post = posts[id]
        post.Title = title
        post.Content = content
    }else {
        newId := helpers.GenerateStringId()
        post = model.NewPost(newId, title, content)
        posts[newId] = post
    }

    http.Redirect(writer, request, "/", 302)
}

func deleteHandler(writer http.ResponseWriter, request *http.Request) {
    id := request.FormValue("id")
    if id == "" {
        http.NotFound(writer, request)
        return
    }

    delete(posts, id)

    http.Redirect(writer, request, "/", 302)
}

func unescape(str string) interface{} {
    return template.HTML(str)
}

func main() {
    // Создаем посты
    posts = make(map[string]*model.Post, 0)

    // Martini
    m := martini.Classic()

    // Система работы с шаблонами
    renderOptions := render.Options{
        Directory: "templates",
        Layout: "layout",
        Extensions: []string{".tmpl", ".html"},
        Funcs: []template.FuncMap{ template.FuncMap{"unescape": unescape} },
        Charset: "UTF-8",
        IndentJSON: true,
    }
    martiniRenderer := render.Render(renderOptions)
    m.Apply(martiniRenderer)

    // Работа со статикой
    staticOptions := martini.StaticOptions{Prefix: "assets"}
    m.Apply(staticOptions)

    // Создаем http сервер
    m.Get("/", indexHandler)
    m.Get("/write", writeHandler)
    m.Get("/edit", editHandler)
    m.Get("/delete", deleteHandler)
    m.Get("/save_post", savePostHandler)

    // Запускаем сервер
}

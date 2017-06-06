package main

import "fmt"
import "net/http"
import "html/template"
import "./model"
import "./helpers"

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

func main() {
    // Загружаем шаблоны
    templates, err := template.ParseFiles(
        "./templates/index.html",
        "./templates/write.html",
        "./templates/header.html",
        "./templates/footer.html")
    if err != nil {
        fmt.Println("Templates loading error")
        return
    }
    loadedTemplates = templates

    // Создаем посты
    posts = make(map[string]*model.Post, 0)

    // Создаем http сервер
    http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/write", writeHandler)
    http.HandleFunc("/edit", editHandler)
    http.HandleFunc("/delete", deleteHandler)
    http.HandleFunc("/save_post", savePostHandler)

    // Запускаем сервер
    http.ListenAndServe(":8000", nil)
}

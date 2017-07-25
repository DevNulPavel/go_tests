package model

type Post struct {
    Id string
    Title string
    Content string
}

func NewPost(id string, title string, content string) *Post  {
    return &Post{id, title, content}
}
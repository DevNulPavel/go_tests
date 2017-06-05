package chat



type Message struct {
    author string
    body string
}

// конвертация в строку
func (this *Message) String() string {
    return this.author + "sad: " + this.body
}
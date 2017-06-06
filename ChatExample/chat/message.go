package chat

type Message struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

// конвертация в строку
func (this *Message) String() string {
	return this.Author + " says " + this.Body
}

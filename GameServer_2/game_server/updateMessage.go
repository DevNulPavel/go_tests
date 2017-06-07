package game_server

type ClienState struct {
	Id int   `json:"id"`
	X  int   `json:"x"`
	Y  int   `json:"y"`
}

// конвертация в строку
func (mes *ClienState) String() string {
	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
}

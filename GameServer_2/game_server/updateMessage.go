package game_server

type ClienState struct {
	Id uint32  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

// конвертация в строку
//func (mes *ClienState) String() string {
//	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
//}

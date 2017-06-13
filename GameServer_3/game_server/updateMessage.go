package gameserver

// ClienState ... Client state structure
type ClienState struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Delta float64 `json:"d"`
	//AX float64 `json:"ax"`
	//AY float64 `json:"ay"`
}

// конвертация в строку
//func (mes *ClienState) String() string {
//	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
//}

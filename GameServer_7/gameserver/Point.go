package gameserver

type Point16 struct {
	X int16
	Y int16
}

func NewPoint16(x, y int16) Point16 {
	return Point16{x, y}
}

func (p *Point16) Div(value int16) Point16 {
	return Point16{p.X / value, p.Y / value}
}

func (p *Point16) Mul(value int16) Point16 {
	return Point16{p.X * value, p.Y * value}
}

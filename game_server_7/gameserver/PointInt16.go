package gameserver

import "math"

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

func (p *Point16) Sub(point Point16) Point16 {
	return Point16{p.X - point.X, p.Y - point.Y}
}

func (p *Point16) Add(point Point16) Point16 {
	return Point16{p.X + point.X, p.Y + point.Y}
}

func (p *Point16) Length() float64 {
	return math.Sqrt(float64(p.X*p.X + p.Y*p.Y))
}

func (p *Point16) Distance(point Point16) float64 {
	subRes := p.Sub(point)
	return subRes.Length()
}

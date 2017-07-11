package gameserver

import "math"

type PointFloat struct {
	X float64
	Y float64
}

func NewPointFloat(x, y float64) PointFloat {
	return PointFloat{x, y}
}

func (p *PointFloat) Div(value float64) PointFloat {
	return PointFloat{p.X / value, p.Y / value}
}

func (p *PointFloat) Mul(value float64) PointFloat {
	return PointFloat{p.X * value, p.Y * value}
}

func (p *PointFloat) Sub(point PointFloat) PointFloat {
	return PointFloat{p.X - point.X, p.Y - point.Y}
}

func (p *PointFloat) Add(point PointFloat) PointFloat {
	return PointFloat{p.X + point.X, p.Y + point.Y}
}

func (p *PointFloat) Length() float64 {
	return math.Sqrt(float64(p.X*p.X + p.Y*p.Y))
}

func (p *PointFloat) Distance(point PointFloat) float64 {
	subRes := p.Sub(point)
	return subRes.Length()
}

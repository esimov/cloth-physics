package main

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Spring struct {
	p1, p2 *Particle
	length float64
	color  color.NRGBA
}

func NewSpring(p1, p2 *Particle, length float64, col color.NRGBA) *Spring {
	return &Spring{
		p1, p2, length, col,
	}
}

func (s *Spring) Update(gtx layout.Context) {
	s.draw(gtx, s.p1, s.p2)
	s.update(s.p1, s.p2)
}

func (s *Spring) draw(gtx layout.Context, p1, p2 *Particle) {
	var path clip.Path

	drawSticks := func(ops *op.Ops) clip.PathSpec {
		path.Begin(gtx.Ops)

		path.MoveTo(f32.Pt(float32(p1.x), float32(p1.y)))
		path.LineTo(f32.Pt(float32(p2.x), float32(p2.y)))
		path.LineTo(f32.Pt(float32(p2.x+1), float32(p2.y)))
		path.LineTo(f32.Pt(float32(p1.x+1), float32(p1.y)))

		path.MoveTo(f32.Pt(float32(p1.x), float32(p1.y)))
		path.LineTo(f32.Pt(float32(p2.x), float32(p2.y)))
		path.LineTo(f32.Pt(float32(p2.x), float32(p2.y+1)))
		path.LineTo(f32.Pt(float32(p1.x), float32(p1.y+1)))
		path.Close()

		return path.End()
	}

	paint.FillShape(gtx.Ops, s.color, clip.Outline{
		Path: drawSticks(gtx.Ops),
	}.Op())
}

func (s *Spring) update(p1, p2 *Particle) {
	dx := p1.x - p2.x
	dy := p1.y - p2.y

	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < s.length {
		return
	}
	diff := (s.length - dist) / dist
	mul := diff * 0.5 * (1 - s.length/dist)

	offsetX, offsetY := dx*mul, dy*mul

	if !p1.pinX {
		p1.x += offsetX
		p1.y += offsetY
	}
	if !p2.pinX {
		p2.x -= offsetX
		p2.y -= offsetY
	}
}

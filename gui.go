package main

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type gui struct {
	ctx layout.Context
}

// drawCircle draws a circle with radius r at the {x, y} position.
func (g *gui) draw(x, y, r float32, col color.NRGBA) {
	var (
		sq   float64
		p1   f32.Point
		p2   f32.Point
		orig = f32.Pt(x-r, y)
	)

	sq = math.Sqrt(float64(r*r) - float64(r*r))
	p1 = f32.Pt(x+float32(sq), y).Sub(orig)
	p2 = f32.Pt(x-float32(sq), y).Sub(orig)

	var path clip.Path
	path.Begin(g.ctx.Ops)
	path.Move(orig)
	path.Arc(p1, p2, 2*math.Pi)
	path.Close()

	defer clip.Outline{Path: path.End()}.Op().Push(g.ctx.Ops).Pop()
	paint.ColorOp{Color: col}.Add(g.ctx.Ops)
	paint.PaintOp{}.Add(g.ctx.Ops)
}

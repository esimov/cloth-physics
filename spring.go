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

type Constraint struct {
	p1, p2   *Particle
	length   float64
	isActive bool
	color    color.NRGBA
}

func NewConstraint(p1, p2 *Particle, length float64, col color.NRGBA) *Constraint {
	return &Constraint{
		p1, p2, length, true, col,
	}
}

func (c *Constraint) Update(gtx layout.Context, cloth *Cloth, mouse *Mouse) {
	c.resolve(c.p1, c.p2, cloth, mouse)
}

func (c *Constraint) Draw(gtx layout.Context) {
	c.draw(gtx, c.p1, c.p2)
}

func (c *Constraint) draw(gtx layout.Context, p1, p2 *Particle) {
	var path clip.Path

	drawLine := func(ops *op.Ops) clip.PathSpec {
		path.Begin(gtx.Ops)

		// We use `clip.Outline` instead of `clip.Stroke` for performance reasons.
		// For this reason we need to draw the full outline of the stroke.
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

	paint.FillShape(gtx.Ops, c.color, clip.Outline{
		Path: drawLine(gtx.Ops),
	}.Op())
}

func (c *Constraint) resolve(p1, p2 *Particle, cloth *Cloth, mouse *Mouse) {
	p1.constraint = c

	dx := p1.x - p2.x
	dy := p1.y - p2.y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < c.length {
		return
	}
	if mouse.getDragging() {
		if dist > 100 {
			c.p1.removeConstraint(cloth)
		}
	}

	diff := (c.length - dist) / dist
	mul := diff * 0.4 * (1 - c.length/dist)

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

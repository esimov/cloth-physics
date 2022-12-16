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

type Particle struct {
	x, y   float64
	px, py float64
	vx, vy float64
	mass   float64
}

type Stick struct {
	p1, p2 *Particle
	length float64
}

type mouse struct {
	x, y      float64
	px, py    float64
	isDown    bool
	threshold float64
}

func NewParticle(x, y, mass float64) *Particle {
	p := &Particle{
		x, y, x, y, 0, 0, mass,
	}
	return p
}

func NewStick(p1, p2 *Particle, length float64) *Stick {
	return &Stick{
		p1, p2, length,
	}
}

func (p *Particle) Update(gtx layout.Context) {
	col := color.NRGBA{R: 0, G: 0, B: 0, A: 0xff}
	p.draw(gtx, float32(p.x), float32(p.y), float32(p.mass), col)
	p.update(gtx)
}

func (p *Particle) draw(gtx layout.Context, x, y, r float32, col color.NRGBA) {
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
	path.Begin(gtx.Ops)
	path.Move(orig)
	path.Arc(p1, p2, 2*math.Pi)
	path.Close()

	defer clip.Outline{Path: path.End()}.Op().Push(gtx.Ops).Pop()
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

// force = acceleration * mass ->
// acceleration = force / mass
// velocity = acceleration * deltaTime
// position = velocity * deltaTime
func (p *Particle) update(gtx layout.Context) {
	dt := 1.5
	force := struct{ x, y float64 }{x: 0.0, y: 0.5}
	ax := force.x / p.mass
	ay := force.y / p.mass

	px, py := p.x, p.y

	// Verlet integration:
	// x(t+Δt)=2x(t)−x(t−Δt)+a(t)Δt2
	p.x = 2*p.x - p.px + ax*(dt*dt)
	p.y = 2*p.y - p.py + ay*(dt*dt)

	p.px = px
	p.py = py

	width, height := gtx.Constraints.Max.X, gtx.Constraints.Max.Y
	if p.x > float64(width)-p.mass {
		p.x = float64(width) - p.mass
	} else if p.x < 0 {
		p.x = p.mass / 2
	}

	if p.y > float64(height)-p.mass {
		p.y = float64(height) - p.mass
	} else if p.y < 0 {
		p.y = p.mass / 2
	}
}

func (s *Stick) Update(gtx layout.Context) {
	col := color.NRGBA{R: 0, G: 0, B: 0, A: 0xff}
	s.draw(gtx, s.p1, s.p2, col)
	s.update(s.p1, s.p2)
}

func (s *Stick) draw(gtx layout.Context, p1, p2 *Particle, col color.NRGBA) {
	var path clip.Path

	drawLine := func(ops *op.Ops) clip.PathSpec {
		path.Begin(gtx.Ops)
		path.MoveTo(f32.Pt(float32(p1.x), float32(p1.y)))
		path.LineTo(f32.Pt(float32(p2.x), float32(p2.y)))
		path.Close()

		return path.End()
	}

	paint.FillShape(gtx.Ops, col, clip.Stroke{
		Path:  drawLine(gtx.Ops),
		Width: gtx.Metric.PxPerDp,
	}.Op())
}

func (s *Stick) update(p1, p2 *Particle) {
	dx := p1.x - p2.x
	dy := p1.y - p2.y

	dist := math.Sqrt(dx*dx + dy*dy)
	df := (s.length - dist) / dist * 0.1
	offsetX, offsetY := dx*df, dy*df

	p1.x += offsetX
	p1.y += offsetY
	p2.x -= offsetX
	p2.y -= offsetY
}

func getDistance(p1, p2 *Particle) float64 {
	dx := p2.x - p1.x
	dy := p2.y - p1.y

	return math.Sqrt(dx*dx + dy*dy)
}

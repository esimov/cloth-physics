package main

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Particle struct {
	x, y     float64
	px, py   float64
	vx, vy   float64
	mass     float64
	friction float64
	pinX     bool
	color    color.NRGBA
}

func NewParticle(x, y, mass float64, col color.NRGBA) *Particle {
	p := &Particle{
		x: x, y: y, px: x, py: y, vx: 0, vy: 0, mass: mass, color: col,
	}
	return p
}

func (p *Particle) Update(gtx layout.Context, delta float64) {
	p.draw(gtx, float32(p.x), float32(p.y), float32(p.mass))
	p.update(gtx, delta)
}

func (p *Particle) draw(gtx layout.Context, x, y, r float32) {
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
	paint.ColorOp{Color: p.color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func (p *Particle) update(gtx layout.Context, dt float64) {
	if p.pinX {
		return
	}
	force := struct{ x, y float64 }{x: 0.0, y: 0.02}

	// Newton's law of motion: force = acceleration * mass
	// acceleration = force / mass
	ax := force.x / p.mass
	ay := force.y / p.mass

	px, py := p.x, p.y
	// velocity = acceleration * deltaTime
	// position = velocity * deltaTime
	posX, posY := ax*(dt*dt), ay*(dt*dt)

	// Verlet integration:
	// x(t+Δt)=2x(t)−x(t−Δt)+a(t)Δt2
	p.x = p.x + (p.x-p.px)*p.friction + posX
	p.y = p.y + (p.y-p.py)*p.friction + posY

	p.px, p.py = px, py

	width, height := gtx.Constraints.Max.X, gtx.Constraints.Max.Y

	if p.x >= float64(width)-p.mass {
		p.x = float64(width) - p.mass
		p.px = p.x
	} else if p.x < 0 {
		p.x = p.mass / 2
		p.px = p.x
	}

	if p.y > float64(height)-p.mass {
		p.y = float64(height) - p.mass
		p.py = p.y
	} else if p.y < 0 {
		p.y = p.mass / 2
		p.py = p.y
	}
}

func (p *Particle) increaseFriction(force float64) {
	p.friction += force
}

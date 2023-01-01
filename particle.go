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
	x, y       float64
	px, py     float64
	size       float64
	friction   float64
	elasticity float64
	dragForce  float64
	pinX       bool
	color      color.NRGBA
	constraint *Constraint
}

func NewParticle(x, y, size float64, col color.NRGBA) *Particle {
	p := &Particle{
		x: x, y: y, px: x, py: y, size: size, color: col,
	}
	p.elasticity = 18.0
	p.dragForce = 1.9

	return p
}

func (p *Particle) Update(gtx layout.Context, mouse *Mouse, delta float64) {
	//p.draw(gtx, float32(p.x), float32(p.y), 2)
	p.update(gtx, mouse, delta)
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

func (p *Particle) update(gtx layout.Context, mouse *Mouse, dt float64) {
	if p.pinX {
		return
	}
	dt *= 1.2

	dx := p.x - mouse.x
	dy := p.y - mouse.y
	dist := math.Sqrt(dx*dx + dy*dy)

	if p.constraint != nil && dist < 60 {
		p.constraint.isActive = true
	}

	if mouse.getDragging() {
		if dist < 80 {
			dx := mouse.x - mouse.px
			dy := mouse.y - mouse.py
			if dx > p.elasticity {
				dx = p.elasticity
			}
			if dy > p.elasticity {
				dy = p.elasticity
			}
			if dx < -p.elasticity {
				dx = -p.elasticity
			}
			if dy < -p.elasticity {
				dy = -p.elasticity
			}
			p.px = p.x - dx*p.dragForce
			p.py = p.y - dy*p.dragForce
		}
	}

	if mouse.getRightMouseButton() {
		if dist < 15 {
			p.constraint.isSelected = false
		}
	}

	force := struct{ x, y float64 }{x: 0.0, y: 0.005}
	// Newton's law of motion: force = acceleration * mass
	// acceleration = force / mass
	ax := force.x / p.size
	ay := force.y / p.size

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

	if p.x >= float64(width)-p.size {
		p.x = float64(width) - p.size
		p.px = p.x
	} else if p.x < 0 {
		p.x = p.size / 2
		p.px = p.x
	}

	if p.y > float64(height)-p.size {
		p.y = float64(height) - p.size
		p.py = p.y
	} else if p.y < 0 {
		p.y = p.size / 2
		p.py = p.y
	}
}

func (p *Particle) removeConstraint(cloth *Cloth) {
	for idx, c := range cloth.constraints {
		if c == p.constraint {
			cloth.constraints = append(cloth.constraints[:idx], cloth.constraints[idx+1:]...)
			break
		}
	}
}

func (p *Particle) increaseFriction(force float64) {
	p.friction += force
}

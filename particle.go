package main

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

const (
	clothPinDist = 4
	defFocusArea = 50
	minFocusArea = 30
	maxFocusArea = 120
)

// Particle holds the basic components of the particle system.
type Particle struct {
	x, y        float64
	px, py      float64
	vx, vy      float64
	friction    float64
	stiffness   float64
	dragForce   float64
	pinX        bool
	isActive    bool
	highlighted bool
	color       color.NRGBA
}

// NewParticle initializes a new Particle.
func NewParticle(x, y float64, hud *Hud, col color.NRGBA) *Particle {
	p := &Particle{
		x: x, y: y, px: x, py: y, color: col,
	}
	hudDragForce := float64(hud.sliders[0].widget.Value)
	hudStiffness := float64(hud.sliders[2].widget.Value)

	p.isActive = true
	p.highlighted = false
	p.dragForce = hudDragForce
	p.stiffness = hudStiffness

	return p
}

// Update updates the particle system using the Verlet integration.
func (p *Particle) Update(gtx layout.Context, mouse *Mouse, hud *Hud, delta float64) {
	p.update(gtx, mouse, hud, delta)
}

// Deprecated: draw draws the particle at the {x, y} position with the radius `r`.
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

// update is an internal method to update the cloth system using Verlet integration.
func (p *Particle) update(gtx layout.Context, mouse *Mouse, hud *Hud, dt float64) {
	p.highlighted = false

	p.friction = float64(hud.sliders[3].widget.Value)
	p.dragForce = float64(hud.sliders[0].widget.Value)
	p.stiffness = float64(hud.sliders[2].widget.Value)

	gravityForce := float64(hud.sliders[1].widget.Value)
	tearDistance := float64(hud.sliders[4].widget.Value)

	if p.pinX {
		// Recalculate the pinned particles position when the window is resized.
		// We need to do this only for the pinned particles, because the rest
		// of the particles will just adjust themselves automatically.
		p.x += hud.winOffsetX
		p.y += hud.winOffsetY
		return
	}

	// Holding the left mouse button will increase the dragging force
	// resulting in a much advanced cloth destruction.
	if mouse.getLeftButton() {
		maxDragForce := float64(hud.sliders[0].max)
		p.increaseForce(mouse, maxDragForce)
	}

	// Window width and height.
	width, height := gtx.Constraints.Max.X, gtx.Constraints.Max.Y

	dx := p.x - mouse.x
	dy := p.y - mouse.y
	dist := math.Sqrt(dx*dx + dy*dy)

	if mouse.getDragging() && dist < float64(tearDistance) {
		dx := mouse.x - mouse.px
		dy := mouse.y - mouse.py
		if dx > p.stiffness {
			dx = p.stiffness
		}
		if dy > p.stiffness {
			dy = p.stiffness
		}
		if dx < -p.stiffness {
			dx = -p.stiffness
		}
		if dy < -p.stiffness {
			dy = -p.stiffness
		}
		p.px = p.x - dx*p.dragForce
		p.py = p.y - dy*p.dragForce
	}

	// Pin up the particle if the mouse is pressed combined with the CTRL key.
	if mouse.getCtrlDown() && dist < clothPinDist {
		p.pinX = true
	}

	// Modify the mouse focus area size on scrolling.
	focusArea := mouse.getScrollY()
	if focusArea > maxFocusArea {
		focusArea = maxFocusArea
	} else if focusArea < minFocusArea {
		focusArea = minFocusArea
	}

	if dist < float64(focusArea) {
		p.highlighted = true
	}

	// With right click we can tear up the cloth at the mouse position.
	if mouse.getRightButton() {
		if dist < float64(focusArea) {
			p.isActive = false
		}
	}

	px, py := p.x, p.y
	p.vy += gravityForce

	// position = velocity * deltaTime
	posX, posY := p.vx*(dt*dt), p.vy*(dt*dt)

	// Verlet integration:
	// x(t+Δt)=2x(t)−x(t−Δt)+a(t)Δt2
	p.x = p.x + (p.x-p.px)*p.friction + posX
	p.y = p.y + (p.y-p.py)*p.friction + posY

	p.px, p.py = px, py

	if p.x >= float64(width) {
		p.x = float64(width)
		p.px = p.x
	} else if p.x < 0 {
		p.x = 0
		p.px = p.x
	}

	if p.y > float64(height) {
		p.y = float64(height)
		p.py = p.y
	} else if p.y < 0 {
		p.y = 0
		p.py = p.y
	}

	p.vx, p.vy = 0.0, 0.0
}

// increaseForce increases the dragging force.
func (p *Particle) increaseForce(m *Mouse, maxDragForce float64) {
	p.dragForce += m.getForce()
	if p.dragForce > maxDragForce {
		p.dragForce = maxDragForce
	}
}

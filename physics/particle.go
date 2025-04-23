package physics

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/esimov/cloth-physics/gui"
	"github.com/esimov/cloth-physics/params"
)

// particle holds the basic components of the particle system.
type particle struct {
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
func NewParticle(x, y float64, hud *gui.Hud, col color.NRGBA) *particle {
	p := &particle{
		x: x, y: y, px: x, py: y, color: col,
	}
	hudDragForce := float64(hud.Sliders[gui.HudSliderDragForce].Widget.Value)
	hudStiffness := float64(hud.Sliders[gui.HudSliderStiffness].Widget.Value)

	p.isActive = true
	p.highlighted = false
	p.dragForce = hudDragForce
	p.stiffness = hudStiffness

	return p
}

// Update updates the particle system using the Verlet integration.
func (p *particle) Update(gtx layout.Context, mouse *Mouse, hud *gui.Hud, delta float64) {
	p.update(gtx, mouse, hud, delta)
}

// Deprecated: draw draws the particle at the {x, y} position with the radius `r`.
func (p *particle) draw(gtx layout.Context, x, y, r float32) {
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
func (p *particle) update(gtx layout.Context, mouse *Mouse, hud *gui.Hud, dt float64) {
	p.highlighted = false

	p.dragForce = float64(hud.Sliders[gui.HudSliderDragForce].Widget.Value)
	gravityForce := float64(hud.Sliders[gui.HudSliderGravityForce].Widget.Value)
	p.stiffness = float64(hud.Sliders[gui.HudSliderStiffness].Widget.Value)

	p.friction = float64(hud.Sliders[gui.HudSliderFriction].Widget.Value)
	tearDistance := float64(hud.Sliders[gui.HudSliderTearDistance].Widget.Value)

	if p.pinX {
		// Recalculate the pinned particles position when the window is resized.
		// We need to do this only for the pinned particles, because the rest
		// of the particles will just adjust themselves automatically.
		p.x += hud.WinOffsetX
		p.y += hud.WinOffsetY
		return
	}

	// Holding the left mouse button will increase the dragging force
	// resulting in a much advanced cloth destruction.
	if mouse.GetLeftButton() {
		maxDragForce := float64(hud.Sliders[gui.HudSliderDragForce].Max)
		p.increaseForce(mouse, maxDragForce)
	}

	// Window width and height.
	width, height := gtx.Constraints.Max.X, gtx.Constraints.Max.Y

	dx := p.x - mouse.x
	dy := p.y - mouse.y
	dist := math.Sqrt(dx*dx + dy*dy)

	if mouse.GetDragging() && dist < float64(tearDistance) {
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
	if mouse.GetCtrlDown() && dist < params.ClothPinDist {
		p.pinX = true
	}

	// Modify the mouse focus area size on scrolling.
	focusArea := mouse.GetScrollY()
	if focusArea > params.MaxFocusArea {
		focusArea = params.MaxFocusArea
	} else if focusArea < params.MinFocusArea {
		focusArea = params.MinFocusArea
	}

	if dist < float64(focusArea) {
		p.highlighted = true
	}

	// With right click we can tear up the cloth at the mouse position.
	if mouse.GetRightButton() {
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
func (p *particle) increaseForce(m *Mouse, maxDragForce float64) {
	p.dragForce += m.GetForce()
	if p.dragForce > maxDragForce {
		p.dragForce = maxDragForce
	}
}

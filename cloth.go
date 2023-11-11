package main

import (
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

const lineWidth = 0.6

type Cloth struct {
	constraints   []*Constraint
	particles     []*Particle
	width         int
	height        int
	spacing       int
	friction      float64
	color         color.NRGBA
	isInitialized bool
}

// NewCloth creates a new cloth which dimension is calculated based on
// the application window width and height and the spacing between the sticks.
func NewCloth(width, height, spacing int, col color.NRGBA) *Cloth {
	return &Cloth{
		width:   width,
		height:  height,
		spacing: spacing,
		color:   col,
	}
}

// Init initializes the cloth where the `posX` and `posY`
// are the {x, y} position of the cloth's the top-left side.
func (c *Cloth) Init(posX, posY int, hud *Hud) {
	clothX := c.width / c.spacing
	clothY := c.height / c.spacing

	// Skip the cloth initialization when the window is resized.
	if c.isInitialized {
		return
	}

	for y := 0; y <= clothY; y++ {
		for x := 0; x <= clothX; x++ {
			px := posX + x*c.spacing
			py := posY + y*c.spacing

			particle := NewParticle(float64(px), float64(py), hud, c.color)
			particle.friction = c.friction

			// Connect the particles with sticks but skip the particles from the first column and row.
			// We connect the particles from the second row and column onward to the particles before.
			if y != 0 {
				top := c.particles[x+(y-1)*(clothX+1)]
				constraint := NewConstraint(top, particle, float64(c.spacing), c.color)
				c.constraints = append(c.constraints, constraint)
			}
			if x != 0 {
				left := c.particles[len(c.particles)-1]
				constraint := NewConstraint(left, particle, float64(c.spacing), c.color)
				c.constraints = append(c.constraints, constraint)
			}

			pinX := x % (clothX / 10)
			if y == 0 && pinX == 0 {
				particle.pinX = true
			}

			c.particles = append(c.particles, particle)
		}
	}
	c.isInitialized = true
}

// Update updates the cloth particles invoked on each frame event of the Gio internal window calls.
// The cloth contraints are solved by using the Verlet integration formulas.
func (cloth *Cloth) Update(gtx layout.Context, mouse *Mouse, hud *Hud, dt float64) {
	dragForce := float32(mouse.getForce() * 0.1)
	clothColor := color.NRGBA{R: 0x55, A: 0xff}
	// Convert the RGB color to HSL based on the applied force over the mouse focus area.
	col := LinearFromSRGB(clothColor).HSLA().Lighten(dragForce).RGBA().SRGB()

	for _, p := range cloth.particles {
		p.Update(gtx, mouse, hud, dt)
	}

	for _, c := range cloth.constraints {
		if c.p1.isActive && c.p2.isActive {
			c.Update(gtx, cloth, mouse)
		}
	}

	var path clip.Path
	path.Begin(gtx.Ops)

	// For performance reasons we draw the sticks as a single clip path instead of multiple clips paths.
	// The performance improvement is considerable compared of drawing each clip path separately.
	for _, c := range cloth.constraints {
		if c.p1.isActive && c.p2.isActive {
			a := f32.Pt(float32(c.p1.x), float32(c.p1.y))
			b := f32.Pt(float32(c.p2.x), float32(c.p2.y))
			addSegment(&path, a, b, lineWidth)
		}
	}
	// We are using `clip.Outline` instead of `clip.Stroke`, because the performance gains
	// are much better, but we need to draw the full outline of the stroke.
	paint.FillShape(gtx.Ops, cloth.color, clip.Outline{
		Path: path.End(),
	}.Op())

	// Here we are drawing the mouse focus area in a separate clip path,
	// because the color used for highlighting the selected area
	// should be different than the cloth's default color.
	for _, c := range cloth.constraints {
		if (c.p1.isActive && c.p1.highlighted) &&
			(c.p2.isActive && c.p2.highlighted) {

			c.color = color.NRGBA{R: col.R, A: col.A}

			path.Begin(gtx.Ops)
			a := f32.Pt(float32(c.p1.x), float32(c.p1.y))
			b := f32.Pt(float32(c.p2.x), float32(c.p2.y))
			addSegment(&path, a, b, lineWidth)

			paint.FillShape(gtx.Ops, c.color, clip.Outline{
				Path: path.End(),
			}.Op())
		}
	}
}

// Reset resets the cloth to the initial state.
func (c *Cloth) Reset(startX, startY int, hud *Hud) {
	c.constraints = nil
	c.particles = nil
	c.isInitialized = false

	c.Init(startX, startY, hud)
}

func addSegment(p *clip.Path, a, b f32.Point, w float32) {
	n := normal(a, b, w)
	p.MoveTo(a.Add(n))
	p.LineTo(b.Add(n))
	p.LineTo(b.Sub(n))
	p.LineTo(a.Sub(n))
	p.Close()
}

// Calculate the scaled normal vector.
func normal(a, b f32.Point, w float32) f32.Point {
	dir := b.Sub(a)
	dir.X, dir.Y = +dir.Y, -dir.X
	d := math.Hypot(float64(dir.X), float64(dir.Y))
	if math.Abs(d) < 1e-5 {
		return f32.Point{}
	}
	return dir.Mul(w / float32(d))
}

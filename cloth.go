package main

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Cloth struct {
	constraints   []*Constraint
	particles     []*Particle
	height        int
	spacing       int
	friction      float64
	width         int
	color         color.NRGBA
	isInitialized bool
}

// NewCloth creates a new cloth which dimension is calculated based on
// the application window width and height and the spacing between the sticks.
func NewCloth(width, height, spacing int, friction float64, col color.NRGBA) *Cloth {
	return &Cloth{
		width:    width,
		height:   height,
		spacing:  spacing,
		friction: friction,
		color:    col,
	}
}

// Init initializes the cloth where the `posX` and `posY`
// are the {x, y} position of the cloth's the top-left side.
func (c *Cloth) Init(posX, posY int, hud *Hud) {
	clothX := c.width / c.spacing
	clothY := c.height / c.spacing

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

			pinX := x % (clothX / 7)
			if y == 0 && pinX == 0 {
				particle.pinX = true
			}

			c.particles = append(c.particles, particle)
		}
	}
	c.isInitialized = true
}

// Update is invoked on each frame event of the Gio internal window calls.
// It updates the cloth particles, which are the basic entities over the
// cloth constraints are applied and solved using Verlet integration.
func (cloth *Cloth) Update(gtx layout.Context, mouse *Mouse, hud *Hud, delta float64) {
	dragForce := float32(mouse.getForce() * 0.75)
	clothColor := color.NRGBA{R: 0x55, A: 0xff}
	// Convert the RGB color to HSL based on the applied force over the mouse focus area.
	col := LinearFromSRGB(clothColor).HSLA().Lighten(dragForce).RGBA().SRGB()

	for _, p := range cloth.particles {
		p.Update(gtx, mouse, hud, delta)
	}

	for _, c := range cloth.constraints {
		if c.p1.isActive {
			c.Update(gtx, cloth, mouse)
		}
	}

	var path clip.Path
	path.Begin(gtx.Ops)

	// For performance reasons we draw the sticks as a single clip path instead of multiple clips paths.
	// The performance improvement is considerable compared of drawing each clip path separately.
	for _, c := range cloth.constraints {
		if c.p1.isActive && c.p2.isActive {
			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)).Add(f32.Point{X: 1.2}))
			path.LineTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)).Add(f32.Point{X: 1.2}))
			path.Close()
		}
	}
	// We are using `clip.Outline` instead of `clip.Stroke`, because the performance gains
	// are much better, but we need to draw the full outline of the stroke.
	paint.FillShape(gtx.Ops, cloth.color, clip.Outline{
		Path: path.End(),
	}.Op())

	path.Begin(gtx.Ops)
	for _, c := range cloth.constraints {
		if c.p1.isActive && c.p2.isActive {

			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)).Add(f32.Point{Y: 1.2}))
			path.LineTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)).Add(f32.Point{Y: 1.2}))
			path.Close()

		}
	}
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
			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)).Add(f32.Point{X: 1}))
			path.LineTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)).Add(f32.Point{X: 1}))
			path.Close()

			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)).Add(f32.Point{Y: 1}))
			path.LineTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)).Add(f32.Point{Y: 1}))
			path.Close()

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

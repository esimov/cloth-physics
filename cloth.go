package main

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Cloth struct {
	width    int
	height   int
	spacing  int
	friction float64
	color    color.NRGBA

	particles   []*Particle
	constraints []*Constraint

	isInitialized bool
}

func NewCloth(width, height, spacing int, friction float64, col color.NRGBA) *Cloth {
	return &Cloth{
		width:    width,
		height:   height,
		spacing:  spacing,
		friction: friction,
		color:    col,
	}
}

func (c *Cloth) Init(startX, startY int) {
	clothX := c.width / c.spacing
	clothY := c.height / c.spacing

	for y := 0; y <= clothY; y++ {
		for x := 0; x <= clothX; x++ {
			px := startX + x*c.spacing
			py := startY + y*c.spacing

			particle := NewParticle(float64(px), float64(py), c.color)
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

func (cloth *Cloth) Update(gtx layout.Context, mouse *Mouse, delta float64) {
	dragForce := float32(mouse.getForce() * 0.75)
	clothColor := color.NRGBA{R: 0x55, A: 0xff}
	col := LinearFromSRGB(clothColor).HSLA().Lighten(dragForce).RGBA().SRGB()

	for _, p := range cloth.particles {
		p.Update(gtx, mouse, delta)
	}

	for _, c := range cloth.constraints {
		if c.isSelected {
			c.Update(gtx, cloth, mouse)
		}
	}

	// For performance reasons we draw the sticks as a single clip path instead of multiple clips paths.
	// The performance improvement is considerable compared to the multiple clip paths rendered separately.
	var path clip.Path
	path.Begin(gtx.Ops)

	for _, c := range cloth.constraints {
		if c.isSelected {
			// We use `clip.Outline` instead of `clip.Stroke` for performance reasons.
			// For this reason we need to draw the full outline of the stroke.
			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x+1), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p1.x+1), float32(c.p1.y)))

			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y+1)))
			path.LineTo(f32.Pt(float32(c.p1.x), float32(c.p1.y+1)))
			path.Close()
		}
	}

	paint.FillShape(gtx.Ops, cloth.color, clip.Outline{
		Path: path.End(),
	}.Op())

	// Here we are drawing the mouse focus area in a separate clip path,
	// because the color used for highlighting the selected area
	// should be different than the cloth default color.
	for _, c := range cloth.constraints {
		if c.isActive && c.isSelected {
			path.Begin(gtx.Ops)

			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x+1), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p1.x+1), float32(c.p1.y)))

			path.MoveTo(f32.Pt(float32(c.p1.x), float32(c.p1.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y)))
			path.LineTo(f32.Pt(float32(c.p2.x), float32(c.p2.y+1)))
			path.LineTo(f32.Pt(float32(c.p1.x), float32(c.p1.y+1)))
			path.Close()

			c.color = color.NRGBA{R: col.R, A: col.A}

			paint.FillShape(gtx.Ops, c.color, clip.Outline{
				Path: path.End(),
			}.Op())

			c.isActive = false
		}
	}
}

func (c *Cloth) Reset(startX, startY int) {
	c.constraints = nil
	c.particles = nil
	c.isInitialized = false

	c.Init(startX, startY)
}

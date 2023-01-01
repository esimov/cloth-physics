package main

import (
	"image/color"

	"gioui.org/layout"
)

type Cloth struct {
	width    int
	height   int
	spacing  int
	friction float64
	mass     float64
	partCol  color.NRGBA
	stickCol color.NRGBA

	particles   []*Particle
	constraints []*Constraint

	isInitialized bool
}

func NewCloth(width, height, spacing int, mass, friction float64, col1, col2 color.NRGBA) *Cloth {
	return &Cloth{
		width:    width,
		height:   height,
		spacing:  spacing,
		friction: friction,
		mass:     mass,
		partCol:  col1,
		stickCol: col2,
	}
}

func (c *Cloth) Init(startX, startY int) {
	clothX := c.width / c.spacing
	clothY := c.height / c.spacing

	for y := 0; y <= clothY; y++ {
		for x := 0; x <= clothX; x++ {
			px := startX + x*c.spacing
			py := startY + y*c.spacing

			particle := NewParticle(float64(px), float64(py), c.mass, c.partCol)
			particle.friction = c.friction

			// Connect the particles with sticks but skip the particles from the first column and row.
			// We connect the particles from the second row and column onward to the particles before.
			if y != 0 {
				top := c.particles[x+(y-1)*(clothX+1)]
				constraint := NewConstraint(top, particle, float64(c.spacing), c.stickCol)
				c.constraints = append(c.constraints, constraint)
			}
			if x != 0 {
				left := c.particles[len(c.particles)-1]
				constraint := NewConstraint(left, particle, float64(c.spacing), c.stickCol)
				c.constraints = append(c.constraints, constraint)
			}

			pin := (x + clothX) % (clothX / 8)
			if y == 0 && (x == 0 || pin == 0) {
				particle.pinX = true
			}

			c.particles = append(c.particles, particle)
		}
	}
	c.isInitialized = true
}

func (cloth *Cloth) Update(gtx layout.Context, mouse *Mouse, delta float64) {
	for _, p := range cloth.particles {
		p.Update(gtx, mouse, delta)
	}

	for _, c := range cloth.constraints {
		if c.isSelected {
			c.Update(gtx, cloth, mouse)
		}
	}

	for _, c := range cloth.constraints {
		if c.isActive {
			c.color = color.NRGBA{R: 0xff, A: 0xcc}
		} else {
			c.color = cloth.stickCol
		}
		if c.isSelected {
			c.Draw(gtx)
		}
	}
}

func (c *Cloth) Reset(startX, startY int) {
	c.constraints = nil
	c.particles = nil
	c.isInitialized = false

	c.Init(startX, startY)
}

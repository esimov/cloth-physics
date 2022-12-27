package main

import (
	"image/color"

	"gioui.org/layout"
)

type Cloth struct {
	width     int
	height    int
	spacing   int
	friction  float64
	mass      float64
	particles []*Particle
	springs   []*Spring
	partCol   color.NRGBA
	springCol color.NRGBA

	isInitialized bool
}

func NewCloth(width, height, spacing int, mass, friction float64, col1, col2 color.NRGBA) *Cloth {
	return &Cloth{
		width:     width,
		height:    height,
		spacing:   spacing,
		friction:  friction,
		mass:      mass,
		partCol:   col1,
		springCol: col2,
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

			// Connect the particles with springs but skip the particles from the first column and row.
			// We connect the particles from the second row and column onward to the particles before.
			if y != 0 {
				top := c.particles[x+(y-1)*(clothX+1)]
				s := NewSpring(top, particle, float64(c.spacing), c.springCol)
				c.springs = append(c.springs, s)
			}
			if x != 0 {
				left := c.particles[len(c.particles)-1]
				s := NewSpring(left, particle, float64(c.spacing), c.springCol)
				c.springs = append(c.springs, s)
			}

			pin := (x + clothX) % (clothX / 6)
			if y == 0 && (x == 0 || pin == 0) {
				particle.pinX = true
			}

			c.particles = append(c.particles, particle)
		}
	}
	c.isInitialized = true
}

func (c *Cloth) Update(gtx layout.Context, delta float64) {
	for _, p := range c.particles {
		p.Update(gtx, delta)
	}

	for _, s := range c.springs {
		s.Update(gtx)
	}
}

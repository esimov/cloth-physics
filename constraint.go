package main

import (
	"image/color"
	"math"

	"gioui.org/layout"
)

type Constraint struct {
	p1, p2     *Particle
	length     float64
	isSelected bool
	isActive   bool
	color      color.NRGBA
}

// NewConstraint creates a new constraint between two points/particles.
// The constraint actually is a stick which connects two points.
func NewConstraint(p1, p2 *Particle, length float64, col color.NRGBA) *Constraint {
	return &Constraint{
		p1, p2, length, true, false, col,
	}
}

// Update updates the stick between two points by resolving the constraints between them.
func (c *Constraint) Update(gtx layout.Context, cloth *Cloth, mouse *Mouse) {
	c.p1.constraint = c

	dx := c.p1.x - c.p2.x
	dy := c.p1.y - c.p2.y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < c.length {
		return
	}
	// Tear up the cloth under the mouse position if the applied force exceeds a certain threshold.
	// The threshold is the distance between the two points.
	if mouse.getDragging() {
		if dist > 150 {
			c.p1.removeConstraint(cloth)
		}
	}

	diff := (c.length - dist) / dist
	mul := diff * 0.4 * (1 - c.length/dist)

	offsetX, offsetY := dx*mul, dy*mul

	if !c.p1.pinX {
		c.p1.x += offsetX
		c.p1.y += offsetY
	}
	if !c.p2.pinX {
		c.p2.x -= offsetX
		c.p2.y -= offsetY
	}
}
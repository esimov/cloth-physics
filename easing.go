package main

import (
	"time"

	"gioui.org/layout"
)

// Easing
type Easing struct {
	initTime time.Time
	duration time.Duration
	delta    time.Duration
}

// Progress calculates the time passed from the first invocation of the time.Now function.
func (e *Easing) Progress() float64 {
	return float64(e.delta) / float64(e.duration)
}

// Update updates the time passed from the initial invocation of the time.Now
// function until the time set as duration is not reached.
func (e *Easing) Update(gtx layout.Context, isActive bool) float64 {
	delta := gtx.Now.Sub(e.initTime)
	e.initTime = gtx.Now

	if isActive {
		if e.delta < e.duration {
			e.delta += delta
			if e.delta > e.duration {
				e.delta = e.duration
			}
		}
	} else {
		if e.delta > 0 {
			e.delta -= delta
			if e.delta < 0 {
				e.delta = 0
			}
		}
	}

	return e.Progress()
}

// InOutBack is the easing function used for the HUD position update.
func (e *Easing) InOutBack(t float64) float64 {
	s := 1.70158
	t *= 2
	if t < 1 {
		s *= 1.525
		return 0.5 * (t * t * ((s+1)*t - s))
	} else {
		t -= 2
		s *= 1.525
		return 0.5 * (t*t*((s+1)*t+s) + 2)
	}
}

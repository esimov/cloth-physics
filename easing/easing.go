package easing

import "gioui.org/layout"

type Easing interface {
	Update(gtx layout.Context, isActive bool) float64
	Animate(formula EasingFormula, t float64) float64
}

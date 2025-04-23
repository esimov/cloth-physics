// Code taken from: github.com/egonelbre/expgio/shadow/f32color

package utils

import "math"

// HSLA is a 32 bit floating point HSL space color, all from 0 to 1.
type HSLA struct{ H, S, L, A float32 }

// HSLA converts color to HSLA.
func (col RGBA) HSLA() HSLA {
	// see https://en.wikipedia.org/wiki/HSL_and_HSV#Hue_and_chroma
	max := max3(col.R, col.G, col.B)
	min := min3(col.R, col.G, col.B)
	c := max - min

	var h float32
	if c == 0 {
		// undefined, we'll default to 0
	} else if max == col.R {
		h = mod6((col.G - col.B) / c)
	} else if max == col.G {
		h = (col.B-col.R)/c + 2
	} else {
		h = (col.R-col.G)/c + 4
	}
	h /= 6.0
	if h < 0 {
		h += 1
	}

	l := (max + min) / 2
	var s float32
	if c != 0 {
		s = c / (1 - abs(2*l-1))
	}

	return HSLA{
		H: h,
		S: s,
		L: l,
		A: col.A,
	}
}

// HSLA converts color to RGBA.
func (col HSLA) RGBA() RGBA {
	if col.S == 0 {
		return RGBA{A: col.A}
	}

	h := mod1(col.H)

	var v2 float32
	if col.L < 0.5 {
		v2 = col.L * (1 + col.S)
	} else {
		v2 = (col.L + col.S) - col.S*col.L
	}

	v1 := 2*col.L - v2

	return RGBA{
		R: hue(v1, v2, h+1.0/3.0),
		G: hue(v1, v2, h),
		B: hue(v1, v2, h-1.0/3.0),
		A: col.A,
	}
}

// IsBright estimates whether the color is considered as bright or
// not based on CIELab perceived lightness.
func (col HSLA) IsBright() bool {
	// See RGBA.IsBright for calculation details.
	const grayPoint = 65.0
	const d = (grayPoint + 16.0) / 116.0
	return col.L > d*d*d
}

// Emphasize darkens light colors and lightens dark colors.
func (col HSLA) Emphasize(ratio float32) HSLA {
	if col.IsBright() {
		return col.Darken(ratio)
	} else {
		return col.Lighten(ratio)
	}
}

// Lighten returns linear color blend with white in HSL colorspace with the specified percentage.
func (col HSLA) Lighten(p float32) HSLA {
	p = clamp1(p)
	col.L = clamp1(col.L + (1-col.L)*p)
	return col
}

// Darken returns linear color blend with black in HSL colorspace with the specified percentage.
func (col HSLA) Darken(p float32) HSLA {
	p = clamp1(p)
	col.L = clamp1(col.L * (1 - p))
	return col
}

func hue(v1, v2, h float32) float32 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	if 6*h < 1 {
		return v1 + (v2-v1)*6*h
	} else if 2*h < 1 {
		return v2
	} else if 3*h < 2 {
		return v1 + (v2-v1)*(2.0/3.0-h)*6
	}

	return v1
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
func mod1(v float32) float32 { return float32(math.Mod(float64(v), 1)) }
func mod6(v float32) float32 { return float32(math.Mod(float64(v), 6)) }

func min3(a, b, c float32) float32 {
	if a < b {
		if a < c {
			return a
		} else {
			return c
		}
	} else {
		if b < c {
			return b
		} else {
			return c
		}
	}
}

func max3(a, b, c float32) float32 {
	if a > b {
		if a > c {
			return a
		} else {
			return c
		}
	} else {
		if b > c {
			return b
		} else {
			return c
		}
	}
}

func clamp1(v float32) float32 {
	if v > 1 {
		return 1
	} else if v < 0 {
		return 0
	}
	return v
}

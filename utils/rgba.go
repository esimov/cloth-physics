// Code taken from: github.com/egonelbre/expgio/shadow/f32color

package utils

import (
	"image/color"
	"math"
)

// RGBA is a 32 bit floating point linear space color.
type RGBA struct {
	R, G, B, A float32
}

// Array returns rgba values in a [4]float32 array.
func (rgba RGBA) Array() [4]float32 {
	return [4]float32{rgba.R, rgba.G, rgba.B, rgba.A}
}

// Float32 returns r, g, b, a values.
func (col RGBA) Float32() (r, g, b, a float32) {
	return col.R, col.G, col.B, col.A
}

// SRGBA converts from linear to sRGB color space.
func (col RGBA) SRGB() color.NRGBA {
	return color.NRGBA{
		R: uint8(linearTosRGB(col.R)*255 + 0.5),
		G: uint8(linearTosRGB(col.G)*255 + 0.5),
		B: uint8(linearTosRGB(col.B)*255 + 0.5),
		A: uint8(col.A*255 + 0.5),
	}
}

// Opaque returns the color without alpha component.
func (col RGBA) Opaque() RGBA {
	col.A = 1.0
	return col
}

// LinearFromSRGB converts from SRGBA to RGBA.
func LinearFromSRGB(col color.NRGBA) RGBA {
	r, g, b, a := col.RGBA()
	return RGBA{
		R: sRGBToLinear(float32(r) / 0xffff),
		G: sRGBToLinear(float32(g) / 0xffff),
		B: sRGBToLinear(float32(b) / 0xffff),
		A: float32(a) / 0xFFFF,
	}
}

// linearTosRGB transforms color value from linear to sRGB.
func linearTosRGB(c float32) float32 {
	// Formula from EXT_sRGB.
	switch {
	case c <= 0:
		return 0
	case 0 < c && c < 0.0031308:
		return 12.92 * c
	case 0.0031308 <= c && c < 1:
		return 1.055*float32(math.Pow(float64(c), 0.41666)) - 0.055
	}

	return 1
}

// sRGBToLinear transforms color value from sRGB to linear.
func sRGBToLinear(c float32) float32 {
	// Formula from EXT_sRGB.
	if c <= 0.04045 {
		return c / 12.92
	} else {
		return float32(math.Pow(float64((c+0.055)/1.055), 2.4))
	}
}

// MulAlpha scales all color components by alpha/255.
func MulAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	// TODO: Optimize. This is pretty slow.
	a := float32(alpha) / 255.
	rgba := LinearFromSRGB(c)
	rgba.A *= a
	rgba.R *= a
	rgba.G *= a
	rgba.B *= a
	return rgba.SRGB()
}

// LightenRGB returns linear color blend with white in RGB colorspace with the specified percentage.
// Returns `(r,g,b) * (1 - p) + (1, 1, 1) * p`.
func (col RGBA) Lighten(p float32) RGBA {
	p = clamp1(p)
	col.R = clamp1(col.R + (1-col.R)*p)
	col.G = clamp1(col.G + (1-col.G)*p)
	col.B = clamp1(col.B + (1-col.B)*p)
	return col
}

// DarkenRGB returns linear color blend with black in RGB colorspace with the specified percentage.
// Returns `(r,g,b) * (1 - p) + (0, 0, 0) * p`.
func (col RGBA) Darken(p float32) RGBA {
	p = clamp1(p)
	col.R = clamp1(col.R * (1 - p))
	col.G = clamp1(col.G * (1 - p))
	col.B = clamp1(col.B * (1 - p))
	return col
}

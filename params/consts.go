package params

import "image/color"

const Version = "v1.0.4"

const (
	HudTimeout = 2.5
	Delta      = 0.022

	WindowSizeX = 1280
	WindowSizeY = 820

	DefaultWindowWidth  = 1024
	DefaultWindowHeigth = 640

	ClothPinDist     = 4
	DefaultFocusArea = 50
	MinFocusArea     = 30
	MaxFocusArea     = 120
)

var (
	WindowWidth  = DefaultWindowWidth
	WindowHeight = DefaultWindowHeigth

	HudDefaultColor = color.NRGBA{R: 0xd9, G: 0x03, B: 0x68, A: 0xff}
)

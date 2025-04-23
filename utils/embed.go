package utils

import (
	_ "embed"
	"log"
	"sync"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
)

//go:embed fonts/AlbertSans-Regular.ttf
var AlbertSansRegular []byte

//go:embed fonts/AlbertSans-Medium.ttf
var AlbertSansMedium []byte

//go:embed fonts/AlbertSans-Light.ttf
var AlbertSansLight []byte

//go:embed fonts/AlbertSans-SemiBold.ttf
var AlbertSansSemiBold []byte

var (
	once       sync.Once
	collection []text.FontFace
)

func Collection() []text.FontFace {
	once.Do(func() {
		register("AlbertSans", font.Font{}, AlbertSansRegular)
		register("AlbertSans", font.Font{Weight: font.Light}, AlbertSansLight)
		register("AlbertSans", font.Font{Weight: font.Medium}, AlbertSansMedium)
		register("AlbertSans", font.Font{Weight: font.SemiBold}, AlbertSansSemiBold)
	})
	return collection
}

func register(typeface string, fnt font.Font, ttf []byte) {
	face, err := opentype.Parse(ttf)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}

	fnt.Typeface = font.Typeface(typeface)
	collection = append(collection, font.FontFace{Font: fnt, Face: face})
}

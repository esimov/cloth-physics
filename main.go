package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

const (
	windowWidth  = 800
	windowHeight = 520
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("Tearable Cloth"),
			app.Size(unit.Dp(windowWidth), unit.Dp(windowHeight)),
		)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	var (
		ops       op.Ops
		deltaTime int64
	)

	partCol := color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xff}
	springCol := color.NRGBA{R: 0x9a, G: 0x9a, B: 0x9a, A: 0xff}

	// p1 := NewParticle(220, 20, 20, partCol)
	// p2 := NewParticle(320, 20, 20, partCol)
	// p3 := NewParticle(220, 120, 20, partCol)
	// p4 := NewParticle(320, 120, 20, partCol)
	// particles := []*Particle{p1, p2, p3, p4}

	// st1 := NewSpring(p1, p2, getDistance(p1, p3), springCol)
	// st2 := NewSpring(p2, p4, getDistance(p1, p2), springCol)
	// st3 := NewSpring(p4, p3, getDistance(p2, p4), springCol)
	// st4 := NewSpring(p3, p1, getDistance(p4, p3), springCol)
	// sticks := []*Spring{st1, st2, st3, st4}

	var width, height int = windowWidth, windowHeight * 0.5

	cloth := NewCloth(width, height, 15, 2, 0.995, partCol, springCol)
	startX := (width/2 - (width/cloth.spacing)/2) / 2

	//renderer->GetWindowWidth() * 0.5f - clothWidth * clothSpacing * 0.5f

	fmt.Println(width)
	fmt.Println(float64(cloth.spacing) / float64(width))
	fmt.Println(width/2 - (width/cloth.spacing)/2)
	fmt.Println("StartX:", startX/2)
	startY := int(float64(height) * 0.2)

	cloth.Init(startX, startY)

	initTime := time.Now()
	mouse := &Mouse{}
	ctrlDown := false

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				pointer.InputOp{
					Tag:   w,
					Types: pointer.Press | pointer.Drag | pointer.Release | pointer.Type(pointer.ButtonPrimary) | pointer.Type(pointer.ButtonSecondary),
				}.Add(gtx.Ops)

				op.InvalidateOp{}.Add(gtx.Ops)
				key.InputOp{Tag: w, Keys: key.NameEscape + "|" + key.NameCtrl + "|" + key.NameAlt}.Add(gtx.Ops)

				for _, ev := range gtx.Queue.Events(w) {
					if e, ok := ev.(key.Event); ok && e.Name == key.NameEscape {
						w.Perform(system.ActionClose)
					}

					switch ev := ev.(type) {
					case pointer.Event:
						switch ev.Type {
						case pointer.Press:
							if ev.Modifiers == key.ModCtrl {
								ctrlDown = true
							}
						case pointer.Release:
							if ev.Modifiers == key.ModCtrl {
								ctrlDown = false
							}
						case pointer.Drag:
							fmt.Println("DRAGGING: ", ctrlDown)
						}
						switch ev.Buttons {
						case pointer.ButtonPrimary:
							mouse.setLeftMousePress()
							fmt.Println("primary")
							pos := mouse.getCurrentPosition(ev)
							fmt.Println(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
							fmt.Println(pos.X, pos.Y)
						case pointer.ButtonSecondary:
							mouse.setRightMousePress()
							fmt.Println("secondary")
						}
					}
				}

				fillBackground(gtx, color.NRGBA{R: 0xf2, G: 0xf2, B: 0xf2, A: 0xff})
				// for _, p := range particles {
				// 	p.draw(gtx, float32(p.x), float32(p.y), float32(p.mass))
				// 	p.update(gtx, deltaTime)
				// }

				// for _, s := range sticks {
				// 	s.Update(gtx)
				// }

				currentTime := time.Now()
				deltaTime = currentTime.Sub(initTime).Milliseconds()

				cloth.Update(gtx, float64(deltaTime)*0.2)
				e.Frame(gtx.Ops)

				initTime = currentTime
			}
		}
	}
}

func fillBackground(gtx layout.Context, col color.NRGBA) {
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

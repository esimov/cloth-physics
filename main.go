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
	windowWidth  = 820
	windowHeight = 540
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
	stickCol := color.NRGBA{R: 0x9a, G: 0x9a, B: 0x9a, A: 0xff}

	var clothW int = windowWidth
	var clothH int = windowHeight * 0.4
	cloth := NewCloth(clothW, clothH, 11, 2, 0.98, partCol, stickCol)

	initTime := time.Now()
	mouse := &Mouse{}
	ctrlDown := false
	isDragging := false

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if !cloth.isInitialized {
					width := gtx.Constraints.Max.X
					height := gtx.Constraints.Max.Y

					startX := width/2 - clothW/2
					startY := int(float64(height) * 0.2)
					cloth.Init(startX, startY)
				}

				pointer.InputOp{
					Tag:   w,
					Types: pointer.Move | pointer.Press | pointer.Drag | pointer.Release | pointer.Type(pointer.ButtonPrimary) | pointer.Type(pointer.ButtonSecondary),
				}.Add(gtx.Ops)

				op.InvalidateOp{}.Add(gtx.Ops)
				key.InputOp{
					Tag:  w,
					Keys: key.NameEscape + "|" + key.NameCtrl + "|" + key.NameAlt + "|" + key.NameSpace,
				}.Add(gtx.Ops)

				for _, ev := range gtx.Queue.Events(w) {
					if e, ok := ev.(key.Event); ok {
						if e.State == key.Press {
							if e.Name == key.NameSpace {
								width := gtx.Constraints.Max.X
								height := gtx.Constraints.Max.Y

								startX := width/2 - clothW/2
								startY := int(float64(height) * 0.2)
								cloth.Reset(startX, startY)
							}
						}
						if e.Name == key.NameEscape {
							w.Perform(system.ActionClose)
						}
					}

					switch ev := ev.(type) {
					case pointer.Event:
						switch ev.Type {
						case pointer.Move:
							pos := mouse.getCurrentPosition(ev)
							mouse.updatePosition(float64(pos.X), float64(pos.Y))
						case pointer.Press:
							fmt.Println("Press")
							if ev.Modifiers == key.ModCtrl {
								ctrlDown = true
							}
						case pointer.Release:
							fmt.Println("Release")
							isDragging = false

							mouse.releaseLeftMouseButton()
							mouse.releaseRightMouseButton()
							mouse.setDragging(isDragging)

							if ev.Modifiers == key.ModCtrl {
								ctrlDown = false
							}
						case pointer.Drag:
							isDragging = true
							fmt.Println("DRAGGING: ", ctrlDown)
						}
						switch ev.Buttons {
						case pointer.ButtonPrimary:
							mouse.setLeftMouseButton()
							pos := mouse.getCurrentPosition(ev)
							mouse.updatePosition(float64(pos.X), float64(pos.Y))
							mouse.setDragging(isDragging)
						case pointer.ButtonSecondary:
							mouse.setRightMouseButton()
							pos := mouse.getCurrentPosition(ev)
							mouse.updatePosition(float64(pos.X), float64(pos.Y))
						}
					}
				}

				fillBackground(gtx, color.NRGBA{R: 0xf2, G: 0xf2, B: 0xf2, A: 0xff})

				currentTime := time.Now()
				deltaTime = currentTime.Sub(initTime).Milliseconds()

				cloth.Update(gtx, mouse, float64(deltaTime)*0.2)
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

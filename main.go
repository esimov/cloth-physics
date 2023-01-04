package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"os"
	"runtime/pprof"
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
	windowWidth  = 840
	windowHeight = 540
)

var (
	cpuprofile string
	f          *os.File
	err        error
)

func main() {
	flag.StringVar(&cpuprofile, "debug-cpuprofile", "", "write CPU profile to this file")
	flag.Parse()

	if cpuprofile != "" {
		f, err = os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
	}

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
		initTime  time.Time
		deltaTime time.Duration
		scrollY   unit.Dp
	)
	if cpuprofile != "" {
		defer pprof.StopCPUProfile()
	}

	col := color.NRGBA{R: 0x9a, G: 0x9a, B: 0x9a, A: 0xff}
	mouse := &Mouse{maxScrollY: unit.Dp(200)}
	isDragging := false

	var clothW int = windowWidth
	var clothH int = windowHeight * 0.4
	cloth := NewCloth(clothW, clothH, 12, 0.99, col)

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				if cpuprofile != "" {
					pprof.StartCPUProfile(f)
				}

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
					Types: pointer.Scroll | pointer.Move | pointer.Press | pointer.Drag | pointer.Release | pointer.Type(pointer.ButtonPrimary) | pointer.Type(pointer.ButtonSecondary),
					ScrollBounds: image.Rectangle{
						Min: image.Point{
							X: 0,
							Y: -30,
						},
						Max: image.Point{
							X: 0,
							Y: 30,
						},
					},
				}.Add(gtx.Ops)

				op.InvalidateOp{}.Add(gtx.Ops)
				key.InputOp{
					Tag:  w,
					Keys: key.NameEscape + "|" + key.NameCtrl + "|" + key.NameAlt + "|" + key.NameSpace,
				}.Add(gtx.Ops)

				if mouse.getLeftButton() {
					deltaTime = time.Now().Sub(initTime)
					mouse.increaseForce(deltaTime.Seconds())
				}

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
						case pointer.Scroll:
							scrollY += unit.Dp(ev.Scroll.Y)
							if scrollY < 0 {
								scrollY = 0
							} else if scrollY > mouse.maxScrollY {
								scrollY = mouse.maxScrollY
							}
							mouse.setScrollY(scrollY)
						case pointer.Move:
							pos := mouse.getCurrentPosition(ev)
							mouse.updatePosition(float64(pos.X), float64(pos.Y))
						case pointer.Press:
							if ev.Modifiers == key.ModCtrl {
								mouse.setCtrlDown(true)
							}
							mouse.setLeftButton()
							initTime = time.Now()
						case pointer.Release:
							isDragging = false

							mouse.resetForce()
							mouse.releaseLeftButton()
							mouse.releaseRightButton()
							mouse.setDragging(isDragging)

							if ev.Modifiers == key.ModCtrl {
								mouse.setCtrlDown(false)
							}
						case pointer.Drag:
							isDragging = true
						}
						switch ev.Buttons {
						case pointer.ButtonPrimary:
							mouse.setLeftButton()
							pos := mouse.getCurrentPosition(ev)
							mouse.updatePosition(float64(pos.X), float64(pos.Y))
							mouse.setDragging(isDragging)
						case pointer.ButtonSecondary:
							mouse.setRightButton()
							pos := mouse.getCurrentPosition(ev)
							mouse.updatePosition(float64(pos.X), float64(pos.Y))
						}
					}
				}

				fillBackground(gtx, color.NRGBA{R: 0xf2, G: 0xf2, B: 0xf2, A: 0xff})

				cloth.Update(gtx, mouse, 0.025)
				e.Frame(gtx.Ops)

				w.Invalidate()
			}
		}
	}
}

func fillBackground(gtx layout.Context, col color.NRGBA) {
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

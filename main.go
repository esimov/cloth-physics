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
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/loov/hrtime"
)

const (
	windowWidth  = 940
	windowHeight = 580
	hudTimeout   = 2.5
)

var (
	profile string
	f       *os.File
	err     error
)

func main() {
	flag.StringVar(&profile, "debug-cpuprofile", "", "write CPU profile to this file")
	flag.Parse()

	if profile != "" {
		f, err = os.Create(profile)
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		w := app.NewWindow(
			app.Title("Gio - 2D Cloth Simulation"),
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
		panelInit time.Time
		scrollY   unit.Dp
	)

	if profile != "" {
		defer pprof.StopCPUProfile()
	}

	defaultColor := color.NRGBA{R: 0x9a, G: 0x9a, B: 0x9a, A: 0xff}

	th := material.NewTheme(gofont.Collection())
	th.TextSize = unit.Sp(12)
	th.Palette.ContrastBg = defaultColor
	th.FingerSize = 15

	mouse := &Mouse{maxScrollY: unit.Dp(200)}
	isDragging := false

	var clothW int = windowWidth * 1.3
	var clothH int = windowHeight * 0.4
	cloth := NewCloth(clothW, clothH, 8, 0.99, defaultColor)
	hud := NewHud()

	var keyTag struct{}

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				start := hrtime.Now()
				gtx := layout.NewContext(&ops, e)
				hud.width = windowWidth
				hud.btnSize = gtx.Dp(35)
				hud.closeBtn = gtx.Dp(25)

				if hud.isActive {
					if !panelInit.IsZero() {
						dt := time.Now().Sub(panelInit).Seconds()
						if dt > hudTimeout {
							hud.isActive = false
						}
					}
				} else {
					panelInit = time.Time{}
				}

				if profile != "" {
					pprof.StartCPUProfile(f)
				}

				if !cloth.isInitialized {
					width := gtx.Constraints.Max.X
					height := gtx.Constraints.Max.Y

					startX := width/2 - clothW/2
					startY := int(float64(height) * 0.2)
					cloth.Init(startX, startY, hud)
				}

				key.InputOp{
					Tag:  &keyTag,
					Keys: key.NameEscape + "|" + key.NameCtrl + "|" + key.NameAlt + "|" + key.NameSpace,
				}.Add(gtx.Ops)

				if mouse.getLeftButton() {
					deltaTime = time.Now().Sub(initTime)
					mouse.increaseForce(deltaTime.Seconds())
				}

				for _, ev := range gtx.Queue.Events(&keyTag) {
					if e, ok := ev.(key.Event); ok {
						if e.State == key.Press {
							if e.Name == key.NameSpace {
								width := gtx.Constraints.Max.X
								height := gtx.Constraints.Max.Y

								startX := width/2 - clothW/2
								startY := int(float64(height) * 0.2)
								cloth.Reset(startX, startY, hud)
							}
						}
						if e.Name == key.NameEscape {
							w.Perform(system.ActionClose)
						}
					}
				}
				fillBackground(gtx, color.NRGBA{R: 0xf2, G: 0xf2, B: 0xf2, A: 0xff})

				layout.Stack{}.Layout(gtx,
					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						// Push a new clip area here so that we can attach a pointer input handler.
						// We listen for these canvas interactions here because we don't want to make
						// this input area the root of the input tree. If it's the root, it will receive
						// copies of all pointer input from its children, and we don't want that.
						defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
						// Add a pointer listener for all of the events that affect the cloth.
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

						// Process pointer events from previous frame.
						for _, ev := range gtx.Queue.Events(w) {
							switch ev := ev.(type) {
							case pointer.Event:
								// We should reset the key focus back to the cloth canvas each time a mouse pointer
								// activity is detected. This is required because if the checkbox or reset button is
								// activated on the slider panel, the focus will be hold on them indefinitely.
								key.FocusOp{Tag: keyTag}.Add(gtx.Ops)
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
									mouse.setCtrlDown(false)
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

						cloth.Update(gtx, mouse, hud, 0.022)
						return layout.Dimensions{}
					}),

					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						if hud.debug.Value {
							layout.Stack{}.Layout(gtx,
								layout.Stacked(func(gtx layout.Context) layout.Dimensions {
									op.Offset(image.Pt(10, 10)).Add(gtx.Ops)
									return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										m := material.Label(th, unit.Sp(15), hrtime.Since(start).String())
										m.Color = defaultColor
										return m.Layout(gtx)
									})
								}),
							)
						}

						if hud.isActive {
							for _, ev := range gtx.Queue.Events(&hud.hudTag) {
								switch ev := ev.(type) {
								case pointer.Event:
									switch ev.Type {
									case pointer.Leave:
										if panelInit.IsZero() {
											panelInit = time.Now()
										}
									case pointer.Move:
										panelInit = time.Time{}
									}
								}
							}

							hud.ShowHideControls(gtx, th, mouse, true)
							hud.DrawCtrlBtn(gtx, th, mouse, true)
						} else {
							hud.DrawCtrlBtn(gtx, th, mouse, false)
							hud.ShowHideControls(gtx, th, mouse, false)
						}

						return layout.Dimensions{}
					}),
				)

				op.InvalidateOp{}.Add(gtx.Ops)
				e.Frame(gtx.Ops)
			}
		}
	}
}

func fillBackground(gtx layout.Context, col color.NRGBA) {
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

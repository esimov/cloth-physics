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
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/esimov/cloth-physics/gui"
	"github.com/esimov/cloth-physics/params"
	"github.com/esimov/cloth-physics/physics"
	"github.com/loov/hrtime"
)

const Version = "v1.0.3"

const (
	hudTimeout = params.HudTimeout
	delta      = params.Delta

	windowSizeX = params.WindowSizeX
	windowSizeY = params.WindowSizeY

	defaultWindowWidth  = params.DefaultWindowWidth
	defaultWindowHeigth = params.DefaultWindowHeigth
)

var (
	windowWidth  = defaultWindowWidth
	windowHeight = defaultWindowHeigth

	// App related variables
	hud    *gui.Hud
	cloth  *physics.Cloth
	mouse  *physics.Mouse
	clothW int
	clothH int

	clothSpacing = 6

	// Gio Ops related variables
	ops          op.Ops
	initTime     time.Time
	deltaTime    time.Duration
	mouseScrollY unit.Dp
	mouseDrag    bool

	// pprof related variables
	profile string
	file    *os.File
	err     error
)

func main() {
	flag.StringVar(&profile, "debug-cpuprofile", "", "write CPU profile to this file")
	flag.Parse()

	if profile != "" {
		file, err = os.Create(profile)
		if err != nil {
			log.Fatal(err)
		}
	}

	hud = gui.NewHud()

	mouse = &physics.Mouse{}
	mouse.SetScrollY(params.DefaultFocusArea)
	mouse.SetMaxScrollY(params.MaxFocusArea)

	go func() {
		w := app.NewWindow(
			app.Title("Gio - 2D Cloth Simulation"),
			app.Size(windowSizeX, windowSizeY),
		)

		// Center the window on the screen.
		w.Perform(system.ActionCenter)

		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

func run(w *app.Window) error {
	var keyTag struct{}

	if profile != "" {
		defer pprof.StopCPUProfile()
	}

	defaultColor := color.NRGBA{R: 0x9a, G: 0x9a, B: 0x9a, A: 0xff}

	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.TextSize = unit.Sp(12)
	th.Palette.ContrastBg = defaultColor
	th.FingerSize = 15

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				start := hrtime.Now()
				gtx := layout.NewContext(&ops, e)
				hud.PanelWidth = gtx.Dp(unit.Dp(windowSizeX))
				hud.BtnSize = gtx.Dp(unit.Dp(40))
				hud.CloseBtn = gtx.Dp(unit.Dp(30))

				if hud.IsActive {
					if !hud.InitPanel.IsZero() {
						dt := time.Since(hud.InitPanel).Seconds()
						if dt > hudTimeout {
							hud.IsActive = false
						}
					}
				} else {
					hud.InitPanel = time.Time{}
				}

				if profile != "" {
					pprof.StartCPUProfile(file)
				}

				// Cloth is not initialized yet.
				if cloth == nil {
					clothW = gtx.Dp(unit.Dp(windowWidth))
					clothH = gtx.Dp(unit.Dp(windowHeight) * 0.33)
					clothSpacing = func() int { // different cloth spacing for hi-res devices.
						if clothW <= windowWidth {
							return clothSpacing
						}
						return 2 * clothSpacing
					}()
					cloth = physics.NewCloth(clothW, clothH, clothSpacing, defaultColor)

					width := gtx.Constraints.Max.X
					height := gtx.Constraints.Max.Y

					startX := int(unit.Dp(width-clothW) / 2)
					startY := int(unit.Dp(height) * 0.2)

					cloth.Init(startX, startY, hud)
				}

				key.InputOp{
					Tag:  &keyTag,
					Keys: key.NameEscape + "|" + key.NameCtrl + "|" + key.NameAlt + "|" + key.NameSpace + "|" + key.NameF1,
				}.Add(gtx.Ops)

				if mouse.GetLeftButton() {
					deltaTime = time.Since(initTime)
					mouse.SetForce(deltaTime.Seconds() * 5)
				}

				for _, ev := range gtx.Queue.Events(&keyTag) {
					if e, ok := ev.(key.Event); ok {
						if e.State == key.Press {
							switch e.Name {
							case key.NameSpace:
								width := gtx.Constraints.Max.X
								height := gtx.Constraints.Max.Y

								startX := (width - clothW) / 2
								startY := int(unit.Dp(height) * 0.2)

								cloth.Width = clothW
								cloth.Height = clothH

								cloth.Reset(startX, startY, hud)
							case key.NameF1:
								hud.ShowHelpPanel = !hud.ShowHelpPanel
								hud.IsActive = false
							}
						}
						if e.Name == key.NameEscape {
							hud.ShowHelpPanel = false
						}
					}
				}

				// Reset the window offsets on resize.
				hud.WinOffsetX = 0
				hud.WinOffsetY = 0

				if defaultWindowWidth != windowWidth {
					hud.WinOffsetX = float64(e.Size.X-windowWidth) * 0.5
				}
				if defaultWindowHeigth != windowHeight {
					hud.WinOffsetY = float64(e.Size.Y-windowHeight) * 0.25
				}

				if e.Size.X != windowWidth || e.Size.Y != windowHeight {
					cloth.Init(windowWidth, windowHeight, hud)

					windowWidth = e.Size.X
					windowHeight = e.Size.Y

					cloth.Width = windowWidth
					cloth.Height = windowHeight

					if e.Size.X < defaultWindowWidth {
						hud.ShowHelpPanel = false
					}
					if e.Size.Y < defaultWindowHeigth {
						hud.ShowHelpPanel = false
					}
				}
				// Fill background
				paint.ColorOp{Color: color.NRGBA{R: 0xf2, G: 0xf2, B: 0xf2, A: 0xff}}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)

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
									mouseScrollY += unit.Dp(ev.Scroll.Y)
									if mouseScrollY < params.MinFocusArea {
										mouseScrollY = params.MinFocusArea
									} else if mouseScrollY > mouse.GetMaxScrollY() {
										mouseScrollY = mouse.GetMaxScrollY()
									}
									mouse.SetScrollY(mouseScrollY)
								case pointer.Move:
									pos := mouse.GetCurrentPosition(ev)
									mouse.UpdatePosition(float64(pos.X), float64(pos.Y))
								case pointer.Press:
									if ev.Modifiers == key.ModCtrl {
										mouse.SetCtrlDown(true)
									}
									mouse.SetLeftButton()
									initTime = time.Now()
									hud.ShowHelpPanel = false
								case pointer.Release:
									mouseDrag = false

									mouse.ResetForce()
									mouse.ReleaseLeftButton()
									mouse.ReleaseRightButton()
									mouse.SetDragging(mouseDrag)
									mouse.SetCtrlDown(false)
								case pointer.Drag:
									mouseDrag = true
								}
								switch ev.Buttons {
								case pointer.ButtonPrimary:
									mouse.SetLeftButton()
									pos := mouse.GetCurrentPosition(ev)
									mouse.UpdatePosition(float64(pos.X), float64(pos.Y))
									mouse.SetDragging(mouseDrag)
								case pointer.ButtonSecondary:
									mouse.SetRightButton()
									pos := mouse.GetCurrentPosition(ev)
									mouse.UpdatePosition(float64(pos.X), float64(pos.Y))
								}
							}
						}
						cloth.Update(gtx, mouse, hud, delta)
						return layout.Dimensions{}
					}),

					layout.Stacked(func(gtx layout.Context) layout.Dimensions {
						if hud.Debug.Value {
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

						if hud.IsActive {
							hud.ShowHelpPanel = false
							for _, ev := range gtx.Queue.Events(&hud.Tag) {
								switch ev := ev.(type) {
								case pointer.Event:
									switch ev.Type {
									case pointer.Leave:
										if hud.InitPanel.IsZero() {
											hud.InitPanel = time.Now()
										}
									case pointer.Move:
										hud.InitPanel = time.Time{}
									}
								}
							}
						}

						hud.DrawCtrlBtn(gtx, th, hud.IsActive)
						hud.ShowControlPanel(gtx, th, hud.IsActive)
						hud.ShowHelpDialog(gtx, th, windowWidth, windowHeight, windowSizeX, hud.ShowHelpPanel)

						return layout.Dimensions{}
					}),
				)

				op.InvalidateOp{}.Add(gtx.Ops)
				e.Frame(gtx.Ops)
			}
		}
	}
}

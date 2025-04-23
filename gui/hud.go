package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/esimov/cloth-physics/easing"
	"github.com/esimov/cloth-physics/params"
)

type (
	D = layout.Dimensions
	C = layout.Context
)

type HudSliderType int

const (
	HudSliderDragForce HudSliderType = iota
	HudSliderGravityForce
	HudSliderStiffness
	HudSliderFriction
	HudSliderTearDistance
)

type Hud struct {
	Tag           struct{}
	Sliders       map[HudSliderType]*slider
	InitPanel     time.Time
	PanelWidth    int
	PanelHeight   int
	WinOffsetX    float64 // stores the X offset on window horizontal resize
	WinOffsetY    float64 // stores the Y offset on window vertical resize
	Debug         widget.Bool
	CloseBtn      int
	BtnSize       int
	IsActive      bool
	ShowHelpPanel bool

	commands  map[int]command
	ctrlPanel easing.Easing
	ctrlBtn   easing.Easing
	reset     widget.Clickable
	list      layout.List
	activator gesture.Click
	closer    gesture.Click
	controls  gesture.Hover
	*helpPanel
}

type slider struct {
	Widget *widget.Float
	Title  string
	Value  float32
	Min    float32
	Max    float32
}

// NewHud creates a new HUD used to interactively change the default settings via sliders and checkboxes.
func NewHud() *Hud {
	hud := &Hud{
		Sliders:  make(map[HudSliderType]*slider),
		commands: make(map[int]command),
		helpPanel: &helpPanel{
			fontType:   "AlbertSans",
			lineHeight: 3,
			h1FontSize: 18,
			h2FontSize: 15,
		},
	}

	sliders := []slider{
		{Title: "Dragging force", Min: 1.1, Value: 2, Max: 15},
		{Title: "Gravity", Min: 100, Value: 250, Max: 500},
		{Title: "Cloth friction", Min: 10, Value: 30, Max: 50},
		{Title: "Cloth stiffness", Min: 0.95, Value: 0.98, Max: 0.99},
		{Title: "Tear distance", Min: 5, Value: 15, Max: 50},
	}

	for idx, slider := range sliders {
		sliderType := HudSliderType(idx)
		hud.addSlider(sliderType, slider)
	}

	commands := []command{
		{"F1": "Toggle the quick help panel"},
		{"Space": "Redraw the cloth"},
		{"Right click": "Tear the cloth at mouse position"},
		{"Click & hold": "Increase cloth tearing force"},
		{"Scroll Up/Down": "Increase/decrease cloth tearing area"},
		{"CTRL+click": "Pin the particle at mouse position"},
	}

	for idx, cmd := range commands {
		hud.commands[idx] = cmd
	}

	slide := &easing.Animation{Duration: 800 * time.Millisecond}
	hover := &easing.Animation{Duration: 600 * time.Millisecond}

	hud.Debug = widget.Bool{}
	hud.Debug.Value = false
	hud.ctrlPanel = slide
	hud.ctrlBtn = hover

	return hud
}

// Add adds a new widget to the list of HUD elements.
func (h *Hud) addSlider(index HudSliderType, s slider) {
	h.list.Axis = layout.Vertical
	s.Widget = &widget.Float{}
	s.Widget.Value = s.Value
	h.Sliders[index] = &s
}

// ShowControlPanel is responsible for showing or hiding the HUD control elements.
func (h *Hud) ShowControlPanel(gtx layout.Context, th *material.Theme, isActive bool) {
	if h.reset.Pressed() {
		for _, s := range h.Sliders {
			s.Widget.Value = s.Value
		}
	}

	progress := h.ctrlPanel.Update(gtx, isActive)
	pos := h.ctrlPanel.Animate(easing.EaseInOutBack, progress) * float64(h.PanelHeight)

	// This offset will apply to the rest of the content laid out in this function.
	defer op.Offset(image.Pt(0, gtx.Constraints.Max.Y+h.CloseBtn-int(pos))).Push(gtx.Ops).Pop()

	{ // Draw HUD main surface area
		var path clip.Path
		path.Begin(gtx.Ops)
		path.MoveTo(f32.Pt(0, 0))
		path.LineTo(f32.Pt(float32(gtx.Constraints.Max.X), 0))
		paint.FillShape(gtx.Ops, color.NRGBA{A: 20}, clip.Stroke{
			Path:  path.End(),
			Width: gtx.Metric.PxPerDp,
		}.Op())

		paint.FillShape(gtx.Ops, color.NRGBA{A: 20}, clip.Rect{
			Max: image.Point{gtx.Constraints.Max.X, gtx.Dp(unit.Dp(1))},
		}.Op())
	}

	// Push this offset, but prepare to pop it after the button is drawn.
	closeOffStack := op.Offset(image.Pt(10, -h.CloseBtn)).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 170},
		clip.Rect{Max: image.Pt(h.CloseBtn, h.CloseBtn)}.Op(),
	)

	{ // Draw close button
		offset := float32(gtx.Dp(unit.Dp(8)))

		var path clip.Path
		path.Begin(gtx.Ops)
		path.MoveTo(f32.Pt(offset, offset))
		path.LineTo(f32.Pt(float32(h.CloseBtn)-offset, float32(h.CloseBtn)-offset))
		path.MoveTo(f32.Pt(float32(h.CloseBtn)-offset, offset))
		path.LineTo(f32.Pt(offset, float32(h.CloseBtn)-offset))

		paint.FillShape(gtx.Ops, color.NRGBA{A: 0xff}, clip.Stroke{
			Path:  path.End(),
			Width: float32(gtx.Dp(unit.Dp(3))),
		}.Op())
	}

	buttonArea := clip.UniformRRect(
		image.Rectangle{Max: image.Pt(h.CloseBtn, h.CloseBtn)}, 0,
	)
	paint.FillShape(gtx.Ops, th.ContrastBg, clip.Stroke{
		Path:  buttonArea.Path(gtx.Ops),
		Width: float32(gtx.Dp(unit.Dp(0.2))),
	}.Op())

	buttonStack := buttonArea.Push(gtx.Ops)
	pointer.CursorPointer.Add(gtx.Ops)
	h.closer.Add(gtx.Ops)
	buttonStack.Pop()

	for _, e := range h.closer.Events(gtx) {
		if e.Type == gesture.ClickType(pointer.Press) {
			h.IsActive = false
			break
		}
	}
	// Pop button-specific offset.
	closeOffStack.Pop()

	r := image.Rectangle{
		Max: image.Point{
			X: gtx.Constraints.Max.X,
			Y: int(pos),
		},
	}

	defer clip.Rect(r).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 170})
	pointer.InputOp{
		Tag:   &h.Tag,
		Types: pointer.Scroll | pointer.Move | pointer.Press | pointer.Drag | pointer.Release | pointer.Leave,
	}.Add(gtx.Ops)
	h.controls.Add(gtx.Ops)

	pointer.CursorPointer.Add(gtx.Ops)

	/* Draw HUD Contents */
	layout.Flex{
		Spacing: layout.SpaceEnd,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.X = h.PanelWidth / 3
			gtx.Constraints.Max.X = gtx.Constraints.Min.X
			layout := layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx C) D {
				return h.list.Layout(gtx, len(h.Sliders),
					func(gtx C, index int) D {
						sliderType := HudSliderType(index)
						if slider, ok := h.Sliders[sliderType]; ok {
							var precisionFmt string
							if slider.Widget.Value > 1 {
								precisionFmt = "%s: %.0f"
							} else {
								precisionFmt = "%s: %.2f"
							}
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(material.Body1(th, fmt.Sprintf(precisionFmt, slider.Title, slider.Widget.Value)).Layout),
								layout.Flexed(1, material.Slider(th, slider.Widget, slider.Min, slider.Max).Layout),
							)
						}
						return D{}
					})
			})
			h.PanelHeight = layout.Size.Y + h.CloseBtn
			return layout
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.UniformInset(unit.Dp(5)).Layout(gtx, material.CheckBox(th, &h.Debug, "Show Frame Rates").Layout)
				}),
				layout.Rigid(func(gtx C) D {
					btnTheme := material.NewTheme()
					btnTheme.Palette.ContrastBg = params.HudDefaultColor
					return layout.UniformInset(unit.Dp(10)).Layout(gtx, material.Button(btnTheme, &h.reset, "Reset").Layout)
				}),
			)
		}),

		layout.Flexed(1, func(gtx C) D {
			w := material.Body1(th, fmt.Sprintf("2D Cloth Simulation %s\nCopyright © 2023, Endre Simo", params.Version))
			w.Alignment = text.End
			w.Color = th.ContrastBg
			w.TextSize = 12
			txtOffs := h.PanelHeight - (3 * h.CloseBtn)

			defer op.Offset(image.Point{Y: txtOffs}).Push(gtx.Ops).Pop()
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return layout.Inset{Bottom: 0, Right: unit.Dp(20)}.Layout(gtx, w.Layout)
				}),
			)
		}),
	)
}

// DrawCtrlBtn draws the button which activates the main HUD area with the sliders.
func (h *Hud) DrawCtrlBtn(gtx layout.Context, th *material.Theme, isActive bool) {
	progress := h.ctrlPanel.Update(gtx, isActive)
	pos := h.ctrlPanel.Animate(easing.EaseInOutBack, progress) * float64(h.PanelHeight)
	offset := gtx.Dp(unit.Dp(60))

	offStack := op.Offset(image.Pt(0, gtx.Constraints.Max.Y-offset+int(pos))).Push(gtx.Ops)
	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx C) D {
				for _, e := range h.activator.Events(gtx) {
					if e.Type == gesture.ClickType(pointer.Press) {
						h.InitPanel = time.Now()
						h.IsActive = true
						break
					}
				}

				progress := h.ctrlBtn.Update(gtx, isActive || h.activator.Hovered())
				btnWidth := h.ctrlBtn.Animate(easing.EaseInOutElastic, progress) * 2.5

				var path clip.Path

				offset := float32(gtx.Dp(unit.Dp(10)))
				btnSize := float32(unit.Dp(h.BtnSize))
				spacing := btnSize / 4
				startX := btnSize/2 - spacing

				// HUD controls button
				for i := float32(0); i < 3; i++ {
					{ // Draw Line
						func(x, y float32) {
							path.Begin(gtx.Ops)
							path.MoveTo(f32.Pt(x, offset))
							path.LineTo(f32.Pt(x, btnSize-offset))
							path.Close()

							paint.FillShape(gtx.Ops, color.NRGBA{A: 0xff}, clip.Stroke{
								Path:  path.End(),
								Width: float32(gtx.Dp(unit.Dp(2.0))),
							}.Op())
						}(startX+(spacing*i), offset)
					}
					{ // Draw Circle
						func(x, y, r float32) {
							orig := f32.Pt(x-r, y)
							sq := math.Sqrt(float64(r*r) - float64(r*r))
							p1 := f32.Pt(x+float32(sq), y).Sub(orig)
							p2 := f32.Pt(x-float32(sq), y).Sub(orig)

							path.Begin(gtx.Ops)
							path.Move(orig)
							path.Arc(p1, p2, 2*math.Pi)
							path.Close()

							defer clip.Outline{Path: path.End()}.Op().Push(gtx.Ops).Pop()
							paint.ColorOp{Color: color.NRGBA{A: 0xff}}.Add(gtx.Ops)
							paint.PaintOp{}.Add(gtx.Ops)
						}(startX+(spacing*i), offset+(spacing*i), float32(gtx.Dp(unit.Dp(4.5))))
					}
				}

				defer clip.Stroke{
					Path: clip.UniformRRect(image.Rectangle{
						Max: image.Pt(h.BtnSize, h.BtnSize),
					}, gtx.Dp(10)).Path(gtx.Ops),
					Width: 2.0 + float32(gtx.Dp(unit.Dp(btnWidth))),
				}.Op().Push(gtx.Ops).Pop()

				pointer.CursorPointer.Add(gtx.Ops)
				h.activator.Add(gtx.Ops)

				paint.ColorOp{Color: params.HudDefaultColor}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)

				return layout.Dimensions{}
			})
		}),
	)
	offStack.Pop()
}

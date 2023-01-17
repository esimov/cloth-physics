package main

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
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const maxIconBorderWidth = unit.Dp(4)

type (
	D = layout.Dimensions
	C = layout.Context
)

type Hud struct {
	width     int
	height    int
	btnSize   int
	closeBtn  int
	isActive  bool
	list      layout.List
	debug     widget.Bool
	activator gesture.Click
	closer    gesture.Click
	controls  gesture.Hover
	reset     widget.Clickable
	sliders   map[int]*slider
	slide     *Easing
	hover     *Easing
}

type slider struct {
	widget *widget.Float
	index  int
	title  string
	min    float32
	value  float32
	max    float32
}

// NewHud creates a new HUD used to interactively change the default settings via sliders and checkboxes.
func NewHud(width, height int) *Hud {
	h := Hud{
		width:   width,
		height:  height,
		sliders: make(map[int]*slider),
	}

	sliders := []slider{
		{title: "Drag force", min: 2, value: 5, max: 25},
		{title: "Gravity force", min: 200, value: 400, max: 1000},
		{title: "Elasticity", min: 10, value: 30, max: 50},
		{title: "Tear distance", min: 10, value: 60, max: 100},
	}

	for idx, s := range sliders {
		h.addSlider(idx, s)
	}

	slide := &Easing{duration: 600 * time.Millisecond}
	hover := &Easing{duration: 300 * time.Millisecond}

	h.debug = widget.Bool{}
	h.debug.Value = false
	h.slide = slide
	h.hover = hover
	h.btnSize = 50
	h.closeBtn = 40

	return &h
}

// Add adds a new widget to the list of HUD elements.
func (h *Hud) addSlider(index int, s slider) {
	h.list.Axis = layout.Vertical
	s.widget = &widget.Float{}
	s.widget.Value = s.value
	h.sliders[index] = &s
}

// ShowHideControls is responsible for showing or hiding the HUD control elements.
// After hovering the mouse over the bottom part of the window a certain amount of time
// it shows the HUD control by invoking an easing function.
func (h *Hud) ShowHideControls(gtx layout.Context, th *material.Theme, m *Mouse, isActive bool) {
	for _, s := range h.sliders {
		if s.widget.Changed() {
			//fmt.Println(s.title, ":", s.widget.Value)
		}
	}

	if h.reset.Pressed() {
		for _, s := range h.sliders {
			s.widget.Value = s.value
		}
	}

	progress := h.slide.Update(gtx, isActive)
	pos := h.slide.InOutBack(progress) * float64(h.height)

	r := image.Rectangle{
		Max: image.Point{
			X: gtx.Constraints.Max.X,
			Y: int(pos),
		},
	}
	op.Offset(image.Pt(0, gtx.Constraints.Max.Y+h.closeBtn-int(pos))).Add(gtx.Ops)

	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			op.Offset(image.Pt(5, -h.closeBtn)).Add(gtx.Ops)
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				clip.Rect{Max: image.Pt(h.closeBtn, h.closeBtn)}.Op(),
			)

			var path clip.Path
			{ // Draw the close button.
				path.Begin(gtx.Ops)
				path.MoveTo(f32.Pt(10, 10))
				path.LineTo(f32.Pt(float32(h.closeBtn)-10, float32(h.closeBtn)-10))
				path.MoveTo(f32.Pt(float32(h.closeBtn)-10, 10))
				path.LineTo(f32.Pt(10, float32(h.closeBtn)-10))
				path.Close()

				paint.FillShape(gtx.Ops, color.NRGBA{A: 0xff}, clip.Stroke{
					Path:  path.End(),
					Width: 5,
				}.Op())
			}

			defer clip.Stroke{
				Path: clip.UniformRRect(
					image.Rectangle{Max: image.Pt(h.closeBtn, h.closeBtn)}, 0,
				).Path(gtx.Ops),
				Width: 0.3,
			}.Op().Push(gtx.Ops).Pop()

			paint.ColorOp{Color: th.ContrastBg}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)

			pointer.CursorPointer.Add(gtx.Ops)
			h.closer.Add(gtx.Ops)

			for _, e := range h.closer.Events(gtx) {
				if e.Type == gesture.ClickType(pointer.Press) {
					h.isActive = false
					break
				}
			}

			return layout.Dimensions{}
		}),
		layout.Expanded(func(gtx C) D {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 127}, clip.Op(clip.Rect(r).Op()))
			stack := clip.Rect(r).Push(gtx.Ops)
			pointer.InputOp{
				Tag:   stack,
				Types: pointer.Scroll | pointer.Move | pointer.Press | pointer.Drag | pointer.Release,
			}.Add(gtx.Ops)

			pointer.CursorPointer.Add(gtx.Ops)
			h.controls.Add(gtx.Ops)

			return layout.Dimensions{Size: r.Max}
		}),
		layout.Stacked(func(gtx C) D {
			border := image.Rectangle{
				Max: image.Point{
					X: gtx.Constraints.Max.X,
					Y: gtx.Dp(unit.Dp(0.5)),
				},
			}
			paint.FillShape(gtx.Ops, color.NRGBA{A: 0x20}, clip.Rect(border).Op())
			return layout.Dimensions{Size: r.Max}
		}),
		layout.Stacked(func(gtx C) D {
			border := image.Rectangle{
				Max: image.Point{
					X: gtx.Constraints.Max.X,
					Y: gtx.Dp(unit.Dp(0.5)),
				},
			}
			paint.FillShape(gtx.Ops, color.NRGBA{A: 0x10}, clip.Rect(border).Op())
			return layout.Dimensions{Size: r.Max}
		}),
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx C) D {
				return h.list.Layout(gtx, len(h.sliders),
					func(gtx C, index int) D {
						if slider, ok := h.sliders[index]; ok {
							gtx.Constraints.Min.X = gtx.Dp(unit.Dp(h.width / 3))
							if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
								gtx.Constraints.Min.X = gtx.Constraints.Max.X
							}
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(material.Body1(th, fmt.Sprintf("%s: %.0f", slider.title, slider.widget.Value)).Layout),
								layout.Flexed(1, material.Slider(th, slider.widget, slider.min, slider.max).Layout),
							)
						}
						return D{}
					})
			})
		}),
	)

	op.Offset(image.Pt(gtx.Dp(unit.Dp(h.width/3)), 0)).Add(gtx.Ops)
	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(material.CheckBox(th, &h.debug, "Show Frame Rates").Layout),
				)
			})
		}),

		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx C) D {
				op.Offset(image.Pt(0, 80)).Add(gtx.Ops)
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(material.Button(th, &h.reset, "Reset").Layout),
				)
			})
		}),
	)
}

func (h *Hud) DrawCtrlBtn(gtx layout.Context, th *material.Theme, m *Mouse, isActive bool) layout.Dimensions {
	op.Offset(image.Pt(0, gtx.Constraints.Max.Y-80)).Add(gtx.Ops)

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx C) D {
				for _, e := range h.activator.Events(gtx) {
					if e.Type == gesture.ClickType(pointer.Press) {
						h.isActive = true
						break
					}
				}

				progress := h.hover.Update(gtx, isActive || h.activator.Hovered())
				width := h.hover.InOutBack(progress) * float64(maxIconBorderWidth)

				var path clip.Path

				spacing := h.btnSize / 4
				startX := h.btnSize/2 - spacing
				for i := 0; i < 3; i++ {
					{ // Draw Line
						func(x, y float32) {
							path.Begin(gtx.Ops)
							path.MoveTo(f32.Pt(float32(startX+(spacing*i)), 10))
							path.LineTo(f32.Pt(float32(startX+(spacing*i)), float32(h.btnSize)-10))
							path.Close()

							paint.FillShape(gtx.Ops, color.NRGBA{A: 0xff}, clip.Stroke{
								Path:  path.End(),
								Width: 3,
							}.Op())
						}(float32(startX+(spacing*i)), 10)
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

							paint.FillShape(gtx.Ops, color.NRGBA{A: 0xff}, clip.Stroke{
								Path:  path.End(),
								Width: 5,
							}.Op())
						}(float32(startX+(spacing*i)), float32(15+spacing*i), 4)
					}
				}

				defer clip.Stroke{
					Path:  clip.UniformRRect(image.Rectangle{Max: image.Pt(h.btnSize, h.btnSize)}, gtx.Dp(10)).Path(gtx.Ops),
					Width: 0.5 + float32(width),
				}.Op().Push(gtx.Ops).Pop()

				pointer.CursorPointer.Add(gtx.Ops)
				h.activator.Add(gtx.Ops)

				paint.ColorOp{Color: color.NRGBA{R: 0xd9, G: 0x03, B: 0x68, A: 0xff}}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)

				return layout.Dimensions{}
			})
		}),
	)
}

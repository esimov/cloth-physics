package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type (
	D = layout.Dimensions
	C = layout.Context
)

type Hud struct {
	width   int
	height  int
	list    layout.List
	debug   widget.Bool
	reset   widget.Clickable
	sliders map[int]*slider
	easing  *Easing
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
		{title: "Drag force", min: 10, value: 20, max: 40},
		{title: "Gravity force", min: 400, value: 600, max: 1000},
		{title: "Elasticity", min: 10, value: 30, max: 50},
		{title: "Tear distance", min: 10, value: 60, max: 100},
	}

	for idx, s := range sliders {
		h.addSlider(idx, s)
	}

	easing := &Easing{duration: 600 * time.Millisecond}
	h.debug = widget.Bool{}
	h.debug.Value = false
	h.easing = easing

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
		fmt.Println("pressed")
		for _, s := range h.sliders {
			s.widget.Value = s.value
		}
	}

	progress := h.easing.Update(gtx, isActive)
	pos := h.easing.InOutBack(progress) * float64(h.height)

	r := image.Rectangle{
		Max: image.Point{
			X: gtx.Constraints.Max.X,
			Y: int(pos),
		},
	}
	op.Offset(image.Pt(0, gtx.Constraints.Max.Y-int(pos))).Add(gtx.Ops)
	layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 127}, clip.Op(clip.Rect(r).Op()))
			stack := clip.Rect(r).Push(gtx.Ops)
			pointer.InputOp{
				Tag:   stack,
				Types: pointer.Scroll | pointer.Move | pointer.Press | pointer.Drag | pointer.Release,
			}.Add(gtx.Ops)
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

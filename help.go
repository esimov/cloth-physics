package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

type help struct {
	fontType   font.Typeface
	lineHeight int
	h1FontSize int
	h2FontSize int
}

type command map[string]string

// ShowHelpDialog activates a dialog panel whith a list of the available key shortcuts.
func (h *Hud) ShowHelpDialog(gtx layout.Context, th *material.Theme, isActive bool) {
	var (
		panelWidth  int
		panelHeight int
	)

	// show the help dialog panel only if it's not yet activated.
	if !isActive {
		return
	}

	paint.FillShape(gtx.Ops, color.NRGBA{R: 127, G: 127, B: 127, A: 70},
		clip.UniformRRect(image.Rectangle{
			Max: image.Point{
				X: gtx.Constraints.Max.X,
				Y: gtx.Constraints.Max.X,
			},
		}, 0).Op(gtx.Ops))

	centerX := windowWidth / 2
	centerY := windowHeight / 2

	switch width := gtx.Dp(unit.Dp(windowWidth)); {
	case width <= windowSizeX:
		panelWidth = gtx.Dp(unit.Dp(windowWidth / 2))
	default:
		panelWidth = gtx.Dp(unit.Dp(windowWidth / 3))
	}
	panelHeight = gtx.Dp(unit.Dp(300))

	px := panelWidth / 2
	py := panelHeight / 2
	dx, dy := centerX-px, centerY-py

	// Limit the applicable constraints to the panel size from this point onward.
	gtx.Constraints.Min.X = panelWidth
	gtx.Constraints.Max.X = panelWidth

	// This offset will apply to the rest of the content laid out in this function.
	defer op.Offset(image.Point{X: dx, Y: dy}).Push(gtx.Ops).Pop()

	layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
				clip.UniformRRect(image.Rectangle{
					Max: image.Point{
						X: panelWidth,
						Y: panelHeight,
					},
				}, gtx.Dp(5)).Op(gtx.Ops))

			paint.FillShape(gtx.Ops, color.NRGBA{A: 127},
				clip.Stroke{
					Path: clip.Rect{Max: image.Point{
						X: panelWidth,
						Y: panelHeight,
					}}.Path(),
					Width: 0.2,
				}.Op(),
			)

			return layout.UniformInset(20).Layout(gtx, func(gtx C) D {
				layout.Center.Layout(gtx, func(gtx C) D {
					h1 := material.H2(th, "Quick help")
					h1.TextSize = unit.Sp(h.h1FontSize)
					h1.Font.Typeface = h.fontType
					h1.Font.Weight = font.SemiBold

					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(h1.Layout),
					)
				})
				gtx.Constraints.Min.X = panelWidth - 220

				defer op.Offset(image.Point{X: 0, Y: 50}).Push(gtx.Ops).Pop()
				h.list.Layout(gtx, len(h.commands),
					func(gtx C, index int) D {
						builder := strings.Builder{}
						if cmd, ok := h.commands[index]; ok {
							for key := range cmd {
								builder.WriteString(fmt.Sprintf("%s\n", key))
							}
						}
						h2 := material.H2(th, builder.String())
						h2.TextSize = unit.Sp(h.h2FontSize)
						h2.Font.Typeface = h.fontType
						h2.Font.Weight = font.Weight(font.SemiBold)

						return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(h2.Layout),
						)
					},
				)
				defer op.Offset(image.Point{X: gtx.Dp(200), Y: 0}).Push(gtx.Ops).Pop()
				h.list.Layout(gtx, len(h.commands),
					func(gtx C, index int) D {
						builder := strings.Builder{}
						if cmd, ok := h.commands[index]; ok {
							for _, desc := range cmd {
								builder.WriteString(fmt.Sprintf("%s\n", desc))
							}
						}
						h2 := material.H2(th, builder.String())
						h2.TextSize = unit.Sp(h.h2FontSize)
						h2.Font.Typeface = h.fontType
						h2.Font.Weight = font.Weight(font.Regular)

						return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(h2.Layout),
						)
					},
				)
				return layout.Dimensions{}
			})
		}),
	)
}

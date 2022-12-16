package main

import (
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func main() {
	width, height := unit.Dp(800), unit.Dp(600)

	go func() {
		w := app.NewWindow(
			app.Title("Tearable Cloth"),
			app.Size(width, height),
		)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func NewGUI() *gui {
	gui := &gui{
		ctx: layout.Context{
			Ops: new(op.Ops),
		},
	}
	return gui
}

func loop(w *app.Window) error {
	var ops op.Ops
	p1 := NewParticle(220, 20, 20)
	p2 := NewParticle(320, 20, 20)
	p3 := NewParticle(220, 120, 20)
	p4 := NewParticle(320, 120, 20)
	particles := []*Particle{p1, p2, p3, p4}

	st1 := NewStick(p1, p2, getDistance(p1, p3))
	st2 := NewStick(p2, p4, getDistance(p1, p2))
	st3 := NewStick(p4, p3, getDistance(p2, p4))
	st4 := NewStick(p3, p1, getDistance(p4, p3))
	sticks := []*Stick{st1, st2, st3, st4}

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				op.InvalidateOp{}.Add(gtx.Ops)
				key.InputOp{Tag: w, Keys: key.NameEscape}.Add(gtx.Ops)

				for _, ev := range gtx.Queue.Events(w) {
					if e, ok := ev.(key.Event); ok && e.Name == key.NameEscape {
						w.Perform(system.ActionClose)
					}
				}

				fillBackground(gtx, color.NRGBA{R: 0xf2, G: 0xf2, B: 0xf2, A: 0xff})
				for _, p := range particles {
					p.Update(gtx)
				}

				for _, st := range sticks {
					st.Update(gtx)
				}

				e.Frame(gtx.Ops)
			}
		}
	}
}

func fillBackground(gtx layout.Context, col color.NRGBA) {
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

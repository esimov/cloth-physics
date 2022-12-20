package main

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
)

type Mouse struct {
	px, py       float32
	prevX, prevY float32
	leftDown     bool
	rightDown    bool
}

func (m *Mouse) updatePosition(x, y float32) {
	m.prevX = m.px
	m.prevY = m.py

	m.px = x
	m.py = y
}

func (m *Mouse) getCurrentPosition(ev pointer.Event) f32.Point {
	return ev.Position
}

func (m *Mouse) setLeftMousePress() {
	m.leftDown = !m.leftDown
}

func (m *Mouse) getLeftMousePress() bool {
	return m.leftDown
}

func (m *Mouse) setRightMousePress() {
	m.rightDown = !m.rightDown
}

func (m *Mouse) getRightMousePress() bool {
	return m.rightDown
}

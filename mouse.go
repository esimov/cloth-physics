package main

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
)

type Mouse struct {
	x, y       float64
	px, py     float64
	leftDown   bool
	rightDown  bool
	isDragging bool
}

func (m *Mouse) updatePosition(x, y float64) {
	m.px = m.x
	m.py = m.y

	m.x = x
	m.y = y
}

func (m *Mouse) getCurrentPosition(ev pointer.Event) f32.Point {
	return ev.Position
}

func (m *Mouse) setLeftMouseButton() {
	m.leftDown = true
}

func (m *Mouse) releaseLeftMouseButton() {
	m.leftDown = false
}

func (m *Mouse) getLeftMouseButton() bool {
	return m.leftDown
}

func (m *Mouse) setRightMouseButton() {
	m.rightDown = true
}

func (m *Mouse) releaseRightMouseButton() {
	m.rightDown = false
}

func (m *Mouse) getRightMouseButton() bool {
	return m.rightDown
}

func (m *Mouse) setDragging(dragging bool) {
	m.isDragging = dragging
}

func (m *Mouse) getDragging() bool {
	return m.isDragging
}

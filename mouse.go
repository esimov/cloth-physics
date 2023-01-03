package main

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
)

type Mouse struct {
	x, y       float64
	px, py     float64
	force      float64
	leftDown   bool
	rightDown  bool
	isDragging bool
	ctrlDown   bool
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

func (m *Mouse) setLeftButton() {
	m.leftDown = true
}

func (m *Mouse) releaseLeftButton() {
	m.leftDown = false
}

func (m *Mouse) getLeftButton() bool {
	return m.leftDown
}

func (m *Mouse) setRightButton() {
	m.rightDown = true
}

func (m *Mouse) releaseRightButton() {
	m.rightDown = false
}

func (m *Mouse) getRightButton() bool {
	return m.rightDown
}

func (m *Mouse) setDragging(dragging bool) {
	m.isDragging = dragging
}

func (m *Mouse) getDragging() bool {
	return m.isDragging
}

func (m *Mouse) setCtrlDown(status bool) {
	m.ctrlDown = status
}

func (m *Mouse) getCtrlDown() bool {
	return m.ctrlDown
}

func (m *Mouse) increaseForce(force float64) {
	m.force = force
}

func (m *Mouse) getForce() float64 {
	return m.force
}

func (m *Mouse) resetForce() {
	m.force = 0
}

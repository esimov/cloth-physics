package physics

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
)

type Mouse struct {
	x, y       float64
	px, py     float64
	force      float64
	scrollY    unit.Dp
	maxScrollY unit.Dp
	leftDown   bool
	rightDown  bool
	isDragging bool
	ctrlDown   bool
}

func (m *Mouse) UpdatePosition(x, y float64) {
	m.px = m.x
	m.py = m.y

	m.x = x
	m.y = y
}

func (m *Mouse) GetCurrentPosition(ev pointer.Event) f32.Point {
	return ev.Position
}

func (m *Mouse) SetLeftButton() {
	m.leftDown = true
}

func (m *Mouse) GetLeftButton() bool {
	return m.leftDown
}

func (m *Mouse) ReleaseLeftButton() {
	m.leftDown = false
}

func (m *Mouse) SetRightButton() {
	m.rightDown = true
}

func (m *Mouse) GetRightButton() bool {
	return m.rightDown
}

func (m *Mouse) ReleaseRightButton() {
	m.rightDown = false
}

func (m *Mouse) SetDragging(dragging bool) {
	m.isDragging = dragging
}

func (m *Mouse) GetDragging() bool {
	return m.isDragging
}

func (m *Mouse) SetCtrlDown(status bool) {
	m.ctrlDown = status
}

func (m *Mouse) GetCtrlDown() bool {
	return m.ctrlDown
}

func (m *Mouse) SetForce(force float64) {
	m.force = force
}

func (m *Mouse) GetForce() float64 {
	return m.force
}

func (m *Mouse) ResetForce() {
	m.force = 0
}

func (m *Mouse) SetScrollY(scrollY unit.Dp) {
	m.scrollY = scrollY
}

func (m *Mouse) GetScrollY() unit.Dp {
	return m.scrollY
}

func (m *Mouse) SetMaxScrollY(maxScrollY unit.Dp) {
	m.maxScrollY = maxScrollY
}

func (m *Mouse) GetMaxScrollY() unit.Dp {
	return m.maxScrollY
}

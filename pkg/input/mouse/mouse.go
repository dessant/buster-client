package mouse

import (
	"github.com/go-vgo/robotgo"
)

func MoveMouse(x, y int) {
	robotgo.MoveMouseSmooth(x, y, 0.5, 1.0, 0)
}

func ClickMouse() {
	robotgo.MouseClick()
}

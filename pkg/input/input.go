package input

import (
	"buster-client/pkg/input/keyboard"
	"buster-client/pkg/input/mouse"
)

func MoveMouse(x, y int) {
	mouse.MoveMouse(x, y)
}

func ClickMouse() {
	mouse.ClickMouse()
}

func TypeText(text string) {
	keyboard.TypeText(text)
}

func PressKey(key string) {
	keyboard.PressKey(key)
}

func ReleaseKey(key string) {
	keyboard.ReleaseKey(key)
}

func TapKey(key string) {
	keyboard.TapKey(key)
}

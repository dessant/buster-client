package keyboard

import (
	"fmt"

	"github.com/go-vgo/robotgo"

	"buster-client/pkg/input/utils"
)

func TypeText(text string) {
	for _, char := range text {
		robotgo.TypeStr(fmt.Sprintf("%c", char))
		utils.RandSleep(60, 80)
	}
}

func PressKey(key string) {
	toggleKey(key, true)
}

func ReleaseKey(key string) {
	toggleKey(key, false)
}

func TapKey(key string) {
	PressKey(key)

	utils.RandSleep(35, 40)

	ReleaseKey(key)
}

func toggleKey(key string, down bool) {
	var d string
	if down {
		d = "down"
	} else {
		d = "up"
	}

	robotgo.KeyToggle(key, d)
}

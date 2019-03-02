package utils

import (
	"os"
	"os/user"
	"runtime"
)

func UserAdmin() (bool, error) {
	if runtime.GOOS == "windows" {
		f, err := os.Open(`\\.\PHYSICALDRIVE0`)
		f.Close()

		if err == nil {
			return true, nil
		}
	} else {
		usr, err := user.Current()
		if err != nil {
			return false, err
		}

		if usr.Username == "root" {
			return true, nil
		}
	}

	return false, nil
}

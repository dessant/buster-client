package utils

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
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

func InitLogger(file string) {
	logPath := filepath.Join(os.TempDir(), file)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err == nil {
		log.SetOutput(logFile)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	} else {
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	}
}

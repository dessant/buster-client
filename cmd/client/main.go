package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/dessant/nativemessaging"
	"github.com/go-vgo/robotgo"

	"buster-client/utils"
)

var buildVersion string

type message struct {
	MessageID  string `json:"id"`
	APIVersion string `json:"apiVersion"`
	Command    string `json:"command"`
	Data       string `json:"data"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

type response struct {
	MessageID  string `json:"id"`
	APIVersion string `json:"apiVersion"`
	Success    bool   `json:"success"`
	Data       string `json:"data"`
}

func installClient(version string) error {
	admin, err := utils.UserAdmin()
	if err != nil {
		log.Println(err)
		return errors.New("cannot inspect current user")
	}
	if admin {
		return errors.New("setup must be run without administrative rights")
	}

	goos := runtime.GOOS
	if goos == "darwin" {
		goos = "macos"
	}
	url := fmt.Sprintf("https://github.com/dessant/buster-client/releases/download/v%s/buster-client-v%s-%s-%s", version, version, goos, runtime.GOARCH)
	if goos == "windows" {
		url += ".exe"
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}

	newExecPath := execPath + ".new"
	currentExecPath := execPath + ".old"

	os.Remove(newExecPath)
	if err := downloadFile(newExecPath, url); err != nil {
		log.Println(err)
		return errors.New("cannot download client")
	}

	os.Remove(currentExecPath)
	if err := os.Rename(execPath, currentExecPath); err != nil {
		log.Println(err)
		return errors.New("cannot rename current client")
	}

	if err := os.Rename(newExecPath, execPath); err != nil {
		log.Println(err)
		if err := os.Rename(currentExecPath, execPath); err != nil {
			log.Println(err)
			return errors.New("cannot undo current client rename")
		}
		return errors.New("cannot rename new client")
	}

	return nil
}

func processMessage(msg *message, rsp *response) error {
	log.Printf("Command: %s", msg.Command)

	if msg.Command == "pressKey" {
		log.Printf("Data: %s", msg.Data)
		robotgo.KeyToggle(msg.Data, "down")
	} else if msg.Command == "releaseKey" {
		log.Printf("Data: %s", msg.Data)
		robotgo.KeyToggle(msg.Data, "up")
	} else if msg.Command == "tapKey" {
		log.Printf("Data: %s", msg.Data)
		robotgo.KeyTap(msg.Data)
	} else if msg.Command == "typeText" {
		log.Printf("Data: %s", msg.Data)
		robotgo.TypeStrDelay(msg.Data, 800)
	} else if msg.Command == "moveMouse" {
		log.Printf("X: %d, Y: %d", msg.X, msg.Y)
		if runtime.GOOS == "windows" {
			robotgo.MoveMouseSmooth(msg.X, msg.Y, 0.5, 1.0, 0)
		} else {
			robotgo.MoveMouseSmooth(msg.X, msg.Y, 1.0, 2.0, 1)
		}
	} else if msg.Command == "clickMouse" {
		robotgo.MouseClick()
	} else if msg.Command == "ping" {
		rsp.Data = "pong"
	} else {
		return errors.New("unknown command")
	}

	return nil
}

func main() {
	utils.InitLogger("buster-client-log.txt")
	log.Printf("Starting client (version: %s)", buildVersion)

	go func() {
		<-time.After(10 * time.Minute)
		log.Println("Closing client (forced)")
		os.Exit(0)
	}()

	decoder := nativemessaging.NewNativeJSONDecoder(os.Stdin)
	encoder := nativemessaging.NewNativeJSONEncoder(os.Stdout)

	for {
		log.Println("Receiving message")

		var msg message
		var rsp response

		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				log.Println("Closing client")
			} else {
				log.Println(err)
			}
			return
		}

		rsp.MessageID = msg.MessageID
		rsp.APIVersion = buildVersion

		if msg.Command == "installClient" {
			log.Printf("Installing client (version: %s)", msg.Data)
			err := installClient(msg.Data)
			if err == nil {
				rsp.Success = true
			} else {
				rsp.Data = err.Error()
			}
		} else {
			if msg.APIVersion == buildVersion {
				log.Println("Processing message")
				err := processMessage(&msg, &rsp)
				if err == nil {
					rsp.Success = true
				} else {
					rsp.Data = err.Error()
				}
			} else {
				log.Printf("Unsupported client version (requested: %s, supported: %s)", msg.APIVersion, buildVersion)
			}
		}

		log.Println("Sending response")

		if err := encoder.Encode(rsp); err != nil {
			log.Println(err)
			return
		}
	}
}

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/dessant/nativemessaging"

	"buster-client/pkg/input"
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
	url := fmt.Sprintf("https://github.com/dessant/buster-client/releases/download/v%s/buster-client-setup-v%s-%s-%s", version, version, goos, runtime.GOARCH)
	if goos == "windows" {
		url += ".exe"
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}
	setupPath := filepath.Join(filepath.Dir(execPath), utils.GetExecName("buster-client-setup"))

	if err := downloadFile(setupPath, url); err != nil {
		log.Println(err)
		return errors.New("cannot download setup")
	}

	setupOutput, err := exec.Command(setupPath, "--update").CombinedOutput()
	if err != nil {
		log.Println(err)
		return errors.New("cannot run setup")
	}

	os.Remove(setupPath)

	errMessage := string(setupOutput)
	if errMessage != "" {
		return errors.New(errMessage)
	}

	return nil
}

func installCleanup() error {
	execPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}

	os.Remove(execPath + ".old")

	return nil
}

func processMessage(msg *message, rsp *response) error {
	log.Printf("Command: %s", msg.Command)

	if msg.Command == "pressKey" {
		log.Printf("Data: %s", msg.Data)
		input.PressKey(msg.Data)
	} else if msg.Command == "releaseKey" {
		log.Printf("Data: %s", msg.Data)
		input.ReleaseKey(msg.Data)
	} else if msg.Command == "tapKey" {
		log.Printf("Data: %s", msg.Data)
		input.TapKey(msg.Data)
	} else if msg.Command == "typeText" {
		log.Printf("Data: %s", msg.Data)
		input.TypeText(msg.Data)
	} else if msg.Command == "moveMouse" {
		log.Printf("X: %d, Y: %d", msg.X, msg.Y)
		input.MoveMouse(msg.X, msg.Y)
	} else if msg.Command == "clickMouse" {
		input.ClickMouse()
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
		<-time.After(12 * time.Minute)
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
		} else if msg.Command == "installCleanup" {
			log.Println("Cleaning up after installation")
			err := installCleanup()
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

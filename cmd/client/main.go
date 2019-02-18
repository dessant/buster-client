package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dessant/nativemessaging"
	"github.com/go-vgo/robotgo"
)

const apiVersion = "1"

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

func initLogger() {
	logPath := filepath.Join(os.TempDir(), "buster-client-log.txt")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(logFile)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	} else {
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	}
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
	initLogger()
	log.Printf("Starting client (version: %s)", buildVersion)

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
		rsp.APIVersion = apiVersion

		if msg.APIVersion == apiVersion {
			log.Println("Processing message")
			if err := processMessage(&msg, &rsp); err == nil {
				rsp.Success = true
			}
		} else {
			log.Printf("Unsupported API version (requested: %s, supported: %s)", msg.APIVersion, apiVersion)
		}

		log.Println("Sending response")

		if err := encoder.Encode(rsp); err != nil {
			log.Println(err)
			return
		}
	}
}

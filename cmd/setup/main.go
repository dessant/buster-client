package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/dessant/open-golang/open"

	"buster-client/utils"
)

const initResponse = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Buster Client - Setup</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
      html,
      body {
        height: 100% !important;
      }
      body {
        margin: 0 !important;
      }
      iframe {
        all: initial !important;
        position: fixed !important;
        top: 0 !important;
        right: 0 !important;
        bottom: 0 !important;
        left: 0 !important;
        width: 100% !important;
        height: 100% !important;
        z-index: 2147483647 !important;
      }
      .hidden {
        display: none !important;
      }
    </style>
  </head>
  <body>
    <div id="notice" class="hidden">Open this page in a browser with the latest version of <a href="https://github.com/dessant/buster" target="_blank" rel="noreferrer">Buster</a> installed.</div>
    <script>
      window.setTimeout(() => {
        document.querySelector('#notice').classList.remove('hidden');
      }, 1000);
    </script>
  </body>
</html>
`

var buildVersion string

var server *http.Server
var shutdown = make(chan bool)
var updating *bool

var session string

type manifest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Path              string   `json:"path"`
	Type              string   `json:"type"`
	AllowedExtensions []string `json:"allowed_extensions,omitempty"`
	AllowedOrigins    []string `json:"allowed_origins,omitempty"`
}

func isValidSession(key string) bool {
	return key == session
}

func getLocation(browser, targetEnv string) (map[string]string, error) {
	admin, err := utils.UserAdmin()
	if err != nil {
		log.Println(err)
		return nil, errors.New("cannot inspect current user")
	}
	if admin {
		return nil, errors.New("setup must be run without administrative rights")
	}

	location := map[string]string{
		"appDir":      "",
		"manifestDir": "",
	}

	if runtime.GOOS == "linux" {
		usr, err := user.Current()
		if err != nil {
			log.Println(err)
			return nil, errors.New("cannot get current user")
		}
		home := usr.HomeDir

		manifestDir := map[string]string{
			"chrome":   filepath.Join(home, ".config/google-chrome/NativeMessagingHosts"),
			"firefox":  filepath.Join(home, ".mozilla/native-messaging-hosts"),
			"opera":    filepath.Join(home, ".config/google-chrome/NativeMessagingHosts"),
			"chromium": filepath.Join(home, ".config/chromium/NativeMessagingHosts"),
		}
		manifest := manifestDir[browser]
		if manifest == "" {
			manifest = manifestDir[targetEnv]
		}

		location["appDir"] = filepath.Join(home, ".local/opt/buster")
		location["manifestDir"] = manifest
	} else if runtime.GOOS == "darwin" {
		usr, err := user.Current()
		if err != nil {
			log.Println(err)
			return nil, errors.New("cannot get current user")
		}
		home := usr.HomeDir

		manifestDir := map[string]string{
			"chrome":   filepath.Join(home, "Library/Application Support/Google/Chrome/NativeMessagingHosts"),
			"firefox":  filepath.Join(home, "Library/Application Support/Mozilla/NativeMessagingHosts"),
			"opera":    filepath.Join(home, "Library/Application Support/Google/Chrome/NativeMessagingHosts"),
			"chromium": filepath.Join(home, "Library/Application Support/Chromium/NativeMessagingHosts"),
		}
		manifest := manifestDir[browser]
		if manifest == "" {
			manifest = manifestDir[targetEnv]
		}

		location["appDir"] = filepath.Join(home, ".local/opt/buster")
		location["manifestDir"] = manifest
	} else if runtime.GOOS == "windows" {
		localAppDataDir := os.Getenv("LOCALAPPDATA")

		location["appDir"] = filepath.Join(localAppDataDir, "buster")
	}

	return location, nil
}

func install(manifestDir, appDir, targetEnv, extension string) error {
	execName := utils.GetExecName("buster-client")
	execPath := filepath.Join(appDir, execName)
	if err := restoreAsset(execPath, execName); err != nil {
		log.Println(err)
		return errors.New("cannot unpack client")
	}

	if runtime.GOOS == "windows" {
		var manifestSubDir string
		if targetEnv == "firefox" {
			manifestSubDir = "firefox"
		} else {
			manifestSubDir = "chrome"
		}
		manifestDir = filepath.Join(appDir, "manifest", manifestSubDir)
	}

	manifestPath := filepath.Join(manifestDir, "org.buster.client.json")
	isPath, err := pathExists(manifestPath)
	if err != nil {
		log.Println(err)
		return errors.New("cannot check manifest file")
	}

	var extensions []string

	if isPath {
		content, err := ioutil.ReadFile(manifestPath)
		if err == nil {
			currentManifestData := manifest{}
			err := json.Unmarshal(content, &currentManifestData)
			if err == nil {
				if targetEnv == "firefox" {
					extensions = append(extensions, currentManifestData.AllowedExtensions...)
				} else {
					extensions = append(extensions, currentManifestData.AllowedOrigins...)
				}
			}
		}
	}

	if !stringInSlice(extensions, extension) {
		extensions = append(extensions, extension)
	}

	manifestData := manifest{
		Name:        "org.buster.client",
		Description: "Buster",
		Type:        "stdio",
		Path:        filepath.Join(appDir, execName),
	}
	if targetEnv == "firefox" {
		manifestData.AllowedExtensions = extensions
	} else {
		manifestData.AllowedOrigins = extensions
	}

	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		log.Println(err)
		return errors.New("cannot create manifest directory")
	}
	manifestJSON, _ := json.MarshalIndent(manifestData, "", "  ")
	if err := ioutil.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		log.Println(err)
		return errors.New("cannot save manifest file")
	}

	if runtime.GOOS == "windows" {
		if err := setManifestRegistry(targetEnv, manifestPath); err != nil {
			log.Println(err)
			return errors.New("cannot set registry value")
		}
	}

	return nil
}

func update() error {
	setupPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}

	execName := utils.GetExecName("buster-client")
	execPath := filepath.Join(filepath.Dir(setupPath), execName)

	newExecPath := execPath + ".new"
	currentExecPath := execPath + ".old"

	if err := restoreAsset(newExecPath, execName); err != nil {
		log.Println(err)
		return errors.New("cannot unpack new client")
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


func writeError(res http.ResponseWriter, error error) {
	response, _ := json.Marshal(map[string]string{"error": error.Error()})
	res.Header().Set("Content-Type", "application/json")
	http.Error(res, string(response), http.StatusBadRequest)
}

func initHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		if isValidSession(req.FormValue("session")) {
			res.Header().Set("Content-Security-Policy", "frame-ancestors 'none';")
			res.Header().Set("Content-Type", "text/html; charset=utf-8")
			log.Println("Loading setup page")
			fmt.Fprint(res, initResponse)
		} else {
			panic(http.ErrAbortHandler)
		}
	} else {
		panic(http.ErrAbortHandler)
	}
}

func locationHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		if isValidSession(req.FormValue("session")) {
			browser := req.FormValue("browser")
			targetEnv := req.FormValue("targetEnv")

			log.Println("Getting install location")
			location, err := getLocation(browser, targetEnv)
			if err != nil {
				writeError(res, err)
				return
			}
			response, _ := json.Marshal(location)

			res.Header().Set("Content-Type", "application/json")
			fmt.Fprint(res, string(response))
		} else {
			panic(http.ErrAbortHandler)
		}
	} else {
		panic(http.ErrAbortHandler)
	}
}

func installHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		if isValidSession(req.FormValue("session")) {
			appDir := req.FormValue("appDir")
			manifestDir := req.FormValue("manifestDir")
			targetEnv := req.FormValue("targetEnv")
			extension := req.FormValue("extension")

			log.Println("Installing client")
			if err := install(manifestDir, appDir, targetEnv, extension); err != nil {
				writeError(res, err)
				return
			}

			// remove legacy client
			os.Remove(filepath.Join(appDir, utils.GetExecName("buster")))

			res.WriteHeader(http.StatusOK)
		} else {
			panic(http.ErrAbortHandler)
		}
	} else {
		panic(http.ErrAbortHandler)
	}
}

func closeHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		if isValidSession(req.FormValue("session")) {
			res.WriteHeader(http.StatusOK)
			log.Println("Closing setup")
			go exit()
		} else {
			panic(http.ErrAbortHandler)
		}
	} else {
		panic(http.ErrAbortHandler)
	}
}

func notFoundHandler(res http.ResponseWriter, req *http.Request) {
	panic(http.ErrAbortHandler)
}

func exit() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(ctx)
	shutdown <- true
}

func main() {
	utils.InitLogger("buster-client-setup-log.txt")
	log.Printf("Starting setup (version: %s)", buildVersion)

	go func() {
		<-time.After(10 * time.Minute)
		log.Println("Closing setup (forced)")
		if *updating {
			// send error to stdout
			fmt.Println("setup timed out")
		}
		exit()
	}()

	updating = flag.Bool("update", false, "update client")
	flag.Parse()
	if *updating {
		log.Println("Updating client")
		if err := update(); err != nil {
			// send error to stdout
			fmt.Println(err)
		}
		log.Println("Closing setup")
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/buster/setup", initHandler)
	mux.HandleFunc("/api/v1/setup/location", locationHandler)
	mux.HandleFunc("/api/v1/setup/install", installHandler)
	mux.HandleFunc("/api/v1/setup/close", closeHandler)
	mux.HandleFunc("/", notFoundHandler)

	server = &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Println(err)
		return
	}

	session = newToken()
	url := fmt.Sprintf("http://%s/buster/setup?session=%s", listener.Addr(), session)
	go open.Start(url)

	if err := server.Serve(listener); err != nil {
		if err == http.ErrServerClosed {
			<-shutdown
		} else {
			log.Println(err)
		}
	}
}

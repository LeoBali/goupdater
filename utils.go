package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	serviceUrl = "https://appsitory.com/updater.json?method=update"
	firstOpenUrl = "https://appsitory.com/?utm_source=updater_1.0"
	desktopUrl = "https://appsitory.com/?utm_source=desktop"
)

func OpenUrl(urlString string) {
	log.Printf("opening url %s", urlString)
	exec.Command("rundll32", "url.dll,FileProtocolHandler", urlString).Start()
}

func GetExecutablePath() string {
    ex, err := os.Executable()
    if err != nil {
        panic(err)
    }
    exPath := filepath.Dir(ex)
    return exPath
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
	   return false
	}
	return !info.IsDir()
}
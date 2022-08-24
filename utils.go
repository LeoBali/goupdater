package main

import (
	"log"
	"os/exec"
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
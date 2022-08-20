package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"time"
)

const appIconResID = 7
var Link string

func main() {
	log.Println("Starting updater")
	tray, err := New()
	if err != nil {
		log.Fatalf("error in new systray %s\r\n", err)
	}

	//err = tray.Show(appIconResID, "Updater")
	err = tray.ShowCustom("updater.ico", "Updater")
	if err != nil {
		log.Fatalf("error in show systray %s\r\n", err)
	}

	// Append more menu items and use tray.AppendSeparator() to separate them.
	tray.AppendMenu("Quit", func() {
		fmt.Println("Quit")
		os.Exit(0)
	})

	go func() {
		time.Sleep(5 * time.Second)
		var uid string
		user, err := user.Current()
		if err == nil {
			uid = user.Uid
		}
		log.Printf("user id: %s\r\n", uid)
		ver, apps := GetVerAndApps()
		count, link := Update(ver, apps, uid)
		tray.SetLink(link)
		if count > 0 {
			if count == 1 {
				tray.ShowMessage("Updater", "You have 1 new update", false)
			} else {
				tray.ShowMessage("Updater", fmt.Sprintf("You have %d new updates", count), false)
			}
		}
		if err != nil {
			log.Println(err)
		}
	}()

	err = tray.Run()
	if err != nil {
		log.Println(err)
	}
}

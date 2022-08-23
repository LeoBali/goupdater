package main

import (
	"fmt"
	"log"
	// "os"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"os/user"
	"time"
)

const (
	SIZE_W = 420
	SIZE_H = 180
)

const appIconResID = 7

//var Link string

func (mw *UpdaterWindow) AddNotifyIcon() {
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}

	icon, err := walk.Resources.Image("updater.ico")
	if err != nil {
		log.Fatal(err)
	}
	mw.SetIcon(icon)
	mw.ni.SetIcon(icon)
	mw.ni.SetToolTip("Appsitory Updater")

	exitAction := walk.NewAction()
	if err := exitAction.SetText("E&xit"); err != nil {
		log.Fatal(err)
	}
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := mw.ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}

	mw.ni.SetVisible(true)

	mw.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			mw.Show()
			win.ShowWindow(mw.Handle(), win.SW_RESTORE)
		}
	})

}

func (mw *UpdaterWindow) removeStyle(style int32) {
	currStyle := win.GetWindowLong(mw.hWnd, win.GWL_STYLE)
	win.SetWindowLong(mw.hWnd, win.GWL_STYLE, currStyle&style)
}

type UpdaterWindow struct {
	*walk.MainWindow
	hWnd win.HWND
	ni   *walk.NotifyIcon
	pb	*walk.ProgressBar
	label *walk.TextLabel
	lnkCancel *walk.LinkLabel
	lnkReadMore *walk.LinkLabel
	lnkRescan *walk.LinkLabel
}

func main() {
	uw := new(UpdaterWindow)
	MainWindow{
		AssignTo: &uw.MainWindow,
		Title:    "Appsitory Updater (Beta)",
		Size:     Size{Width: SIZE_W, Height: SIZE_H},
		Font:     Font{Family: "Segoe UI", PointSize: 10},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget {
					TextLabel{
						AssignTo: &uw.label,
						//Text: "Scanning computer...",
						ColumnSpan: 1,
					},
					LinkLabel{
						AssignTo: &uw.lnkRescan,
						Text: `<a id="this" href="#">Rescan</a>`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							log.Printf("id: '%s', url: '%s'\n", link.Id(), link.URL())
						},
						Alignment: AlignHFarVNear,
						Visible: false,
					},
					VSpacer{
					 	ColumnSpan: 2,
					 	Size:       8,
					},
					LinkLabel{
						AssignTo: &uw.lnkReadMore,
						Text: `<a id="this" href="#">View results...</a>`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							log.Printf("id: '%s', url: '%s'\n", link.Id(), link.URL())
						},
						Alignment: AlignHNearVNear,
						ColumnSpan: 2,
						Visible: false,
					},
					ProgressBar{
						AssignTo: &uw.pb,
						MinValue: 0,
						Value:    50,
						MaxValue: 100,
						MaxSize: Size{Height: 20},
					},
					LinkLabel{
						AssignTo: &uw.lnkCancel,
						Text: `<a id="this" href="#">Cancel</a>`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							log.Printf("id: '%s', url: '%s'\n", link.Id(), link.URL())
						},
					},
					VSpacer{
					 	ColumnSpan: 2,
					 	Size:       8,
					},
					LinkLabel{
						Text: `<a id="this" href="#">Settings</a>`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							log.Printf("id: '%s', url: '%s'\n", link.Id(), link.URL())
						},
						Alignment: AlignHNearVNear,
						ColumnSpan: 2,
					},
				},
			},
		},
		/*Children: []Widget{
			HSpacer{},
			TextLabel{
				Text: "Scanning computer...",
			},
			HSpacer{},
			HSplitter{
				MaxSize: Size{Width: 100, Height: 20},
				Children: []Widget{
					ProgressBar{
						MinValue: 0,
						Value:    50,
						MaxValue: 100,
					},
					HSpacer{Size: 30},
					LinkLabel{
						Text: `<a id="this" href="#">Cancel</a>`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							log.Printf("id: '%s', url: '%s'\n", link.Id(), link.URL())
						},
					},
				},
			},
			HSpacer{},
			LinkLabel{
				Text: `<a id="this" href="#">Settings</a>`,
				OnLinkActivated: func(link *walk.LinkLabelLink) {
					log.Printf("id: '%s', url: '%s'\n", link.Id(), link.URL())
				},
				Alignment: AlignHNearVNear,
			},
			HSpacer{},
		},*/
	}.Create()
	uw.hWnd = uw.Handle()
	uw.AddNotifyIcon()
	uw.removeStyle(^win.WS_SIZEBOX)
	//mw.removeStyle(^win.WS_MINIMIZEBOX)
	uw.removeStyle(^win.WS_MAXIMIZEBOX)
	//hMenu := win.GetSystemMenu(mw.hWnd, false)
	//win.RemoveMenu(hMenu, win.SC_CLOSE, win.MF_BYCOMMAND)
	/*
			xScreen := win.GetSystemMetrics(win.SM_CXSCREEN);
		    yScreen := win.GetSystemMetrics(win.SM_CYSCREEN);
		    win.SetWindowPos(
		        mw.Handle(),
		        0,
		        (xScreen - SIZE_W)/2,
		        (yScreen - SIZE_H)/2,
		        SIZE_W,
		        SIZE_H,
		        win.SWP_FRAMECHANGED,
		    )*/

	go func() {
		uw.label.SetText("Scanning computer...")
		for i := 0; i < 50; i++ {
			uw.pb.SetValue(i)
			time.Sleep(10 * time.Millisecond)
		}
		uw.label.SetText("Sending data for analysis...")
		uw.pb.SetValue(75)
		time.Sleep(1 * time.Second)
		uw.label.SetText("Opening results...")	
		uw.pb.SetValue(99)	
		time.Sleep(1 * time.Second)
		uw.label.SetText("You have 12 new updates available.")
		uw.pb.SetVisible(false)
		uw.lnkCancel.SetVisible(false)
		uw.lnkReadMore.SetVisible(true)
		uw.lnkRescan.SetVisible(true)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		var uid string
		user, err := user.Current()
		if err == nil {
			uid = user.Uid
		}
		log.Printf("user id: %s\r\n", uid)
		ver, apps := GetVerAndApps()
		count, _ := Update(ver, apps, uid)

		if count > 0 {
			if count == 1 {
				// Now that the icon is visible, we can bring up an info balloon.
				if err := uw.ni.ShowInfo("We have found one update for your apps", "Click here to view details"); err != nil {
					log.Fatal(err)
				}
			} else {
				// Now that the icon is visible, we can bring up an info balloon.
				if err := uw.ni.ShowInfo(fmt.Sprintf("We have found %d updates for your apps", count), "Click here to view details"); err != nil {
					log.Fatal(err)
				}
			}
		}
		if err != nil {
			log.Println(err)
		}
	}()

	uw.Run()
}

/*func mainOld() {
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
}*/

package main

import (
	"fmt"
	"log"
	"syscall"

	// "os"
	"os/user"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"gopkg.in/toast.v1"
)

func (mw *UpdaterWindow) addNotifyIcon() {
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Println("error in NewNotifyIcon: %v\r\n", err)
	}

	mw.SetIcon(icon)
	mw.ni.SetIcon(icon)
	mw.ni.SetToolTip(appsitoryUpdater)

	mw.ni.MessageClicked().Attach(func() {
		log.Println("balloon clicked")
		OpenUrl(link)
	})

	visitAppsitory := walk.NewAction()
	visitAppsitory.SetText("&Visit appsitory.com...")
	visitAppsitory.Triggered().Attach(func() { OpenUrl(desktopUrl) })
	mw.ni.ContextMenu().Actions().Add(visitAppsitory)

	mw.ni.ContextMenu().Actions().Add(walk.NewSeparatorAction())

	scanForUpdates := walk.NewAction()
	scanForUpdates.SetText("Scan for &updates...")
	scanForUpdates.Triggered().Attach(func() {
		mw.Show()
		win.ShowWindow(mw.Handle(), win.SW_RESTORE)
		go scan(false)
	})
	mw.ni.ContextMenu().Actions().Add(scanForUpdates)

	settingsAction := walk.NewAction()
	settingsAction.SetText("&Settings...")
	settingsAction.Triggered().Attach(func() {
		settings()
	})
	mw.ni.ContextMenu().Actions().Add(settingsAction)

	mw.ni.ContextMenu().Actions().Add(walk.NewSeparatorAction())

	exitAction := walk.NewAction()
	exitAction.SetText("E&xit")
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	mw.ni.ContextMenu().Actions().Add(exitAction)

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

func (uw *UpdaterWindow) hideButtons() {
	uw.removeStyle(^win.WS_SIZEBOX)
	uw.removeStyle(^win.WS_MINIMIZEBOX)
	uw.removeStyle(^win.WS_MAXIMIZEBOX)
	//hMenu := win.GetSystemMenu(mw.hWnd, false)
	//win.RemoveMenu(hMenu, win.SC_CLOSE, win.MF_BYCOMMAND)
}

func (uw *UpdaterWindow) interceptWndProc() {
	var prevWndProcPtr uintptr
	prevWndProcPtr = win.SetWindowLongPtr(uw.hWnd, win.GWL_WNDPROC,
		syscall.NewCallback(func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
			if msg == win.WM_CLOSE {
				win.ShowWindow(hWnd, win.SW_HIDE)
				return 0
			} else if msg == win.WM_APP+0 {
				switch lParam {
				case win.NIN_BALLOONSHOW:
					log.Println("wnd proc NIN_BALLOONSHOW")
				case win.NIN_BALLOONHIDE:
					log.Println("wnd proc NIN_BALLOONHIDE")
				case win.NIN_BALLOONTIMEOUT:
					log.Println("wnd proc NIN_BALLOONTIMEOUT")
				case win.NIN_BALLOONUSERCLICK:
					log.Println("wnd proc NIN_BALLOONUSERCLICK")
				}
			}
			return win.CallWindowProc(prevWndProcPtr, hWnd, msg, wParam, lParam)
		}))
}

func (uw *UpdaterWindow) centerWindow() {
	var rect win.RECT
	win.GetWindowRect(uw.hWnd, &rect)
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	xScreen := win.GetSystemMetrics(win.SM_CXSCREEN)
	yScreen := win.GetSystemMetrics(win.SM_CYSCREEN)
	win.SetWindowPos(
		uw.hWnd,
		0,
		(xScreen-width)/2,
		(yScreen-height)/2,
		width,
		height,
		win.SWP_FRAMECHANGED,
	)
}

type UpdaterWindow struct {
	*walk.MainWindow
	hWnd            win.HWND
	ni              *walk.NotifyIcon
	pb              *walk.ProgressBar
	label           *walk.TextLabel
	lnkCancelRescan *walk.LinkLabel
	lnkReadMore     *walk.LinkLabel
	//lnkRescan   *walk.LinkLabel
}

func settings() {
	Dialog{
		FixedSize: true,
		Icon:      icon,
		Title:     "Settings",
		MinSize:   Size{Width: 400, Height: 250},
		Font:      Font{Family: "Segoe UI", PointSize: 10},
		Layout:    VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "Automatic Update",
				Layout: VBox{},
				Children: []Widget{
					CheckBox{
						Text:      "Notify me when new version of Update Detector is available",
						Alignment: AlignHNearVCenter,
						MinSize:   Size{Width: 400},
					},
					CheckBox{
						Text:      "Automatically update to the last version",
						Checked:   true,
						Alignment: AlignHNearVCenter,
						MinSize:   Size{Width: 400},
					},
				},
				MinSize:   Size{Width: 400},
				Alignment: AlignHNearVCenter,
			},
			VSpacer{
				Size: 8,
			},
			GroupBox{
				Title:  "Windows Starts",
				Layout: VBox{},
				Children: []Widget{
					CheckBox{
						Text:      "Run Software Update when Windows starts",
						Checked:   true,
						Alignment: AlignHNearVCenter,
						MinSize:   Size{Width: 400},
					},
				},
				MinSize:   Size{Width: 400},
				Alignment: AlignHNearVCenter,
			},
		},
	}.Run(uw)
}

func scan(openResults bool) {
	isScan = true
	stop = false
	uw.label.SetText("Scanning computer...")
	uw.lnkCancelRescan.SetText(`<a id="this" href="#">Cancel</a>`)
	uw.lnkReadMore.SetVisible(false)
	uw.pb.SetVisible(true)
	for i := 0; i < 50; i++ {
		uw.pb.SetValue(i)
		if stop {
			isScan = false
			uw.pb.SetVisible(false)
			uw.lnkCancelRescan.SetText(`<a id="this" href="#">Rescan</a>`)
			uw.lnkReadMore.SetVisible(true)
			uw.lnkReadMore.SetEnabled(false)
			uw.lnkReadMore.SetVisible(false)
			uw.label.SetText("Scan stopped.")
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	ver, apps := GetVerAndApps()
	uw.label.SetText("Sending data for analysis...")
	uw.pb.SetValue(75)
	var count int
	var err error
	count, link, err = Update(ver, apps, uid)
	time.Sleep(1 * time.Second)
	if err != nil {
		isScan = false
		uw.pb.SetVisible(false)
		uw.lnkCancelRescan.SetText(`<a id="this" href="#">Rescan</a>`)
		uw.lnkReadMore.SetEnabled(false)
		uw.lnkReadMore.SetVisible(false)
		uw.label.SetText("Error connecting to Update Service.")
		return
	}
	if count > 0 && openResults {
		uw.label.SetText("Opening results...")
		OpenUrl(link)
		uw.pb.SetValue(99)
		time.Sleep(1 * time.Second)
	} else {
		uw.pb.SetValue(99)
		time.Sleep(50 * time.Millisecond)
	}

	isScan = false
	uw.pb.SetVisible(false)
	uw.lnkCancelRescan.SetText(`<a id="this" href="#">Rescan</a>`)
	if count > 0 {
		uw.lnkReadMore.SetEnabled(true)
		uw.lnkReadMore.SetVisible(true)
		uw.label.SetText(fmt.Sprintf("You have %v new updates available.", count))
	} else {
		uw.lnkReadMore.SetEnabled(false)
		uw.lnkReadMore.SetVisible(false)
		uw.label.SetText("You don't have any updates available.")
	}
}

func notification(title string, message string) {
	log.Println()
	if err := uw.ni.ShowCustom(title, message, icon); err != nil {
		log.Printf("error showing balloon %v\r\n", err)
	}
}

func notificationWin10(title string, message string, button string, url string) {
	log.Printf("pushing toast notification %v %v %v %v\r\n", title, message, button, url)
	icoFile := GetExecutablePath() + "\\" + updaterIco
	log.Printf("ico file %v\r\n", icoFile)
	var notification toast.Notification
	if FileExists(icoFile) {
		notification = toast.Notification{
			AppID:   appsitoryUpdaterBeta,
			Title:   title,
			Message: message,
			Icon:    icoFile,
			Actions: []toast.Action{
				{Type: "protocol", Label: button, Arguments: url},
			},
			ActivationArguments: url,
			Duration:            "long",
		}
	} else {
		notification = toast.Notification{
			AppID:   appsitoryUpdaterBeta,
			Title:   title,
			Message: message,
			Actions: []toast.Action{
				{Type: "protocol", Label: button, Arguments: url},
			},
			ActivationArguments: url,
			Duration:            "long",
		}
	}
	err := notification.Push()
	if err != nil {
		log.Printf("error pushing toast notification: %v\r\n", err)
	}
}

const (
	appsitoryUpdaterBeta = "Appsitory Updater (Beta)"
	appsitoryUpdater     = "Appsitory Updater"
	updaterIco           = "updater.ico"
)

var icon walk.Image
var link string
var uw *UpdaterWindow
var isScan bool
var stop bool
var uid string
var isWin10 bool

func main() {
	var err error
	icon, err = walk.Resources.Image("8") // 8 is an icon id
	if err != nil {
		log.Printf("error opening icon from resource: %v\r\n", err)
	}

	isFirstStart := ReadFirstStart()
	if isFirstStart {
		log.Println("first start")
		OpenUrl(firstOpenUrl)
	} else {
		log.Println("not a first start")
	}

	major, minor, err := GetCurrentVersion()
	if err != nil {
		log.Printf("error reading current version %s, assuming not win10\r\n", err)
		isWin10 = false
	}
	if major >= 10 {
		log.Printf("current version %d.%d, assuming win10\r\n", major, minor)
		isWin10 = true
	} else {
		log.Printf("current version %d.%d, assuming not win10\r\n", major, minor)
		isWin10 = false
	}

	user, err := user.Current()
	if err == nil {
		uid = user.Uid
	}
	log.Printf("user id: %s\r\n", uid)

	uw = new(UpdaterWindow)
	children := []Widget{
		Composite{
			Layout: Grid{Columns: 2},
			Children: []Widget{
				TextLabel{
					AssignTo:   &uw.label,
					Text:       "Press Scan to start a scan",
					ColumnSpan: 2,
				},
				VSpacer{
					ColumnSpan: 2,
					Size:       8,
				},
				LinkLabel{
					AssignTo: &uw.lnkReadMore,
					Text:     `<a id="this" href="#">View results...</a>`,
					OnLinkActivated: func(_ *walk.LinkLabelLink) {
						OpenUrl(link)
					},
					Alignment:  AlignHNearVNear,
					ColumnSpan: 2,
					Visible:    false,
					Enabled:    false,
				},
				ProgressBar{
					AssignTo: &uw.pb,
					MinValue: 0,
					//Value:    50,
					MaxValue:   100,
					MaxSize:    Size{Height: 20},
					ColumnSpan: 2,
					Visible:    false,
				},
				VSpacer{
					ColumnSpan: 2,
					Size:       16,
				},
				HSpacer{},
				LinkLabel{
					AssignTo: &uw.lnkCancelRescan,
					Text:     `<a id="this" href="#">Scan</a>`,
					OnLinkActivated: func(link *walk.LinkLabelLink) {
						if isScan {
							log.Printf("stopping scan\r\n")
							stop = true
						} else {
							go scan(false)
						}
					},
					Alignment: AlignHFarVNear,
				},
			},
		},
	}
	if isFirstStart {
		MainWindow{
			AssignTo: &uw.MainWindow,
			Title:    appsitoryUpdaterBeta,
			Size:     Size{Width: 420, Height: 200},
			Font:     Font{Family: "Segoe UI", PointSize: 10},
			Layout:   VBox{},
			Children: children,
		}.Create()
	} else {
		MainWindow{
			Visible:  false,
			AssignTo: &uw.MainWindow,
			Title:    appsitoryUpdaterBeta,
			Size:     Size{Width: 420, Height: 200},
			Font:     Font{Family: "Segoe UI", PointSize: 10},
			Layout:   VBox{},
			Children: children,
		}.Create()
	}

	uw.hWnd = uw.Handle()
	uw.addNotifyIcon()
	uw.hideButtons()
	uw.centerWindow()
	uw.interceptWndProc()

	if isFirstStart {
		go scan(true)
	}

	go func() {
		time.Sleep(1 * time.Minute)
		log.Println("background check")
		ver, apps := GetVerAndApps()
		var count int
		count, link, err = Update(ver, apps, uid)
		if err != nil {
			return
		}
		if count > 0 {
			var title string
			if count == 1 {
				title = "We have found one update for your apps"
			} else {
				title = fmt.Sprintf("We have found %d updates for your apps", count)
			}
			if isWin10 {
				notificationWin10(title, "Click to view details", "View details", link)
			} else {
				notification(title, "Click here to view details")
			}

			/*log.Println("showing balloon")
			if count == 1 {
				if err := uw.ni.ShowCustom("We have found one update for your apps", "Click here to view details", icon); err != nil {
					log.Printf("error showing balloon %v\r\n", err)
				}
			} else {
				if err := uw.ni.ShowCustom(fmt.Sprintf("We have found %d updates for your apps", count), "Click here to view details", icon); err != nil {
					log.Printf("error showing balloon %v\r\n", err)
				}
			}*/
		}

	}()

	uw.Run()
}

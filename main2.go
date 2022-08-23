package main

import (
	"log"

	"github.com/lxn/win"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MyWindow struct {
	*walk.MainWindow
	hWnd        win.HWND
	minimizeBox *walk.CheckBox
	maximizeBox *walk.CheckBox
	closeBox    *walk.CheckBox
	sizeBox     *walk.CheckBox
	ni          *walk.NotifyIcon
}

func (mw *MyWindow) SetMinimizeBox() {
	if mw.minimizeBox.Checked() {
		mw.addStyle(win.WS_MINIMIZEBOX)
		return
	}
	mw.removeStyle(^win.WS_MINIMIZEBOX)
}

func (mw *MyWindow) SetMaximizeBox() {
	if mw.maximizeBox.Checked() {
		mw.addStyle(win.WS_MAXIMIZEBOX)
		return
	}
	mw.removeStyle(^win.WS_MAXIMIZEBOX)
}

func (mw *MyWindow) SetSizePersistent() {
	if mw.sizeBox.Checked() {
		mw.addStyle(win.WS_SIZEBOX)
		return
	}
	mw.removeStyle(^win.WS_SIZEBOX)
}

func (mw *MyWindow) addStyle(style int32) {
	currStyle := win.GetWindowLong(mw.hWnd, win.GWL_STYLE)
	win.SetWindowLong(mw.hWnd, win.GWL_STYLE, currStyle|style)
}

func (mw *MyWindow) removeStyle(style int32) {
	currStyle := win.GetWindowLong(mw.hWnd, win.GWL_STYLE)
	win.SetWindowLong(mw.hWnd, win.GWL_STYLE, currStyle&style)
}

func (mw *MyWindow) SetCloseBox() {
	if mw.closeBox.Checked() {
		win.GetSystemMenu(mw.hWnd, true)
		return
	}
	hMenu := win.GetSystemMenu(mw.hWnd, false)
	win.RemoveMenu(hMenu, win.SC_CLOSE, win.MF_BYCOMMAND)
}

func (mw *MyWindow) AddNotifyIcon() {
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}

	// icon, err := walk.Resources.Image("img/show.ico")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// mw.SetIcon(icon)
	// mw.ni.SetIcon(icon)
	mw.ni.SetVisible(true)

	mw.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			mw.Show()
			win.ShowWindow(mw.Handle(), win.SW_RESTORE)
		}
	})

}

func main2() {
	mw := new(MyWindow)
	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "notify icon",
		Size:     Size{Width: 550, Height: 380},
		Layout:   VBox{MarginsZero: true},
		OnSizeChanged: func() {
			if win.IsIconic(mw.Handle()) {
				mw.Hide()
				mw.ni.SetVisible(true)
			}
		},
		Children: []Widget{
			CheckBox{
				AssignTo:            &mw.minimizeBox,
				Text:                "Display minimize Button " ,
				Checked:             true,
				OnCheckStateChanged: mw.SetMinimizeBox,
			},
			CheckBox{
				AssignTo:            &mw.maximizeBox,
				Text:                "Show maximize button" ,
				Checked:             true,
				OnCheckStateChanged: mw.SetMaximizeBox,
			},
			CheckBox{
				AssignTo:            &mw.closeBox,
				Text:                " Show close button" ,
				Checked:             true,
				OnCheckStateChanged: mw.SetCloseBox,
			},
			CheckBox{
				AssignTo:            &mw.sizeBox,
				Text:                "Allow to modify the size" ,
				Checked:             true,
				OnCheckStateChanged: mw.SetSizePersistent,
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	mw.hWnd = mw.Handle()
	mw.AddNotifyIcon()
	mw.removeStyle(^win.WS_SIZEBOX)
	mw.removeStyle(^win.WS_MINIMIZEBOX)
	mw.removeStyle(^win.WS_MAXIMIZEBOX)
	hMenu := win.GetSystemMenu(mw.hWnd, false)
	win.RemoveMenu(hMenu, win.SC_CLOSE, win.MF_BYCOMMAND)
	mw.Run()
}
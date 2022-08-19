package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/user"

	//"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const TrayIconMsg = WM_APP + 1
const QuitMsg = WM_APP + 2

var urlChan chan string
var ti *TrayIcon

func buildMenu() uintptr {
	hMenu := CreateMenu()
	AppendMenu(hMenu, MF_STRING, QuitMsg, "&Quit")
	hMenubar := CreateMenu()
	AppendMenu(hMenubar, MF_POPUP, hMenu, "_Parent")
	return GetSubMenu(hMenubar, 0)
}

func menuShow(hWndParent uintptr) {
	hMenu := buildMenu()
	x, y := GetCursorPos()
	SetForegroundWindow(hWndParent)
	if !TrackPopupMenu(hMenu, TPM_LEFTALIGN, x, y, hWndParent) {
		log.Println("track popup menu failed")
	}
	PostMessage(hWndParent, 0, 0, 0)
}

func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		switch nmsg := LOWORD(uint32(wParam)); nmsg {
		case QuitMsg:
			log.Println("quit from tray menu")
			ti.Dispose()
			os.Exit(0)
		}
	case TrayIconMsg:
		switch nmsg := LOWORD(uint32(lParam)); nmsg {
		case NIN_BALLOONUSERCLICK:
			log.Println("user clicked the balloon notification")
			url := <-urlChan
			log.Printf("opening url %s", url)
			exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		case WM_CONTEXTMENU:
			menuShow(hWnd)
			return 0
		}
	case WM_DESTROY:
		PostQuitMessage(0)
	default:
		r, _ := DefWindowProc(hWnd, msg, wParam, lParam)
		return r
	}
	return 0
}

func createMainWindow() (uintptr, error) {
	hInstance, err := GetModuleHandle(nil)
	if err != nil {
		return 0, err
	}

	wndClass := windows.StringToUTF16Ptr("UpdaterWindow")

	var wcex WNDCLASSEX

	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.LpfnWndProc = windows.NewCallback(wndProc)
	wcex.HInstance = hInstance
	wcex.LpszClassName = wndClass
	if _, err := RegisterClassEx(&wcex); err != nil {
		return 0, err
	}

	hwnd, err := CreateWindowEx(
		0,
		wndClass,
		windows.StringToUTF16Ptr("Updater Window"),
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		400,
		300,
		0,
		0,
		hInstance,
		nil)
	if err != nil {
		return 0, err
	}

	return hwnd, nil
}

func newGUID() GUID {
	var buf [16]byte
	rand.Read(buf[:])
	return *(*GUID)(unsafe.Pointer(&buf[0]))
}

type TrayIcon struct {
	hwnd uintptr
	guid GUID
}

func NewTrayIcon(hwnd uintptr) (*TrayIcon, error) {
	ti := &TrayIcon{hwnd: hwnd, guid: newGUID()}
	data := ti.initData()
	// ..uFlags = NIF_ICON | NIF_TIP | NIF_MESSAGE | NIF_SHOWTIP | NIF_GUID
	data.UFlags |= NIF_ICON | NIF_TIP | NIF_MESSAGE | NIF_SHOWTIP
	data.UCallbackMessage = TrayIconMsg
	if _, err := Shell_NotifyIcon(NIM_ADD, data); err != nil {
		return nil, err
	}
	// nid.uVersion = 4;
	// Shell_NotifyIcon(NIM_SETVERSION, _nid);
	data.UVersion = 4
	Shell_NotifyIcon(NIM_SETVERSION, data)
	return ti, nil
}

func (ti *TrayIcon) initData() *NOTIFYICONDATA {
	var data NOTIFYICONDATA
	data.CbSize = uint32(unsafe.Sizeof(data))
	data.UFlags = NIF_GUID
	data.HWnd = ti.hwnd
	data.GUIDItem = ti.guid
	return &data
}

func (ti *TrayIcon) Dispose() error {
	_, err := Shell_NotifyIcon(NIM_DELETE, ti.initData())
	return err
}

func (ti *TrayIcon) SetIcon(icon uintptr) error {
	data := ti.initData()
	data.UFlags |= NIF_ICON
	data.HIcon = icon
	_, err := Shell_NotifyIcon(NIM_MODIFY, data)
	return err
}

func (ti *TrayIcon) SetTooltip(tooltip string) error {
	data := ti.initData()
	data.UFlags |= NIF_TIP
	copy(data.SzTip[:], windows.StringToUTF16(tooltip))
	_, err := Shell_NotifyIcon(NIM_MODIFY, data)
	return err
}

func (ti *TrayIcon) ShowBalloonNotification(title, text string) error {
	data := ti.initData()
	data.UFlags |= NIF_INFO
	if title != "" {
		copy(data.SzInfoTitle[:], windows.StringToUTF16(title))
	}
	copy(data.SzInfo[:], windows.StringToUTF16(text))
	_, err := Shell_NotifyIcon(NIM_MODIFY, data)
	return err
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Println("starting updater")
	hwnd, err := createMainWindow()
	if err != nil {
		log.Fatalf("error in createMainWindow %s\r\n", err)
	}

	icon, err := LoadImage(
		0,
		windows.StringToUTF16Ptr("updater.ico"),
		IMAGE_ICON,
		0,
		0,
		LR_DEFAULTSIZE|LR_LOADFROMFILE)
	if err != nil {
		log.Fatalf("error in loadImage %s\r\n", err)
	}

	ti, err = NewTrayIcon(hwnd)
	if err != nil {
		log.Fatalf("error in createMainWindow %s\r\n", err)
	}
	defer ti.Dispose()

	ti.SetIcon(icon)
	ti.SetTooltip("Updater")

	urlChan = make(chan string)

	go func(url chan string) {
		var uid string
		user, err := user.Current()
		if err == nil {
			uid = user.Uid
		}
		log.Printf("user id: %s\r\n", user.Uid);
		ver, apps := GetVerAndApps()
		count, link := Update(ver, apps, uid)
		if count > 0 {
			if count == 1 {
				ti.ShowBalloonNotification("Updater", "You have 1 new update")
			} else {
				ti.ShowBalloonNotification("Updater", fmt.Sprintf("You have %d new updates", count))
			}
			url <- link
		}
	}(urlChan)

	ShowWindow(hwnd, windows.SW_HIDE)

	var msg MSG
	for {
		if r, err := GetMessage(&msg, 0, 0, 0); err != nil {
			log.Fatalf("error in getMessage %s\r\n", err)
		} else if r == 0 {
			break
		}
		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
}

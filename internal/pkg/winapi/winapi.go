package winapi

import (
	"github.com/lxn/win"
	"syscall"
	"unsafe"
)

func NewWindowWithClassName(windowClassName string) win.HWND {
	return NewWindow(windowClassName, win.DefWindowProc)
}

func NewWindow(windowClassName string, windowProcedure WindowProcedure) win.HWND {
	return newWindowWithClassName(windowClassName, windowProcedure)
}

func ActivateWindowByPid(pid int) {
	if hWnd := FindWindowByPid(pid); hWnd != win.HWND_TOP {
		win.ShowWindow(hWnd, win.SW_SHOWMAXIMIZED)
		win.SetForegroundWindow(hWnd)
	}
}

func MustUTF16PtrFromString(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

type WindowProcedure = func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr

func newWindowWithClassName(windowClassName string, windowProcedure WindowProcedure) win.HWND {

	wndProc := func(hWnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case win.WM_DESTROY:
			win.PostQuitMessage(0)
		default:
			return windowProcedure(hWnd, msg, wParam, lParam)
		}
		return 0
	}

	mustRegisterWindowClassWithWndProcPtr(
		windowClassName, syscall.NewCallback(wndProc))

	return win.CreateWindowEx(
		0,
		MustUTF16PtrFromString(windowClassName),
		nil,
		0,
		0,
		0,
		0,
		0,
		win.HWND_TOP,
		0,
		win.GetModuleHandle(nil),
		nil)
}

func FindWindowByPid(thePid int) (foundWindow win.HWND) {

	win.EnumChildWindows(0, syscall.NewCallback(func(hWnd win.HWND, lParam uintptr) uintptr {
		var pid uint32
		win.GetWindowThreadProcessId(hWnd, &pid)
		if int(pid) == thePid {
			foundWindow = hWnd
			return 0
		}
		return 1
	}), 0)
	return
}

func mustRegisterWindowClassWithWndProcPtr(className string, wndProcPtr uintptr) {

	hInst := win.GetModuleHandle(nil)
	if hInst == 0 {
		panic("GetModuleHandle")
	}

	hIcon := win.LoadIcon(hInst, win.MAKEINTRESOURCE(7)) // rsrc uses 7 for app icon
	if hIcon == 0 {
		hIcon = win.LoadIcon(0, win.MAKEINTRESOURCE(win.IDI_APPLICATION))
	}
	if hIcon == 0 {
		panic("LoadIcon")
	}

	hCursor := win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	if hCursor == 0 {
		panic("LoadCursor")
	}

	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = wndProcPtr
	wc.HInstance = hInst
	wc.HIcon = hIcon
	wc.HCursor = hCursor
	wc.HbrBackground = win.COLOR_BTNFACE + 1
	wc.LpszClassName = MustUTF16PtrFromString(className)
	wc.Style = 0

	if atom := win.RegisterClassEx(&wc); atom == 0 {
		panic("RegisterClassEx")
	}

}

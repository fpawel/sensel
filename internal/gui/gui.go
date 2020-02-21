package gui

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fpawel/sensel/internal/pkg/winapi"
	"github.com/fpawel/sensel/internal/pkg/winapi/copydata"
	"github.com/lxn/win"
	"os"
	"sync/atomic"
)

type MsgCopyData = uintptr

const (
	MsgNewCommTransaction MsgCopyData = iota
)

func SetHWndTargetSendMessage(hWnd win.HWND) {
	setHWndTargetSendMessage(hWnd)
}

func NotifyNewCommTransaction(c CommTransaction) bool {
	return copyData().SendJson(MsgNewCommTransaction, c)
}

func sendMessage(msg uint32, wParam uintptr, lParam uintptr) uintptr {
	return win.SendMessage(hWndTargetSendMessage(), msg, wParam, lParam)
}

func writeBinary(buf *bytes.Buffer, data interface{}) {
	if err := binary.Write(buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}
}

func copyData() copydata.W {
	return copydata.New(hWndSourceSendMessage, hWndTargetSendMessage())
}

var (
	hWndSourceSendMessage = winapi.NewWindowWithClassName(os.Args[0] + "33BCE8B6-E14D-4060-97C9-2B7E79719195")

	hWndTargetSendMessage,
	setHWndTargetSendMessage = func() (func() win.HWND, func(win.HWND)) {
		hWnd := int64(win.HWND_TOP)
		return func() win.HWND {
				return win.HWND(atomic.LoadInt64(&hWnd))
			}, func(x win.HWND) {
				atomic.StoreInt64(&hWnd, int64(x))
			}
	}()
)

func init() {

	go func() {
		for {
			var msg win.MSG
			if win.GetMessage(&msg, 0, 0, 0) == 0 {
				fmt.Println("выход из цикла оконных сообщений")
				return
			}
			fmt.Printf("%+v\n", msg)
			win.TranslateMessage(&msg)
			win.DispatchMessage(&msg)
		}
	}()
}

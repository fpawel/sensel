package copydata

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/atool/internal"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"os"
	"reflect"
	"sync/atomic"
	"unicode/utf16"
	"unsafe"
)

type W struct {
	HWndSrc, HWndDest win.HWND
	logDestinationMustBeSetAtomic *int32
}

func New(HWndSrc, HWndDest win.HWND) W {
	return W{
		HWndSrc:  HWndSrc,
		HWndDest: HWndDest,
		logDestinationMustBeSetAtomic: new(int32),
	}
}

func (x W) SendString(msg uintptr, s string) bool {

	if x.HWndDest == win.HWND_TOP {
		log.Debug("WM_COPYDATA: destination must be set",
			"window_source", x.HWndSrc,
			"msg_copy_data", msg,
			"copy_data", s)
		return false
	}

	if sendMessage(x.HWndSrc, x.HWndDest, msg, utf16FromString(s)) == 0 {
		log.PrintErr("WM_COPYDATA failed",
			"window_source", x.HWndSrc,
			"window_destination", x.HWndDest,
			"msg_copy_data", msg,
			"copy_data", s)
		return false
	}
	return true
}

func (x W) SendJson(msg uintptr, param interface{}) bool {
	b, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	return x.SendString(msg, string(b))
}

func (x W) SendBytes(msg uintptr, b []byte) bool {

	if x.HWndDest == win.HWND_TOP {
		log.PrintErr("WM_COPYDATA: destination must be set",
			"window_source", x.HWndSrc,
			"msg_copy_data", msg,
			"copy_data_length", len(b))
		return false
	}

	if sendMessage(x.HWndSrc, x.HWndDest, msg, b) == 0 {
		x.logDestinationMustBeSet(msg , b)
		return false
	}
	return true
}

type copyData struct {
	DwData uintptr
	CbData uint32
	LpData uintptr
}

func utf16FromString(s string) (b []byte) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			panic(fmt.Sprintf("%q[%d] is 0", s, i))
		}
	}
	for _, v := range utf16.Encode([]rune(s)) {
		b = append(b, byte(v), byte(v>>8))
	}
	return
}

func sendMessage(hWndSrc, hWndDest win.HWND, wParam uintptr, b []byte) uintptr {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	cd := copyData{
		CbData: uint32(header.Len),
		LpData: header.Data,
		DwData: uintptr(hWndSrc),
	}
	return win.SendMessage(hWndDest, win.WM_COPYDATA, wParam, uintptr(unsafe.Pointer(&cd)))
}

func getCopyData(ptr unsafe.Pointer) (uintptr, []byte) {
	cd := (*copyData)(ptr)
	p := ptrSliceFrom(unsafe.Pointer(cd.LpData), int(cd.CbData))
	return cd.DwData, *(*[]byte)(p)
}

func ptrSliceFrom(p unsafe.Pointer, s int) unsafe.Pointer {
	return unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(p), Len: s, Cap: s})
}

func (x W) logDestinationMustBeSet(msg uintptr, b []byte){
	if atomic.LoadInt32(x.logDestinationMustBeSetAtomic) > 0 {
		return
	}
	atomic.AddInt32(x.logDestinationMustBeSetAtomic, 1, )
	log.PrintErr("WM_COPYDATA failed",
		"window_source", x.HWndSrc,
		"window_destination", x.HWndDest,
		"msg_copy_data", msg,
		"copy_data_length", len(b))
}

var (
	log = structlog.New()

)

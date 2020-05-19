package winapi

import "C"
import (
	"fmt"
	"github.com/fpawel/sensel/internal/pkg/must"
	"reflect"
	"syscall"
	"unsafe"
)

func GetDefaultPrinter() (string, error) {
	var pcchBuffer uintptr
	_, _, errNo := syscall.Syscall(hGetDefaultPrinter.Addr(), 2,
		0,
		uintptr(unsafe.Pointer(&pcchBuffer)),
		0)
	if errNo != 122 {
		return "", fmt.Errorf("error 122 expected, got %d:%+v", errNo, errNo)
	}
	pBuffer := make([]byte, pcchBuffer)
	ret, _, errNo := syscall.Syscall(hGetDefaultPrinter.Addr(), 2,
		uintptr(unsafe.Pointer(&pBuffer[0])),
		uintptr(unsafe.Pointer(&pcchBuffer)),
		0)
	if ret == 0 {
		return "", errNo
	}
	return utf16PtrToString(uintptr(unsafe.Pointer(&pBuffer[0]))), nil
}

func EnumPrinters() ([]string, error) {
	const (
		PRINTER_ENUM_CONNECTIONS = 4
		PRINTER_ENUM_LOCAL       = 2
		flags                    = PRINTER_ENUM_CONNECTIONS | PRINTER_ENUM_LOCAL
		Level                    = 4
	)
	var pcbNeeded, pcReturned uintptr
	_, _, errNo := syscall.Syscall9(hEnumPrinters.Addr(), 7,
		uintptr(flags),
		0,
		uintptr(Level),
		0,
		0,
		uintptr(unsafe.Pointer(&pcbNeeded)),
		uintptr(unsafe.Pointer(&pcReturned)),
		0,
		0)
	if errNo != 122 {
		return nil, fmt.Errorf("error 122 expected, got %d:%+v", errNo, errNo)
	}
	p := make([]printerInfo4, pcbNeeded/sizeOfPrinterInfo4)

	ret, _, errNo := syscall.Syscall9(hEnumPrinters.Addr(), 7,
		uintptr(flags),
		0,
		uintptr(Level),
		uintptr(unsafe.Pointer(&p[0])),
		pcbNeeded,
		uintptr(unsafe.Pointer(&pcbNeeded)),
		uintptr(unsafe.Pointer(&pcReturned)),
		0,
		0)
	if ret == 0 {
		return nil, errNo
	}

	var xs []string
	for i := 0; i < int(pcReturned); i++ {
		pStr := p[i].pPrinterName
		if pStr != 0 {
			xs = append(xs, utf16PtrToString(pStr))
		}
	}
	return xs, nil
}

func utf16PtrToString(pStr uintptr) string {
	Len := getStrLen(pStr)
	p := unsafe.Pointer(&reflect.SliceHeader{
		Data: pStr,
		Len:  Len,
		Cap:  Len,
	})
	pUInt16 := (*[]uint16)(p)
	s := syscall.UTF16ToString(*pUInt16)
	return s
}

func getStrLen(pStr uintptr) (n int) {
	for *(*uint16)(unsafe.Pointer(pStr)) != 0 {
		n++
		pStr += 2
	}
	return n
}

type printerInfo4 struct {
	pPrinterName uintptr
	pServerName  uintptr
	attributes   uintptr
}

var (
	sizeOfPrinterInfo4 = func() uintptr {
		var x printerInfo4
		return unsafe.Sizeof(x.attributes) + unsafe.Sizeof(x.pServerName) + unsafe.Sizeof(x.pPrinterName)
	}()

	winspoolDrv = func() *syscall.DLL {
		h, err := syscall.LoadDLL("Winspool.drv")
		must.PanicIf(err)
		return h
	}()
	hEnumPrinters = func() *syscall.Proc {
		h, err := winspoolDrv.FindProc("EnumPrintersW")
		must.PanicIf(err)
		return h
	}()
	hGetDefaultPrinter = func() *syscall.Proc {
		h, err := winspoolDrv.FindProc("GetDefaultPrinterW")
		must.PanicIf(err)
		return h
	}()
)

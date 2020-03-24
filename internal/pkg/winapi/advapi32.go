package winapi

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/pkg/must"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"syscall"
)

func NotifyRegChangeComport(handler func([]string)) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `hardware\devicemap\serialcomm`, syscall.KEY_NOTIFY)
	if err != nil {
		return err
	}
	for {
		if err := regNotifyChangeKeyValue(k, 0, 0x00000001|0x00000004, 0, 0); err != nil {
			return err
		}
		comports, err := comport.Ports()
		if err != nil {
			return err
		}
		handler(comports)
	}
}

func regNotifyChangeKeyValue(regKey registry.Key, watchSubtree uintptr, dwNotifyFilter uintptr, hEvent windows.Handle, asynchronous uintptr) error {

	advApi32, err := syscall.LoadDLL("Advapi32.dll")
	must.PanicIf(err)

	regNotifyChangeKeyValue, err := advApi32.FindProc("RegNotifyChangeKeyValue")
	must.PanicIf(err)

	_, _, err = regNotifyChangeKeyValue.Call(uintptr(regKey), watchSubtree, dwNotifyFilter, uintptr(hEvent), asynchronous)

	switch int(err.(syscall.Errno)) {
	case 0:
		return nil
	default:
		return err
	}
}

package app

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/pkg/winapi"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sort"
)

func ComboBoxComport(getComportNameFunc func() string, setComportNameFunc func(string)) ComboBox {
	var comboBoxComport *walk.ComboBox
	ports, _ := comport.Ports()
	sort.Strings(ports)
	getCurrentIndex := func() int {
		n := -1
		for i, s := range ports {
			if s == getComportNameFunc() {
				n = i
				break
			}
		}
		return n
	}
	comboboxComports = append(comboboxComports, &comboBoxComport)
	return ComboBox{
		Editable:     true,
		AssignTo:     &comboBoxComport,
		MaxSize:      Size{100, 0},
		Model:        ports,
		CurrentIndex: getCurrentIndex(),
		OnCurrentIndexChanged: func() {
			setComportNameFunc(comboBoxComport.Text())
		},
	}
}

func trackRegChangeComport() {
	_ = winapi.NotifyRegChangeComport(func(ports []string) {
		appWindow.Synchronize(func() {
			for _, cb := range comboboxComports {
				cb := cb
				_ = (*cb).SetModel(ports)
			}
		})
	})
}

var comboboxComports []**walk.ComboBox

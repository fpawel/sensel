package app

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/pkg/winapi"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sort"
)

func (x *cbComport) Combobox() ComboBox {

	ports, _ := comport.Ports()
	sort.Strings(ports)
	getCurrentIndex := func() int {
		n := -1
		c := cfg.Get()
		for i, s := range ports {
			if s == x.fnGet(c) {
				n = i
				break
			}
		}
		return n
	}
	handleChanged := func() {
		if disableComboboxComportTextChanged {
			return
		}
		disableComboboxComportTextChanged = true
		c := cfg.Get()
		x.fnSet(&c, x.cb.Text())
		must.PanicIf(cfg.Set(c))
		disableComboboxComportTextChanged = false
	}
	return ComboBox{
		Editable:              true,
		AssignTo:              &x.cb,
		MaxSize:               Size{100, 0},
		Model:                 ports,
		CurrentIndex:          getCurrentIndex(),
		OnCurrentIndexChanged: handleChanged,
		OnTextChanged:         handleChanged,
	}
}

func trackRegChangeComport() {
	_ = winapi.NotifyRegChangeComport(func(ports []string) {
		appWindow.Synchronize(func() {
			disableComboboxComportTextChanged = true
			c := cfg.Get()
			for _, x := range cbComports {
				_ = x.cb.SetModel(ports)
				_ = x.cb.SetText(x.fnGet(c))
			}
			disableComboboxComportTextChanged = false
		})
	})
}

var disableComboboxComportTextChanged bool

const (
	nCbControlSheet = iota
	nCbGas
	nCbVoltmeter
)

var cbComports = []*cbComport{
	{
		fnGet: func(c cfg.Config) string {
			return c.ControlSheet.Comport
		},
		fnSet: func(c *cfg.Config, s string) {
			c.ControlSheet.Comport = s
		},
	},
	{
		fnGet: func(c cfg.Config) string {
			return c.Gas.Comport
		},
		fnSet: func(c *cfg.Config, s string) {
			c.Gas.Comport = s
		},
	},
	{
		fnGet: func(c cfg.Config) string {
			return c.Voltmeter.Comport
		},
		fnSet: func(c *cfg.Config, s string) {
			c.Voltmeter.Comport = s
		},
	},
}

type cbComport struct {
	cb    *walk.ComboBox
	fnGet func(cfg.Config) string
	fnSet func(*cfg.Config, string)
}

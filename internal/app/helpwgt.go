package app

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sort"
)

func WidgetsConfig() []Widget {
	updCfg := func(f func(c *cfg.Config)) {
		c := cfg.Get()
		f(&c)
		must.PanicIf(cfg.Set(c))
	}

	var productTypes []string
	for _, s := range prodTypes.ListProductTypes() {
		productTypes = append(productTypes, s)
	}
	sort.Strings(productTypes)

	var (
		leMeasurementName *walk.LineEdit
	)

	return []Widget{
		Label{Text: "СОМ порт вольтметра"},
		ComboBoxComport(func() string {
			return cfg.Get().Voltmeter.Comport
		}, func(s string) {
			updCfg(func(c *cfg.Config) {
				c.Voltmeter.Comport = s
			})
		}),
		Label{Text: "СОМ порт газового блока"},
		ComboBoxComport(func() string {
			return cfg.Get().Gas.Comport
		}, func(s string) {
			updCfg(func(c *cfg.Config) {
				c.Gas.Comport = s
			})
		}),
		Label{Text: "Исполнение"},
		ComboBoxWithList(productTypes, func() string {
			return measurement.ProductType
		}, func(s string) {
			measurement.ProductType = s
			setMeasurementViewModel(measurement)
		}),
		Label{Text: "Наименование обмера"},
		LineEdit{
			Text:     measurement.Name,
			AssignTo: &leMeasurementName,
			OnTextChanged: func() {
				measurement.Name = leMeasurementName.Text()
			},
		},
	}
}

func DialogAppConfig() Dialog {
	return Dialog{
		Title: "Настройки",
		Font: Font{
			Family:    appWindow.Font().Family(),
			PointSize: appWindow.Font().PointSize(),
		},
		Layout: VBox{
			Alignment: AlignHNearVNear,
		},
		Children: WidgetsConfig(),
	}
}

func ComboBoxWithList(values []string, getFunc func() string, setFunc func(string)) ComboBox {
	var cb *walk.ComboBox
	var n int
	for i, s := range values {
		if s == getFunc() {
			n = i
			break
		}
	}
	return ComboBox{
		//Editable:true,
		AssignTo:     &cb,
		MaxSize:      Size{100, 0},
		Model:        values,
		CurrentIndex: n,
		OnCurrentIndexChanged: func() {
			setFunc(cb.Text())
		},
	}
}

func ComboBoxComport(getComportNameFunc func() string, setComportNameFunc func(string)) ComboBox {
	ports, _ := comport.Ports()
	sort.Strings(ports)
	return ComboBoxWithList(ports, getComportNameFunc, setComportNameFunc)
}

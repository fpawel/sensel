package app

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
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
	//for _, s := range prodTypes.ListProductTypes() {
	//	productTypes = append(productTypes, s)
	//}
	//sort.Strings(productTypes)

	var (
		leMeasurementName *walk.LineEdit
	)

	measurement := getMainTableViewModel().GetMeasurement()

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

		Label{Text: "Прибор"},
		ComboBoxWithList(productTypes, func() string {
			return getMainTableViewModel().GetMeasurement().Device
		}, func(s string) {
			measurement.Device = s
			must.PanicIf(data.SaveMeasurement(db, &measurement))
			setMeasurementViewModel(measurement)
		}),

		Label{Text: "Тип"},
		ComboBoxWithList(productTypes, func() string {
			return getMainTableViewModel().GetMeasurement().Kind
		}, func(s string) {
			measurement.Kind = s
			must.PanicIf(data.SaveMeasurement(db, &measurement))
			setMeasurementViewModel(measurement)
		}),

		Label{Text: "Наименование обмера"},
		LineEdit{
			Text:     measurement.Name,
			AssignTo: &leMeasurementName,
			OnTextChanged: func() {
				measurement.Name = leMeasurementName.Text()
				must.PanicIf(data.SaveMeasurement(db, &measurement))
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
	n := -1
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

package app

import (
	"github.com/fpawel/comm/comport"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/pkg/winapi"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sort"
	"time"
)

func runDialogMeasurement() (data.Measurement, bool) {
	var (
		dialog           *walk.Dialog
		pbOk, pbCancel   *walk.PushButton
		cbDevice, cbKind *walk.ComboBox
		edName           *walk.LineEdit
		neGas            [4]*walk.NumberEdit
		nDevice, nKind   int
	)

	var m data.Measurement
	_ = data.GetLastMeasurement(db, &m)
	m.Name = ""

	devices := Calc.ListDevices()
	for i := range devices {
		if devices[i] == m.Device {
			nDevice = i
		}
	}
	kinds := Calc.ListKinds(m.Device)
	for i := range kinds {
		if kinds[i] == m.Kind {
			nKind = i
		}
	}
	pgs := func(n int) float64 {
		if n < len(m.Pgs) {
			return m.Pgs[n]
		}
		return 0
	}
	onEditPgs := func(n int) func() {
		return func() {
			for n >= len(m.Pgs) {
				m.Pgs = append(m.Pgs, 0)
			}
			m.Pgs[n] = neGas[n].Value()
		}
	}

	dlg := Dialog{
		AssignTo: &dialog,
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		Title:  "Новое измерение",
		Layout: HBox{},
		Children: []Widget{
			Composite{
				MinSize: Size{150, 0},
				Layout:  VBox{},
				Children: []Widget{
					Label{Text: "Исполнение"},
					ComboBox{
						AssignTo:     &cbDevice,
						Model:        devices,
						CurrentIndex: nDevice,
						OnCurrentIndexChanged: func() {
							m.Device = cbDevice.Text()
							model := Calc.ListKinds(cbDevice.Text())
							must.PanicIf(cbKind.SetModel(model))
							if len(model) > 0 {
								must.PanicIf(cbKind.SetCurrentIndex(0))
							}
						},
					},
					ComboBox{
						AssignTo:     &cbKind,
						CurrentIndex: nKind,
						Model:        kinds,
						OnCurrentIndexChanged: func() {
							m.Kind = cbKind.Text()
						},
					},

					Label{Text: "ПГС"},
					NumberEdit{
						Decimals:       2,
						AssignTo:       &neGas[0],
						Value:          pgs(0),
						OnValueChanged: onEditPgs(0),
					},
					NumberEdit{
						Decimals:       2,
						AssignTo:       &neGas[1],
						Value:          pgs(1),
						OnValueChanged: onEditPgs(1),
					},
					NumberEdit{
						Decimals:       2,
						AssignTo:       &neGas[2],
						Value:          pgs(2),
						OnValueChanged: onEditPgs(2),
					},
					NumberEdit{
						Decimals:       2,
						AssignTo:       &neGas[3],
						Value:          pgs(3),
						OnValueChanged: onEditPgs(3),
					},
					Label{Text: "Наименование"},
					LineEdit{
						AssignTo: &edName,
						OnTextChanged: func() {
							m.Name = edName.Text()
						},
					},
				},
			},
			ScrollView{
				HorizontalFixed: true,
				MinSize:         Size{120, 0},
				Layout:          VBox{},
				Children: []Widget{
					PushButton{
						AssignTo: &pbOk,
						Text:     "Ок",
						OnClicked: func() {
							dialog.Accept()
						},
					},
					PushButton{
						AssignTo: &pbCancel,
						Text:     "Отмена",
						OnClicked: func() {
							dialog.Cancel()
						},
					},
				},
			},
		},
	}

	r, err := dlg.Run(appWindow)
	must.PanicIf(err)
	m.Samples = nil
	m.MeasurementID = 0
	m.CreatedAt = time.Now()
	return m, r == walk.DlgCmdOK
}

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

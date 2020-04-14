package app

import (
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"time"
)

func runDialogFloat1(value float64, title string, caption string, decimals int, min, max float64) (float64, bool) {
	var (
		dialog         *walk.Dialog
		ne             *walk.NumberEdit
		pbOk, pbCancel *walk.PushButton
	)

	dlg := Dialog{
		AssignTo: &dialog,
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		Title:  title,
		Layout: VBox{},
		Children: []Widget{
			Label{Text: caption},
			NumberEdit{
				Decimals: decimals,
				AssignTo: &ne,
				Value:    value,
				MinValue: min,
				MaxValue: max,
				OnValueChanged: func() {
					value = ne.Value()
				},
			},
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
	}

	r, err := dlg.Run(appWindow)
	must.PanicIf(err)
	return value, r == walk.DlgCmdOK
}

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
					Label{Text: "Комменттарий"},
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

func runAppSettingsDialog() {
	c := cfg.Get()
	var (
		edFontSizePixels, edCellHorizSpaceMM, edRowHeightMM *walk.NumberEdit
		dlg                                                 *walk.Dialog
	)

	r, err := Dialog{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		MinSize:  Size{280, 180},
		MaxSize:  Size{280, 180},
		AssignTo: &dlg,
		Title:    "Настройки",
		Layout:   VBox{},
		Children: []Widget{

			Label{Text: "Высота строки таблицы, мм"},
			NumberEdit{
				AssignTo: &edRowHeightMM,
				OnValueChanged: func() {
					c.Table.RowHeightMM = edRowHeightMM.Value()
				},
				Value:    c.Table.RowHeightMM,
				Decimals: 2,
			},

			Label{Text: "Отступ ячеек таблицы, мм"},
			NumberEdit{
				AssignTo: &edCellHorizSpaceMM,
				OnValueChanged: func() {
					c.Table.CellHorizSpaceMM = edCellHorizSpaceMM.Value()
				},
				Value:    c.Table.CellHorizSpaceMM,
				Decimals: 2,
			},

			Label{Text: "Размер шрифта таблицы, пиксели"},
			NumberEdit{
				AssignTo: &edFontSizePixels,
				OnValueChanged: func() {
					c.Table.FontSizePixels = edFontSizePixels.Value()
				},
				Value:    c.Table.FontSizePixels,
				Decimals: 2,
			},

			PushButton{
				Text: "Применить",
				OnClicked: func() {
					dlg.Accept()
				},
			},
			PushButton{
				Text: "Отмена",
				OnClicked: func() {
					dlg.Cancel()
				},
			},
		},
	}.Run(appWindow)
	must.PanicIf(err)
	if r != walk.DlgCmdOK {
		return
	}
	if err := cfg.Set(c); err != nil {
		panic(err)
	}
}

func runCurrentMeasurementNameDialog() {
	measurementID, err := getSelectedMeasurementID()
	if err != nil {
		return
	}
	var m data.Measurement
	m.MeasurementID = measurementID
	if err := data.GetMeasurement(db, &m); err != nil {
		return
	}

	var (
		ed  *walk.LineEdit
		dlg *walk.Dialog
	)

	r, err := Dialog{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		MinSize:  Size{280, 180},
		MaxSize:  Size{280, 180},
		AssignTo: &dlg,
		Title:    "Ввод комментария обмера",
		Layout:   VBox{},
		Children: []Widget{
			LineEdit{
				AssignTo: &ed,
				OnTextChanged: func() {
					m.Name = ed.Text()
				},
				Text: m.Name,
			},
			PushButton{
				Text: "Применить",
				OnClicked: func() {
					dlg.Accept()
				},
			},
			PushButton{
				Text: "Отмена",
				OnClicked: func() {
					dlg.Cancel()
				},
			},
		},
	}.Run(appWindow)
	must.PanicIf(err)
	if r != walk.DlgCmdOK {
		return
	}
	must.PanicIf(data.SaveMeasurement(db, &m))

	xs := comboboxMeasurements.Model().([]string)
	n := comboboxMeasurements.CurrentIndex()
	if n < 0 || n >= len(xs) {
		return
	}
	xs[n] = formatMeasureInfo(m.MeasurementInfo)
	handleComboboxMeasurements = false
	must.PanicIf(comboboxMeasurements.SetModel(xs))
	must.PanicIf(comboboxMeasurements.SetCurrentIndex(n))
	handleComboboxMeasurements = true

}

package app

import (
	"context"
	"fmt"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/comports"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view/viewarch"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	lua "github.com/yuin/gopher-lua"
	"io/ioutil"
	luar "layeh.com/gopher-luar"
	"sync"
	"time"
)

func executeConsole() {
	L := lua.NewState()
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	L.SetContext(ctx)
	defer func() {
		cancel()
		wg.Wait()
		comports.CloseAllComports()
		L.Close()
	}()
	L.SetGlobal("go", luar.New(L, &luaConsole{L: L}))

	const helpStr = "go:Gas(1)\r\n" +
		"go:SetTension(10)\r\n" +
		"go:SetCurrent(0.05)\r\n" +
		"go:SetConnection(0xFFFF)\r\n" +
		"go:SetConnectionB('1111111111111111')\r\n" +
		"go:ReadGasConsumption()\r\n"

	var (
		edCmd    *walk.TextEdit
		edStatus *walk.LineEdit
		edHelp   *walk.TextEdit
		pb       *walk.PushButton
		dlg      *walk.Dialog
	)

	setStatus := func(ok bool, text string) {
		var color walk.Color
		if ok {
			color = walk.RGB(0, 0, 128)
		} else {
			color = walk.RGB(255, 0, 0)
		}
		text = fmt.Sprintf("%s %s", time.Now().Format("15:04:05"), text)
		edStatus.SetTextColor(color)
		must.PanicIf(edStatus.SetText(text))
	}

	srcByes, _ := ioutil.ReadFile("console.lua")

	Dlg := Dialog{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		AssignTo: &dlg,
		Title:    "Ввод команд",
		Layout:   VBox{},
		MinSize:  Size{Width: 700, Height: 400},
		MaxSize:  Size{Width: 700, Height: 400},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					TextEdit{
						AssignTo: &edCmd,
						Text:     string(srcByes),
					},
					TextEdit{
						AssignTo: &edHelp,
						Text:     helpStr,
						ReadOnly: true,
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					LineEdit{
						AssignTo:   &edStatus,
						ReadOnly:   true,
						TextColor:  walk.RGB(255, 0, 0),
						Background: SolidColorBrush{Color: walk.RGB(233, 233, 233)},
					},
					PushButton{
						Text:     "Выполнить",
						AssignTo: &pb,
						MaxSize:  Size{Width: 90},
						MinSize:  Size{Width: 90},
						OnClicked: func() {
							setStatus(true, "выполняется...")
							pb.SetEnabled(false)
							L.SetContext(ctx)
							go func() {
								s := edCmd.Text()
								must.PanicIf(ioutil.WriteFile("console.lua", []byte(s), 0666))
								err := L.DoString(s)
								appWindow.Synchronize(func() {
									if ctx.Err() != nil {
										return
									}
									if err == nil {
										setStatus(true, "OK")
									} else {
										setStatus(false, err.Error())
									}
									pb.SetEnabled(true)
								})
							}()
						},
					},
				},
			},
		},
	}
	must.PanicIf(Dlg.Create(appWindow))
	dlg.Run()
}

func executeDialogFloat1(value float64, title string, caption string, decimals int, min, max float64) (float64, bool) {
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
	)
	nDevice := -1
	nKind := -1

	var m data.Measurement
	_ = data.GetLastMeasurement(db, &m)
	m.Name = ""
	if len(m.Pgs) < 4 {
		m.Pgs = make([]float64, 4)
	}

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
							pbOk.SetEnabled(nDevice >= 0 && nKind >= 0)
						},
					},
					ComboBox{
						AssignTo:     &cbKind,
						CurrentIndex: nKind,
						Model:        kinds,
						OnCurrentIndexChanged: func() {
							m.Kind = cbKind.Text()
							pbOk.SetEnabled(nDevice >= 0 && nKind >= 0)

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
						Enabled: nDevice >= 0 && nKind >= 0,
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
		edFontSizePixels,
		edCellHorizSpaceMM,
		edRowHeightMM *walk.NumberEdit
		cbIncSamps *walk.CheckBox
		dlg        *walk.Dialog
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

			CheckBox{
				Text:     "Включить в отчёт измеренные напряжения",
				AssignTo: &cbIncSamps,
				Checked:  c.Table.IncludeSamples,
				OnCheckedChanged: func() {
					c.Table.IncludeSamples = cbIncSamps.Checked()
				},
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

func runFilterArchiveDialog() (int, time.Month, int, bool) {
	var (
		dlg       *walk.Dialog
		tv        *walk.TreeView
		year, day int
		month     time.Month
	)
	var arch []data.MeasurementInfo
	must.PanicIf(data.ListArchive(db, &arch, 0))

	var tmArch []time.Time
	for _, x := range arch {
		tmArch = append(tmArch, x.CreatedAt)
	}

	r, err := Dialog{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		MinSize:  Size{280, 180},
		MaxSize:  Size{280, 180},
		AssignTo: &dlg,
		Title:    "Выбрать дату",
		Layout:   VBox{},
		Children: []Widget{

			TreeView{
				Model:    viewarch.NewTreeViewModel(tmArch),
				AssignTo: &tv,
				MinSize:  Size{Height: 300},
				OnItemActivated: func() {
					var ok bool
					year, month, day, ok = viewarch.GetItemDate(tv.CurrentItem())
					if ok {
						dlg.Accept()
					}
				},
			},
		},
	}.Run(appWindow)
	must.PanicIf(err)
	return year, month, day, r == walk.DlgCmdOK
}

func errorDialog(errStr string) {
	var (
		dlg *walk.Dialog
		pb  *walk.PushButton
	)

	Dlg := Dialog{
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		AssignTo: &dlg,
		Title:    "Произошла ошибка",
		Layout:   HBox{},
		MinSize:  Size{Width: 600, Height: 300},
		MaxSize:  Size{Width: 600, Height: 300},

		CancelButton:  &pb,
		DefaultButton: &pb,

		Children: []Widget{

			TextEdit{
				TextColor: walk.RGB(255, 0, 0),
				ReadOnly:  true,
				Text:      errStr,
			},
			ScrollView{
				Layout:          VBox{},
				HorizontalFixed: true,
				Children: []Widget{
					PushButton{
						AssignTo: &pb,
						Text:     "Продолжить",
						OnClicked: func() {
							dlg.Accept()
						},
					},
					ImageView{
						Image: "assets/img/error_80.png",
					},
				},
			},
		},
	}
	must.PanicIf(Dlg.Create(appWindow))
	must.PanicIf(pb.SetFocus())
	dlg.Run()
}

package app

import (
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/comports"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view/viewarch"
	"github.com/fpawel/sensel/internal/view/viewmeasure"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sync"
)

func newApplicationWindow() MainWindow {

	updCfg := func(f func(c *cfg.Config)) {
		c := cfg.Get()
		f(&c)
		must.PanicIf(cfg.Set(c))
	}

	bgWhite := SolidColorBrush{Color: walk.RGB(255, 255, 255)}

	const widthStatusLabelCaption = 140

	return MainWindow{
		AssignTo:   &appWindow,
		Title:      "ЧЭ лаборатория 74",
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 12,
		},
		Layout: VBox{
			Alignment: AlignHFarVNear,
		},
		MenuItems: []MenuItem{
			Action{
				AssignTo:    &menuRunInterrogate,
				Text:        "Опрос",
				OnTriggered: runReadSample,
			},
			Action{
				AssignTo:    &menuRunMeasure,
				Text:        "Обмер",
				OnTriggered: runMeasure,
			},
			Action{
				AssignTo: &menuStop,
				Text:     "Прервать",
				Visible:  false,
				OnTriggered: func() {
					interruptWorkFunc()
				},
			},
			Action{
				Text:     "Журнал",
				AssignTo: &menuJournal,
				OnTriggered: func() {
					groupBoxJournal.SetVisible(!groupBoxJournal.Visible())
				},
			},
		},
		Children: []Widget{

			ScrollView{
				AssignTo:      &scrollViewMeasurement,
				VerticalFixed: true,
				MaxSize:       Size{Height: 40, Width: 0},
				MinSize:       Size{Height: 40, Width: 0},
				Layout: HBox{
					Alignment:   AlignHCenterVCenter,
					MarginsZero: true,
				},
				Children: []Widget{
					Label{Text: "Исполнение"},
					ComboBox{
						AssignTo: &comboBoxDevice,
						Editable: true,
						MaxSize:  Size{100, 0},
						OnCurrentIndexChanged: func() {
							m := Calc.ListKinds(comboBoxDevice.Text())
							must.PanicIf(comboBoxKind.SetModel(m))
							if len(m) > 0 {
								must.PanicIf(comboBoxKind.SetCurrentIndex(0))
							}
						},
					},
					ComboBox{
						AssignTo: &comboBoxKind,
						Editable: true,
						MaxSize:  Size{150, 0},
					},

					Label{Text: "ПГС1"},
					NumberEdit{
						Decimals: 2,
						AssignTo: &numberEditC[0],
						MaxSize:  Size{90, 0},
						MinSize:  Size{90, 0},
					},

					Label{Text: "ПГС2"},
					NumberEdit{
						Decimals: 2,
						AssignTo: &numberEditC[1],
						MaxSize:  Size{90, 0},
						MinSize:  Size{90, 0},
					},

					Label{Text: "ПГС3"},
					NumberEdit{
						Decimals: 2,
						AssignTo: &numberEditC[2],
						MaxSize:  Size{90, 0},
						MinSize:  Size{90, 0},
					},

					Label{Text: "ПГС4"},
					NumberEdit{
						Decimals: 2,
						AssignTo: &numberEditC[3],
						MaxSize:  Size{90, 0},
						MinSize:  Size{90, 0},
					},

					Label{Text: "Наименование обмера"},
					LineEdit{
						AssignTo: &lineEditMeasureName,
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					GroupBox{
						AssignTo: &groupBoxJournal,
						Visible:  false,
						Title:    "Журнал",
						Layout:   Grid{MarginsZero: true, SpacingZero: true},
						Children: []Widget{
							TableView{
								AssignTo:                 &tableViewArch,
								MaxSize:                  Size{500, 0},
								MinSize:                  Size{500, 0},
								ColumnsOrderable:         false,
								ColumnsSizable:           true,
								LastColumnStretched:      true,
								MultiSelection:           true,
								NotSortableByHeaderClick: true,
								Model:                    new(viewarch.TableViewModel),
								Columns: []TableViewColumn{
									{
										Title: "№",
										Width: 40,
									},
									{
										Title: "Год",
										Width: 60,
									},
									{
										Title: "Месяц",
										Width: 60,
									},
									{
										Title: "День",
										Width: 60,
									},
									{
										Title: "Время",
										Width: 60,
									},
									{
										Title: "Исполнение",
									},
									{
										Title: "Наименование",
									},
								},
								OnCurrentIndexChanged: func() {
									archive := getArchiveTableViewModel().ViewData()
									n := tableViewArch.CurrentIndex()
									if n < 0 || n >= len(archive) {
										return
									}
									var m data.Measurement
									m.MeasurementID = archive[n].MeasurementID
									must.PanicIf(data.GetMeasurement(db, &m))
									setMeasurement(m)
								},
							},
						},
					},
					Composite{
						Layout: VBox{
							MarginsZero: true,
							//SpacingZero: true,
						},
						Children: []Widget{
							TableView{
								AssignTo:                 &tableViewMeasure,
								ColumnsOrderable:         false,
								ColumnsSizable:           true,
								LastColumnStretched:      false,
								MultiSelection:           true,
								NotSortableByHeaderClick: true,
							},
							LineEdit{
								ReadOnly:   true,
								Background: bgWhite,
								AssignTo:   &labelCalcErr,
								TextColor:  walk.RGB(255, 0, 0),
							},

							Composite{
								Layout: HBox{
									MarginsZero: true,
								},
								Children: []Widget{
									Label{
										Text:          "Статус выполнения",
										TextAlignment: AlignFar,
										TextColor:     walk.RGB(0, 0, 128),
										MinSize:       Size{Width: 180},
									},
									LineEdit{
										AssignTo:   &labelWorkStatus,
										ReadOnly:   true,
										Background: bgWhite,
										TextColor:  walk.RGB(0, 0, 128),
									},
									Label{
										TextAlignment: AlignFar,
										Text:          "Плата управления",
										TextColor:     walk.RGB(0, 0, 128),
										MinSize:       Size{Width: widthStatusLabelCaption},
									},
									ComboBoxComport(func() string {
										return cfg.Get().ControlSheet.Comport
									}, func(s string) {
										updCfg(func(c *cfg.Config) {
											c.ControlSheet.Comport = s
										})
									}),
									LineEdit{
										ReadOnly:   true,
										Background: bgWhite,
										AssignTo:   &labelControlSheet,
										TextColor:  walk.RGB(0, 0, 128),
									},
								},
							},
							Composite{
								Layout: HBox{
									MarginsZero: true,
								},
								Children: []Widget{
									Label{
										TextAlignment: AlignFar,
										Text:          "Вольтметр",
										TextColor:     walk.RGB(0, 0, 128),
									},
									ComboBoxComport(func() string {
										return cfg.Get().Voltmeter.Comport
									}, func(s string) {
										updCfg(func(c *cfg.Config) {
											c.Voltmeter.Comport = s
										})
									}),
									LineEdit{
										Background: bgWhite,
										ReadOnly:   true,
										AssignTo:   &labelVoltmeter,
										TextColor:  walk.RGB(0, 0, 128),
									},
									Label{
										TextAlignment: AlignFar,
										Text:          "Газовый блок",
										TextColor:     walk.RGB(0, 0, 128),
										MinSize:       Size{Width: widthStatusLabelCaption},
									},
									ComboBoxComport(func() string {
										return cfg.Get().Gas.Comport
									}, func(s string) {
										updCfg(func(c *cfg.Config) {
											c.Gas.Comport = s
										})
									}),
									LineEdit{
										Background: bgWhite,
										ReadOnly:   true,
										AssignTo:   &labelGasBlock,
										TextColor:  walk.RGB(0, 0, 128),
									},
								},
							},

							Composite{
								Layout: HBox{
									MarginsZero: true,
								},
								Children: []Widget{
									LineEdit{
										TextAlignment: AlignFar,
										ReadOnly:      true,
										Background:    bgWhite,
										AssignTo:      &labelCurrentDelay,
										TextColor:     walk.RGB(0, 0, 128),
									},
									ProgressBar{
										AssignTo:      &progressBarCurrentWork,
										MaxValue:      100,
										StretchFactor: 2,
									},
								},
							},

							Composite{
								Layout: HBox{
									MarginsZero: true,
								},
								Children: []Widget{
									LineEdit{
										TextAlignment: AlignFar,
										ReadOnly:      true,
										Background:    bgWhite,
										AssignTo:      &labelTotalDelay,
										TextColor:     walk.RGB(0, 0, 128),
										Text:          "Общий прогресс выполнения",
									},
									ProgressBar{
										AssignTo:      &progressBarTotalWork,
										MaxValue:      100,
										StretchFactor: 2,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func runWork(work func(ctx context.Context) error) {

	setStatusOk(labelWorkStatus, "")
	setStatusOk(labelVoltmeter, "")
	setStatusOk(labelGasBlock, "")
	setStatusOk(labelControlSheet, "")

	setupWidgets := func(run bool) {
		must.PanicIf(menuStop.SetVisible(run))
		must.PanicIf(menuRunMeasure.SetVisible(!run))
		must.PanicIf(menuRunInterrogate.SetVisible(!run))
		must.PanicIf(menuJournal.SetVisible(!run))
		if run {
			groupBoxJournal.SetVisible(false)
		}
		scrollViewMeasurement.SetEnabled(!run)
	}

	setupWidgets(true)
	var ctx context.Context
	ctx, interruptWorkFunc = context.WithCancel(appCtx)
	wgWork.Add(1)
	go func() {
		err := work(ctx)
		interruptWorkFunc()
		wgWork.Done()
		comports.CloseAllComports()
		appWindow.Synchronize(func() {
			setupWidgets(false)
			if err == nil || merry.Is(err, context.Canceled) {
				setStatusOk(labelWorkStatus, "выполнение окончено успешно")
				return
			}
			setStatusErr(labelWorkStatus, err)
			walk.MsgBox(appWindow, "Произошла ошибка", err.Error(), walk.MsgBoxIconError)
		})
	}()
}

func getMeasureTableViewModel() *viewmeasure.TableViewModel {
	return tableViewMeasure.Model().(*viewmeasure.TableViewModel)
}

func getArchiveTableViewModel() *viewarch.TableViewModel {
	return tableViewArch.Model().(*viewarch.TableViewModel)
}

func setMeasurement(m data.Measurement) {

	must.PanicIf(comboBoxDevice.SetModel(Calc.ListDevices()))
	must.PanicIf(comboBoxKind.SetModel(Calc.ListKinds(m.Device)))

	must.PanicIf(appWindow.SetTitle(fmt.Sprintf("Обмер №%d %s",
		m.MeasurementID, m.CreatedAt.Format("02.01.06 15:04"))))

	must.PanicIf(comboBoxDevice.SetText(m.Device))
	must.PanicIf(comboBoxKind.SetText(m.Kind))
	must.PanicIf(lineEditMeasureName.SetText(m.Name))

	for i := 0; i < 4; i++ {
		var value float64
		if i < len(m.Pgs) {
			value = m.Pgs[i]
		}
		must.PanicIf(numberEditC[i].SetValue(value))
	}
	calcCols, err := Calc.CalculateMeasure(m)
	if err != nil {
		must.PanicIf(labelCalcErr.SetText(err.Error()))
		labelCalcErr.SetVisible(true)
	} else {
		labelCalcErr.SetVisible(false)
	}
	getMeasureTableViewModel().SetViewData(m, calcCols)
}

var (
	menuJournal,
	menuStop,
	menuRunMeasure,
	menuRunInterrogate *walk.Action
	tableViewMeasure *walk.TableView
	tableViewArch    *walk.TableView
	labelCalcErr     *walk.LineEdit

	scrollViewMeasurement *walk.ScrollView

	labelWorkStatus,
	labelControlSheet,
	labelVoltmeter,
	labelGasBlock *walk.LineEdit

	labelCurrentDelay      *walk.LineEdit
	labelTotalDelay        *walk.LineEdit
	progressBarCurrentWork *walk.ProgressBar
	progressBarTotalWork   *walk.ProgressBar

	lineEditMeasureName *walk.LineEdit
	groupBoxJournal     *walk.GroupBox
	comboBoxDevice      *walk.ComboBox
	comboBoxKind        *walk.ComboBox

	numberEditC [4]*walk.NumberEdit

	appWindow *walk.MainWindow

	wgWork            sync.WaitGroup
	interruptWorkFunc = func() {}
)

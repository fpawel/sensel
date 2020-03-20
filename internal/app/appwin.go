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
				AssignTo: &menuRunMeasure,
				Text:     "Обмер",
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
				Text: "Журнал",
				OnTriggered: func() {
					groupBoxJournal.SetVisible(!groupBoxJournal.Visible())
				},
			},
		},
		Children: []Widget{

			//ScrollView{
			//	MaxSize:       Size{Height: 30, Width: 0},
			//	MinSize:       Size{Height: 30, Width: 0},
			//	VerticalFixed: true,
			//	Layout: HBox{
			//		Alignment: AlignHCenterVCenter,
			//		MarginsZero: true,
			//	},
			//	Children: []Widget{},
			//},

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
							SpacingZero: true,
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

							Label{
								AssignTo:  &labelCalcErr,
								TextColor: walk.RGB(255, 0, 0),
							},
						},
					},
					ScrollView{
						MaxSize:         Size{Height: 0, Width: 220},
						MinSize:         Size{Height: 0, Width: 220},
						HorizontalFixed: true,
						Layout: VBox{
							Alignment: AlignHCenterVCenter,
							//SpacingZero:true,
							MarginsZero: true,
						},
						Children: []Widget{

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

							Label{Text: "СОМ порт платы управления"},
							ComboBoxComport(func() string {
								return cfg.Get().Control.Comport
							}, func(s string) {
								updCfg(func(c *cfg.Config) {
									c.Control.Comport = s
								})
							}),

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
								MaxSize:  Size{40, 0},
							},

							Label{Text: "ПГС2"},
							NumberEdit{
								Decimals: 2,
								AssignTo: &numberEditC[1],
								MaxSize:  Size{40, 0},
							},

							Label{Text: "ПГС3"},
							NumberEdit{
								Decimals: 2,
								AssignTo: &numberEditC[2],
								MaxSize:  Size{40, 0},
							},

							Label{Text: "ПГС4"},
							NumberEdit{
								Decimals: 2,
								AssignTo: &numberEditC[3],
								MaxSize:  Size{40, 0},
							},

							Label{Text: "Наименование обмера"},
							LineEdit{
								AssignTo: &lineEditMeasureName,
							},
						},
					},
				},
			},
		},
	}
}

func runWork(work func(ctx context.Context) error) {

	setupWidgets := func(run bool) {
		must.PanicIf(menuStop.SetVisible(run))
		must.PanicIf(menuRunMeasure.SetVisible(!run))
		must.PanicIf(menuRunInterrogate.SetVisible(!run))
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
				return
			}
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
	menuStop,
	menuRunMeasure,
	menuRunInterrogate *walk.Action
	tableViewMeasure    *walk.TableView
	tableViewArch       *walk.TableView
	labelCalcErr        *walk.Label
	lineEditMeasureName *walk.LineEdit
	groupBoxJournal     *walk.GroupBox
	comboBoxDevice      *walk.ComboBox
	comboBoxKind        *walk.ComboBox

	numberEditC [4]*walk.NumberEdit

	appWindow *walk.MainWindow

	wgWork            sync.WaitGroup
	interruptWorkFunc = func() {}
)

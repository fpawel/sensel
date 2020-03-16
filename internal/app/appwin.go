package app

import (
	"context"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"sync"
)

func newApplicationWindow() MainWindow {
	return MainWindow{
		AssignTo:   &appWindow,
		Title:      "ЧЭ лаборатория 74",
		Background: SolidColorBrush{Color: walk.RGB(255, 255, 255)},
		Font: Font{
			Family:    "Segoe UI",
			PointSize: 10,
		},
		Layout: VBox{
			Alignment: AlignHFarVNear,
		},
		MenuItems: []MenuItem{
			Action{
				AssignTo: &menuRunInterrogate,
				Text:     "Опрос",
				OnTriggered: func() {
					runWork(func(ctx context.Context) error {
						<-ctx.Done()
						return nil
					})
				},
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
				Text: "Настройки",
				OnTriggered: func() {
					_, err := DialogAppConfig().Run(appWindow)
					must.PanicIf(err)
				},
			},
			Menu{
				Text: "Конфигурация",
			},
			Menu{
				Text: "Сценарий",
			},
		},
		Children: []Widget{
			ScrollView{
				MaxSize:       Size{Height: 50, Width: 0},
				MinSize:       Size{Height: 50, Width: 0},
				VerticalFixed: true,
				Layout:        HBox{Alignment: AlignHCenterVCenter},
				Children: []Widget{

					CheckBox{
						Text:    "Журнал",
						Checked: false,
						OnCheckedChanged: func() {
							groupBoxJournal.SetVisible(!groupBoxJournal.Visible())
						},
					},

					RadioButton{
						AssignTo: &radioButtonCalc,
						Text:     "Расчёт",
						Value:    true,
						OnClicked: func() {
							getMainTableViewModel().SetShowCalc(true)
						},
					},
					RadioButton{
						Text:  "Снятие",
						Value: false,
						OnClicked: func() {
							getMainTableViewModel().SetShowCalc(false)
						},
					},
				},
			},
			Label{
				AssignTo:  &labelCalcErr,
				TextColor: walk.RGB(255, 0, 0),
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					GroupBox{
						AssignTo: &groupBoxJournal,
						Visible:  false,
						Title:    "Журнал",
						MinSize:  Size{400, 0},
						MaxSize:  Size{400, 0},
						Layout:   Grid{},
						Children: []Widget{
							TableView{
								ColumnsOrderable:         false,
								ColumnsSizable:           true,
								LastColumnStretched:      false,
								MultiSelection:           true,
								NotSortableByHeaderClick: true,
							},
						},
					},

					GroupBox{
						Title:  "Обмер",
						Layout: Grid{},
						Children: []Widget{
							TableView{
								AssignTo:                 &mainTableView,
								ColumnsOrderable:         false,
								ColumnsSizable:           true,
								LastColumnStretched:      false,
								MultiSelection:           true,
								NotSortableByHeaderClick: true,
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
		_ = work(ctx)
		interruptWorkFunc()
		wgWork.Done()
		appWindow.Synchronize(func() {
			setupWidgets(false)
		})
	}()
}

var (
	menuStop,
	menuRunMeasure,
	menuRunInterrogate *walk.Action
	radioButtonCalc *walk.RadioButton
	mainTableView   *walk.TableView
	labelCalcErr    *walk.Label
	groupBoxJournal *walk.GroupBox

	appWindow *walk.MainWindow

	wgWork            sync.WaitGroup
	interruptWorkFunc = func() {}
)

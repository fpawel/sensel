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
					RadioButton{
						Text: "Измерено",
						OnClicked: func() {
							measurementViewModel.SetShowCalc(false)
						},
					},
					RadioButton{
						Text: "Расчитано",
						OnClicked: func() {
							measurementViewModel.SetShowCalc(true)
						},
					},
				},
			},
			TableView{
				Columns:                  measurementViewModel.Columns(),
				Model:                    measurementViewModel,
				ColumnsOrderable:         false,
				ColumnsSizable:           true,
				LastColumnStretched:      false,
				MultiSelection:           true,
				NotSortableByHeaderClick: true,
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
	menuStop, menuRunMeasure, menuRunInterrogate *walk.Action
	wgWork                                       sync.WaitGroup
	interruptWorkFunc                            = func() {}
)

package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pdf"
	"github.com/fpawel/sensel/internal/pkg/comports"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view/viewmeasure"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"strconv"
	"strings"
	"sync"
)

func newApplicationWindow() MainWindow {

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
				AssignTo: &menuStop,
				Text:     "Прервать",
				Visible:  false,
				OnTriggered: func() {
					interruptWorkFunc()
				},
			},

			Action{
				AssignTo:    &menuRunMeasure,
				Text:        "Обмер",
				OnTriggered: startMeasure,
			},
			Menu{
				AssignTo: &menuControl,
				Text:     "Управление",
				Items: []MenuItem{
					Action{
						Text:        "Проверка связи",
						OnTriggered: runCheckConnection,
					},
					Action{
						Text:        "Опрос вольтметра",
						OnTriggered: runReadVoltmeter,
					},
					Action{
						Text:        "Поиск обрыва",
						OnTriggered: runSearchBreak,
					},
					Action{
						Text:        "Ввод команд",
						OnTriggered: executeConsole,
					},
				},
			},
			Action{
				Text:        "Отчёт",
				OnTriggered: newReport,
			},
			Action{
				Text:        "Настройка",
				OnTriggered: runAppSettingsDialog,
			},
		},
		Children: []Widget{

			Composite{
				Layout: VBox{
					MarginsZero: true,
					//SpacingZero: true,
				},
				Children: []Widget{
					ScrollView{
						VerticalFixed: true,
						MaxSize:       Size{Height: 50, Width: 0},
						MinSize:       Size{Height: 50, Width: 0},
						Layout: HBox{
							Alignment:   AlignHCenterVCenter,
							MarginsZero: true,
						},
						Children: []Widget{
							Label{
								Text: "Измерение:",
							},
							ComboBox{
								AssignTo:              &comboboxMeasurements,
								OnCurrentIndexChanged: comboboxMeasurementsCurrentIndexChanged,
								ContextMenuItems: []MenuItem{
									Action{
										Text:        "Коментарий...",
										OnTriggered: runCurrentMeasurementNameDialog,
									},
									Action{
										Text:        "Удалить",
										OnTriggered: deleteCurrentMeasurement,
									},
									Action{
										AssignTo:    &actionArchiveFilterLast,
										OnTriggered: executeDialogFilterLastMeasurement,
									},
									Action{
										AssignTo: &actionArchiveFilterData,
										Text:     "Фильтр: дата...",
										OnTriggered: func() {
											setupArchiveFilterData()
										},
									},
								},
							},
						},
					},
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
								MinSize:       Size{Width: 182},
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
							cbComports[nCbControlSheet].Combobox(),
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
							cbComports[nCbVoltmeter].Combobox(),
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
							cbComports[nCbGas].Combobox(),
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
	}
}

func executeDialogFilterLastMeasurement() {
	c := cfg.Get()
	v, ok := executeDialogFloat1(float64(c.LastMeasurementsCount), "Фильтр", "Количество измерений", 0, 0, 0)
	if !ok {
		return
	}
	if v < 10 {
		v = 10
	}
	c.LastMeasurementsCount = int(v)
	must.PanicIf(cfg.Set(c))
	setupArchiveFilterLastMeasurements(c.LastMeasurementsCount)
}

func setupArchiveFilterData() {
	y, m, d, ok := runFilterArchiveDialog()
	if !ok {
		return
	}
	var arch []data.MeasurementInfo
	must.PanicIf(data.ListArchiveDay(db, &arch, y, m, d))
	setupArchiveComboBox(arch)
	must.PanicIf(actionArchiveFilterLast.SetText("Фильтр: последние измерения"))
	must.PanicIf(actionArchiveFilterData.SetText(fmt.Sprintf("Фильтр: %02d.%02d.%d", d, m, y)))
	must.PanicIf(actionArchiveFilterLast.SetChecked(false))
	must.PanicIf(actionArchiveFilterData.SetChecked(true))
}

func setupArchiveFilterLastMeasurements(count int) {
	if count < 10 {
		count = 10
	}
	var arch []data.MeasurementInfo
	must.PanicIf(data.ListArchive(db, &arch, count))
	setupArchiveComboBox(arch)
	must.PanicIf(actionArchiveFilterLast.SetText(fmt.Sprintf("Фильтр: последние %d измерений", count)))
	must.PanicIf(actionArchiveFilterData.SetText("Фильтр: дата"))
	must.PanicIf(actionArchiveFilterLast.SetChecked(true))
	must.PanicIf(actionArchiveFilterData.SetChecked(false))
}

func setupArchiveComboBox(arch []data.MeasurementInfo) {
	handleComboboxMeasurements = false
	defer func() {
		handleComboboxMeasurements = true
	}()
	var cbm []string
	for _, x := range arch {
		cbm = append(cbm, formatMeasureInfo(x))
	}
	must.PanicIf(comboboxMeasurements.SetModel(cbm))
	if len(arch) == 0 {
		setSampleView(data.Sample{})
		return
	}
	must.PanicIf(comboboxMeasurements.SetCurrentIndex(0))
	m := data.Measurement{
		MeasurementInfo: arch[0],
	}
	must.PanicIf(data.GetMeasurement(db, &m))
	setMeasurementView(m)
}

func comboboxMeasurementsCurrentIndexChanged() {
	if !handleComboboxMeasurements {
		return
	}
	measurementID, err := getSelectedMeasurementID()
	if err != nil {
		setMeasurementView(data.Measurement{})
		return
	}

	var m data.Measurement
	m.MeasurementID = measurementID
	must.PanicIf(data.GetMeasurement(db, &m))
	setMeasurementView(m)
}

func newReport() {
	if err := func() error {
		measurementID, err := getSelectedMeasurementID()
		if err != nil {
			return err
		}
		var m data.Measurement
		m.MeasurementID = measurementID
		if err := data.GetMeasurement(db, &m); err != nil {
			return err
		}
		calcCols, err := Calc.CalculateMeasure(m)
		if err != nil {
			return err
		}
		c := cfg.Get().Table
		if err := pdf.New(m, calcCols, pdf.TableConfig{
			RowHeight:      c.RowHeightMM,
			CellHorizSpace: c.CellHorizSpaceMM,
			FontSize:       c.FontSizePixels,
		}, c.IncludeSamples); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		errorDialog(err.Error())
	}
}

func deleteCurrentMeasurement() {
	measurementID, err := getSelectedMeasurementID()
	if err != nil {
		return
	}
	db.MustExec(`DELETE FROM measurement WHERE measurement_id=?`, measurementID)
	setupArchiveFilterLastMeasurements(50)
}

func getSelectedMeasurementID() (int64, error) {
	s := comboboxMeasurements.Text()
	xs := strings.Fields(s)
	if len(xs) == 0 {
		return 0, errors.New("not a measurement id")
	}
	return strconv.ParseInt(xs[0], 10, 64)
}

func runWork(work func(ctx context.Context) error) {
	must.PanicIf(labelWorkStatus.SetText(""))
	must.PanicIf(labelVoltmeter.SetText(""))
	must.PanicIf(labelGasBlock.SetText(""))
	must.PanicIf(labelControlSheet.SetText(""))

	setupWidgets := func(run bool) {
		must.PanicIf(menuRunMeasure.SetVisible(!run))
		must.PanicIf(menuStop.SetVisible(run))
		for i := 0; i < menuControl.Actions().Len(); i++ {
			must.PanicIf(menuControl.Actions().At(i).SetVisible(!run))
		}
	}

	setupWidgets(true)
	var ctx context.Context
	ctx, interruptWorkFunc = context.WithCancel(appCtx)
	wgWork.Add(1)
	go func() {

		defer panicWithSaveRecoveredErrorToFile()

		err := work(ctx)

		interruptWorkFunc()
		wgWork.Done()
		comports.CloseAllComports()
		appWindow.Synchronize(func() {
			setupWidgets(false)
			if err == nil {
				setStatusOkSync(labelWorkStatus, "выполнение окончено успешно")
				return
			}
			if merry.Is(err, context.Canceled) {
				setStatusErrSync(labelWorkStatus, errors.New("выполнение прервано пользователем"))
				return
			}
			setStatusErr(labelWorkStatus, err)
			saveErrorToFile(err.Error())
			errorDialog(err.Error())
		})
	}()
}

func getMeasureTableViewModel() *viewmeasure.TableViewModel {
	return tableViewMeasure.Model().(*viewmeasure.TableViewModel)
}

func startMeasure() {
	m, ok := runDialogMeasurement()
	if !ok {
		return
	}
	must.PanicIf(data.SaveMeasurement(db, &m))
	xs := comboboxMeasurements.Model().([]string)
	xs = append([]string{formatMeasureInfo(m.MeasurementInfo)}, xs...)
	handleComboboxMeasurements = false
	must.PanicIf(comboboxMeasurements.SetModel(xs))
	handleComboboxMeasurements = true
	setMeasurementView(m)
	runMeasure(m)
}

func setSampleView(smp data.Sample) {
	m := data.Measurement{
		MeasurementData: data.MeasurementData{
			Samples: []data.Sample{smp},
		},
	}
	must.PanicIf(appWindow.SetTitle("УКЧЭ"))
	getMeasureTableViewModel().SetViewData(m, nil)
	labelCalcErr.SetVisible(false)
	handleComboboxMeasurements = false
	must.PanicIf(comboboxMeasurements.SetCurrentIndex(-1))
	handleComboboxMeasurements = true
}

func setSampleViewUISafe(smp data.Sample) {
	m := data.Measurement{
		MeasurementData: data.MeasurementData{
			Samples: []data.Sample{smp},
		},
	}
	appWindow.Synchronize(func() {
		must.PanicIf(appWindow.SetTitle("УКЧЭ"))
		getMeasureTableViewModel().SetViewData(m, nil)
		labelCalcErr.SetVisible(false)
		handleComboboxMeasurements = false
		must.PanicIf(comboboxMeasurements.SetCurrentIndex(-1))
		handleComboboxMeasurements = true
	})
}

func setMeasurementViewUISafe(m data.Measurement) {
	appWindow.Synchronize(func() {
		setMeasurementView(m)
	})
}

func setMeasurementView(m data.Measurement) {

	if m.MeasurementID != 0 {
		must.PanicIf(appWindow.SetTitle(fmt.Sprintf("УКЧЭ. Измерение №%d %s",
			m.MeasurementID, m.CreatedAt.Format("02.01.06 15:04"))))
	} else {
		must.PanicIf(appWindow.SetTitle("УКЧЭ"))
	}

	calcCols, err := Calc.CalculateMeasure(m)
	if err != nil {
		must.PanicIf(labelCalcErr.SetText("Расчёт: " + err.Error()))
		labelCalcErr.SetVisible(true)
	} else {
		labelCalcErr.SetVisible(false)
	}
	getMeasureTableViewModel().SetViewData(m, calcCols)

	for i, x := range comboboxMeasurements.Model().([]string) {
		if x == formatMeasureInfo(m.MeasurementInfo) {
			handleComboboxMeasurements = false
			must.PanicIf(comboboxMeasurements.SetCurrentIndex(i))
			handleComboboxMeasurements = true
			return
		}
	}
	handleComboboxMeasurements = false
	must.PanicIf(comboboxMeasurements.SetCurrentIndex(-1))
	handleComboboxMeasurements = true
}

var (
	menuStop,
	menuRunMeasure *walk.Action
	menuControl      *walk.Menu
	tableViewMeasure *walk.TableView
	labelCalcErr     *walk.LineEdit

	actionArchiveFilterLast,
	actionArchiveFilterData *walk.Action

	labelWorkStatus,
	labelControlSheet,
	labelVoltmeter,
	labelGasBlock *walk.LineEdit

	labelCurrentDelay      *walk.LineEdit
	labelTotalDelay        *walk.LineEdit
	progressBarCurrentWork *walk.ProgressBar
	progressBarTotalWork   *walk.ProgressBar

	comboboxMeasurements *walk.ComboBox

	handleComboboxMeasurements bool

	appWindow *walk.MainWindow

	wgWork            sync.WaitGroup
	interruptWorkFunc = func() {}
)

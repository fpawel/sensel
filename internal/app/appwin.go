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
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func newApplicationWindow() MainWindow {

	bgWhite := SolidColorBrush{Color: walk.RGB(255, 255, 255)}

	const widthStatusLabelCaption = 140

	setGas := func(gas int) func() {
		return func() {
			go func() {
				chGas <- gas
			}()
		}
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
		MenuItems: []MenuItem{},
		Children: []Widget{

			Composite{
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					ScrollView{
						AssignTo:      &scrollViewSelectMeasure,
						VerticalFixed: true,
						MaxSize:       Size{Height: 80, Width: 0},
						MinSize:       Size{Height: 80, Width: 0},
						Layout: HBox{
							Alignment:   AlignHCenterVCenter,
							MarginsZero: true,
						},
						Children: []Widget{
							PushButton{
								Image:    "assets/img/cancel25.png",
								AssignTo: &buttonStop,
								Visible:  false,
								Text:     " Прервать",
								MaxSize:  Size{Width: 120},
								OnClicked: func() {
									interruptWorkFunc()
									log.Debug("work interrupted")
								},
							},
							SplitButton{
								AssignTo:  &buttonRun,
								Image:     "assets/img/rs232_25.png",
								Text:      " Обмер",
								MaxSize:   Size{Width: 120},
								OnClicked: startMeasure,
								MenuItems: []MenuItem{
									Action{
										Text:        "Проверить связь",
										OnTriggered: runCheckConnection,
									},
									Action{
										Text:        "Газ",
										OnTriggered: runReadConsumption,
									},
									Action{
										Text:        "Опрос вольтметра",
										OnTriggered: runReadVoltmeter,
									},
									Action{
										Text:        "Опрос установки контроля",
										OnTriggered: runReadSample,
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

							SplitButton{
								Text:    " Отчёт",
								Image:   "assets/img/pdf.png",
								MaxSize: Size{Width: 120},
								MenuItems: []MenuItem{
									Action{
										Text:        "Печать",
										OnTriggered: printCurrentReport,
									},
									Action{
										Text:        "Настройка",
										OnTriggered: runReportSettingsDialog,
									},
								},
								OnClicked: showCurrentReport,
							},

							Label{
								Text: "Измерение:",
							},
							ComboBox{
								MaxSize:               Size{Width: 500},
								MinSize:               Size{Width: 500},
								AssignTo:              &comboboxMeasurements,
								OnCurrentIndexChanged: comboboxMeasurementsCurrentIndexChanged,
								ContextMenuItems: []MenuItem{
									Action{
										Text:        "Комментарий...",
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

					ScrollView{
						AssignTo: &scrollViewCheckConsumption,
						Visible:  false,
						Layout: VBox{
							MarginsZero: true,
						},
						HorizontalFixed: true,
						Children: []Widget{

							Composite{
								Layout: HBox{
									MarginsZero: true,
								},
								Children: []Widget{
									PushButton{
										Text:      "Газ 1",
										OnClicked: setGas(1),
									},
									PushButton{
										Text:      "Газ 2",
										OnClicked: setGas(2),
									},
									PushButton{
										Text:      "Газ 3",
										OnClicked: setGas(3),
									},
									PushButton{
										Text:      "Газ 4",
										OnClicked: setGas(4),
									},
									PushButton{
										Text:      "Выкл.",
										OnClicked: setGas(0),
									},
									PushButton{
										Text: "Закрыть",
										OnClicked: func() {
											interruptWorkFunc()
											log.Debug("work interrupted")
										},
									},
								},
							},

							Composite{
								Layout: HBox{MarginsZero: true},
								Children: []Widget{
									Label{
										Text: "Газ:",
										Font: Font{Family: "Segoe UI", PointSize: 40},
									},
									Label{
										AssignTo:  &labelGas,
										Text:      "1",
										Font:      Font{Family: "Segoe UI", PointSize: 40},
										TextColor: walk.RGB(0, 0, 128),
									},
								},
							},

							Composite{
								Layout: HBox{MarginsZero: true},
								Children: []Widget{
									Label{
										Text: "Расход:",
										Font: Font{Family: "Segoe UI", PointSize: 40},
									},
									Label{
										AssignTo:  &labelConsumption,
										Text:      "0.12",
										Font:      Font{Family: "Segoe UI", PointSize: 40},
										TextColor: walk.RGB(0, 0, 128),
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

func newPdf(m data.Measurement) (string, error) {
	calcCols, err := Calc.CalculateMeasure(m)
	if err != nil {
		return "", err
	}
	c := cfg.Get().Table

	return pdf.NewFile(m, calcCols, pdf.TableConfig{
		RowHeight:      c.RowHeightMM,
		CellHorizSpace: c.CellHorizSpaceMM,
		FontSize:       c.FontSizePixels,
	}, c.IncludeSamples)
}

func withErrorDialog(err error) {
	if err != nil {
		errorDialog(err)
	}
}

func showCurrentReport() {
	withErrorDialog(func() error {
		m, err := getSelectedMeasurement()
		if err != nil {
			return err
		}
		filename, err := newPdf(m)
		if err != nil {
			return err
		}
		return exec.Command("explorer.exe", filename).Start()
	}())
}

func printCurrentReport() {
	withErrorDialog(func() error {
		m, err := getSelectedMeasurement()
		if err != nil {
			return err
		}
		return printMeasurement(m)
	}())
}

func deleteCurrentMeasurement() {
	measurementID, err := getSelectedMeasurementID()
	if err != nil {
		return
	}
	db.MustExec(`DELETE FROM measurement WHERE measurement_id=?`, measurementID)
	setupArchiveFilterLastMeasurements(50)
}

func getSelectedMeasurement() (data.Measurement, error) {
	measurementID, err := getSelectedMeasurementID()
	if err != nil {
		return data.Measurement{}, err
	}
	var m data.Measurement
	m.MeasurementID = measurementID
	if err := data.GetMeasurement(db, &m); err != nil {
		return data.Measurement{}, err
	}
	return m, nil
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
		buttonStop.SetVisible(run)
		buttonRun.SetVisible(!run)
		comboboxMeasurements.SetEnabled(!run)
	}

	setupWidgets(true)
	var ctx context.Context
	ctx, interruptWorkFunc = context.WithCancel(appCtx)
	wgWork.Add(1)
	go func() {

		defer func() {
			panicWithSaveRecoveredErrorToFile()
			log.Debug("work done")
		}()

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
			saveErrorToFile(err)
			errorDialog(err)
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
	buttonStop *walk.PushButton
	buttonRun  *walk.SplitButton

	tableViewMeasure *walk.TableView
	labelCalcErr     *walk.LineEdit

	labelConsumption *walk.Label
	labelGas         *walk.Label

	actionArchiveFilterLast,
	actionArchiveFilterData *walk.Action

	labelWorkStatus,
	labelControlSheet,
	labelVoltmeter,
	labelGasBlock *walk.LineEdit

	scrollViewCheckConsumption, scrollViewSelectMeasure *walk.ScrollView

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

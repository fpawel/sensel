package app

import (
	"context"
	"github.com/fpawel/sensel/internal/calcsens"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
	"time"
)

func Main() {
	defer func() {
		x := recover()
		panicMsgBox(x)
		if x != nil {
			panic(x)
		}
	}()

	exeDir := filepath.Dir(os.Args[0])

	var err error
	prodTypes, err = calcsens.NewProductTypes(filepath.Join(exeDir, "lua", "sensel.lua"))

	// общий контекст приложения с прерыванием
	var interrupt context.CancelFunc
	appCtx, interrupt = context.WithCancel(context.Background())

	// соединение с базой данных
	dbFilename := filepath.Join(exeDir, "sensel.sqlite")
	log.Debug("open database: " + dbFilename)

	db, err = data.Open(dbFilename)
	must.PanicIf(err)

	must.PanicIf(newApplicationWindow().Create())

	// инициализация модели представления
	initMeasurementViewModel()

	radioButtonCalc.SetChecked(true)

	if !win.ShowWindow(appWindow.Handle(), win.SW_SHOWMAXIMIZED) {
		panic("can`t show window")
	}
	appWindow.Run()

	log.Debug("прервать все фоновые горутины")
	interrupt()

	// дождаться завершения выполняемых горутин
	wgWork.Wait()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func getMainTableViewModel() *view.MainTableViewModel {
	return mainTableView.Model().(*view.MainTableViewModel)
}

func initMeasurementViewModel() {

	t := prodTypes.GetFirstProductType()

	samples := make([]data.Sample, len(t.Columns))
	for i, m := range t.Columns {
		samples[i].Name = m.Name
		samples[i].CreatedAt = time.Now().Add(-time.Minute * time.Duration(i))
	}
	data.RandSamples(samples)
	measurement = data.Measurement{
		ProductType: t.Name,
		Samples:     samples,
		Pgs:         []float64{1, 2, 3, 4, 5},
	}
	view.NewMainTableViewModel(mainTableView)
	setMeasurementViewModel(measurement)
}

func setMeasurementViewModel(measurement data.Measurement) {
	calcColumns, t, err := prodTypes.CalcSamples(measurement)
	if err != nil {
		must.PanicIf(labelCalcErr.SetText(err.Error()))
		labelCalcErr.SetVisible(true)
	} else {
		labelCalcErr.SetVisible(false)
	}
	getMainTableViewModel().SetViewData(view.MainTableViewData{
		D:  measurement,
		Cs: calcColumns,
		Pt: t,
	})
}

var (
	log         = structlog.New()
	db          *sqlx.DB
	appCtx      context.Context
	prodTypes   calcsens.C
	measurement data.Measurement
)

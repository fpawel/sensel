package app

import (
	"context"
	"github.com/fpawel/sensel/internal/calcsens"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
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

	// инициализация модели представления
	initMeasurementViewModel()

	must.PanicIf(newApplicationWindow().Create())

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

// initMeasurementViewModel инициализация модели представления
func initMeasurementViewModel() {

	t, _ := prodTypes.GetProductTypeByName(prodTypes.ListProductTypeNames()[0])

	samples := make([]data.Sample, len(t.Columns))
	for i, m := range t.Columns {
		samples[i].Name = m.Name
		samples[i].CreatedAt = time.Now().Add(-time.Minute * time.Duration(i))
	}
	data.RandSamples(samples)
	measurementViewModel = &MeasurementViewModel{
		m: data.Measurement{
			ProductType: t.Name,
			Samples:     samples,
			Pgs:         []float64{1, 2, 3, 4, 5},
		},
	}
	measurementViewModel.c, measurementViewModel.calcErr = prodTypes.CalcSamples(measurementViewModel.m)
	if measurementViewModel.calcErr != nil {
		panic(measurementViewModel.calcErr)
	}
}

var (
	log    = structlog.New()
	db     *sqlx.DB
	appCtx context.Context

	measurementViewModel *MeasurementViewModel
	prodTypes            calcsens.C
)

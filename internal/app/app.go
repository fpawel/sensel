package app

import (
	"context"
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/walk"
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

	initProductTypes()

	var err error

	// общий контекст приложения с прерыванием
	var interrupt context.CancelFunc
	appCtx, interrupt = context.WithCancel(context.Background())

	// соединение с базой данных
	dbFilename := filepath.Join(filepath.Dir(os.Args[0]), "sensel.sqlite")
	log.Debug("open database: " + dbFilename)

	db, err = data.Open(dbFilename)
	must.PanicIf(err)

	// инициализация модели представления
	initMeasurementViewModel()

	must.PanicIf(newApplicationWindow().Create())

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
	samples := make([]data.Sample, 10)
	for i := range samples {
		samples[i].Name = fmt.Sprintf("X%d", i)
		samples[i].CreatedAt = time.Now().Add(-time.Minute * time.Duration(i))
	}
	data.RandSamples(samples)
	measurementViewModel = &MeasurementViewModel{
		M: data.Measurement{
			Samples: samples,
		},
	}
	for s := range ProductTypes {
		measurementViewModel.M.ProductType = s
	}
}

var (
	log                  = structlog.New()
	db                   *sqlx.DB
	appCtx               context.Context
	appWindow            *walk.MainWindow
	measurementViewModel *MeasurementViewModel
)

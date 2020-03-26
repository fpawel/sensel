package app

import (
	"context"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view/viewmeasure"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
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
	//prodTypes, err = calcsens.NewProductTypes(filepath.Join(exeDir, "lua", "sensel.lua"))

	// общий контекст приложения с прерыванием
	var interrupt context.CancelFunc
	appCtx, interrupt = context.WithCancel(context.Background())

	// соединение с базой данных
	dbFilename := filepath.Join(exeDir, "sensel.sqlite")
	log.Debug("open database: " + dbFilename)

	db, err = data.Open(dbFilename)
	must.PanicIf(err)

	Calc, err = calc.New(filepath.Join(exeDir, "lua", "sensel.lua"))
	must.PanicIf(err)

	must.PanicIf(newApplicationWindow().Create())

	// инициализация виджетов
	labelCurrentDelay.SetVisible(false)
	labelTotalDelay.SetVisible(false)
	progressBarCurrentWork.SetVisible(false)
	progressBarTotalWork.SetVisible(false)
	go trackRegChangeComport()

	// инициализация модели представления
	{
		var arch []data.MeasurementInfo
		must.PanicIf(data.ListArchive(db, &arch))

		var cbm []string
		for _, x := range arch {
			cbm = append(cbm, formatMeasureInfo(x))
		}
		must.PanicIf(comboboxMeasure.SetModel(cbm))
	}

	var measurement data.Measurement
	_ = data.GetLastMeasurement(db, &measurement)
	viewmeasure.NewMainTableViewModel(tableViewMeasure)

	setMeasurement(measurement)

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

var (
	log    = structlog.New()
	db     *sqlx.DB
	appCtx context.Context
	Calc   calc.C
)

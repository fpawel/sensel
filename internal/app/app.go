package app

import (
	"context"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/fpawel/sensel/internal/view/viewmeasure"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
)

func Main() {
	defer panicWithSaveRecoveredErrorToFile()

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

	// связывание TableView с моделью
	viewmeasure.NewMainTableViewModel(tableViewMeasure)

	appWindow.Closing().Attach(func(*bool, walk.CloseReason) {
		cs := tableViewMeasure.Columns()
		c := cfg.Get()
		xs := &c.AppWindow.TableViewMeasure.ColumnWidths
		*xs = nil
		for i := 0; i < cs.Len(); i++ {
			*xs = append(*xs, cs.At(i).Width())
		}
		must.PanicIf(cfg.Set(c))
	})

	// инициализация виджетов
	labelCurrentDelay.SetVisible(false)
	labelTotalDelay.SetVisible(false)
	progressBarCurrentWork.SetVisible(false)
	progressBarTotalWork.SetVisible(false)
	go trackRegChangeComport()

	// инициализация модели представления
	setupArchiveFilterLastMeasurements(cfg.Get().LastMeasurementsCount)

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

	chGas = make(chan int)
)

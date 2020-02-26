package app

import (
	"context"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/jmoiron/sqlx"
	"github.com/lxn/walk"
	"github.com/lxn/win"
	"github.com/powerman/structlog"
	"os"
	"path/filepath"
)

func Main() {
	var err error

	// общий контекст приложения с прерыванием
	var interrupt context.CancelFunc
	appCtx, interrupt = context.WithCancel(context.Background())

	// соединение с базой данных
	dbFilename := filepath.Join(filepath.Dir(os.Args[0]), "atool.sqlite")
	log.Debug("open database: " + dbFilename)

	db, err = data.Open(dbFilename)
	must.PanicIf(err)

	must.PanicIf(newApplicationWindow().Create())

	if !win.ShowWindow(appWindow.Handle(), win.SW_SHOWMAXIMIZED) {
		panic("can`t show window")
	}
	appWindow.Run()

	log.Debug("прервать все фоновые горутины")
	interrupt()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

var (
	log                  = structlog.New()
	db                   *sqlx.DB
	appCtx               context.Context
	appWindow            *walk.MainWindow
	measurementViewModel *MeasurementViewModel
)

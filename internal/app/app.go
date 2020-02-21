package app

import (
	"context"
	"fmt"
	"github.com/fpawel/comm"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/gui"
	"github.com/fpawel/sensel/internal/pkg/logfile"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/jmoiron/sqlx"
	"github.com/powerman/structlog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
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

	// журнал СОМ порта
	comportLogfile, err = logfile.New(".comport")
	must.PanicIf(err)

	// инициализация отправки оповещений с посылками СОМ порта в gui
	comm.SetNotify(notifyComm)

	// старт сервера
	//stopApiServer := runApiServer()

	if envVarDevModeSet() {
		log.Printf("waiting system signal because of %s=%s", envVarDevMode, os.Getenv(envVarDevMode))
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-done
		log.Debug("system signal: " + sig.String())
	} else {
		cmd := exec.Command(filepath.Join(filepath.Dir(os.Args[0]), "atoolgui.exe"))
		log.ErrIfFail(cmd.Start)
		log.ErrIfFail(cmd.Wait)
		log.Debug("gui was closed.")
	}

	log.Debug("прервать все фоновые горутины")
	interrupt()
	//guiwork.Interrupt()
	//guiwork.Wait()

	log.Debug("остановка сервера api")
	//stopApiServer()

	log.Debug("закрыть соединение с базой данных")
	log.ErrIfFail(db.Close)

	log.Debug("закрыть журнал СОМ порта")
	log.ErrIfFail(comportLogfile.Close)

	log.Debug("закрыть журнал")
	//log.ErrIfFail(guiwork.CloseJournal)

	// записать в лог что всё хорошо
	log.Debug("all canceled and closed")
}

func notifyComm(x comm.Info) {
	ct := gui.CommTransaction{
		Port:     x.Port,
		Request:  fmt.Sprintf("% X", x.Request),
		Response: fmt.Sprintf("% X", x.Response),
		Ok:       x.Err == nil,
	}
	if x.Err != nil {
		if len(x.Response) > 0 {
			ct.Response += " "
		}
		ct.Response += x.Err.Error()
	}
	ct.Response += " " + x.Duration.String()
	if x.Attempt > 0 {
		ct.Response += fmt.Sprintf(" попытка %d", x.Attempt+1)
	}
	go gui.NotifyNewCommTransaction(ct)

	_, err := fmt.Fprintf(comportLogfile, "%s %s % X -> % X", time.Now().Format("15:04:05.000"), x.Port, x.Request, x.Response)
	must.PanicIf(err)
	if x.Err != nil {
		_, err := fmt.Fprintf(comportLogfile, " %s", x.Err)
		must.PanicIf(err)
	}
	_, err = comportLogfile.WriteString("\n")
	must.PanicIf(err)
}

func envVarDevModeSet() bool {
	return os.Getenv(envVarDevMode) == "true"
}

const (
	envVarApiPort = "SENSEL_API_PORT"
	envVarDevMode = "SENSEL_DEV_MODE"
)

var (
	log            = structlog.New()
	db             *sqlx.DB
	appCtx         context.Context
	comportLogfile *os.File
)

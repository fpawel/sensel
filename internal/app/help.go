package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

func saveErrorToFile(saveErr string) {
	file, err := os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "errors.log"), os.O_CREATE|os.O_APPEND, 0666)
	must.FatalIf(err)
	defer func() {
		must.FatalIf(file.Close())
	}()
	_, err = file.WriteString(time.Now().Format("2006.01.02 15:04:05") + " " + strings.TrimSpace(saveErr) + "\n")
	must.FatalIf(err)
}

func formatMeasureInfo(m data.MeasurementInfo) string {
	return fmt.Sprintf("%4d - %s - %s %s - %q",
		m.MeasurementID, m.CreatedAt.Format("2006.01.02 15:04"), m.Device, m.Kind, m.Name)
}

func panicWithSaveRecoveredErrorToFile() {
	msgBoxErr := func(msg string) {
		dir := filepath.Dir(os.Args[0])
		msg = strings.ReplaceAll(msg, dir+"\\", "")
		walk.MsgBox(nil, "Установка контроля ЧЭ", msg,
			walk.MsgBoxIconError|walk.MsgBoxOK|walk.MsgBoxSystemModal)
	}

	x := recover()
	if x == nil {
		return
	}
	errStr := fmt.Sprintf("panic: %+v", x)
	saveErrorToFile(errStr + ": " + string(debug.Stack()))
	msgBoxErr(errStr)
	panic(x)
}

func setStatusOkSync(label *walk.LineEdit, text string) {
	appWindow.Synchronize(func() {
		setStatusOk(label, text)
	})
}

func setStatusErrSync(label *walk.LineEdit, err error) {
	appWindow.Synchronize(func() {
		setStatusErr(label, err)
	})
}

func setStatusOk(label *walk.LineEdit, text string) {
	setStatusText(label, true, text)
}

func setStatusErr(label *walk.LineEdit, err error) {
	setStatusText(label, false, err.Error())
}

func setStatusText(label *walk.LineEdit, ok bool, text string) {
	if err := label.SetText(time.Now().Format("15:04:05") + " " + text); err != nil {
		return
	}

	var color walk.Color
	if ok {
		color = walk.RGB(0, 0, 128)
	} else {
		color = walk.RGB(255, 0, 0)
	}
	label.SetTextColor(color)
}

func pause(chDone <-chan struct{}, d time.Duration) {
	timer := time.NewTimer(d)
	for {
		select {
		case <-timer.C:
			return
		case <-chDone:
			timer.Stop()
			return
		}
	}
}

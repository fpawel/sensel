package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func saveErrorToFile(saveErr error) {
	file, err := os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "errors.log"), os.O_CREATE|os.O_APPEND, 0666)
	must.FatalIf(err)
	defer func() {
		must.FatalIf(file.Close())
	}()
	_, err = file.WriteString(time.Now().Format("2006.01.02 15:04:05") + " " + formatError(saveErr) + "\n")
	must.FatalIf(err)
	_, err = file.WriteString(formatMerryStacktrace(saveErr))
	must.FatalIf(err)
}

func formatMeasureInfo(m data.MeasurementInfo) string {
	return fmt.Sprintf("%4d - %s - %s %s - %q",
		m.MeasurementID, m.CreatedAt.Format("2006.01.02 15:04"), m.Device, m.Kind, m.Name)
}

func panicWithSaveRecoveredErrorToFile() {
	x := recover()
	if x == nil {
		return
	}
	err := fmt.Errorf("panic: %+v", x)
	saveErrorToFile(err)
	walk.MsgBox(nil, "Установка контроля ЧЭ", formatError(err),
		walk.MsgBoxIconError|walk.MsgBoxOK|walk.MsgBoxSystemModal)
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

func pause(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-timer.C:
			return
		case <-ctx.Done():
			return
		}
	}
}

func formatMerryStacktrace(e error) string {
	s := merry.Stack(e)
	if len(s) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for i, fp := range s {
		fnc := runtime.FuncForPC(fp)
		if fnc != nil {
			f, l := fnc.FileLine(fp)
			name := filepath.Base(fnc.Name())
			ident := " "
			if i > 0 {
				ident = "\t"
			}
			buf.WriteString(fmt.Sprintf("%s%s:%d %s\n", ident, f, l, name))
		}
	}
	return buf.String()
}

func formatError(err error) string {
	dir := filepath.Dir(os.Args[0])
	return strings.ReplaceAll(err.Error(), dir+"\\", "")
}

package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func formatMeasureInfo(m data.MeasurementInfo) string {
	return fmt.Sprintf("%4d - %s - %s %s - %q",
		m.MeasurementID, m.CreatedAt.Format("2006.01.02 15:04"), m.Device, m.Kind, m.Name)
}

func msgBoxErr(msg string) {
	dir := filepath.Dir(os.Args[0])
	msg = strings.ReplaceAll(msg, dir+"\\", "")
	walk.MsgBox(nil, "Установка контроля ЧЭ", msg,
		walk.MsgBoxIconError|walk.MsgBoxOK|walk.MsgBoxSystemModal)
}

func panicMsgBox(x interface{}) {
	switch x := x.(type) {
	case nil:
		return
	case error:
		msgBoxErr(x.Error())
	default:
		msgBoxErr(fmt.Sprintf("%+v", x))
	}
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
	must.PanicIf(label.SetText(time.Now().Format("15:04:05") + " " + text))

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

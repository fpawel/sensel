package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"os"
	"path/filepath"
	"strings"
)

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

func newTableViewColumn(tvc declarative.TableViewColumn) *walk.TableViewColumn {
	w := walk.NewTableViewColumn()
	must.PanicIf(w.SetAlignment(walk.Alignment1D(tvc.Alignment)))
	w.SetDataMember(tvc.DataMember)
	if tvc.Format != "" {
		must.PanicIf(w.SetFormat(tvc.Format))
	}
	must.PanicIf(w.SetPrecision(tvc.Precision))
	w.SetName(tvc.Name)
	must.PanicIf(w.SetTitle(tvc.Title))

	must.PanicIf(w.SetVisible(!tvc.Hidden))
	must.PanicIf(w.SetFrozen(tvc.Frozen))
	must.PanicIf(w.SetWidth(tvc.Width))
	w.SetLessFunc(tvc.LessFunc)
	w.SetFormatFunc(tvc.FormatFunc)
	return w
}

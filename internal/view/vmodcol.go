package view

import (
	"github.com/fpawel/sensel/internal/pkg/must"
	. "github.com/lxn/walk/declarative"
)

func (x *MainTableViewModel) setupTableViewColumns() {
	must.PanicIf(x.tv.Columns().Clear())
	for _, c := range x.columns() {
		must.PanicIf(c.Create(x.tv))
	}
}

func (x *MainTableViewModel) columns() (xs []TableViewColumn) {
	appendResult := func(c TableViewColumn) {
		xs = append(xs, c)
	}
	appendResult(TableViewColumn{
		Name:  "Датчик",
		Title: "Датчик",
	})

	if x.showCalc {
		for _, m := range x.d.Pt.Columns {
			appendResult(TableViewColumn{
				Name:  m.Name,
				Title: m.Name,
			})
		}
		return
	}
	for _, smp := range x.d.D.Samples {
		appendResult(TableViewColumn{
			Name:  smp.Name,
			Title: smp.Name,
		})
	}
	return
}

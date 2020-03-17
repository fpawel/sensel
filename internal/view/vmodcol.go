package view

import (
	"fmt"
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
		return
	}
	for i := range x.d.D.Samples {
		appendResult(TableViewColumn{
			Name:  fmt.Sprintf("column%d", i),
			Title: fmt.Sprintf("%d", i+1),
		})
	}
	return
}

package viewmeasure

import (
	"fmt"
	"github.com/fpawel/sensel/internal/pkg/must"
	. "github.com/lxn/walk/declarative"
)

func (x *TableViewModel) setupTableViewColumns() {
	must.PanicIf(x.tv.Columns().Clear())
	for _, c := range x.columns() {
		must.PanicIf(c.Create(x.tv))
	}
}

func (x *TableViewModel) columns() (xs []TableViewColumn) {
	appendResult := func(c TableViewColumn) {
		xs = append(xs, c)
	}
	appendResult(TableViewColumn{
		Name:  "Датчик",
		Title: "Датчик",
	})

	for i := range x.d.Samples {
		appendResult(TableViewColumn{
			Name:  fmt.Sprintf("column%d", i),
			Title: fmt.Sprintf("%d", i+1),
		})
	}
	return
}

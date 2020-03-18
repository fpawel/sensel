package viewcalc

import (
	"fmt"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

type TableViewModel struct {
	walk.TableModelBase
	d  []calc.Column
	tv *walk.TableView
}

var _ walk.TableModel = new(TableViewModel)

func (x *TableViewModel) RowCount() int {
	return len(rows)
}

func (x *TableViewModel) Value(row, col int) interface{} {
	mRo := rows[row]
	if col == 0 {
		return mRo.Title
	}
	if col-1 >= len(x.d) {
		return ""
	}
	return mRo.F(x, col)
}

func New(tv *walk.TableView) walk.TableModel {
	x := &TableViewModel{
		tv: tv,
	}
	must.PanicIf(tv.SetModel(x))
	return x
}

func (x *TableViewModel) ViewData() []calc.Column {
	return x.d
}

func (x *TableViewModel) SetViewData(d []calc.Column) {
	x.d = d
	x.setupTableViewColumns()
	x.PublishRowsReset()
}

func (x *TableViewModel) setupTableViewColumns() {
	must.PanicIf(x.tv.Columns().Clear())
	for _, c := range x.columns() {
		must.PanicIf(c.Create(x.tv))
	}
}

type tableViewColumn = declarative.TableViewColumn

func (x *TableViewModel) columns() (xs []tableViewColumn) {
	appendResult := func(c tableViewColumn) {
		xs = append(xs, c)
	}
	appendResult(tableViewColumn{
		Name:  "Датчик",
		Title: "Датчик",
	})

	for _, c := range x.d {
		appendResult(tableViewColumn{
			Name:      fmt.Sprintf("column%d%s", c.Index, c.Name),
			Title:     fmt.Sprintf("%s", c.Name),
			Precision: 3,
		})
	}
	return
}

type rowType struct {
	Title string
	F     rowFunc
}

type rowFunc = func(x *TableViewModel, col int) interface{}

var rows = func() (xs []rowType) {
	type M = *TableViewModel
	type Row = rowType
	for i := 0; i < 16; i++ {
		i := i
		xs = append(xs, Row{
			Title: fmt.Sprintf("%d", i),
			F: func(x M, col int) interface{} {
				return x.d[col-1].Values[i].Value
			},
		})
	}
	return
}()

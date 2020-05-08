package viewmeasure

import (
	"fmt"
	"github.com/fpawel/sensel/internal/calc"
	"github.com/fpawel/sensel/internal/cfg"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"math"
)

var _ walk.TableModel = new(TableViewModel)

type TableViewModel struct {
	walk.TableModelBase
	d    data.Measurement
	cs   []calc.Column
	tv   *walk.TableView
	rows [][]interface{}
}

func NewMainTableViewModel(tv *walk.TableView) *TableViewModel {
	x := &TableViewModel{
		tv: tv,
	}
	must.PanicIf(tv.SetModel(x))
	return x
}

func (x *TableViewModel) ViewData() data.Measurement {
	return x.d
}

func (x *TableViewModel) SetViewData(d data.Measurement, cs []calc.Column) {
	x.d = d
	x.cs = cs
	x.rows = nil

	type smp = data.Sample

	addRow := func(title string, f func(smp) interface{}, fc func(calc.Column) interface{}) {
		row := []interface{}{title}
		for _, smp := range x.d.Samples {
			smp := smp
			row = append(row, f(smp))
		}

		for _, c := range x.cs {
			c := c
			row = append(row, fc(c))
		}

		x.rows = append(x.rows, row)
	}

	addRowNoCalc := func(title string, f func(smp) interface{}) {
		addRow(title, f, func(calc.Column) interface{} {
			return ""
		})
	}

	for i := 0; i < 16; i++ {
		addRow(fmt.Sprintf("%d", i+1), func(s smp) interface{} {
			return pkg.FormatFloat(s.U[i], 3)
		}, func(c calc.Column) interface{} {
			value := c.Values[i].Value
			if math.IsNaN(value) {
				return ""
			}
			return pkg.FormatFloat(value, c.Precision)
		})
	}
	addRowNoCalc("", func(s smp) interface{} {
		return ""
	})

	addRowNoCalc("Газ", func(s smp) interface{} {
		return s.Gas
	})
	addRowNoCalc("Q,мл/мин", func(s smp) interface{} {
		return pkg.FormatFloat(s.Q, 2)
	})
	addRowNoCalc("I,А", func(s smp) interface{} {
		return pkg.FormatFloat(s.I*1000., 3)
	})
	addRowNoCalc("U,В", func(s smp) interface{} {
		return pkg.FormatFloat(s.Ub, 2)
	})
	addRowNoCalc("Т⁰C", func(s smp) interface{} {
		return s.T
	})
	addRowNoCalc("Время", func(s smp) interface{} {
		return s.Tm.Format("15:04:05")
	})

	x.setupTableViewColumns()
	x.PublishRowsReset()
}

func (x *TableViewModel) RowCount() int {
	return len(x.rows)
}

func (x *TableViewModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(x.rows) {
		return ""
	}
	Row := x.rows[row]
	if col < 0 || col >= len(Row) {
		return ""
	}

	return Row[col]
}

func (x *TableViewModel) StyleCell(s *walk.CellStyle) {

	if s.Col() < 0 || s.Row() < 0 || s.Row() >= 16 {
		return
	}
	if s.Col() == 0 && s.Row() < 16 {
		for _, smp := range x.d.Samples {
			if smp.Br[s.Row()] {
				s.Image = "assets/img/error_circle.png"
				return
			}
		}
		for _, c := range x.cs {
			if c.IsErr(s.Row()) {
				s.Image = "assets/img/error.png"
				return
			}
		}
		return
	}
	nSamp := s.Col() - 1
	if s.Row() < 16 && nSamp >= 0 && nSamp < len(x.d.Samples) {
		if x.d.Samples[nSamp].Br[s.Row()] {
			s.Image = "assets/img/error_circle.png"
		}
		return
	}

	nCalc := s.Col() - len(x.d.Samples) - 1
	if nCalc < 0 || nCalc >= len(x.cs) {
		return
	}
	if x.cs[nCalc].IsErr(s.Row()) {
		s.Image = "assets/img/error.png"
		s.TextColor = walk.RGB(255, 0, 0)
	} else {
		s.TextColor = walk.RGB(0, 0, 128)
	}
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

	for i := range x.d.Samples {
		appendResult(tableViewColumn{
			Name:  fmt.Sprintf("column%d", i),
			Title: fmt.Sprintf("Обмер №%d", i+1),
			Width: 120,
		})
	}

	cws := cfg.Get().AppWindow.TableViewMeasure.ColumnWidths
	for i, c := range x.cs {
		w := 80
		if i < len(cws) {
			w = cws[i]
		}
		appendResult(tableViewColumn{
			Name:  fmt.Sprintf("column%d%s", c.Index, c.Name),
			Title: c.Name,
			Width: w,
		})
	}
	return
}

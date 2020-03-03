package view

import (
	"github.com/fpawel/sensel/internal/calcsens"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
)

var _ walk.TableModel = new(MainTableViewModel)

type MainTableViewModel struct {
	walk.TableModelBase
	d        MainTableViewData
	showCalc bool
	tv       *walk.TableView
}

type MainTableViewData struct {
	D  data.Measurement
	Cs []calcsens.ColumnCalculated
	Pt calcsens.ProductType
}

func NewMainTableViewModel(tv *walk.TableView) *MainTableViewModel {
	x := &MainTableViewModel{
		tv: tv,
	}
	must.PanicIf(tv.SetModel(x))
	return x
}

func (x *MainTableViewModel) ProductType() string {
	return x.d.Pt.Name
}

func (x *MainTableViewModel) SetViewData(d MainTableViewData) {
	x.d = d
	x.showCalc = true
	x.setupTableViewColumns()
	x.PublishRowsReset()
}

func (x *MainTableViewModel) SetShowCalc(v bool) {
	if v == x.showCalc {
		return
	}
	x.showCalc = v
	x.setupTableViewColumns()
	x.PublishRowsReset()
}

func (x *MainTableViewModel) RowCount() int {
	return len(measurementRows)
}

func (x *MainTableViewModel) Value(row, col int) interface{} {
	mRo := measurementRows[row]
	if col == 0 {
		return mRo.Title
	}
	if x.showCalc {
		if col-1 >= len(x.d.Cs) {
			return ""
		}
	} else {
		if col-1 >= len(x.d.D.Samples) {
			return ""
		}
	}
	return mRo.F(x, col)

}

func (x *MainTableViewModel) StyleCell(s *walk.CellStyle) {
	//mRo := measurementRows[s.Row()]
	if s.Col() < 0 || s.Row() < 0 {
		return
	}
	for nSmp, smp := range x.d.D.Samples {
		for i, p := range smp.Productions {
			if p.Break {
				if i == s.Row() {
					s.BackgroundColor = walk.RGB(240, 240, 240)
					switch s.Col() {
					case 0:
						s.Image = "img/error.png"
					case nSmp + 1:
						s.Image = "img/error_circle.png"
					}
				}
			}
		}
	}
}

package view

import (
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
	D data.Measurement
}

func NewMainTableViewModel(tv *walk.TableView) *MainTableViewModel {
	x := &MainTableViewModel{
		tv: tv,
	}
	must.PanicIf(tv.SetModel(x))
	return x
}

func (x *MainTableViewModel) Device() string {
	return x.d.D.Device
}

func (x *MainTableViewModel) Kind() string {
	return x.d.D.Kind
}

func (x *MainTableViewModel) GetMeasurement() data.Measurement {
	return x.d.D
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
		return ""
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

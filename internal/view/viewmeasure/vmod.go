package viewmeasure

import (
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
)

var _ walk.TableModel = new(TableViewModel)

type TableViewModel struct {
	walk.TableModelBase
	d  data.Measurement
	tv *walk.TableView
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

func (x *TableViewModel) SetViewData(d data.Measurement) {
	x.d = d
	x.setupTableViewColumns()
	x.PublishRowsReset()
}

func (x *TableViewModel) RowCount() int {
	return len(measurementRows)
}

func (x *TableViewModel) Value(row, col int) interface{} {
	mRo := measurementRows[row]
	if col == 0 {
		return mRo.Title
	}
	if col-1 >= len(x.d.Samples) {
		return ""
	}
	return mRo.F(x, col)

}

func (x *TableViewModel) StyleCell(s *walk.CellStyle) {
	//mRo := measurementRows[s.Row()]
	if s.Col() < 0 || s.Row() < 0 {
		return
	}
	for nSmp, smp := range x.d.Samples {
		for i, br := range smp.Br {
			if br {
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

package app

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"github.com/fpawel/sensel/internal/pkg/must"
	"github.com/lxn/walk"
)

type MeasurementViewModel struct {
	walk.TableModelBase
	M        data.Measurement
	ShowCalc bool
}

func (x *MeasurementViewModel) SetShowCalc(v bool) {
	if v == x.ShowCalc {
		return
	}
	x.ShowCalc = v
	x.PublishRowsReset()
}

func (x *MeasurementViewModel) RowCount() int {
	return len(measurementRows)
}

func (x *MeasurementViewModel) Value(row, col int) interface{} {
	mRo := measurementRows[row]
	if col == 0 {
		return mRo.Title
	}
	smp := x.M.Samples[col-1]
	return mRo.F(smp)

}

func (x *MeasurementViewModel) StyleCell(s *walk.CellStyle) {
	//mRo := measurementRows[s.Row()]
	if s.Col() < 0 || s.Row() < 0 {
		return
	}
	for nSmp, smp := range x.M.Samples {
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

func (x *MeasurementViewModel) SetupTableViewColumns(tableView *walk.TableView) {
	must.PanicIf(tableView.Columns().Clear())
	for _, c := range x.Columns() {
		must.PanicIf(c.Create(tableView))
	}
}

type measurementRow struct {
	Title string
	F     func(smp data.Sample) interface{}
}

var measurementRows = func() (xs []measurementRow) {
	appendResult := func(title string, F func(data.Sample) interface{}) {
		xs = append(xs, measurementRow{
			Title: title,
			F:     F,
		})
	}
	for i := 0; i < 16; i++ {
		i := i
		appendResult(fmt.Sprintf("%d", i), func(smp data.Sample) interface{} {
			return smp.Productions[i].Value
		})
	}
	appendResult("Газ", func(smp data.Sample) interface{} {
		if smp.Gas == 0 {
			return ""
		}
		return smp.Gas
	})
	appendResult("Расход", func(smp data.Sample) interface{} {
		return smp.Consumption
	})
	appendResult("Ток", func(smp data.Sample) interface{} {
		return smp.Current
	})
	appendResult("Температура", func(smp data.Sample) interface{} {
		return smp.Temperature
	})
	appendResult("Время", func(smp data.Sample) interface{} {
		return smp.CreatedAt.Format("15:04:05")
	})
	return
}()

var _ walk.TableModel = new(MeasurementViewModel)

package view

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
	"math"
)

type measurementRow struct {
	Title string
	F     measurementRowFunc
}

type measurementRowFunc = func(x *MainTableViewModel, col int) interface{}

var measurementRows = func() (xs []measurementRow) {
	type M = *MainTableViewModel
	type Row = measurementRow

	appendResult := func(title string, F measurementRowFunc) {
		xs = append(xs, Row{
			Title: title,
			F:     F,
		})
	}

	smp := func(x M, col int) data.Sample {
		return x.d.D.Samples[col-1]
	}

	fmtNoNaN := func(v float64) interface{} {
		if math.IsNaN(v) {
			return ""
		}
		return v
	}

	for i := 0; i < 16; i++ {
		i := i
		appendResult(fmt.Sprintf("%d", i), func(x M, col int) interface{} {
			if x.showCalc {
				if col-1 < len(x.d.Cs) {
					return fmtNoNaN(x.d.Cs[col-1].Values[i].Float)
				}
				return ""
			}
			return smp(x, col).Productions[i].Value
		})
	}
	appendResult("Газ", func(x M, col int) interface{} {
		smp := smp(x, col)
		if smp.Gas == 0 {
			return ""
		}
		return smp.Gas
	})
	appendResult("Расход", func(x M, col int) interface{} {
		return smp(x, col).Consumption
	})
	appendResult("Ток", func(x M, col int) interface{} {
		return smp(x, col).Current
	})
	appendResult("Температура", func(x M, col int) interface{} {
		return smp(x, col).Temperature
	})
	appendResult("Время", func(x M, col int) interface{} {
		return smp(x, col).CreatedAt.Format("15:04:05")
	})
	return
}()
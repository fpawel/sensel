package viewmeasure

import (
	"fmt"
	"github.com/fpawel/sensel/internal/data"
)

type measurementRow struct {
	Title string
	F     measurementRowFunc
}

type measurementRowFunc = func(x *TableViewModel, col int) interface{}

var measurementRows = func() (xs []measurementRow) {
	type M = *TableViewModel
	type Row = measurementRow

	appendResult := func(title string, F measurementRowFunc) {
		xs = append(xs, Row{
			Title: title,
			F:     F,
		})
	}

	smp := func(x M, col int) data.Sample {
		return x.d.Samples[col-1]
	}

	for i := 0; i < 16; i++ {
		i := i
		appendResult(fmt.Sprintf("%d", i), func(x M, col int) interface{} {
			return smp(x, col).U[i]
		})
	}
	appendResult("Газ", func(x M, col int) interface{} {
		smp := smp(x, col)
		return smp.Gas
	})
	appendResult("Расход", func(x M, col int) interface{} {
		return smp(x, col).Q
	})
	appendResult("Ток", func(x M, col int) interface{} {
		return smp(x, col).I
	})
	appendResult("Напряжение", func(x M, col int) interface{} {
		return smp(x, col).T
	})
	appendResult("Температура", func(x M, col int) interface{} {
		return smp(x, col).T
	})
	appendResult("Время", func(x M, col int) interface{} {
		return smp(x, col).Tm.Format("15:04:05")
	})
	return
}()
